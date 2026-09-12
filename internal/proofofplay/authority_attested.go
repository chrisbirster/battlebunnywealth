package proofofplay

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

func DeviceWeightPercent(ordinal int) int {
	switch ordinal {
	case 0: return 100
	case 1: return 25
	case 2: return 10
	default: return 2
	}
}

func (s *AuthorityService) IssueWeightedMission(accountID, deviceID, publicKeySPKI, attestationStatus, checkpointHash string, epoch uint64, deviceWeightPercent int) (AuthoritySnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	accountID = strings.TrimSpace(accountID)
	deviceID = strings.TrimSpace(deviceID)
	if accountID == "" || deviceID == "" || checkpointHash == "" || !validDevicePublicKey(publicKeySPKI) { return AuthoritySnapshot{}, ErrInvalidMissionDevice }
	if deviceWeightPercent <= 0 || deviceWeightPercent > 100 { return AuthoritySnapshot{}, ErrInvalidMissionDevice }
	for i := range s.state.Missions {
		mission := &s.state.Missions[i]
		if mission.Status == MissionPending && !mission.ExpiresAt.After(now) { mission.Status = MissionExpired }
		if mission.AccountID == accountID && mission.DeviceID == deviceID && mission.Status == MissionPending && mission.ExpiresAt.After(now) { return s.snapshotLocked(accountID,now),nil }
	}
	used := s.dailyMissionsUsedLocked(accountID,now)
	if used >= s.config.MaxMissionsPerDay { return AuthoritySnapshot{}, ErrMissionLimit }
	missionID,err:=randomAuthorityToken(18);if err!=nil{return AuthoritySnapshot{},err}
	challenge,err:=randomAuthorityToken(32);if err!=nil{return AuthoritySnapshot{},err}
	template:=missionTemplateFor(accountID,deviceID,epoch,used)
	award:=s.config.MissionAward*int64(deviceWeightPercent)/100;if award<1{award=1}
	mission:=NetworkMission{ID:missionID,AccountID:accountID,DeviceID:deviceID,DevicePublicKeySPKI:publicKeySPKI,DeviceAttestation:attestationStatus,Kind:template.Kind,NPC:template.NPC,Title:template.Title,Briefing:template.Briefing,Operation:template.Operation,Epoch:epoch,Challenge:challenge,CheckpointHash:checkpointHash,AuthorityAward:award,Status:MissionPending,IssuedAt:now,ExpiresAt:now.Add(s.config.MissionTTL)}
	mission.SignedPayload=canonicalMissionPayload(mission)
	s.state.Missions=append(s.state.Missions,mission)
	if err:=s.store.Save(s.state);err!=nil{return AuthoritySnapshot{},err}
	return s.snapshotLocked(accountID,now),nil
}

func (s *AuthorityService) SnapshotWithAttestation(accountID string, productionEligible bool) (AuthoritySnapshot,error) {
	snapshot,err:=s.Snapshot(accountID);if err!=nil{return AuthoritySnapshot{},err}
	snapshot.Authority.ProductionEligible=snapshot.Authority.PrototypeEligible&&productionEligible
	if snapshot.Authority.ProductionEligible { snapshot.AttestationGate="attested device policy satisfied for permissioned-testnet committee research" } else { snapshot.AttestationGate="production committee eligibility requires at least one active hardware-backed verified device" }
	return snapshot,nil
}

func (s *AuthorityService) SelectProductionCommittee(epoch uint64, seed string, size int, eligibleAccounts map[string]bool) []CommitteeMember {
	s.mu.Lock();defer s.mu.Unlock();if size<=0{return nil};now:=s.now().UTC();candidates:=[]CommitteeMember{}
	for i:=range s.state.Authority{record:=&s.state.Authority[i];s.applyDecayLocked(record,now);view:=s.authorityViewLocked(*record,now);if !view.PrototypeEligible||view.CommitteeWeight<=0||!eligibleAccounts[record.AccountID]{continue};candidates=append(candidates,CommitteeMember{AccountID:record.AccountID,Authority:record.Score,CommitteeWeight:view.CommitteeWeight,ProductionEligible:true})}
	sort.Slice(candidates,func(i,j int)bool{return candidates[i].AccountID<candidates[j].AccountID});if size>len(candidates){size=len(candidates)};selected:=make([]CommitteeMember,0,size)
	for seat:=0;seat<size&&len(candidates)>0;seat++{var total uint64;for _,candidate:=range candidates{total+=uint64(candidate.CommitteeWeight)};if total==0{break};digest:=sha256.Sum256([]byte(fmt.Sprintf("pop-production-committee/v1|%s|%d|%d",seed,epoch,seat)));draw:=binary.BigEndian.Uint64(digest[:8])%total;var cumulative uint64;pick:=0;for i,candidate:=range candidates{cumulative+=uint64(candidate.CommitteeWeight);if draw<cumulative{pick=i;break}};selected=append(selected,candidates[pick]);candidates=append(candidates[:pick],candidates[pick+1:]...)}
	return selected
}
