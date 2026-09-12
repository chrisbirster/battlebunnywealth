package proofofplay

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"
)

var appleAppAttestNonceOID = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 2}

// AppleAppAttestCertificateValidator performs the server-side cryptographic checks
// described by Apple's App Attest validation flow. Roots must contain Apple's App
// Attest trust anchor; callers intentionally provide it so the trust root can be
// rotated without shipping a stale certificate in the binary.
type AppleAppAttestCertificateValidator struct {
	Roots *x509.CertPool
	Now   func() time.Time
}

func NewAppleAppAttestCertificateValidator(rootPEM []byte) (*AppleAppAttestCertificateValidator, error) {
	roots := x509.NewCertPool()
	if len(rootPEM) == 0 || !roots.AppendCertsFromPEM(rootPEM) {
		return nil, errors.New("Apple App Attest root CA PEM is required")
	}
	return &AppleAppAttestCertificateValidator{Roots: roots, Now: time.Now}, nil
}

func (v *AppleAppAttestCertificateValidator) ValidateAppAttest(_ context.Context, req AppleValidationRequest) (AppleValidationResult, error) {
	if v == nil || v.Roots == nil {
		return AppleValidationResult{}, errors.New("Apple App Attest trust roots not configured")
	}
	if req.BundleID == "" || req.TeamID == "" || req.ClientDataHash == "" || req.KeyID == "" {
		return AppleValidationResult{}, errors.New("incomplete Apple App Attest request")
	}
	objectBytes, err := base64.StdEncoding.DecodeString(req.AttestationObject)
	if err != nil {
		objectBytes, err = base64.RawStdEncoding.DecodeString(req.AttestationObject)
	}
	if err != nil {
		return AppleValidationResult{}, fmt.Errorf("decode App Attest object: %w", err)
	}
	decoded, rest, err := decodeCBOR(objectBytes, 0)
	if err != nil || len(rest) != 0 {
		return AppleValidationResult{}, fmt.Errorf("decode App Attest CBOR: %w", err)
	}
	top, ok := decoded.(map[any]any)
	if !ok {
		return AppleValidationResult{}, errors.New("App Attest object is not a CBOR map")
	}
	format, _ := top["fmt"].(string)
	if format != "apple-appattest" {
		return AppleValidationResult{}, fmt.Errorf("unexpected App Attest format %q", format)
	}
	authData, ok := top["authData"].([]byte)
	if !ok || len(authData) < 55 {
		return AppleValidationResult{}, errors.New("missing or short App Attest authenticator data")
	}
	stmt, ok := top["attStmt"].(map[any]any)
	if !ok {
		return AppleValidationResult{}, errors.New("missing App Attest statement")
	}
	chainValues, ok := stmt["x5c"].([]any)
	if !ok || len(chainValues) < 1 {
		return AppleValidationResult{}, errors.New("missing App Attest certificate chain")
	}
	certs := make([]*x509.Certificate, 0, len(chainValues))
	for _, value := range chainValues {
		der, ok := value.([]byte)
		if !ok {
			return AppleValidationResult{}, errors.New("invalid App Attest certificate entry")
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return AppleValidationResult{}, fmt.Errorf("parse App Attest certificate: %w", err)
		}
		certs = append(certs, cert)
	}
	leaf := certs[0]
	intermediates := x509.NewCertPool()
	for _, cert := range certs[1:] {
		intermediates.AddCert(cert)
	}
	now := time.Now()
	if v.Now != nil {
		now = v.Now()
	}
	if _, err := leaf.Verify(x509.VerifyOptions{Roots: v.Roots, Intermediates: intermediates, CurrentTime: now, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny}}); err != nil {
		return AppleValidationResult{}, fmt.Errorf("verify App Attest certificate chain: %w", err)
	}
	clientHash, err := decodeBase64Flexible(req.ClientDataHash)
	if err != nil || len(clientHash) != sha256.Size {
		return AppleValidationResult{}, errors.New("invalid App Attest clientDataHash")
	}
	nonceInput := make([]byte, 0, len(authData)+len(clientHash))
	nonceInput = append(nonceInput, authData...)
	nonceInput = append(nonceInput, clientHash...)
	nonce := sha256.Sum256(nonceInput)
	certNonce, err := appleNonceFromCertificate(leaf)
	if err != nil {
		return AppleValidationResult{}, err
	}
	if !bytes.Equal(certNonce, nonce[:]) {
		return AppleValidationResult{}, errors.New("App Attest nonce mismatch")
	}
	appID := req.TeamID + "." + req.BundleID
	rp := sha256.Sum256([]byte(appID))
	if !bytes.Equal(authData[:32], rp[:]) {
		return AppleValidationResult{}, errors.New("App Attest relying-party identifier mismatch")
	}
	if authData[32]&0x40 == 0 {
		return AppleValidationResult{}, errors.New("App Attest credential-data flag not set")
	}
	counter := binary.BigEndian.Uint32(authData[33:37])
	if counter != 0 {
		return AppleValidationResult{}, fmt.Errorf("App Attest initial counter is %d, want 0", counter)
	}
	expectedAAGUID, err := appleAAGUID(req.Environment)
	if err != nil {
		return AppleValidationResult{}, err
	}
	if !bytes.Equal(authData[37:53], expectedAAGUID) {
		return AppleValidationResult{}, errors.New("App Attest environment AAGUID mismatch")
	}
	credentialLen := int(binary.BigEndian.Uint16(authData[53:55]))
	if credentialLen <= 0 || len(authData) < 55+credentialLen {
		return AppleValidationResult{}, errors.New("invalid App Attest credential id")
	}
	credentialID := authData[55 : 55+credentialLen]
	keyIDBytes, err := decodeBase64Flexible(req.KeyID)
	if err != nil {
		return AppleValidationResult{}, errors.New("invalid App Attest key id")
	}
	if !bytes.Equal(credentialID, keyIDBytes) {
		return AppleValidationResult{}, errors.New("App Attest credential id does not match key id")
	}
	pub, ok := leaf.PublicKey.(*ecdsa.PublicKey)
	if !ok || pub.Curve != elliptic.P256() {
		return AppleValidationResult{}, errors.New("App Attest leaf key is not P-256")
	}
	x963 := marshalP256(pub)
	keyHash := sha256.Sum256(x963)
	if !bytes.Equal(keyHash[:], keyIDBytes) {
		return AppleValidationResult{}, errors.New("App Attest key id does not match attested public key")
	}
	coseAny, _, err := decodeCBOR(authData[55+credentialLen:], 0)
	if err != nil {
		return AppleValidationResult{}, fmt.Errorf("decode App Attest credential public key: %w", err)
	}
	cose, ok := coseAny.(map[any]any)
	if !ok {
		return AppleValidationResult{}, errors.New("invalid App Attest credential public key")
	}
	x, okX := cose[int64(-2)].([]byte)
	y, okY := cose[int64(-3)].([]byte)
	if !okX || !okY || !bytes.Equal(pad32(pub.X), x) || !bytes.Equal(pad32(pub.Y), y) {
		return AppleValidationResult{}, errors.New("App Attest credential public key mismatch")
	}
	receipt, ok := stmt["receipt"].([]byte)
	if !ok || len(receipt) == 0 {
		return AppleValidationResult{}, errors.New("missing App Attest risk receipt")
	}
	receiptDigest := sha256.Sum256(receipt)
	return AppleValidationResult{KeyID: req.KeyID, ReceiptDigest: hex.EncodeToString(receiptDigest[:]), AssertionCounter: uint64(counter), HardwareBacked: true}, nil
}

func appleNonceFromCertificate(cert *x509.Certificate) ([]byte, error) {
	for _, ext := range cert.Extensions {
		if !ext.Id.Equal(appleAppAttestNonceOID) {
			continue
		}
		var seq asn1.RawValue
		if _, err := asn1.Unmarshal(ext.Value, &seq); err != nil {
			return nil, fmt.Errorf("decode App Attest nonce extension: %w", err)
		}
		if seq.Tag != asn1.TagSequence {
			return nil, errors.New("invalid App Attest nonce extension sequence")
		}
		var tagged asn1.RawValue
		if _, err := asn1.Unmarshal(seq.Bytes, &tagged); err != nil {
			return nil, fmt.Errorf("decode App Attest nonce field: %w", err)
		}
		if tagged.Class != 2 || tagged.Tag != 1 {
			return nil, errors.New("invalid App Attest nonce extension field")
		}
		var nonce []byte
		if _, err := asn1.Unmarshal(tagged.Bytes, &nonce); err != nil {
			return nil, fmt.Errorf("decode App Attest nonce value: %w", err)
		}
		if len(nonce) != sha256.Size {
			return nil, errors.New("invalid App Attest nonce length")
		}
		return nonce, nil
	}
	return nil, errors.New("App Attest nonce certificate extension not found")
}
func appleAAGUID(environment string) ([]byte, error) {
	switch environment {
	case "production":
		return []byte{'a', 'p', 'p', 'a', 't', 't', 'e', 's', 't', 0, 0, 0, 0, 0, 0, 0}, nil
	case "development", "":
		return []byte("appattestdevelop"), nil
	default:
		return nil, fmt.Errorf("unsupported App Attest environment %q", environment)
	}
}
func marshalP256(pub *ecdsa.PublicKey) []byte {
	out := make([]byte, 65)
	out[0] = 4
	copy(out[1:33], pad32(pub.X))
	copy(out[33:], pad32(pub.Y))
	return out
}
func pad32(v *big.Int) []byte {
	raw := v.Bytes()
	out := make([]byte, 32)
	if len(raw) > 32 {
		raw = raw[len(raw)-32:]
	}
	copy(out[32-len(raw):], raw)
	return out
}
func decodeBase64Flexible(s string) ([]byte, error) {
	encodings := []*base64.Encoding{base64.RawURLEncoding, base64.URLEncoding, base64.StdEncoding, base64.RawStdEncoding}
	var last error
	for _, enc := range encodings {
		raw, err := enc.DecodeString(s)
		if err == nil {
			return raw, nil
		}
		last = err
	}
	return nil, last
}

func decodeCBOR(data []byte, depth int) (any, []byte, error) {
	if depth > 16 {
		return nil, nil, errors.New("CBOR nesting too deep")
	}
	if len(data) == 0 {
		return nil, nil, ioCBOR("unexpected end of CBOR")
	}
	head := data[0]
	major := head >> 5
	value, n, err := cborArgument(head&0x1f, data[1:])
	if err != nil {
		return nil, nil, err
	}
	rest := data[1+n:]
	switch major {
	case 0:
		return int64(value), rest, nil
	case 1:
		if value > uint64(^uint64(0)>>1) {
			return nil, nil, errors.New("CBOR negative integer overflow")
		}
		return -1 - int64(value), rest, nil
	case 2:
		if uint64(len(rest)) < value {
			return nil, nil, ioCBOR("short CBOR byte string")
		}
		return append([]byte(nil), rest[:value]...), rest[value:], nil
	case 3:
		if uint64(len(rest)) < value {
			return nil, nil, ioCBOR("short CBOR text string")
		}
		return string(rest[:value]), rest[value:], nil
	case 4:
		items := make([]any, 0, int(value))
		cursor := rest
		for i := uint64(0); i < value; i++ {
			item, next, err := decodeCBOR(cursor, depth+1)
			if err != nil {
				return nil, nil, err
			}
			items = append(items, item)
			cursor = next
		}
		return items, cursor, nil
	case 5:
		m := map[any]any{}
		cursor := rest
		for i := uint64(0); i < value; i++ {
			key, next, err := decodeCBOR(cursor, depth+1)
			if err != nil {
				return nil, nil, err
			}
			val, next2, err := decodeCBOR(next, depth+1)
			if err != nil {
				return nil, nil, err
			}
			switch key.(type) {
			case string, int64:
			default:
				return nil, nil, errors.New("unsupported CBOR map key")
			}
			m[key] = val
			cursor = next2
		}
		return m, cursor, nil
	default:
		return nil, nil, fmt.Errorf("unsupported CBOR major type %d", major)
	}
}
func cborArgument(info byte, data []byte) (uint64, int, error) {
	switch {
	case info < 24:
		return uint64(info), 0, nil
	case info == 24:
		if len(data) < 1 {
			return 0, 0, ioCBOR("short CBOR uint8")
		}
		return uint64(data[0]), 1, nil
	case info == 25:
		if len(data) < 2 {
			return 0, 0, ioCBOR("short CBOR uint16")
		}
		return uint64(binary.BigEndian.Uint16(data[:2])), 2, nil
	case info == 26:
		if len(data) < 4 {
			return 0, 0, ioCBOR("short CBOR uint32")
		}
		return uint64(binary.BigEndian.Uint32(data[:4])), 4, nil
	case info == 27:
		if len(data) < 8 {
			return 0, 0, ioCBOR("short CBOR uint64")
		}
		return binary.BigEndian.Uint64(data[:8]), 8, nil
	default:
		return 0, 0, errors.New("indefinite CBOR values are not supported")
	}
}
func ioCBOR(msg string) error { return errors.New(msg) }
