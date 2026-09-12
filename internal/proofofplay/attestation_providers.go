package proofofplay

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
)

type AppleValidationRequest struct {
	ChallengeNonce string
	AttestationObject string
	KeyID string
	BundleID string
	TeamID string
	Environment string
}

type AppleValidationResult struct {
	KeyID string
	ReceiptDigest string
	AssertionCounter uint64
	HardwareBacked bool
}

type AppleAttestationValidator interface {
	ValidateAppAttest(ctx context.Context, request AppleValidationRequest) (AppleValidationResult,error)
}

type AppleAppAttestVerifier struct {
	Config AttestationConfig
	Validator AppleAttestationValidator
}
func(v AppleAppAttestVerifier)Provider()string{return ProviderAppleAppAttest}
func(v AppleAppAttestVerifier)VerifyEnrollment(ctx context.Context,challenge AttestationChallenge,evidence AttestationEvidence)(AttestationResult,error){
	if v.Validator==nil{return AttestationResult{},errors.New("Apple App Attest validator not configured")}
	if evidence.KeyID==""||evidence.Payload==""{return AttestationResult{},errors.New("missing Apple App Attest key or object")}
	result,err:=v.Validator.ValidateAppAttest(ctx,AppleValidationRequest{ChallengeNonce:challenge.Nonce,AttestationObject:evidence.Payload,KeyID:evidence.KeyID,BundleID:v.Config.AppleBundleID,TeamID:v.Config.AppleTeamID,Environment:v.Config.AppleEnvironment})
	if err!=nil{return AttestationResult{},err}
	if result.KeyID!=""&&result.KeyID!=evidence.KeyID{return AttestationResult{},errors.New("Apple App Attest key mismatch")}
	digest:=result.ReceiptDigest
	if digest==""{sum:=sha256.Sum256([]byte(evidence.Payload));digest=hex.EncodeToString(sum[:])}
	return AttestationResult{Provider:ProviderAppleAppAttest,ProviderKeyID:evidence.KeyID,EvidenceDigest:digest,IntegrityLabels:[]string{"APP_ATTEST_VERIFIED"},HardwareBacked:result.HardwareBacked,ProductionEligible:result.HardwareBacked,AssertionCounter:result.AssertionCounter},nil
}

type PlayIntegrityVerdict struct {
	RequestHash string
	PackageName string
	CertificateDigests []string
	AppRecognitionVerdict string
	DeviceIntegrity []string
	RecentDeviceActivity string
}

type PlayIntegrityDecoder interface {
	DecodeIntegrityToken(ctx context.Context,packageName,token string)(PlayIntegrityVerdict,error)
}

type GooglePlayIntegrityVerifier struct {
	Config AttestationConfig
	Decoder PlayIntegrityDecoder
}
func(v GooglePlayIntegrityVerifier)Provider()string{return ProviderGooglePlayIntegrity}
func(v GooglePlayIntegrityVerifier)VerifyEnrollment(ctx context.Context,challenge AttestationChallenge,evidence AttestationEvidence)(AttestationResult,error){
	if v.Decoder==nil{return AttestationResult{},errors.New("Play Integrity decoder not configured")}
	if evidence.Payload==""{return AttestationResult{},errors.New("missing Play Integrity token")}
	expectedHash:=playIntegrityRequestHash(challenge)
	verdict,err:=v.Decoder.DecodeIntegrityToken(ctx,v.Config.AndroidPackageName,evidence.Payload);if err!=nil{return AttestationResult{},err}
	if verdict.RequestHash!=expectedHash{return AttestationResult{},errors.New("Play Integrity request hash mismatch")}
	if v.Config.AndroidPackageName!=""&&verdict.PackageName!=v.Config.AndroidPackageName{return AttestationResult{},errors.New("Play Integrity package mismatch")}
	if verdict.AppRecognitionVerdict!="PLAY_RECOGNIZED"{return AttestationResult{},fmt.Errorf("Play Integrity app verdict %q",verdict.AppRecognitionVerdict)}
	if len(v.Config.AndroidAllowedCertificates)>0&&!intersects(verdict.CertificateDigests,v.Config.AndroidAllowedCertificates){return AttestationResult{},errors.New("Play Integrity certificate mismatch")}
	labels:=append([]string(nil),verdict.DeviceIntegrity...);sort.Strings(labels)
	hardwareBacked:=contains(labels,"MEETS_DEVICE_INTEGRITY")||contains(labels,"MEETS_STRONG_INTEGRITY")
	if v.Config.AndroidRequireStrongIntegrity&&!contains(labels,"MEETS_STRONG_INTEGRITY"){return AttestationResult{},errors.New("Play Integrity strong integrity required")}
	if !hardwareBacked{return AttestationResult{},errors.New("Play Integrity device integrity not met")}
	sum:=sha256.Sum256([]byte(evidence.Payload))
	return AttestationResult{Provider:ProviderGooglePlayIntegrity,EvidenceDigest:hex.EncodeToString(sum[:]),IntegrityLabels:labels,HardwareBacked:true,ProductionEligible:true},nil
}

func playIntegrityRequestHash(challenge AttestationChallenge)string{sum:=sha256.Sum256([]byte("bbw-attest/v1|"+challenge.ID+"|"+challenge.DeviceID+"|"+challenge.Nonce));return base64.RawURLEncoding.EncodeToString(sum[:])}
func contains(items []string,want string)bool{for _,item:=range items{if item==want{return true}};return false}
func intersects(a,b []string)bool{set:=map[string]struct{}{};for _,item:=range a{set[item]=struct{}{}};for _,item:=range b{if _,ok:=set[item];ok{return true}};return false}

type DevelopmentAttestationVerifier struct{}
func(DevelopmentAttestationVerifier)Provider()string{return ProviderDevelopment}
func(DevelopmentAttestationVerifier)VerifyEnrollment(_ context.Context,challenge AttestationChallenge,evidence AttestationEvidence)(AttestationResult,error){
	if evidence.Payload!="allow:"+challenge.Nonce{return AttestationResult{},errors.New("development attestation rejected")}
	sum:=sha256.Sum256([]byte(evidence.Payload))
	return AttestationResult{Provider:ProviderDevelopment,EvidenceDigest:hex.EncodeToString(sum[:]),IntegrityLabels:[]string{"DEVELOPMENT_ONLY"},HardwareBacked:false,ProductionEligible:false},nil
}
