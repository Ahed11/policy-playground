package internal

import (
	"gopkg.in/yaml.v2"
)

type Policy struct {
	Policy_id string `yaml:"policy_id"`
	Name string `yaml:"name"`
	Severity string `yaml:"severity"`
	Description string `yaml:"description"`
	Condition Condition `yaml:"condition"`
}