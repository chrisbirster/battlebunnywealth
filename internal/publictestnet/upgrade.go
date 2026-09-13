package publictestnet

import("crypto/sha256";"encoding/hex";"encoding/json";"errors";"sync")
const MinUpgradeNoticeBlocks uint64=1008
var ErrInvalidUpgrade=errors.New("invalid public-testnet protocol upgrade")
type UpgradePlan struct{NetworkID string `json:"networkId"`;FromVersion int `json:"fromVersion"`;ToVersion int `json:"toVersion"`;ActivationHeight uint64 `json:"activationHeight"`;PolicyHash string `json:"policyHash"`;MinimumSoftware string `json:"minimumSoftware"`;Hash string `json:"hash"`}
func NewUpgradePlan(networkID string,from,to int,activationHeight uint64,policyHash,minimumSoftware string)UpgradePlan{p:=UpgradePlan{NetworkID:networkID,FromVersion:from,ToVersion:to,ActivationHeight:activationHeight,PolicyHash:policyHash,MinimumSoftware:minimumSoftware};p.Hash=upgradeHash(p);return p}
func(p UpgradePlan)Validate(networkID string,currentVersion int,currentHeight uint64)error{if p.NetworkID!=networkID||p.FromVersion!=currentVersion||p.ToVersion!=currentVersion+1||p.MinimumSoftware==""||len(p.PolicyHash)!=64{return ErrInvalidUpgrade};if p.ActivationHeight<currentHeight+MinUpgradeNoticeBlocks||p.Hash!=upgradeHash(p){return ErrInvalidUpgrade};return nil}
func upgradeHash(p UpgradePlan)string{p.Hash="";raw,_:=json.Marshal(p);sum:=sha256.Sum256(raw);return hex.EncodeToString(sum[:])}
type UpgradeManager struct{mu sync.Mutex;network string;version int;plan *UpgradePlan}
func NewUpgradeManager(networkID string,currentVersion int)*UpgradeManager{return &UpgradeManager{network:networkID,version:currentVersion}}
func(m *UpgradeManager)Stage(plan UpgradePlan,currentHeight uint64)error{m.mu.Lock();defer m.mu.Unlock();if m.plan!=nil||plan.Validate(m.network,m.version,currentHeight)!=nil{return ErrInvalidUpgrade};p:=plan;m.plan=&p;return nil}
func(m *UpgradeManager)VersionAt(height uint64)int{m.mu.Lock();defer m.mu.Unlock();if m.plan!=nil&&height>=m.plan.ActivationHeight{return m.plan.ToVersion};return m.version}
func(m *UpgradeManager)Plan()*UpgradePlan{m.mu.Lock();defer m.mu.Unlock();if m.plan==nil{return nil};p:=*m.plan;return &p}
