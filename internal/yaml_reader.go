package policy

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)

func ReadScenarioYAML(path string) (Scenario, error) {
	scenarioData, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, fmt.Errorf("чтение файла: %w", err)
	}
	var scenario Scenario
	err = yaml.Unmarshal(scenarioData, &scenario)
	if err != nil {
		return Scenario{}, fmt.Errorf("парсинг YAML: %w", err)
	}

	return scenario, nil
}

func ReadPoliciesYAML(path string) (Policies, error) {
	policyData, err := os.ReadFile(path)
	if err != nil {
		return Policies{}, fmt.Errorf("чтение файла: %w", err)
	}
	var policies Policies
	err = yaml.Unmarshal(policyData, &policies)
	if err != nil {
		return Policies{}, fmt.Errorf("парсинг YAML: %w", err)
	}

	return policies, nil
}