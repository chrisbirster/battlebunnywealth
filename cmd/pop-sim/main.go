package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chrisbirster/battlebunnywealth/internal/popsim"
)

func main() {
	var (
		scenarioName = flag.String("scenario", "baseline", "built-in scenario name, all, or attack-sweep")
		seed = flag.Int64("seed", 42, "deterministic simulation seed")
		days = flag.Int("days", 0, "override scenario duration in days")
		trials = flag.Int("trials", 0, "override committee Monte Carlo trials")
		committee = flag.Int("committee", 0, "override committee size")
		format = flag.String("format", "json", "output format: json or csv")
		outPath = flag.String("out", "", "optional output path; stdout when omitted")
		list = flag.Bool("list", false, "list built-in scenarios")
	)
	flag.Parse()

	if *list {
		for _, name := range popsim.ScenarioNames() { fmt.Println(name) }
		fmt.Println("attack-sweep")
		fmt.Println("all")
		return
	}

	scenarios, err := resolveScenarios(*scenarioName, *seed)
	if err != nil { fatal(err) }
	for i := range scenarios {
		if *days > 0 { scenarios[i].Days = *days }
		if *trials > 0 { scenarios[i].CommitteeTrials = *trials }
		if *committee > 0 { scenarios[i].Policy.CommitteeSize = *committee }
	}

	reports := make([]popsim.Report, 0, len(scenarios))
	for _, scenario := range scenarios {
		report, err := popsim.Run(scenario)
		if err != nil { fatal(fmt.Errorf("run %s: %w", scenario.Name, err)) }
		reports = append(reports, report)
	}

	writer, closeFn, err := outputWriter(*outPath)
	if err != nil { fatal(err) }
	defer closeFn()
	switch strings.ToLower(strings.TrimSpace(*format)) {
	case "json":
		err = popsim.WriteJSON(writer, reports)
	case "csv":
		err = popsim.WriteCSV(writer, reports)
	default:
		err = fmt.Errorf("unsupported format %q", *format)
	}
	if err != nil { fatal(err) }
}

func resolveScenarios(name string, seed int64) ([]popsim.Scenario, error) {
	switch name {
	case "attack-sweep":
		return popsim.AttackSweep(seed, nil)
	case "all":
		out := []popsim.Scenario{}
		for i, builtin := range popsim.ScenarioNames() {
			if builtin == "smoke" { continue }
			s, err := popsim.BuiltinScenario(builtin, seed+int64(i))
			if err != nil { return nil, err }
			out = append(out, s)
		}
		return out, nil
	default:
		s, err := popsim.BuiltinScenario(name, seed)
		if err != nil { return nil, err }
		return []popsim.Scenario{s}, nil
	}
}

func outputWriter(path string) (io.Writer, func(), error) {
	if strings.TrimSpace(path) == "" { return os.Stdout, func(){}, nil }
	file, err := os.Create(path)
	if err != nil { return nil, func(){}, err }
	return file, func(){ _ = file.Close() }, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pop-sim:", err)
	os.Exit(1)
}
