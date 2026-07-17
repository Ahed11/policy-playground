package main

import (
	"flag"
	"fmt"
	"os"
)

import (
	"github.com/Ahed11/policy-playground/internal"
)

type RunConfig struct {
	ScenarioPath string
	PoliciesPath string
	OutPath      string
}

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: policy-playground <command> [options]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		if err := runCmd(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runCmd(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)

	var cfg RunConfig

	fs.StringVar(&cfg.ScenarioPath, "scenario", "", "path to scenario yaml file")
	fs.StringVar(&cfg.PoliciesPath, "policies", "", "path to policies yaml file")
	fs.StringVar(&cfg.OutPath, "out", "alerts.jsonl", "path to alerts JSONL file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cfg.ScenarioPath == "" {
		return fmt.Errorf("--scenario is required")
	}

	if cfg.PoliciesPath == "" {
		return fmt.Errorf("--policies is required")
	}

	return run(cfg)
}

func run(cfg RunConfig) error {
	scenario, err := policy.ReadScenarioYAML(cfg.ScenarioPath)

	if err != nil {
		return err
	}

	policies, err := policy.ReadPoliciesYAML(cfg.PoliciesPath)

	if err != nil {
		return err
	}

	alertFile, err := os.OpenFile(cfg.OutPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		return err
	}

	defer alertFile.Close()

	for x := range scenario.Events {
		for y := range policies.Policies {
			alert, created, err := policy.CreateAlert(policies.Policies[y], scenario.Events[x])

			if err != nil {
				return fmt.Errorf("событие: %v, политика: %v\n%w", scenario.Events[x].EventID, policies.Policies[y].PolicyID, err)
			}

			if created == false {
				continue
			}

			if err := policy.WriteAlert(alertFile, alert); err != nil {
				return err
			}
		}
	}

	return nil
}