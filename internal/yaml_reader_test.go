package policy

import "testing"

func TestReadScenarioYAML(t *testing.T) {
	scenario, err := ReadScenarioYAML("../testdata/control/scenario.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(scenario.Users) != 1 {
		t.Fatal("количество пользователей неверное")
	}

	if len(scenario.Events) != 2 {
		t.Fatal("количество событий неверное")
	}
}

func TestReadPoliciesYAML(t *testing.T) {
	policies, err := ReadPoliciesYAML("../testdata/control/policies.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(policies.Policies) != 1{
		t.Fatal("количество политик неверное")
	}

	if len(policies.Policies[0].Condition.All) != 3 {
		t.Fatal("количество учитываемых колонок неверное")
	}
    // for i, p := range policies.Policies {
    //     t.Logf("policy[%d]: %+v", i, p)
    // }
}