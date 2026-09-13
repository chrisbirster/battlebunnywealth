package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/chrisbirster/battlebunnywealth/internal/popsim"
)

func main() {
	phones := flag.Int("phones", 100, "real attested attacker phones")
	days := flag.Int("days", 180, "simulation days")
	immediate := flag.Bool("immediate", false, "disable maturation/activation throttling to show the unsafe baseline")
	gate := flag.Bool("gate", false, "run the v0.9 bootstrap security regression gate")
	flag.Parse()
	if *gate {
		if err := popsim.BootstrapGate(*phones); err != nil {
			fmt.Fprintln(os.Stderr, "bootstrap gate:", err)
			os.Exit(1)
		}
		fmt.Println("bootstrap gate: ok")
		return
	}
	p := popsim.DefaultBootstrapAttackPolicy(*phones)
	p.SimulationDays = *days
	p.ImmediateActivation = *immediate
	r, err := popsim.SimulateBootstrapAttack(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bootstrap sim:", err)
		os.Exit(1)
	}
	if err := popsim.WriteBootstrapJSON(os.Stdout, []popsim.BootstrapAttackReport{r}); err != nil {
		fmt.Fprintln(os.Stderr, "bootstrap sim:", err)
		os.Exit(1)
	}
}
