package identity

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"testing"
	"time"
)

type fakeResolver struct{}
func (fakeResolver) Resolve(context.Context,string)(ResolvedDID,error){return ResolvedDID{DID:"did:plc:abcdefghijklmnopqrstuvwx",Handle:"bunny.example",PDS:"https://pds.example"},nil}

func TestPasskeyRegistrationLoginAndSession(t *testing.T){
	now:=time.Date(2026,9,11,12,0,0,0,time.UTC);svc,err:=NewService(func()time.Time{return now},NewMemoryStore(),fakeResolver{},Config{RPID:"localhost",AllowedOrigins:[]string{"http://localhost:5173"}});if err!=nil{t.Fatal(err)}
	privateKey,err:=ecdsa.GenerateKey(elliptic.P256(),rand.Reader);if err!=nil{t.Fatal(err)};begin,err:=svc.BeginRegistration("http://localhost:5173","Boomtail","");if err!=nil{t.Fatal(err)};credentialID:=[]byte("credential-one");account,token,err:=svc.FinishRegistration(begin.Ceremony,makeRegistrationCredential(t,privateKey,credentialID,begin.PublicKey.Challenge,"http://localhost:5173","localhost"));if err!=nil{t.Fatal(err)};if account.DisplayName!="Boomtail"||account.PasskeyCount!=1||token==""{t.Fatalf("account=%+v token=%q",account,token)};if _,err:=svc.Authenticate(token);err!=nil{t.Fatal(err)}
	login,err:=svc.BeginLogin("http://localhost:5173");if err!=nil{t.Fatal(err)};assertion:=makeAssertion(t,privateKey,credentialID,login.PublicKey.Challenge,"http://localhost:5173","localhost",1,"");loggedIn,secondToken,err:=svc.FinishLogin(login.Ceremony,assertion);if err!=nil{t.Fatal(err)};if loggedIn.ID!=account.ID||secondToken==""{t.Fatalf("loggedIn=%+v token=%q",loggedIn,secondToken)}
}

func TestDeviceEnrollmentAndATProtoLinkAreSeparateFromAuthentication(t *testing.T){
	now:=time.Date(2026,9,11,12,0,0,0,time.UTC);svc,err:=NewService(func()time.Time{return now},NewMemoryStore(),fakeResolver{},Config{RPID:"localhost",AllowedOrigins:[]string{"http://localhost:5173"}});if err!=nil{t.Fatal(err)};privateKey,_:=ecdsa.GenerateKey(elliptic.P256(),rand.Reader);begin,_:=svc.BeginRegistration("http://localhost:5173","Boomtail","");account,_,err:=svc.FinishRegistration(begin.Ceremony,makeRegistrationCredential(t,privateKey,[]byte("credential-two"),begin.PublicKey.Challenge,"http://localhost:5173","localhost"));if err!=nil{t.Fatal(err)}
	deviceKey,_:=ecdsa.GenerateKey(elliptic.P256(),rand.Reader);spki,_:=x509.MarshalPKIXPublicKey(&deviceKey.PublicKey);withDevice,err:=svc.EnrollDevice(account.ID,"Chrome on Mac","web",base64.RawURLEncoding.EncodeToString(spki));if err!=nil{t.Fatal(err)};if len(withDevice.Devices)!=1||withDevice.Devices[0].AttestationStatus!="unattested"{t.Fatalf("devices=%+v",withDevice.Devices)};linked,err:=svc.LinkATProto(context.Background(),account.ID,"did:plc:abcdefghijklmnopqrstuvwx");if err!=nil{t.Fatal(err)};if linked.ATProto==nil||linked.ATProto.Status!="resolved-unverified"{t.Fatalf("atproto=%+v",linked.ATProto)};revoked,err:=svc.RevokeDevice(account.ID,withDevice.Devices[0].ID);if err!=nil{t.Fatal(err)};if revoked.Devices[0].Status!=DeviceStatusRevoked{t.Fatalf("device=%+v",revoked.Devices[0])}
}

func makeRegistrationCredential(t *testing.T,key *ecdsa.PrivateKey,credentialID []byte,challenge,origin,rpID string) RegistrationCredential { t.Helper();client,_:=json.Marshal(clientData{Type:"webauthn.create",Challenge:challenge,Origin:origin});rpHash:=sha256.Sum256([]byte(rpID));auth:=append([]byte(nil),rpHash[:]...);auth=append(auth,0x41,0,0,0,0);auth=append(auth,make([]byte,16)...);length:=make([]byte,2);binary.BigEndian.PutUint16(length,uint16(len(credentialID)));auth=append(auth,length...);auth=append(auth,credentialID...);x:=key.PublicKey.X.FillBytes(make([]byte,32));y:=key.PublicKey.Y.FillBytes(make([]byte,32));cose:=cborMap([]cborPair{{cborInt(1),cborInt(2)},{cborInt(3),cborInt(-7)},{cborInt(-1),cborInt(1)},{cborInt(-2),cborBytes(x)},{cborInt(-3),cborBytes(y)}});auth=append(auth,cose...);attestation:=cborMap([]cborPair{{cborText("fmt"),cborText("none")},{cborText("attStmt"),cborMap(nil)},{cborText("authData"),cborBytes(auth)}});var out RegistrationCredential;out.ID=encodeB64(credentialID);out.RawID=out.ID;out.Type="public-key";out.Response.ClientDataJSON=encodeB64(client);out.Response.AttestationObject=encodeB64(attestation);return out }
func makeAssertion(t *testing.T,key *ecdsa.PrivateKey,credentialID []byte,challenge,origin,rpID string,count uint32,userHandle string) AssertionCredential { t.Helper();client,_:=json.Marshal(clientData{Type:"webauthn.get",Challenge:challenge,Origin:origin});rpHash:=sha256.Sum256([]byte(rpID));auth:=append([]byte(nil),rpHash[:]...);auth=append(auth,0x01,0,0,0,0);binary.BigEndian.PutUint32(auth[33:37],count);clientHash:=sha256.Sum256(client);message:=append(append([]byte(nil),auth...),clientHash[:]...);digest:=sha256.Sum256(message);sig,err:=ecdsa.SignASN1(rand.Reader,key,digest[:]);if err!=nil{t.Fatal(err)};var out AssertionCredential;out.ID=encodeB64(credentialID);out.RawID=out.ID;out.Type="public-key";out.Response.ClientDataJSON=encodeB64(client);out.Response.AuthenticatorData=encodeB64(auth);out.Response.Signature=encodeB64(sig);out.Response.UserHandle=userHandle;return out }
type cborPair struct{key,value []byte}
func cborMap(pairs []cborPair)[]byte{out:=cborHead(5,uint64(len(pairs)));for _,p:=range pairs{out=append(out,p.key...);out=append(out,p.value...)};return out};func cborText(s string)[]byte{out:=cborHead(3,uint64(len(s)));return append(out,[]byte(s)...)};func cborBytes(b []byte)[]byte{out:=cborHead(2,uint64(len(b)));return append(out,b...)};func cborInt(v int64)[]byte{if v>=0{return cborHead(0,uint64(v))};return cborHead(1,uint64(-1-v))};func cborHead(major byte,n uint64)[]byte{if n<24{return []byte{major<<5|byte(n)}};if n<256{return []byte{major<<5|24,byte(n)}};if n<65536{return []byte{major<<5|25,byte(n>>8),byte(n)}};panic("test CBOR too large")}
