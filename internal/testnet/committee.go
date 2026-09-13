package testnet

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
)

func ActiveValidators(validators []Validator) []Validator {
	out := make([]Validator, 0, len(validators))
	for _, v := range validators {
		if v.Active {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func SelectCommittee(validators []Validator, target int, previousHash string, height uint64, round uint32) []Validator {
	candidates := ActiveValidators(validators)
	if target <= 0 || target >= len(candidates) {
		return candidates
	}
	selected := make([]Validator, 0, target)
	for seat := 0; seat < target && len(candidates) > 0; seat++ {
		var total uint64
		for _, c := range candidates {
			w := c.Authority
			if w < 1 {
				w = 1
			}
			total += uint64(w)
		}
		digest := sha256.Sum256([]byte(fmt.Sprintf("bbw-pop-committee/v1|%s|%d|%d|%d", previousHash, height, round, seat)))
		draw := binary.BigEndian.Uint64(digest[:8]) % total
		var cumulative uint64
		pick := 0
		for i, c := range candidates {
			w := c.Authority
			if w < 1 {
				w = 1
			}
			cumulative += uint64(w)
			if draw < cumulative {
				pick = i
				break
			}
		}
		selected = append(selected, candidates[pick])
		candidates = append(candidates[:pick], candidates[pick+1:]...)
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].ID < selected[j].ID })
	return selected
}

func CommitteeHash(committee []Validator) string {
	ids := make([]string, 0, len(committee))
	for _, v := range committee {
		ids = append(ids, v.ID)
	}
	sort.Strings(ids)
	sum := sha256.Sum256([]byte(fmt.Sprintf("bbw-pop-committee/v1|%v", ids)))
	return hex.EncodeToString(sum[:])
}

func QuorumFor(cfg ProtocolConfig, committeeSize int) int {
	if committeeSize <= 0 {
		return 0
	}
	num, den := cfg.QuorumNumerator, cfg.QuorumDenominator
	if committeeSize <= cfg.BootstrapMaxValidators {
		num, den = cfg.BootstrapQuorumNumerator, cfg.BootstrapQuorumDenominator
	}
	if num <= 0 || den <= 0 || num > den {
		num, den = 2, 3
	}
	return (committeeSize*num + den - 1) / den
}

func ExpectedProposer(committee []Validator, height uint64, round uint32) (Validator, bool) {
	if len(committee) == 0 {
		return Validator{}, false
	}
	idx := int((height - 1 + uint64(round)) % uint64(len(committee)))
	return committee[idx], true
}
