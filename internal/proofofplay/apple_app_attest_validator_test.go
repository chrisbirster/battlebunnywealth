package proofofplay

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func TestAppleAppAttestCertificateValidator(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	rootKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	rootTemplate := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test App Attest root"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	rootCert, _ := x509.ParseCertificate(rootDER)
	leafKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	x963 := marshalP256(&leafKey.PublicKey)
	keyHash := sha256.Sum256(x963)
	keyID := base64.StdEncoding.EncodeToString(keyHash[:])
	appID := "TEAM123456.com.example.bbw"
	rp := sha256.Sum256([]byte(appID))
	clientHash := sha256.Sum256([]byte("server-bound-client-data"))
	cose := cborMap([]cborPair{{cborInt(1), cborInt(2)}, {cborInt(3), cborInt(-7)}, {cborInt(-1), cborInt(1)}, {cborInt(-2), cborBytes(pad32(leafKey.X))}, {cborInt(-3), cborBytes(pad32(leafKey.Y))}})
	auth := make([]byte, 0, 55+32+len(cose))
	auth = append(auth, rp[:]...)
	auth = append(auth, 0x41)
	auth = append(auth, 0, 0, 0, 0)
	auth = append(auth, []byte("appattestdevelop")...)
	var l [2]byte
	binary.BigEndian.PutUint16(l[:], uint16(len(keyHash)))
	auth = append(auth, l[:]...)
	auth = append(auth, keyHash[:]...)
	auth = append(auth, cose...)
	nonceInput := append(append([]byte(nil), auth...), clientHash[:]...)
	nonce := sha256.Sum256(nonceInput)
	nonceDER := append([]byte{0x30, 0x24, 0xa1, 0x22, 0x04, 0x20}, nonce[:]...)
	leafTemplate := &x509.Certificate{SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "attested key"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtraExtensions: []pkix.Extension{{Id: appleAppAttestNonceOID, Value: nonceDER}}}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTemplate, rootCert, &leafKey.PublicKey, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	object := cborMap([]cborPair{{cborText("fmt"), cborText("apple-appattest")}, {cborText("attStmt"), cborMap([]cborPair{{cborText("x5c"), cborArray([][]byte{cborBytes(leafDER)})}, {cborText("receipt"), cborBytes([]byte("receipt"))}})}, {cborText("authData"), cborBytes(auth)}})
	rootPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootDER})
	validator, err := NewAppleAppAttestCertificateValidator(rootPEM)
	if err != nil {
		t.Fatal(err)
	}
	validator.Now = func() time.Time { return now }
	result, err := validator.ValidateAppAttest(context.Background(), AppleValidationRequest{ClientDataHash: base64.RawURLEncoding.EncodeToString(clientHash[:]), AttestationObject: base64.StdEncoding.EncodeToString(object), KeyID: keyID, BundleID: "com.example.bbw", TeamID: "TEAM123456", Environment: "development"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.HardwareBacked || result.KeyID != keyID || result.ReceiptDigest == "" {
		t.Fatalf("result=%+v", result)
	}
}

type cborPair struct{ k, v []byte }

func cborMap(items []cborPair) []byte {
	out := cborLen(5, len(items))
	for _, p := range items {
		out = append(out, p.k...)
		out = append(out, p.v...)
	}
	return out
}
func cborArray(items [][]byte) []byte {
	out := cborLen(4, len(items))
	for _, item := range items {
		out = append(out, item...)
	}
	return out
}
func cborText(s string) []byte  { return append(cborLen(3, len(s)), []byte(s)...) }
func cborBytes(b []byte) []byte { return append(cborLen(2, len(b)), b...) }
func cborInt(v int64) []byte {
	if v >= 0 {
		return cborUint(0, uint64(v))
	}
	return cborUint(1, uint64(-1-v))
}
func cborLen(major byte, n int) []byte { return cborUint(major, uint64(n)) }
func cborUint(major byte, n uint64) []byte {
	if n < 24 {
		return []byte{major<<5 | byte(n)}
	}
	if n <= 255 {
		return []byte{major<<5 | 24, byte(n)}
	}
	if n <= 65535 {
		return []byte{major<<5 | 25, byte(n >> 8), byte(n)}
	}
	panic("test CBOR length too large")
}
