package publictestnet

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

func TestDirectoryAppliesTrustedASNAndProviderLimits(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	directory := NewDirectory("public", 64, 4)
	directory.SetNetworkClassifier(NetworkClassifierFunc(func(ip net.IP) PeerNetworkMetadata {
		v4 := ip.To4()
		if v4 == nil {
			return PeerNetworkMetadata{}
		}
		if v4[0] == 203 {
			return PeerNetworkMetadata{ASN: 64500, Provider: "cloud-a"}
		}
		return PeerNetworkMetadata{ASN: 64600 + uint32(v4[2]), Provider: "cloud-b"}
	}), 3, 4)

	for i := 1; i <= 3; i++ {
		key, _ := testnet.GenerateNodeKey()
		announcement, _ := BuildAnnouncement("public", fmt.Sprintf("http://203.0.%d.1:9000", i), fmt.Sprintf("n%d", i), key, now, time.Hour)
		if err := directory.Upsert(announcement, now); err != nil {
			t.Fatal(err)
		}
	}
	key, _ := testnet.GenerateNodeKey()
	fourthSameASN, _ := BuildAnnouncement("public", "http://203.0.9.1:9000", "same-asn", key, now, time.Hour)
	if err := directory.Upsert(fourthSameASN, now); !errors.Is(err, ErrPeerASNLimit) {
		t.Fatalf("same ASN err=%v", err)
	}

	for i := 1; i <= 4; i++ {
		key, _ := testnet.GenerateNodeKey()
		announcement, _ := BuildAnnouncement("public", fmt.Sprintf("http://198.18.%d.1:9000", i), fmt.Sprintf("p%d", i), key, now, time.Hour)
		if err := directory.Upsert(announcement, now); err != nil {
			t.Fatal(err)
		}
	}
	key, _ = testnet.GenerateNodeKey()
	fifthSameProvider, _ := BuildAnnouncement("public", "http://198.18.9.1:9000", "same-provider", key, now, time.Hour)
	if err := directory.Upsert(fifthSameProvider, now); !errors.Is(err, ErrPeerProviderLimit) {
		t.Fatalf("same provider err=%v", err)
	}
}
