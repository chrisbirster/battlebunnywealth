package testnet

import (
	"errors"
	"fmt"
)

var ErrInvalidEvidence = errors.New("invalid consensus evidence")

type Evidence struct {
	Kind        string    `json:"kind"`
	NetworkID   string    `json:"networkId"`
	Height      uint64    `json:"height"`
	Round       uint32    `json:"round"`
	ValidatorID string    `json:"validatorId"`
	FirstHash   string    `json:"firstHash"`
	SecondHash  string    `json:"secondHash"`
	ProposalA   *Proposal `json:"proposalA,omitempty"`
	ProposalB   *Proposal `json:"proposalB,omitempty"`
	VoteA       *Vote     `json:"voteA,omitempty"`
	VoteB       *Vote     `json:"voteB,omitempty"`
}

func BuildProposalEquivocationEvidence(a, b Proposal, validators []Validator) (Evidence, error) {
	if a.Block.NetworkID == "" || a.Block.NetworkID != b.Block.NetworkID || a.Block.Height != b.Block.Height || a.Block.Round != b.Block.Round || a.Block.ProposerID == "" || a.Block.ProposerID != b.Block.ProposerID || a.Block.Hash == b.Block.Hash {
		return Evidence{}, ErrInvalidEvidence
	}
	validator, ok := findValidator(validators, a.Block.ProposerID)
	if !ok {
		return Evidence{}, ErrUnknownValidator
	}
	if hashBlock(a.Block) != a.Block.Hash || hashBlock(b.Block) != b.Block.Hash {
		return Evidence{}, ErrInvalidEvidence
	}
	if !VerifySignature(validator.Algorithm, validator.PublicKey, proposalSigningMessage(a.Block.Hash), a.Signature) || !VerifySignature(validator.Algorithm, validator.PublicKey, proposalSigningMessage(b.Block.Hash), b.Signature) {
		return Evidence{}, ErrInvalidSignature
	}
	return Evidence{Kind: "proposal-equivocation", NetworkID: a.Block.NetworkID, Height: a.Block.Height, Round: a.Block.Round, ValidatorID: a.Block.ProposerID, FirstHash: a.Block.Hash, SecondHash: b.Block.Hash, ProposalA: &a, ProposalB: &b}, nil
}

func BuildVoteEquivocationEvidence(a, b Vote, validators []Validator) (Evidence, error) {
	if a.NetworkID == "" || a.NetworkID != b.NetworkID || a.Height != b.Height || a.Round != b.Round || a.ValidatorID == "" || a.ValidatorID != b.ValidatorID || a.BlockHash == b.BlockHash || a.Decision != "commit" || b.Decision != "commit" {
		return Evidence{}, ErrInvalidEvidence
	}
	validator, ok := findValidator(validators, a.ValidatorID)
	if !ok {
		return Evidence{}, ErrUnknownValidator
	}
	if !VerifySignature(validator.Algorithm, validator.PublicKey, voteSigningMessage(a), a.Signature) || !VerifySignature(validator.Algorithm, validator.PublicKey, voteSigningMessage(b), b.Signature) {
		return Evidence{}, ErrInvalidSignature
	}
	return Evidence{Kind: "vote-equivocation", NetworkID: a.NetworkID, Height: a.Height, Round: a.Round, ValidatorID: a.ValidatorID, FirstHash: a.BlockHash, SecondHash: b.BlockHash, VoteA: &a, VoteB: &b}, nil
}

func VerifyEvidence(e Evidence, validators []Validator) error {
	switch e.Kind {
	case "proposal-equivocation":
		if e.ProposalA == nil || e.ProposalB == nil {
			return ErrInvalidEvidence
		}
		verified, err := BuildProposalEquivocationEvidence(*e.ProposalA, *e.ProposalB, validators)
		if err != nil {
			return err
		}
		if verified.ValidatorID != e.ValidatorID || verified.FirstHash != e.FirstHash || verified.SecondHash != e.SecondHash {
			return ErrInvalidEvidence
		}
		return nil
	case "vote-equivocation":
		if e.VoteA == nil || e.VoteB == nil {
			return ErrInvalidEvidence
		}
		verified, err := BuildVoteEquivocationEvidence(*e.VoteA, *e.VoteB, validators)
		if err != nil {
			return err
		}
		if verified.ValidatorID != e.ValidatorID || verified.FirstHash != e.FirstHash || verified.SecondHash != e.SecondHash {
			return ErrInvalidEvidence
		}
		return nil
	default:
		return fmt.Errorf("%w: unknown kind %q", ErrInvalidEvidence, e.Kind)
	}
}
