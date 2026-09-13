package testnet

// SyncFromWithStore performs the same independently verified peer catch-up as
// SyncFrom and durably appends every newly imported finalized block. This keeps
// a long-running node from losing peer-synced history on its next restart.
func SyncFromWithStore(peerURL string, client HTTPClient, engine *Engine, store *Store) error {
	before := engine.Height()
	if err := SyncFrom(peerURL, client, engine); err != nil {
		return err
	}
	if store == nil || engine.Height() == before {
		return nil
	}
	blocks := engine.Finalized()
	for i := before; i < uint64(len(blocks)); i++ {
		if err := store.Append(blocks[i]); err != nil {
			return err
		}
	}
	// Imported finality supersedes any stale in-progress round snapshot.
	return store.ClearRoundProgress()
}
