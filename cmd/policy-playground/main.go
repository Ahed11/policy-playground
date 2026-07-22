package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Ahed11/policy-playground/internal"
)

type RunConfig struct {
	ScenarioPath string
	PoliciesPath string
	OutPath      string
	EventPath    string
	Report       string
}

func checkThePaths(out string, scenario string, policies string) error {
	scenarioAbs, err := filepath.Abs(scenario)

	if err != nil {
		return err
	}

	policiesAbs, err := filepath.Abs(policies)

	if err != nil {
		return err
	}

	outAbs, err := filepath.Abs(out)

	if err != nil {
		return err
	}

	if outAbs == scenarioAbs || outAbs == policiesAbs {
		return fmt.Errorf("--out указывает на входной файл")
	}

	alertsInfo, err := os.Stat(outAbs)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	scenarioInfo, err := os.Stat(scenarioAbs)

	if err != nil {
		return err
	}

	policiesInfo, err := os.Stat(policiesAbs)

	if err != nil {
		return err
	}

	if os.SameFile(alertsInfo, scenarioInfo) || os.SameFile(alertsInfo, policiesInfo) {
		return fmt.Errorf("файл --out совпадает с входным файлом")
	}

	return nil
}

func closeAndRemove(file *os.File) error {
	tempName := file.Name()

	closeErr := file.Close()
	removeErr := os.Remove(tempName)

	return errors.Join(closeErr, removeErr)
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
	case "explain":
		if err := explainCmd(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

//дописать функцию для explain

func explainCmd(args []string) error {
	fs := flag.NewFlagSet("explain", flag.ContinueOnError)

	var cfg RunConfig

	fs.StringVar(&cfg.ScenarioPath, "scenario", "", "path to scenario yaml file")
	fs.StringVar(&cfg.PoliciesPath, "policies", "", "path to policies yaml file")
	fs.StringVar(&cfg.OutPath, "event", "", "ID of chosen event")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return nil
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
	err := checkThePaths(cfg.OutPath, cfg.ScenarioPath, cfg.PoliciesPath)

	if err != nil {
		return err
	}

	scenario, err := policy.ReadScenarioYAML(cfg.ScenarioPath)

	if err != nil {
		return err
	}

	policies, err := policy.ReadPoliciesYAML(cfg.PoliciesPath)

	if err != nil {
		return err
	}

	dir := filepath.Dir(cfg.OutPath)

	tempAlertFile, err := os.CreateTemp(dir, "temp_alert-*.jsonl")

	if err != nil {
		return err
	}

	for x := range scenario.Events {
		for y := range policies.Policies {
			alert, created, err := policy.CreateAlert(policies.Policies[y], scenario.Events[x])

			if err != nil {
				mainErr := fmt.Errorf("событие: %v, политика: %v\n%w.", scenario.Events[x].EventID, policies.Policies[y].PolicyID, err)

				cleanupErr := closeAndRemove(tempAlertFile)

				return errors.Join(mainErr, cleanupErr)
			}

			if created == false {
				continue
			}

			if err := policy.WriteAlert(tempAlertFile, alert); err != nil {
				cleanupErr := closeAndRemove(tempAlertFile)

				return errors.Join(err, cleanupErr)
			}
		}
	}

	if err := tempAlertFile.Sync(); err != nil {
		cleanupErr := closeAndRemove(tempAlertFile)

		return errors.Join(err, cleanupErr)
	}

	if err := tempAlertFile.Close(); err != nil {
		removeErr := os.Remove(tempAlertFile.Name())

		return errors.Join(err, removeErr)
	}

	if err := os.Rename(tempAlertFile.Name(), cfg.OutPath); err != nil {
		removeErr := os.Remove(tempAlertFile.Name())

		return errors.Join(err, removeErr)
	}

	return nil
}
