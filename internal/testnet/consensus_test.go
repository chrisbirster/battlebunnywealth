package testnet

import "testing"

// TestProtocolV3HarnessCompiles is a temporary compatibility sentinel while the
// legacy v0.9 test harness is migrated to protocol-v3 state-machine semantics.
func TestProtocolV3HarnessCompiles(t *testing.T) {
	if ProtocolVersion != 3 {
		t.Fatalf("protocol version=%d", ProtocolVersion)
	}
}
