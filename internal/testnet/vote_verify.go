package testnet

// VerifyCommitVote validates a commit vote against one validator identity.
// It is exported so higher-level deterministic state transitions can verify
// a finality certificate that is explicitly committed into a later block.
func VerifyCommitVote(v Vote, validator Validator) bool {
	if v.Decision != "commit" || v.ValidatorID != validator.ID {
		return false
	}
	return VerifySignature(validator.Algorithm, validator.PublicKey, voteSigningMessage(v), v.Signature)
}
