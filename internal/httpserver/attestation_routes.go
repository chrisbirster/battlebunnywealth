package httpserver

import (
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/identity"
	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
)

func WithAttestation(base http.Handler, protocol proofOfPlay, authority *proofofplay.AuthorityService, attestation *proofofplay.AttestationService, auth *identity.Service) http.Handler {
	mux:=http.NewServeMux()
	mux.HandleFunc("GET /api/v1/proof-of-play/attestations",func(w http.ResponseWriter,r *http.Request){account,err:=currentAccount(auth,r);if err!=nil{writeIdentityError(w,err);return};writeJSON(w,http.StatusOK,attestation.Records(account.ID))})
	mux.HandleFunc("POST /api/v1/proof-of-play/attestations/begin",func(w http.ResponseWriter,r *http.Request){
		account,err:=currentAccount(auth,r);if err!=nil{writeIdentityError(w,err);return};var body struct{DeviceID string `json:"deviceId"`;Provider string `json:"provider"`};if err:=decodeJSON(w,r,&body);err!=nil{writeJSON(w,http.StatusBadRequest,map[string]string{"error":err.Error()});return};device,err:=activeDevice(account,body.DeviceID);if err!=nil{writeIdentityError(w,err);return};if !providerMatchesPlatform(body.Provider,device.Platform){writeJSON(w,http.StatusBadRequest,map[string]string{"error":"attestation provider does not match device platform"});return};challenge,err:=attestation.Begin(account.ID,device.ID,device.PublicKeySPKI,body.Provider);if err!=nil{writeAttestationError(w,err);return};writeJSON(w,http.StatusOK,challenge)
	})
	mux.HandleFunc("POST /api/v1/proof-of-play/attestations/complete",func(w http.ResponseWriter,r *http.Request){
		account,err:=currentAccount(auth,r);if err!=nil{writeIdentityError(w,err);return};var evidence proofofplay.AttestationEvidence;if err:=decodeJSON(w,r,&evidence);err!=nil{writeJSON(w,http.StatusBadRequest,map[string]string{"error":err.Error()});return};device,err:=activeDevice(account,evidence.DeviceID);if err!=nil{writeIdentityError(w,err);return};if !providerMatchesPlatform(evidence.Provider,device.Platform){writeJSON(w,http.StatusBadRequest,map[string]string{"error":"attestation provider does not match device platform"});return};record,err:=attestation.Complete(r.Context(),account.ID,device.ID,evidence);if err!=nil{writeAttestationError(w,err);return};updated,err:=auth.SetDeviceAttestation(account.ID,device.ID,record.Provider,record.Status,record.HardwareBacked,record.ProductionEligible,record.VerifiedAt);if err!=nil{writeIdentityError(w,err);return};writeJSON(w,http.StatusOK,map[string]any{"attestation":record,"account":updated})
	})
	mux.HandleFunc("GET /api/v1/proof-of-play/me",func(w http.ResponseWriter,r *http.Request){account,err:=currentAccount(auth,r);if err!=nil{writeIdentityError(w,err);return};snapshot,err:=authority.SnapshotWithAttestation(account.ID,hasProductionEligibleDevice(account));if err!=nil{writeProofError(w,err);return};writeJSON(w,http.StatusOK,snapshot)})
	mux.HandleFunc("POST /api/v1/proof-of-play/missions",func(w http.ResponseWriter,r *http.Request){
		account,err:=currentAccount(auth,r);if err!=nil{writeIdentityError(w,err);return};var body struct{DeviceID string `json:"deviceId"`};if err:=decodeJSON(w,r,&body);err!=nil{writeJSON(w,http.StatusBadRequest,map[string]string{"error":err.Error()});return};device,err:=activeDevice(account,body.DeviceID);if err!=nil{writeIdentityError(w,err);return};ordinal:=activeDeviceOrdinal(account,device.ID);weight:=proofofplay.DeviceWeightPercent(ordinal);now:=time.Now().UTC();snapshot,err:=authority.IssueWeightedMission(account.ID,device.ID,device.PublicKeySPKI,device.AttestationStatus,protocol.Head().Hash,protocol.CurrentEpoch(now),weight);if err!=nil{writeProofError(w,err);return};writeJSON(w,http.StatusOK,snapshot)
	})
	mux.Handle("/",base)
	return mux
}

func providerMatchesPlatform(provider,platform string)bool{switch provider{case proofofplay.ProviderAppleAppAttest:return platform=="ios";case proofofplay.ProviderGooglePlayIntegrity:return platform=="android";case proofofplay.ProviderDevelopment:return true;default:return false}}
func hasProductionEligibleDevice(account identity.AccountView)bool{for _,d:=range account.Devices{if d.Status==identity.DeviceStatusActive&&d.AttestationStatus==proofofplay.AttestationStateVerified&&d.HardwareBacked&&d.ProductionEligible{return true}};return false}
func activeDeviceOrdinal(account identity.AccountView,deviceID string)int{devices:=make([]identity.Device,0,len(account.Devices));for _,d:=range account.Devices{if d.Status==identity.DeviceStatusActive{devices=append(devices,d)}};sort.SliceStable(devices,func(i,j int)bool{if devices[i].EnrolledAt.Equal(devices[j].EnrolledAt){return devices[i].ID<devices[j].ID};return devices[i].EnrolledAt.Before(devices[j].EnrolledAt)});for i,d:=range devices{if d.ID==deviceID{return i}};return len(devices)}
func writeAttestationError(w http.ResponseWriter,err error){status:=http.StatusInternalServerError;switch{case errors.Is(err,proofofplay.ErrAttestationChallengeNotFound):status=http.StatusNotFound;case errors.Is(err,proofofplay.ErrAttestationChallengeExpired):status=http.StatusGone;case errors.Is(err,proofofplay.ErrAttestationChallengeUsed):status=http.StatusConflict;case errors.Is(err,proofofplay.ErrAttestationBinding),errors.Is(err,proofofplay.ErrAttestationProvider),errors.Is(err,proofofplay.ErrAttestationRejected):status=http.StatusBadRequest};message:=err.Error();if status==http.StatusInternalServerError{message="internal server error"};writeJSON(w,status,map[string]string{"error":message})}
