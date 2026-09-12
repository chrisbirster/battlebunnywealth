package proofofplay

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const AttestationSchemaVersion = 1

const (
	AttestationStatePending  = "pending"
	AttestationStateVerified = "verified"
	AttestationStateRejected = "rejected"
	AttestationStateExpired  = "expired"

	ProviderAppleAppAttest      = "apple-app-attest"
	ProviderGooglePlayIntegrity = "google-play-integrity"
	ProviderDevelopment         = "development"
)

var (
	ErrAttestationChallengeNotFound = errors.New("attestation challenge not found")
	ErrAttestationChallengeExpired  = errors.New("attestation challenge expired")
	ErrAttestationChallengeUsed     = errors.New("attestation challenge already used")
	ErrAttestationBinding           = errors.New("attestation does not match account or device")
	ErrAttestationProvider          = errors.New("unsupported attestation provider")
	ErrAttestationRejected          = errors.New("device attestation rejected")
)

type AttestationConfig struct {
	ChallengeTTL                  time.Duration
	AppleBundleID                 string
	AppleTeamID                   string
	AppleEnvironment              string
	AndroidPackageName            string
	AndroidAllowedCertificates    []string
	AndroidRequireStrongIntegrity bool
}

func DefaultAttestationConfig() AttestationConfig { return AttestationConfig{ChallengeTTL:5*time.Minute,AppleEnvironment:"development"} }

type AttestationChallenge struct {
	ID                  string     `json:"id"`
	AccountID           string     `json:"-"`
	DeviceID            string     `json:"deviceId"`
	Provider            string     `json:"provider"`
	Nonce               string     `json:"nonce"`
	DevicePublicKeyHash string     `json:"devicePublicKeyHash"`
	BindingPayload      string     `json:"bindingPayload"`
	RequestHash         string     `json:"requestHash"`
	IssuedAt            time.Time  `json:"issuedAt"`
	ExpiresAt           time.Time  `json:"expiresAt"`
	UsedAt              *time.Time `json:"usedAt,omitempty"`
}

type DeviceAttestationRecord struct {
	AccountID string `json:"accountId"`; DeviceID string `json:"deviceId"`; Provider string `json:"provider"`; Status string `json:"status"`
	VerifiedAt time.Time `json:"verifiedAt,omitempty"`; EvidenceDigest string `json:"evidenceDigest,omitempty"`; ProviderKeyID string `json:"providerKeyId,omitempty"`; AssertionCounter uint64 `json:"assertionCounter,omitempty"`
	IntegrityLabels []string `json:"integrityLabels,omitempty"`; HardwareBacked bool `json:"hardwareBacked"`; ProductionEligible bool `json:"productionEligible"`; Reason string `json:"reason,omitempty"`
}

type AttestationState struct { SchemaVersion int `json:"schemaVersion"`; Challenges []AttestationChallenge `json:"challenges"`; Devices []DeviceAttestationRecord `json:"devices"` }
type AttestationEvidence struct { Provider string `json:"provider"`; ChallengeID string `json:"challengeId"`; DeviceID string `json:"deviceId"`; Payload string `json:"payload"`; KeyID string `json:"keyId,omitempty"`; CertificateChain []string `json:"certificateChain,omitempty"` }
type AttestationResult struct { Provider string; ProviderKeyID string; EvidenceDigest string; IntegrityLabels []string; HardwareBacked bool; ProductionEligible bool; AssertionCounter uint64 }
type AttestationProviderVerifier interface { Provider() string; VerifyEnrollment(context.Context,AttestationChallenge,AttestationEvidence)(AttestationResult,error) }
type AttestationStore interface { Load()(AttestationState,error); Save(AttestationState)error }

type AttestationService struct { mu sync.Mutex; now func()time.Time; store AttestationStore; config AttestationConfig; state AttestationState; providers map[string]AttestationProviderVerifier }

func NewAttestationService(now func()time.Time,store AttestationStore,config AttestationConfig,providers ...AttestationProviderVerifier)(*AttestationService,error){
	if now==nil{now=time.Now};if config.ChallengeTTL<=0{config.ChallengeTTL=5*time.Minute};state,err:=store.Load();if errors.Is(err,ErrAttestationStateNotFound){state=AttestationState{SchemaVersion:AttestationSchemaVersion,Challenges:[]AttestationChallenge{},Devices:[]DeviceAttestationRecord{}};if err:=store.Save(state);err!=nil{return nil,err}}else if err!=nil{return nil,err};if state.SchemaVersion!=AttestationSchemaVersion{return nil,fmt.Errorf("attestation schema %d unsupported; want %d",state.SchemaVersion,AttestationSchemaVersion)};registry:=map[string]AttestationProviderVerifier{};for _,provider:=range providers{if provider!=nil{registry[provider.Provider()]=provider}};return &AttestationService{now:now,store:store,config:config,state:state,providers:registry},nil
}

func(s *AttestationService)Begin(accountID,deviceID,publicKeySPKI,provider string)(AttestationChallenge,error){
	s.mu.Lock();defer s.mu.Unlock();provider=strings.TrimSpace(provider);if _,ok:=s.providers[provider];!ok{return AttestationChallenge{},ErrAttestationProvider};if !validDevicePublicKey(publicKeySPKI){return AttestationChallenge{},ErrAttestationBinding};now:=s.now().UTC();for i:=range s.state.Challenges{c:=&s.state.Challenges[i];if c.AccountID==accountID&&c.DeviceID==deviceID&&c.Provider==provider&&c.UsedAt==nil&&c.ExpiresAt.After(now){return publicAttestationChallenge(*c),nil}}
	id,err:=attestationToken(18);if err!=nil{return AttestationChallenge{},err};nonce,err:=attestationToken(32);if err!=nil{return AttestationChallenge{},err};keyHash:=sha256.Sum256([]byte(publicKeySPKI));keyHashText:=base64.RawURLEncoding.EncodeToString(keyHash[:]);binding:=fmt.Sprintf("bbw-attestation/v1\nchallenge=%s\ndevice=%s\npublicKeySha256=%s",nonce,deviceID,keyHashText);requestHashBytes:=sha256.Sum256([]byte(binding));requestHash:=base64.RawURLEncoding.EncodeToString(requestHashBytes[:]);challenge:=AttestationChallenge{ID:id,AccountID:accountID,DeviceID:deviceID,Provider:provider,Nonce:nonce,DevicePublicKeyHash:keyHashText,BindingPayload:binding,RequestHash:requestHash,IssuedAt:now,ExpiresAt:now.Add(s.config.ChallengeTTL)};s.state.Challenges=append(s.state.Challenges,challenge);if err:=s.store.Save(s.state);err!=nil{return AttestationChallenge{},err};return publicAttestationChallenge(challenge),nil
}

func(s *AttestationService)Complete(ctx context.Context,accountID,deviceID string,evidence AttestationEvidence)(DeviceAttestationRecord,error){
	s.mu.Lock();idx:=-1;for i:=range s.state.Challenges{if s.state.Challenges[i].ID==evidence.ChallengeID{idx=i;break}};if idx<0{s.mu.Unlock();return DeviceAttestationRecord{},ErrAttestationChallengeNotFound};challenge:=s.state.Challenges[idx];now:=s.now().UTC();if challenge.UsedAt!=nil{s.mu.Unlock();return DeviceAttestationRecord{},ErrAttestationChallengeUsed};if !challenge.ExpiresAt.After(now){s.mu.Unlock();return DeviceAttestationRecord{},ErrAttestationChallengeExpired};if challenge.AccountID!=accountID||challenge.DeviceID!=deviceID||evidence.DeviceID!=deviceID||evidence.Provider!=challenge.Provider{s.mu.Unlock();return DeviceAttestationRecord{},ErrAttestationBinding};provider:=s.providers[challenge.Provider];s.state.Challenges[idx].UsedAt=&now;if err:=s.store.Save(s.state);err!=nil{s.mu.Unlock();return DeviceAttestationRecord{},err};s.mu.Unlock()
	result,err:=provider.VerifyEnrollment(ctx,challenge,evidence)
	s.mu.Lock();defer s.mu.Unlock();verifiedAt:=s.now().UTC();record:=DeviceAttestationRecord{AccountID:accountID,DeviceID:deviceID,Provider:challenge.Provider,Status:AttestationStateRejected,VerifiedAt:verifiedAt,Reason:"provider verification failed"};if err==nil{record.Status=AttestationStateVerified;record.ProviderKeyID=result.ProviderKeyID;record.EvidenceDigest=result.EvidenceDigest;record.IntegrityLabels=append([]string(nil),result.IntegrityLabels...);record.HardwareBacked=result.HardwareBacked;record.ProductionEligible=result.ProductionEligible;record.AssertionCounter=result.AssertionCounter;record.Reason=""};updated:=false;for i:=range s.state.Devices{if s.state.Devices[i].AccountID==accountID&&s.state.Devices[i].DeviceID==deviceID{s.state.Devices[i]=record;updated=true;break}};if !updated{s.state.Devices=append(s.state.Devices,record)};if saveErr:=s.store.Save(s.state);saveErr!=nil{return DeviceAttestationRecord{},saveErr};if err!=nil{return record,fmt.Errorf("%w: %v",ErrAttestationRejected,err)};return record,nil
}

func(s *AttestationService)Device(accountID,deviceID string)(DeviceAttestationRecord,bool){s.mu.Lock();defer s.mu.Unlock();for _,record:=range s.state.Devices{if record.AccountID==accountID&&record.DeviceID==deviceID{return record,true}};return DeviceAttestationRecord{},false}
func(s *AttestationService)Records(accountID string)[]DeviceAttestationRecord{s.mu.Lock();defer s.mu.Unlock();out:=[]DeviceAttestationRecord{};for _,record:=range s.state.Devices{if record.AccountID==accountID{out=append(out,record)}};return out}
func publicAttestationChallenge(c AttestationChallenge)AttestationChallenge{c.AccountID="";return c}
func attestationToken(n int)(string,error){b:=make([]byte,n);if _,err:=rand.Read(b);err!=nil{return "",err};return base64.RawURLEncoding.EncodeToString(b),nil}
