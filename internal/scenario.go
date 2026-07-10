package policy

type Scenario struct {
	ScenarioID string `yaml:"scenario_id"`
	Name string `yaml:"name"`
	Users []User `yaml:"users"`
	Events []Event `yaml:"events"`
}