package proofofplay

import (
	"testing"
	"time"
)

func TestWeightedMissionAwardsDiminishAdditionalDevices(t *testing.T){
	now:=time.Date(2026,9,12,12,0,0,0,time.UTC);config:=DefaultAuthorityConfig();config.MaxMissionsPerDay=10;service,err:=NewAuthorityService(func()time.Time{return now},NewMemoryAuthorityStore(),config);if err!=nil{t.Fatal(err)}
	key1:=newTestDeviceKey(t);snapshot,err:=service.IssueWeightedMission("account-a","device-1",key1.spki,"verified","checkpoint",1,DeviceWeightPercent(0));if err!=nil{t.Fatal(err)};m:=snapshot.CurrentMission;if m.AuthorityAward!=25{t.Fatalf("first award=%d",m.AuthorityAward)};if _,err:=service.CompleteMission("account-a","device-1",key1.spki,m.ID,signMission(t,key1.private,m.SignedPayload));err!=nil{t.Fatal(err)}
	key2:=newTestDeviceKey(t);snapshot,err=service.IssueWeightedMission("account-a","device-2",key2.spki,"verified","checkpoint",2,DeviceWeightPercent(1));if err!=nil{t.Fatal(err)};m=snapshot.CurrentMission;if m.AuthorityAward!=6{t.Fatalf("second award=%d want 6",m.AuthorityAward)};completed,err:=service.CompleteMission("account-a","device-2",key2.spki,m.ID,signMission(t,key2.private,m.SignedPayload));if err!=nil{t.Fatal(err)}
	if completed.Authority.Score!=31{t.Fatalf("authority=%d want 31",completed.Authority.Score)}
}

func TestAttestedSnapshotAndProductionCommitteeRequireExternalEligibility(t *testing.T){
	now:=time.Date(2026,9,12,12,0,0,0,time.UTC);config:=DefaultAuthorityConfig();config.EligibilityThreshold=25;service,err:=NewAuthorityService(func()time.Time{return now},NewMemoryAuthorityStore(),config);if err!=nil{t.Fatal(err)};key:=newTestDeviceKey(t);completeOne(t,service,"account-a","device-a",key,1)
	plain,err:=service.SnapshotWithAttestation("account-a",false);if err!=nil{t.Fatal(err)};if plain.Authority.ProductionEligible{t.Fatal("unattested account must not be production eligible")}
	attested,err:=service.SnapshotWithAttestation("account-a",true);if err!=nil{t.Fatal(err)};if !attested.Authority.ProductionEligible{t.Fatal("attested account above threshold should be eligible for permissioned-testnet research")}
	if got:=service.SelectProductionCommittee(1,"seed",1,map[string]bool{"account-a":false});len(got)!=0{t.Fatalf("ineligible committee=%+v",got)}
	got:=service.SelectProductionCommittee(1,"seed",1,map[string]bool{"account-a":true});if len(got)!=1||!got[0].ProductionEligible{t.Fatalf("committee=%+v",got)}
}
