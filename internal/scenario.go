package internal

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v2"
)

type Scenario struct {
	Scenario_id string `yaml:"scenario_id"`
	Name string `yaml:"name"`
	Users []User `yaml:"users"`
	Events []Event `yaml:"events"`
}

func fillScenario() {
	data, _ := os.ReadFile("scenario.yaml")
	var s Scenario
	err := yaml.Unmarshal(data, &s)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("%+v\n", s)
}