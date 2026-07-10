package main

import (
	"flag"
	"fmt"
	//"log"
	"os"
)

type RunConfig struct {
	ScenarioPath string
	PoliciesPath	string
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
	fs.StringVar(&cfg.PoliciesPath, "polices", "", "path to polices yaml file")
	fs.StringVar(&cfg.OutPath, "out", "alets.jsonl", "path to scenario yaml file")

	if err := fs.Parse(args [2:]); err != nil {
		return err
	}	

	if cfg.ScenarioPath == "" {
		return fmt.Errorf("--scenario is required")
	}

	if cfg.PoliciesPath == "" {
		return fmt.Errorf("--polices is required")
	}

	return run(cfg) 
}

func run(cfg RunConfig) error {
	fmt.Println("scenario:", cfg.ScenarioPath)
	fmt.Println("policies:", cfg.PoliciesPath)
	fmt.Println("out:", cfg.OutPath)

	//дальше надо будет прописать чтение параметров(сценария, политик) и запуск симуляции. После попытаться написать alerts.jsonl

	return nil
}