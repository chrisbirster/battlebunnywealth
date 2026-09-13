package main
import("encoding/json";"flag";"fmt";"os";"github.com/chrisbirster/battlebunnywealth/internal/publictestnet")
func main(){attackers:=flag.Int("attackers",100,"fake nodes / phone-farm candidates / invalid transactions");gate:=flag.Bool("gate",false,"exit non-zero if adversarial containment expectations fail");flag.Parse();r,err:=publictestnet.RunSmoke(*attackers);if err!=nil{fatal(err)};_=json.NewEncoder(os.Stdout).Encode(r);if *gate&&(r.PeersAccepted>4||r.InvalidTransactions!=*attackers||r.ActivatedAfterMaturity>1||r.ValidTransactions!=1){fatal(fmt.Errorf("public-testnet adversarial gate failed"))}}
func fatal(err error){fmt.Fprintln(os.Stderr,"pop-public-smoke:",err);os.Exit(1)}
