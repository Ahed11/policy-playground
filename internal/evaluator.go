package policy

import (
	"fmt"
)

func CheckIfEquals() {
	policies, err := ReadPoliciesYAML("../testdata/control/policies.yaml")
	
	if err != nil {
		fmt.Println(err)
		return 
	}

	allPolicies := policies.Policies
	
	conditions := []struct{}

	for i, v := range allPolicies {
		conditions = append(conditions, allPolicies[i].Condition.All)
	}
	
	scenario, err := ReadScenarioYAML("../testdata/control/scenario.yaml")
	
	if err != nil {
		fmt.Println(err)
		return 
	}

	events := scenario.Events
	
	// for i, p := range events {
	// 	if 
	// }
}