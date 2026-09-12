package testnet

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	AlgorithmEd25519  = "ed25519"
	AlgorithmP256SPKI = "p256-spki"
)

func GenerateValidatorKey(algorithm string) (publicKey string, privateKey any, err error) {
	switch algorithm {
	case AlgorithmEd25519:
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return "", nil, err
		}
		return base64.RawURLEncoding.EncodeToString(pub), priv, nil
	case AlgorithmP256SPKI:
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return "", nil, err
		}
		der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			return "", nil, err
		}
		return base64.RawURLEncoding.EncodeToString(der), priv, nil
	default:
		return "", nil, fmt.Errorf("unsupported key algorithm %q", algorithm)
	}
}

func validatorID(publicKey string) string {
	sum := sha256.Sum256([]byte(publicKey))
	return hex.EncodeToString(sum[:16])
}

func Sign(algorithm string, privateKey any, message string) (string, error) {
	digest := sha256.Sum256([]byte(message))
	switch algorithm {
	case AlgorithmEd25519:
		priv, ok := privateKey.(ed25519.PrivateKey)
		if !ok {
			return "", errors.New("ed25519 private key required")
		}
		return base64.RawURLEncoding.EncodeToString(ed25519.Sign(priv, []byte(message))), nil
	case AlgorithmP256SPKI:
		priv, ok := privateKey.(*ecdsa.PrivateKey)
		if !ok {
			return "", errors.New("P-256 private key required")
		}
		sig, err := ecdsa.SignASN1(rand.Reader, priv, digest[:])
		if err != nil {
			return "", err
		}
		return base64.RawURLEncoding.EncodeToString(sig), nil
	default:
		return "", fmt.Errorf("unsupported key algorithm %q", algorithm)
	}
}

func VerifySignature(algorithm, publicKey, message, signature string) bool {
	sig, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	switch algorithm {
	case AlgorithmEd25519:
		pub, err := base64.RawURLEncoding.DecodeString(publicKey)
		if err != nil || len(pub) != ed25519.PublicKeySize {
			return false
		}
		return ed25519.Verify(ed25519.PublicKey(pub), []byte(message), sig)
	case AlgorithmP256SPKI:
		der, err := base64.RawURLEncoding.DecodeString(publicKey)
		if err != nil {
			return false
		}
		parsed, err := x509.ParsePKIXPublicKey(der)
		if err != nil {
			return false
		}
		pub, ok := parsed.(*ecdsa.PublicKey)
		if !ok || pub.Curve != elliptic.P256() {
			return false
		}
		digest := sha256.Sum256([]byte(message))
		return ecdsa.VerifyASN1(pub, digest[:], sig)
	default:
		return false
	}
}

func proposalSigningMessage(blockHash string) string {
	return "bbw-pop-proposal/v1\nblock=" + blockHash
}
func voteSigningMessage(v Vote) string {
	return fmt.Sprintf("bbw-pop-vote/v1\nnetwork=%s\nheight=%d\nround=%d\nblock=%s\ndecision=%s", v.NetworkID, v.Height, v.Round, v.BlockHash, v.Decision)
}

// NodeKey is distinct from player/device validator keys. Running more nodes never creates more consensus votes.
type NodeKey struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

func GenerateNodeKey() (NodeKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return NodeKey{}, err
	}
	return NodeKey{Public: pub, Private: priv}, nil
}

func (k NodeKey) ID() string {
	sum := sha256.Sum256(k.Public)
	return hex.EncodeToString(sum[:16])
}

func (k NodeKey) PublicText() string { return base64.RawURLEncoding.EncodeToString(k.Public) }

func SaveNodeKey(path string, key NodeKey) error {
	if len(key.Private) != ed25519.PrivateKeySize {
		return errors.New("invalid node private key")
	}
	body := "BBW-POP-NODE-PRIVATE-KEY-v1\n" + base64.RawURLEncoding.EncodeToString(key.Private) + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func LoadNodeKey(path string) (NodeKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return NodeKey{}, err
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 2 || lines[0] != "BBW-POP-NODE-PRIVATE-KEY-v1" {
		return NodeKey{}, errors.New("invalid node key file")
	}
	priv, err := base64.RawURLEncoding.DecodeString(lines[1])
	if err != nil || len(priv) != ed25519.PrivateKeySize {
		return NodeKey{}, errors.New("invalid node private key")
	}
	key := ed25519.PrivateKey(priv)
	pub := key.Public().(ed25519.PublicKey)
	return NodeKey{Public: pub, Private: key}, nil
}
