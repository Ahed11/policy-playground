package policy

type Policy struct {
	PolicyID string `yaml:"policy_id"`
	Name string `yaml:"name"`
	Severity string `yaml:"severity"`
	Description string `yaml:"description"`
	Condition Condition `yaml:"condition"`
}