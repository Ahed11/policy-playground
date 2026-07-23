package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Ahed11/policy-playground/internal"
)

type RunConfig struct {
	ScenarioPath string
	PoliciesPath string
	OutPath      string
}

type ExplainConfig struct {
	ScenarioPath string
	PoliciesPath string
	EventID    string
}

type ExplainData struct {
	EventID string
	UserID string
	Action string
	Channel string
	DestinationType string
	ContentClasses *[]string
}

type MatchedPolicy struct {
	PolicyID string
	Name string
	Severity string
	Reasons  []string
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

func outputExplainData(w io.Writer, event policy.Event) error {
	if _, err := fmt.Fprintf(w, "event_id: %v\n", event.EventID); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "user_id: %v\n", event.UserID); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "action: %v\n", event.Action); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "channel: %v\n", event.Channel); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w,	"destination_type: %v\n", event.DestinationType); err != nil {
		return err
	}

	if event.ContentClasses != nil {
		if _, err := fmt.Fprintf(w, "content_classes: %v\n", *event.ContentClasses); err != nil {
			return err
		}
	}

	return nil
}

func outputPolicy(w io.Writer, policy MatchedPolicy) error {
	if _, err := fmt.Fprintf(w, "- policy_id: %v\n", policy.PolicyID); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "  name: %v\n", policy.Name); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "  severity: %v\n", policy.Severity); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "  reasons:"); err != nil {
		return err
	}

	for _, reason := range policy.Reasons {
		if _, err := fmt.Fprintf(w, "  - %v\n", reason); err != nil {
			return err
		}
	}

	return nil
}

func outputExplainResult(w io.Writer, event policy.Event, policies []MatchedPolicy) error {
	if err := outputExplainData(w, event); err != nil {
		return err
	}

	if len(policies) == 0 {
		_, err := fmt.Fprintln(w, "matched policies: none")
		return err
	}

	if _, err := fmt.Fprintln(w, "matched policies:"); err != nil {
		return err
	}

	for _, matchedPolicy := range policies {
		if err := outputPolicy(w, matchedPolicy); err != nil {
			return err
		}
	}

	return nil
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

func explainCmd(args []string) error {
	fs := flag.NewFlagSet("explain", flag.ContinueOnError)

	var cfg ExplainConfig

	fs.StringVar(&cfg.ScenarioPath, "scenario", "", "path to scenario yaml file")
	fs.StringVar(&cfg.PoliciesPath, "policies", "", "path to policies yaml file")
	fs.StringVar(&cfg.EventID, "event", "", "ID of chosen event")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cfg.ScenarioPath == "" {
		return fmt.Errorf("--scenario is required")
	}

	if cfg.PoliciesPath == "" {
		return fmt.Errorf("--policies is required")
	}

	if cfg.EventID == "" {
		return fmt.Errorf("--event is required")
	}

	return explain(cfg)
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

	for s := range scenario.Events {
		for p := range policies.Policies {
			alert, created, err := policy.CreateAlert(policies.Policies[p], scenario.Events[s])

			if err != nil {
				mainErr := fmt.Errorf("событие: %v, политика: %v\n%w.", scenario.Events[s].EventID, policies.Policies[p].PolicyID, err)

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

func explain(cfg ExplainConfig) error {
	
	
	scenario, err := policy.ReadScenarioYAML(cfg.ScenarioPath)

	if err != nil {
		return err
	}

	policies, err := policy.ReadPoliciesYAML(cfg.PoliciesPath)

	if err != nil {
		return err
	}

	var event policy.Event
	var foundEvent bool

	for i := range scenario.Events {
		if cfg.EventID == scenario.Events[i].EventID {
			if foundEvent == false {
				event = scenario.Events[i]
				foundEvent = true
			} else {
				return fmt.Errorf("два события имеют одинаковый ID")
			}
		}
	}

	if !foundEvent {
		return fmt.Errorf("событие %v не было найдено", cfg.EventID)
	}

	var matchedPolicies []MatchedPolicy

	for i := range policies.Policies {
		alert, created, err := policy.CreateAlert(policies.Policies[i], event)

		if err != nil {
			return fmt.Errorf("событие: %v, политика: %v\n%w.", event.EventID, policies.Policies[i].PolicyID, err)
		}

		if created == false {
				continue
		}

		matchedPolicies = append(
			matchedPolicies, 
			MatchedPolicy{
				alert.PolicyID,
				alert.PolicyName,
				alert.Severity,
				alert.Reasons,
			},
		)
	}

	return outputExplainResult(os.Stdout, event, matchedPolicies)
}