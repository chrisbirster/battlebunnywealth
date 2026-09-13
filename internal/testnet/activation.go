package testnet

import (
	"sort"
	"sync"
	"time"
)

type ActivationPolicy struct {
	MinAuthority            int64         `json:"minAuthority"`
	Maturation              time.Duration `json:"maturation"`
	Window                  time.Duration `json:"window"`
	MaxActivationsPerWindow int           `json:"maxActivationsPerWindow"`
}

func DefaultActivationPolicy() ActivationPolicy {
	return ActivationPolicy{MinAuthority: 100, Maturation: 30 * 24 * time.Hour, Window: 7 * 24 * time.Hour, MaxActivationsPerWindow: 1}
}

type ValidatorCandidate struct {
	Validator Validator `json:"validator"`
	Provider  string    `json:"provider"`
	Attested  bool      `json:"attested"`
	JoinedAt  time.Time `json:"joinedAt"`
}

type ActivationEvent struct {
	ValidatorID string    `json:"validatorId"`
	ActivatedAt time.Time `json:"activatedAt"`
}

type ActivationRegistry struct {
	mu         sync.Mutex
	policy     ActivationPolicy
	active     map[string]Validator
	candidates map[string]ValidatorCandidate
	history    []ActivationEvent
}

func NewActivationRegistry(policy ActivationPolicy, genesis []Validator) *ActivationRegistry {
	r := &ActivationRegistry{policy: policy, active: map[string]Validator{}, candidates: map[string]ValidatorCandidate{}}
	for _, v := range genesis {
		if v.Active {
			r.active[v.ID] = v
		}
	}
	return r
}
func (r *ActivationRegistry) AddCandidate(c ValidatorCandidate) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.active[c.Validator.ID]; exists {
		return
	}
	r.candidates[c.Validator.ID] = c
}
func (r *ActivationRegistry) Active() []Validator {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Validator, 0, len(r.active))
	for _, v := range r.active {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (r *ActivationRegistry) Pending() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.candidates)
}
func (r *ActivationRegistry) ActivateEligible(now time.Time) []Validator {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := now.Add(-r.policy.Window)
	recent := 0
	for _, e := range r.history {
		if e.ActivatedAt.After(cutoff) {
			recent++
		}
	}
	budget := r.policy.MaxActivationsPerWindow - recent
	if budget <= 0 {
		return nil
	}
	candidates := make([]ValidatorCandidate, 0, len(r.candidates))
	for _, c := range r.candidates {
		if !c.Attested || c.Validator.Authority < r.policy.MinAuthority || now.Sub(c.JoinedAt) < r.policy.Maturation {
			continue
		}
		candidates = append(candidates, c)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].JoinedAt.Equal(candidates[j].JoinedAt) {
			return candidates[i].Validator.ID < candidates[j].Validator.ID
		}
		return candidates[i].JoinedAt.Before(candidates[j].JoinedAt)
	})
	if budget > len(candidates) {
		budget = len(candidates)
	}
	activated := make([]Validator, 0, budget)
	for i := 0; i < budget; i++ {
		c := candidates[i]
		v := c.Validator
		v.Active = true
		v.Genesis = false
		v.ActivatedAt = now.UTC()
		r.active[v.ID] = v
		delete(r.candidates, v.ID)
		r.history = append(r.history, ActivationEvent{ValidatorID: v.ID, ActivatedAt: now.UTC()})
		activated = append(activated, v)
	}
	return activated
}
