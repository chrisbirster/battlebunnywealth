package proofofplay

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"
)

const AuthoritySchemaVersion = 1

const (
	MissionPending   = "pending"
	MissionCompleted = "completed"
	MissionExpired   = "expired"
)

var (
	ErrMissionLimit            = errors.New("daily Proof-of-Play mission limit reached")
	ErrMissionNotFound         = errors.New("Proof-of-Play mission not found")
	ErrMissionExpired          = errors.New("Proof-of-Play mission expired")
	ErrMissionCompleted        = errors.New("Proof-of-Play mission already completed")
	ErrMissionBinding          = errors.New("Proof-of-Play mission does not match account or device")
	ErrInvalidMissionSignature = errors.New("invalid Proof-of-Play device signature")
	ErrInvalidMissionDevice    = errors.New("invalid Proof-of-Play device key")
)

type AuthorityConfig struct {
	MissionTTL           time.Duration
	MaxMissionsPerDay    int
	MissionAward         int64
	MaxAuthority         int64
	DecayGrace           time.Duration
	DecayPerDayBPS       int64
	EligibilityThreshold int64
	NewcomerRamp         time.Duration
}

func DefaultAuthorityConfig() AuthorityConfig {
	return AuthorityConfig{
		MissionTTL:           5 * time.Minute,
		MaxMissionsPerDay:    4,
		MissionAward:         25,
		MaxAuthority:         1000,
		DecayGrace:           72 * time.Hour,
		DecayPerDayBPS:       50,
		EligibilityThreshold: 100,
		NewcomerRamp:         14 * 24 * time.Hour,
	}
}

type NetworkMission struct {
	ID                    string     `json:"id"`
	AccountID             string     `json:"-"`
	DeviceID              string     `json:"deviceId"`
	DevicePublicKeySPKI   string     `json:"-"`
	DeviceAttestation     string     `json:"deviceAttestation"`
	Kind                  string     `json:"kind"`
	NPC                   string     `json:"npc"`
	Title                 string     `json:"title"`
	Briefing              string     `json:"briefing"`
	Operation             string     `json:"operation"`
	Epoch                 uint64     `json:"epoch"`
	Challenge             string     `json:"challenge"`
	CheckpointHash        string     `json:"checkpointHash"`
	SignedPayload         string     `json:"signedPayload"`
	AuthorityAward        int64      `json:"authorityAward"`
	Status                string     `json:"status"`
	IssuedAt              time.Time  `json:"issuedAt"`
	ExpiresAt             time.Time  `json:"expiresAt"`
	CompletedAt           *time.Time `json:"completedAt,omitempty"`
}

type MissionCompletion struct {
	MissionID       string    `json:"missionId"`
	AccountID       string    `json:"-"`
	DeviceID        string    `json:"deviceId"`
	Epoch           uint64    `json:"epoch"`
	Kind            string    `json:"kind"`
	AuthorityAward  int64     `json:"authorityAward"`
	SignatureDigest string    `json:"signatureDigest"`
	CompletedAt     time.Time `json:"completedAt"`
}

type AuthorityRecord struct {
	AccountID           string    `json:"accountId"`
	Score               int64     `json:"score"`
	MissionsCompleted   int       `json:"missionsCompleted"`
	FirstMissionAt      time.Time `json:"firstMissionAt"`
	LastMissionAt       time.Time `json:"lastMissionAt"`
	DecayAppliedThrough time.Time `json:"decayAppliedThrough"`
}

type AuthorityStoreState struct {
	SchemaVersion int                 `json:"schemaVersion"`
	Missions      []NetworkMission    `json:"missions"`
	Authority     []AuthorityRecord   `json:"authority"`
	Completions   []MissionCompletion `json:"completions"`
}

type AuthorityView struct {
	Score                  int64     `json:"score"`
	Maximum                int64     `json:"maximum"`
	MissionsCompleted      int       `json:"missionsCompleted"`
	EligibilityThreshold   int64     `json:"eligibilityThreshold"`
	PrototypeEligible      bool      `json:"prototypeEligible"`
	ProductionEligible     bool      `json:"productionEligible"`
	CommitteeWeight        int64     `json:"committeeWeight"`
	NewcomerWeightPercent  int       `json:"newcomerWeightPercent"`
	DecayGraceSeconds      int64     `json:"decayGraceSeconds"`
	DecayPerDayBasisPoints int64     `json:"decayPerDayBasisPoints"`
	FirstMissionAt         time.Time `json:"firstMissionAt,omitempty"`
	LastMissionAt          time.Time `json:"lastMissionAt,omitempty"`
}

type AuthoritySnapshot struct {
	Authority         AuthorityView       `json:"authority"`
	CurrentMission    *NetworkMission     `json:"currentMission,omitempty"`
	RecentCompletions []MissionCompletion `json:"recentCompletions"`
	DailyMissionsUsed int                 `json:"dailyMissionsUsed"`
	DailyMissionLimit int                 `json:"dailyMissionLimit"`
	AttestationGate   string              `json:"attestationGate"`
}

type CommitteeMember struct {
	AccountID          string `json:"accountId"`
	Authority          int64  `json:"authority"`
	CommitteeWeight    int64  `json:"committeeWeight"`
	ProductionEligible bool   `json:"productionEligible"`
}

type AuthorityService struct {
	mu     sync.Mutex
	now    func() time.Time
	store  AuthorityStore
	config AuthorityConfig
	state  AuthorityStoreState
}

func NewAuthorityService(now func() time.Time, store AuthorityStore, config AuthorityConfig) (*AuthorityService, error) {
	if now == nil {
		now = time.Now
	}
	if config.MissionTTL <= 0 || config.MaxMissionsPerDay <= 0 || config.MissionAward <= 0 || config.MaxAuthority <= 0 {
		return nil, errors.New("invalid Proof-of-Play authority config")
	}
	if config.DecayGrace < 0 || config.DecayPerDayBPS < 0 || config.DecayPerDayBPS >= 10000 || config.EligibilityThreshold < 0 || config.NewcomerRamp <= 0 {
		return nil, errors.New("invalid Proof-of-Play authority policy")
	}
	state, err := store.Load()
	if errors.Is(err, ErrAuthorityStateNotFound) {
		state = AuthorityStoreState{SchemaVersion: AuthoritySchemaVersion, Missions: []NetworkMission{}, Authority: []AuthorityRecord{}, Completions: []MissionCompletion{}}
		if err := store.Save(state); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if state.SchemaVersion != AuthoritySchemaVersion {
		return nil, fmt.Errorf("Proof-of-Play authority schema %d unsupported; want %d", state.SchemaVersion, AuthoritySchemaVersion)
	}
	return &AuthorityService{now: now, store: store, config: config, state: state}, nil
}

func (s *AuthorityService) IssueMission(accountID, deviceID, publicKeySPKI, attestationStatus, checkpointHash string, epoch uint64) (AuthoritySnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	accountID = strings.TrimSpace(accountID)
	deviceID = strings.TrimSpace(deviceID)
	if accountID == "" || deviceID == "" || checkpointHash == "" || !validDevicePublicKey(publicKeySPKI) {
		return AuthoritySnapshot{}, ErrInvalidMissionDevice
	}
	for i := range s.state.Missions {
		mission := &s.state.Missions[i]
		if mission.Status == MissionPending && !mission.ExpiresAt.After(now) {
			mission.Status = MissionExpired
		}
		if mission.AccountID == accountID && mission.DeviceID == deviceID && mission.Status == MissionPending && mission.ExpiresAt.After(now) {
			return s.snapshotLocked(accountID, now), nil
		}
	}
	used := s.dailyMissionsUsedLocked(accountID, now)
	if used >= s.config.MaxMissionsPerDay {
		return AuthoritySnapshot{}, ErrMissionLimit
	}
	missionID, err := randomAuthorityToken(18)
	if err != nil {
		return AuthoritySnapshot{}, err
	}
	challenge, err := randomAuthorityToken(32)
	if err != nil {
		return AuthoritySnapshot{}, err
	}
	template := missionTemplateFor(accountID, deviceID, epoch, used)
	mission := NetworkMission{
		ID:                  missionID,
		AccountID:           accountID,
		DeviceID:            deviceID,
		DevicePublicKeySPKI: publicKeySPKI,
		DeviceAttestation:   attestationStatus,
		Kind:                template.Kind,
		NPC:                 template.NPC,
		Title:               template.Title,
		Briefing:            template.Briefing,
		Operation:           template.Operation,
		Epoch:               epoch,
		Challenge:           challenge,
		CheckpointHash:      checkpointHash,
		AuthorityAward:      s.config.MissionAward,
		Status:              MissionPending,
		IssuedAt:            now,
		ExpiresAt:           now.Add(s.config.MissionTTL),
	}
	mission.SignedPayload = canonicalMissionPayload(mission)
	s.state.Missions = append(s.state.Missions, mission)
	if err := s.store.Save(s.state); err != nil {
		return AuthoritySnapshot{}, err
	}
	return s.snapshotLocked(accountID, now), nil
}

func (s *AuthorityService) CompleteMission(accountID, deviceID, publicKeySPKI, missionID, signature string) (AuthoritySnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	idx := s.missionIndexLocked(missionID)
	if idx < 0 {
		return AuthoritySnapshot{}, ErrMissionNotFound
	}
	mission := &s.state.Missions[idx]
	if mission.AccountID != accountID || mission.DeviceID != deviceID || mission.DevicePublicKeySPKI != publicKeySPKI {
		return AuthoritySnapshot{}, ErrMissionBinding
	}
	if mission.Status == MissionCompleted {
		return AuthoritySnapshot{}, ErrMissionCompleted
	}
	if mission.Status == MissionExpired || !mission.ExpiresAt.After(now) {
		mission.Status = MissionExpired
		_ = s.store.Save(s.state)
		return AuthoritySnapshot{}, ErrMissionExpired
	}
	if s.dailyCompletionsLocked(accountID, now) >= s.config.MaxMissionsPerDay {
		return AuthoritySnapshot{}, ErrMissionLimit
	}
	if err := verifyMissionDeviceSignature(publicKeySPKI, mission.SignedPayload, signature); err != nil {
		return AuthoritySnapshot{}, err
	}
	mission.Status = MissionCompleted
	mission.CompletedAt = &now
	authority := s.authorityRecordLocked(accountID)
	s.applyDecayLocked(authority, now)
	if authority.FirstMissionAt.IsZero() {
		authority.FirstMissionAt = now
	}
	authority.Score += mission.AuthorityAward
	if authority.Score > s.config.MaxAuthority {
		authority.Score = s.config.MaxAuthority
	}
	authority.MissionsCompleted++
	authority.LastMissionAt = now
	authority.DecayAppliedThrough = now
	digest := sha256.Sum256([]byte(signature))
	s.state.Completions = append(s.state.Completions, MissionCompletion{
		MissionID: mission.ID, AccountID: accountID, DeviceID: deviceID, Epoch: mission.Epoch, Kind: mission.Kind,
		AuthorityAward: mission.AuthorityAward, SignatureDigest: hex.EncodeToString(digest[:]), CompletedAt: now,
	})
	if err := s.store.Save(s.state); err != nil {
		return AuthoritySnapshot{}, err
	}
	return s.snapshotLocked(accountID, now), nil
}

func (s *AuthorityService) Snapshot(accountID string) (AuthoritySnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	if idx := s.authorityIndexLocked(accountID); idx >= 0 {
		s.applyDecayLocked(&s.state.Authority[idx], now)
		if err := s.store.Save(s.state); err != nil {
			return AuthoritySnapshot{}, err
		}
	}
	return s.snapshotLocked(accountID, now), nil
}

func (s *AuthorityService) SelectCommittee(epoch uint64, seed string, size int) []CommitteeMember {
	s.mu.Lock()
	defer s.mu.Unlock()
	if size <= 0 {
		return nil
	}
	now := s.now().UTC()
	candidates := make([]CommitteeMember, 0, len(s.state.Authority))
	for i := range s.state.Authority {
		record := &s.state.Authority[i]
		s.applyDecayLocked(record, now)
		view := s.authorityViewLocked(*record, now)
		if !view.PrototypeEligible || view.CommitteeWeight <= 0 {
			continue
		}
		candidates = append(candidates, CommitteeMember{AccountID: record.AccountID, Authority: record.Score, CommitteeWeight: view.CommitteeWeight, ProductionEligible: false})
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].AccountID < candidates[j].AccountID })
	if size > len(candidates) {
		size = len(candidates)
	}
	selected := make([]CommitteeMember, 0, size)
	for seat := 0; seat < size && len(candidates) > 0; seat++ {
		var total uint64
		for _, candidate := range candidates {
			total += uint64(candidate.CommitteeWeight)
		}
		if total == 0 {
			break
		}
		digest := sha256.Sum256([]byte(fmt.Sprintf("pop-committee/v1|%s|%d|%d", seed, epoch, seat)))
		draw := binary.BigEndian.Uint64(digest[:8]) % total
		var cumulative uint64
		pick := 0
		for i, candidate := range candidates {
			cumulative += uint64(candidate.CommitteeWeight)
			if draw < cumulative {
				pick = i
				break
			}
		}
		selected = append(selected, candidates[pick])
		candidates = append(candidates[:pick], candidates[pick+1:]...)
	}
	return selected
}

func (s *AuthorityService) snapshotLocked(accountID string, now time.Time) AuthoritySnapshot {
	var record AuthorityRecord
	if idx := s.authorityIndexLocked(accountID); idx >= 0 {
		record = s.state.Authority[idx]
	}
	var current *NetworkMission
	for i := len(s.state.Missions) - 1; i >= 0; i-- {
		m := s.state.Missions[i]
		if m.AccountID == accountID && m.Status == MissionPending && m.ExpiresAt.After(now) {
			copy := publicMission(m)
			current = &copy
			break
		}
	}
	recent := make([]MissionCompletion, 0, 8)
	for i := len(s.state.Completions) - 1; i >= 0 && len(recent) < 8; i-- {
		if s.state.Completions[i].AccountID == accountID {
			copy := s.state.Completions[i]
			copy.AccountID = ""
			recent = append(recent, copy)
		}
	}
	return AuthoritySnapshot{
		Authority:         s.authorityViewLocked(record, now),
		CurrentMission:    current,
		RecentCompletions: recent,
		DailyMissionsUsed: s.dailyMissionsUsedLocked(accountID, now),
		DailyMissionLimit: s.config.MaxMissionsPerDay,
		AttestationGate:   "v0.6 authority is prototype-only until v0.7 device attestation; production committee eligibility is always false",
	}
}

func (s *AuthorityService) authorityViewLocked(record AuthorityRecord, now time.Time) AuthorityView {
	percent := 10
	if !record.FirstMissionAt.IsZero() {
		elapsed := now.Sub(record.FirstMissionAt)
		if elapsed < 0 {
			elapsed = 0
		}
		if elapsed >= s.config.NewcomerRamp {
			percent = 100
		} else {
			percent = 10 + int((90*elapsed)/s.config.NewcomerRamp)
		}
	}
	weight := record.Score * int64(percent) / 100
	eligible := record.Score >= s.config.EligibilityThreshold
	if !eligible {
		weight = 0
	}
	return AuthorityView{
		Score: record.Score, Maximum: s.config.MaxAuthority, MissionsCompleted: record.MissionsCompleted,
		EligibilityThreshold: s.config.EligibilityThreshold, PrototypeEligible: eligible, ProductionEligible: false,
		CommitteeWeight: weight, NewcomerWeightPercent: percent, DecayGraceSeconds: int64(s.config.DecayGrace / time.Second),
		DecayPerDayBasisPoints: s.config.DecayPerDayBPS, FirstMissionAt: record.FirstMissionAt, LastMissionAt: record.LastMissionAt,
	}
}

func (s *AuthorityService) applyDecayLocked(record *AuthorityRecord, now time.Time) {
	if record.LastMissionAt.IsZero() || record.Score <= 0 {
		return
	}
	start := record.LastMissionAt.Add(s.config.DecayGrace)
	if !now.After(start) {
		return
	}
	cursor := record.DecayAppliedThrough
	if cursor.Before(start) {
		cursor = start
	}
	days := int(now.Sub(cursor) / (24 * time.Hour))
	if days <= 0 {
		return
	}
	for i := 0; i < days && record.Score > 0; i++ {
		reduction := record.Score * s.config.DecayPerDayBPS / 10000
		if reduction < 1 {
			reduction = 1
		}
		record.Score -= reduction
		if record.Score < 0 {
			record.Score = 0
		}
	}
	record.DecayAppliedThrough = cursor.Add(time.Duration(days) * 24 * time.Hour)
}

func (s *AuthorityService) authorityRecordLocked(accountID string) *AuthorityRecord {
	if idx := s.authorityIndexLocked(accountID); idx >= 0 {
		return &s.state.Authority[idx]
	}
	s.state.Authority = append(s.state.Authority, AuthorityRecord{AccountID: accountID})
	return &s.state.Authority[len(s.state.Authority)-1]
}

func (s *AuthorityService) authorityIndexLocked(accountID string) int {
	for i := range s.state.Authority {
		if s.state.Authority[i].AccountID == accountID {
			return i
		}
	}
	return -1
}

func (s *AuthorityService) missionIndexLocked(id string) int {
	for i := range s.state.Missions {
		if s.state.Missions[i].ID == id {
			return i
		}
	}
	return -1
}

func (s *AuthorityService) dailyMissionsUsedLocked(accountID string, now time.Time) int {
	day := now.UTC().Format("2006-01-02")
	used := 0
	for _, mission := range s.state.Missions {
		if mission.AccountID == accountID && mission.IssuedAt.UTC().Format("2006-01-02") == day {
			used++
		}
	}
	return used
}

func (s *AuthorityService) dailyCompletionsLocked(accountID string, now time.Time) int {
	day := now.UTC().Format("2006-01-02")
	count := 0
	for _, completion := range s.state.Completions {
		if completion.AccountID == accountID && completion.CompletedAt.UTC().Format("2006-01-02") == day {
			count++
		}
	}
	return count
}

type missionTemplate struct{ Kind, NPC, Title, Briefing, Operation string }

func missionTemplateFor(accountID, deviceID string, epoch uint64, ordinal int) missionTemplate {
	digest := sha256.Sum256([]byte(fmt.Sprintf("pop-mission-kind/v1|%s|%s|%d|%d", accountID, deviceID, epoch, ordinal)))
	switch int(digest[0] % 3) {
	case 0:
		return missionTemplate{"recon-patrol", "Private Stuffy", "Recon Patrol", "Command needs a fresh checkpoint read. Stuffy insists this is reconnaissance and not wandering around with a clipboard.", "verify-checkpoint"}
	case 1:
		return missionTemplate{"secure-supply-line", "Captain Cashmere", "Secure the Supply Line", "Cashmere wants the current chain checkpoint bound to your enrolled device before the paperwork starts breeding.", "bind-device-checkpoint"}
	default:
		return missionTemplate{"verify-intel", "First Sergeant Hard-as-Nails", "Verify Intel", "Sign the unpredictable network challenge. Hard-as-Nails would like proof you were actually here, recruit.", "sign-epoch-challenge"}
	}
}

func canonicalMissionPayload(m NetworkMission) string {
	return fmt.Sprintf("pop-mission/v1\nmission=%s\naccount=%s\ndevice=%s\nepoch=%d\nchallenge=%s\ncheckpoint=%s\noperation=%s", m.ID, m.AccountID, m.DeviceID, m.Epoch, m.Challenge, m.CheckpointHash, m.Operation)
}

func publicMission(m NetworkMission) NetworkMission {
	m.AccountID = ""
	m.DevicePublicKeySPKI = ""
	return m
}

func validDevicePublicKey(encoded string) bool {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}
	parsed, err := x509.ParsePKIXPublicKey(raw)
	if err != nil {
		return false
	}
	pub, ok := parsed.(*ecdsa.PublicKey)
	return ok && pub.Curve == elliptic.P256()
}

func verifyMissionDeviceSignature(encodedKey, payload, encodedSignature string) error {
	rawKey, err := base64.RawURLEncoding.DecodeString(encodedKey)
	if err != nil {
		return ErrInvalidMissionDevice
	}
	parsed, err := x509.ParsePKIXPublicKey(rawKey)
	if err != nil {
		return ErrInvalidMissionDevice
	}
	pub, ok := parsed.(*ecdsa.PublicKey)
	if !ok || pub.Curve != elliptic.P256() {
		return ErrInvalidMissionDevice
	}
	sig, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return ErrInvalidMissionSignature
	}
	digest := sha256.Sum256([]byte(payload))
	valid := false
	if len(sig) == 64 {
		r := new(big.Int).SetBytes(sig[:32])
		s := new(big.Int).SetBytes(sig[32:])
		valid = ecdsa.Verify(pub, digest[:], r, s)
	} else {
		valid = ecdsa.VerifyASN1(pub, digest[:], sig)
	}
	if !valid {
		return ErrInvalidMissionSignature
	}
	return nil
}

func randomAuthorityToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
