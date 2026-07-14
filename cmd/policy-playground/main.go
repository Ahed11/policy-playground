package main

import (
	"github.com/Ahed11/policy-playground/internal"
	"flag"
	"fmt"
	"os"
)

type RunConfig struct {
	ScenarioPath string
	PoliciesPath string
	OutPath	string
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

	events := len(scenario.Events)
	CountOfpolicies := len(policies.Policies)

	fmt.Printf(" Количество событий %v\n Количество политик %v\n", events, CountOfpolicies)

	return nil
}