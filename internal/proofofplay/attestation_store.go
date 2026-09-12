package proofofplay

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var ErrAttestationStateNotFound = errors.New("proof-of-play attestation state not found")

type AttestationFileStore struct{ Path string }

func NewAttestationFileStore(path string) *AttestationFileStore { return &AttestationFileStore{Path:path} }

func (s *AttestationFileStore) Load() (AttestationState,error) {
	data,err:=os.ReadFile(s.Path)
	if errors.Is(err,os.ErrNotExist){return AttestationState{},ErrAttestationStateNotFound}
	if err!=nil{return AttestationState{},fmt.Errorf("read attestation state: %w",err)}
	var state AttestationState
	if err:=json.Unmarshal(data,&state);err!=nil{return AttestationState{},fmt.Errorf("decode attestation state: %w",err)}
	return state,nil
}
func (s *AttestationFileStore) Save(state AttestationState) error {
	if err:=os.MkdirAll(filepath.Dir(s.Path),0o755);err!=nil{return fmt.Errorf("create attestation state directory: %w",err)}
	data,err:=json.MarshalIndent(state,"","  ");if err!=nil{return fmt.Errorf("encode attestation state: %w",err)}
	tmp:=s.Path+".tmp";if err:=os.WriteFile(tmp,append(data,'\n'),0o600);err!=nil{return fmt.Errorf("write attestation state: %w",err)}
	if err:=os.Rename(tmp,s.Path);err!=nil{return fmt.Errorf("replace attestation state: %w",err)}
	return nil
}

type MemoryAttestationStore struct{mu sync.Mutex;state *AttestationState}
func NewMemoryAttestationStore()*MemoryAttestationStore{return &MemoryAttestationStore{}}
func(s *MemoryAttestationStore)Load()(AttestationState,error){s.mu.Lock();defer s.mu.Unlock();if s.state==nil{return AttestationState{},ErrAttestationStateNotFound};return cloneAttestationState(*s.state),nil}
func(s *MemoryAttestationStore)Save(state AttestationState)error{s.mu.Lock();defer s.mu.Unlock();copy:=cloneAttestationState(state);s.state=&copy;return nil}
func cloneAttestationState(state AttestationState)AttestationState{state.Challenges=append([]AttestationChallenge(nil),state.Challenges...);state.Devices=append([]DeviceAttestationRecord(nil),state.Devices...);for i:=range state.Devices{state.Devices[i].IntegrityLabels=append([]string(nil),state.Devices[i].IntegrityLabels...)};return state}
