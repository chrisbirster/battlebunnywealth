package identity

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrUnauthorized       = errors.New("authentication required")
	ErrInvalidCeremony    = errors.New("invalid or expired passkey ceremony")
	ErrInvalidDisplayName = errors.New("display name must be 2-40 characters")
	ErrCredentialExists   = errors.New("passkey already registered")
	ErrUnknownCredential  = errors.New("unknown passkey")
	ErrInvalidDevice      = errors.New("invalid device enrollment")
	ErrUnknownDevice      = errors.New("unknown device")
)

type Config struct { RPID string; RPName string; AllowedOrigins []string; SessionTTL time.Duration }
type ceremony struct { Kind string; Challenge string; Origin string; ExpiresAt time.Time; AccountID string; UserHandle string; DisplayName string }
type BeginRegistrationResult struct { Ceremony string `json:"ceremony"`; PublicKey RegistrationOptions `json:"publicKey"` }
type BeginLoginResult struct { Ceremony string `json:"ceremony"`; PublicKey LoginOptions `json:"publicKey"` }

type Service struct { mu sync.Mutex; now func() time.Time; store Store; resolver DIDResolver; config Config; state State; ceremonies map[string]ceremony; origins map[string]struct{} }

func NewService(now func() time.Time, store Store, resolver DIDResolver, config Config) (*Service,error) {
	if now==nil { now=time.Now }; if resolver==nil { resolver=NewPLCResolver(nil) }; if config.RPID=="" { return nil,errors.New("RP ID required") }; if config.RPName=="" { config.RPName="Battle Bunny Wealth" }; if config.SessionTTL<=0 { config.SessionTTL=30*24*time.Hour }
	origins:=map[string]struct{}{}; for _,origin:=range config.AllowedOrigins { origin=strings.TrimRight(strings.TrimSpace(origin),"/"); if origin!="" { origins[origin]=struct{}{} } }; if len(origins)==0 { return nil,errors.New("at least one allowed origin required") }
	state,err:=store.Load(); if errors.Is(err,ErrNotFound) { state=State{SchemaVersion:SchemaVersion,Accounts:[]Account{},Sessions:[]Session{}}; if err:=store.Save(state);err!=nil{return nil,err} } else if err!=nil{return nil,err}; if state.SchemaVersion!=SchemaVersion{return nil,fmt.Errorf("identity schema %d unsupported; want %d",state.SchemaVersion,SchemaVersion)}
	return &Service{now:now,store:store,resolver:resolver,config:config,state:state,ceremonies:map[string]ceremony{},origins:origins},nil
}

func (s *Service) AllowedOrigin(origin string) bool { origin=strings.TrimRight(strings.TrimSpace(origin),"/"); _,ok:=s.origins[origin]; return ok }

func (s *Service) BeginRegistration(origin,displayName,accountID string) (BeginRegistrationResult,error) {
	s.mu.Lock(); defer s.mu.Unlock(); origin=strings.TrimRight(origin,"/"); if !s.AllowedOrigin(origin){return BeginRegistrationResult{},ErrInvalidCeremony}; displayName=strings.TrimSpace(displayName); var userHandle string
	if accountID!="" { idx:=s.accountIndexLocked(accountID); if idx<0{return BeginRegistrationResult{},ErrUnauthorized}; if displayName==""{displayName=s.state.Accounts[idx].DisplayName}; userHandle=s.state.Accounts[idx].UserHandle } else { if len(displayName)<2||len(displayName)>40{return BeginRegistrationResult{},ErrInvalidDisplayName}; var err error; userHandle,err=randomToken(32);if err!=nil{return BeginRegistrationResult{},err} }
	challenge,err:=randomToken(32);if err!=nil{return BeginRegistrationResult{},err}; ceremonyID,err:=randomToken(24);if err!=nil{return BeginRegistrationResult{},err}; s.ceremonies[ceremonyID]=ceremony{Kind:"register",Challenge:challenge,Origin:origin,ExpiresAt:s.now().UTC().Add(5*time.Minute),AccountID:accountID,UserHandle:userHandle,DisplayName:displayName}
	var options RegistrationOptions; options.Challenge=challenge; options.RP.ID=s.config.RPID; options.RP.Name=s.config.RPName; options.User.ID=userHandle; options.User.Name="bunny-"+strings.ToLower(strings.ReplaceAll(displayName," ","-")); options.User.DisplayName=displayName; options.PubKeyCredParams=[]struct{Type string `json:"type"`;Alg int `json:"alg"`}{{Type:"public-key",Alg:-7}}; options.Timeout=60000; options.Attestation="none"; options.AuthenticatorSelection.ResidentKey="required"; options.AuthenticatorSelection.RequireResidentKey=true; options.AuthenticatorSelection.UserVerification="preferred"
	if accountID!="" { idx:=s.accountIndexLocked(accountID); for _,cred:=range s.state.Accounts[idx].Credentials { options.ExcludeCredentials=append(options.ExcludeCredentials,CredentialDescriptor{Type:"public-key",ID:cred.ID,Transports:cred.Transports}) } }
	return BeginRegistrationResult{Ceremony:ceremonyID,PublicKey:options},nil
}

func (s *Service) FinishRegistration(ceremonyID string,input RegistrationCredential) (AccountView,string,error) {
	s.mu.Lock(); defer s.mu.Unlock(); c,ok:=s.takeCeremonyLocked(ceremonyID,"register");if !ok{return AccountView{},"",ErrInvalidCeremony}; credential,err:=verifyRegistration(s.config.RPID,c.Origin,c.Challenge,input);if err!=nil{return AccountView{},"",err}; now:=s.now().UTC();credential.CreatedAt=now;if ai,_:=s.credentialIndexLocked(credential.ID);ai>=0{return AccountView{},"",ErrCredentialExists}; var idx int
	if c.AccountID=="" { accountID,err:=randomToken(18);if err!=nil{return AccountView{},"",err};s.state.Accounts=append(s.state.Accounts,Account{ID:accountID,UserHandle:c.UserHandle,DisplayName:c.DisplayName,CreatedAt:now,Credentials:[]PasskeyCredential{credential},Devices:[]Device{}});idx=len(s.state.Accounts)-1 } else { idx=s.accountIndexLocked(c.AccountID);if idx<0{return AccountView{},"",ErrUnauthorized};s.state.Accounts[idx].Credentials=append(s.state.Accounts[idx].Credentials,credential) }
	token,err:=s.newSessionLocked(s.state.Accounts[idx].ID);if err!=nil{return AccountView{},"",err};if err:=s.store.Save(s.state);err!=nil{return AccountView{},"",err};return view(s.state.Accounts[idx]),token,nil
}

func (s *Service) BeginLogin(origin string) (BeginLoginResult,error) { s.mu.Lock();defer s.mu.Unlock();origin=strings.TrimRight(origin,"/");if !s.AllowedOrigin(origin){return BeginLoginResult{},ErrInvalidCeremony};challenge,err:=randomToken(32);if err!=nil{return BeginLoginResult{},err};ceremonyID,err:=randomToken(24);if err!=nil{return BeginLoginResult{},err};s.ceremonies[ceremonyID]=ceremony{Kind:"login",Challenge:challenge,Origin:origin,ExpiresAt:s.now().UTC().Add(5*time.Minute)};return BeginLoginResult{Ceremony:ceremonyID,PublicKey:LoginOptions{Challenge:challenge,RPID:s.config.RPID,Timeout:60000,UserVerification:"preferred"}},nil }
func (s *Service) FinishLogin(ceremonyID string,input AssertionCredential) (AccountView,string,error) { s.mu.Lock();defer s.mu.Unlock();c,ok:=s.takeCeremonyLocked(ceremonyID,"login");if !ok{return AccountView{},"",ErrInvalidCeremony};accountIdx,credentialIdx:=s.credentialIndexLocked(input.ID);if accountIdx<0{return AccountView{},"",ErrUnknownCredential};account:=&s.state.Accounts[accountIdx];stored:=account.Credentials[credentialIdx];newCount,err:=verifyAssertion(s.config.RPID,c.Origin,c.Challenge,account.UserHandle,stored,input);if err!=nil{return AccountView{},"",err};account.Credentials[credentialIdx].SignCount=newCount;account.Credentials[credentialIdx].LastUsedAt=s.now().UTC();token,err:=s.newSessionLocked(account.ID);if err!=nil{return AccountView{},"",err};if err:=s.store.Save(s.state);err!=nil{return AccountView{},"",err};return view(*account),token,nil }

func (s *Service) Authenticate(token string) (AccountView,error) { s.mu.Lock();defer s.mu.Unlock();if token==""{return AccountView{},ErrUnauthorized};hash:=hashToken(token);now:=s.now().UTC();changed:=false;for i:=len(s.state.Sessions)-1;i>=0;i--{if !s.state.Sessions[i].ExpiresAt.After(now){s.state.Sessions=append(s.state.Sessions[:i],s.state.Sessions[i+1:]...);changed=true}};if changed{_ = s.store.Save(s.state)};for _,session:=range s.state.Sessions{if session.TokenHash==hash{idx:=s.accountIndexLocked(session.AccountID);if idx>=0{return view(s.state.Accounts[idx]),nil}}};return AccountView{},ErrUnauthorized }
func (s *Service) Logout(token string) error { s.mu.Lock();defer s.mu.Unlock();hash:=hashToken(token);for i:=range s.state.Sessions{if s.state.Sessions[i].TokenHash==hash{s.state.Sessions=append(s.state.Sessions[:i],s.state.Sessions[i+1:]...);return s.store.Save(s.state)}};return nil }

func (s *Service) LinkATProto(ctx context.Context,accountID,did string) (AccountView,error) { resolved,err:=s.resolver.Resolve(ctx,did);if err!=nil{return AccountView{},err};s.mu.Lock();defer s.mu.Unlock();idx:=s.accountIndexLocked(accountID);if idx<0{return AccountView{},ErrUnauthorized};s.state.Accounts[idx].ATProto=&ATProtoBinding{DID:resolved.DID,Handle:resolved.Handle,PDS:resolved.PDS,Status:"resolved-unverified",BoundAt:s.now().UTC()};if err:=s.store.Save(s.state);err!=nil{return AccountView{},err};return view(s.state.Accounts[idx]),nil }
func (s *Service) UnlinkATProto(accountID string) (AccountView,error) { s.mu.Lock();defer s.mu.Unlock();idx:=s.accountIndexLocked(accountID);if idx<0{return AccountView{},ErrUnauthorized};s.state.Accounts[idx].ATProto=nil;if err:=s.store.Save(s.state);err!=nil{return AccountView{},err};return view(s.state.Accounts[idx]),nil }

func (s *Service) EnrollDevice(accountID,name,platform,publicKeySPKI string) (AccountView,error) { name=strings.TrimSpace(name);platform=strings.ToLower(strings.TrimSpace(platform));if len(name)<2||len(name)>48||!allowedPlatform(platform){return AccountView{},ErrInvalidDevice};keyRaw,err:=base64.RawURLEncoding.DecodeString(publicKeySPKI);if err!=nil{return AccountView{},ErrInvalidDevice};parsed,err:=x509.ParsePKIXPublicKey(keyRaw);if err!=nil{return AccountView{},ErrInvalidDevice};pub,ok:=parsed.(*ecdsa.PublicKey);if !ok||pub.Curve!=elliptic.P256(){return AccountView{},ErrInvalidDevice};deviceID,err:=randomToken(18);if err!=nil{return AccountView{},err};s.mu.Lock();defer s.mu.Unlock();idx:=s.accountIndexLocked(accountID);if idx<0{return AccountView{},ErrUnauthorized};s.state.Accounts[idx].Devices=append(s.state.Accounts[idx].Devices,Device{ID:deviceID,Name:name,Platform:platform,PublicKeySPKI:publicKeySPKI,Status:DeviceStatusActive,AttestationStatus:"unattested",EnrolledAt:s.now().UTC()});if err:=s.store.Save(s.state);err!=nil{return AccountView{},err};return view(s.state.Accounts[idx]),nil }
func (s *Service) RevokeDevice(accountID,deviceID string) (AccountView,error) { s.mu.Lock();defer s.mu.Unlock();idx:=s.accountIndexLocked(accountID);if idx<0{return AccountView{},ErrUnauthorized};for i:=range s.state.Accounts[idx].Devices{if s.state.Accounts[idx].Devices[i].ID==deviceID{if s.state.Accounts[idx].Devices[i].Status!=DeviceStatusRevoked{now:=s.now().UTC();s.state.Accounts[idx].Devices[i].Status=DeviceStatusRevoked;s.state.Accounts[idx].Devices[i].RevokedAt=&now};if err:=s.store.Save(s.state);err!=nil{return AccountView{},err};return view(s.state.Accounts[idx]),nil}};return AccountView{},ErrUnknownDevice }

func (s *Service) takeCeremonyLocked(id,kind string) (ceremony,bool) { c,ok:=s.ceremonies[id];delete(s.ceremonies,id);if !ok||c.Kind!=kind||!c.ExpiresAt.After(s.now().UTC()){return ceremony{},false};return c,true }
func (s *Service) newSessionLocked(accountID string) (string,error) { token,err:=randomToken(32);if err!=nil{return "",err};now:=s.now().UTC();s.state.Sessions=append(s.state.Sessions,Session{TokenHash:hashToken(token),AccountID:accountID,CreatedAt:now,ExpiresAt:now.Add(s.config.SessionTTL)});return token,nil }
func (s *Service) accountIndexLocked(id string) int { for i:=range s.state.Accounts{if s.state.Accounts[i].ID==id{return i}};return -1 }
func (s *Service) credentialIndexLocked(id string) (int,int) { for ai:=range s.state.Accounts{for ci:=range s.state.Accounts[ai].Credentials{if s.state.Accounts[ai].Credentials[ci].ID==id{return ai,ci}}};return -1,-1 }
func randomToken(n int) (string,error) { b:=make([]byte,n);if _,err:=rand.Read(b);err!=nil{return "",err};return base64.RawURLEncoding.EncodeToString(b),nil }
func hashToken(token string) string { sum:=sha256.Sum256([]byte(token));return hex.EncodeToString(sum[:]) }
func allowedPlatform(v string) bool { return v=="web"||v=="ios"||v=="android"||v=="desktop" }
func view(account Account) AccountView { devices:=append([]Device(nil),account.Devices...);var binding *ATProtoBinding;if account.ATProto!=nil{b:=*account.ATProto;binding=&b};return AccountView{ID:account.ID,DisplayName:account.DisplayName,CreatedAt:account.CreatedAt,PasskeyCount:len(account.Credentials),ATProto:binding,Devices:devices} }
