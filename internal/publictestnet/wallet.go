package publictestnet

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"strings"
)

const AddressPrefix = "tcarrot1"

type Wallet struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

func GenerateWallet() (Wallet, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil { return Wallet{}, err }
	return Wallet{Public: pub, Private: priv}, nil
}
func (w Wallet) PublicText() string { return base64.RawURLEncoding.EncodeToString(w.Public) }
func (w Wallet) Address() string { sum:=sha256.Sum256(w.Public); return AddressPrefix+hex.EncodeToString(sum[:20]) }
func AddressFromPublicText(publicText string) (string,error) { pub,err:=base64.RawURLEncoding.DecodeString(publicText); if err!=nil||len(pub)!=ed25519.PublicKeySize{return "",errors.New("invalid TEST-CARROT public key")}; sum:=sha256.Sum256(pub); return AddressPrefix+hex.EncodeToString(sum[:20]),nil }
func ValidAddress(address string) bool { if !strings.HasPrefix(address,AddressPrefix)||len(address)!=len(AddressPrefix)+40{return false}; _,err:=hex.DecodeString(address[len(AddressPrefix):]); return err==nil }
func SaveWallet(path string,w Wallet) error { if len(w.Private)!=ed25519.PrivateKeySize{return errors.New("invalid TEST-CARROT private key")}; body:="BBW-TEST-CARROT-PRIVATE-KEY-v1\n"+base64.RawURLEncoding.EncodeToString(w.Private)+"\n"; if err:=os.WriteFile(path,[]byte(body),0o600);err!=nil{return err}; return os.Chmod(path,0o600) }
func LoadWallet(path string)(Wallet,error){raw,err:=os.ReadFile(path);if err!=nil{return Wallet{},err};lines:=strings.Split(strings.TrimSpace(string(raw)),"\n");if len(lines)!=2||lines[0]!="BBW-TEST-CARROT-PRIVATE-KEY-v1"{return Wallet{},errors.New("invalid TEST-CARROT wallet file")};priv,err:=base64.RawURLEncoding.DecodeString(lines[1]);if err!=nil||len(priv)!=ed25519.PrivateKeySize{return Wallet{},errors.New("invalid TEST-CARROT private key")};p:=ed25519.PrivateKey(priv);return Wallet{Public:p.Public().(ed25519.PublicKey),Private:p},nil}
