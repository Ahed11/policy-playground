package policy

import (
	"fmt"
)

func CreateAlert(policy Policy, event Event) (alert Alert, created bool, err error) {
	localReasons := []string{}
	var ok bool
	var localErr error
	var reason string
	var reasons []string

	switch {
	case policy.Condition.Field == "" && policy.Condition.All == nil && policy.Condition.Any == nil:
		return alert, false, fmt.Errorf("нет ни простого, ни составного условия")
	case policy.Condition.Field != "" && len(policy.Condition.All) != 0 && len(policy.Condition.Any) != 0:
		return alert, false, fmt.Errorf("одновременно заполнены простое условие и группы all и any")
	case policy.Condition.Field != "" && len(policy.Condition.All) != 0:
		return alert, false, fmt.Errorf("одновременно заполнены простое условие и группа all")
	case policy.Condition.Field != "" && len(policy.Condition.Any) != 0:
		return alert, false, fmt.Errorf("одновременно заполнены простое условие и группа any")
	case len(policy.Condition.All) != 0 && len(policy.Condition.Any) != 0:
		return alert, false, fmt.Errorf("одновременно заполнены группы all и any")
	case policy.Condition.All != nil &&  len(policy.Condition.All) == 0:
		return alert, false, fmt.Errorf("группа all существует, но не содержит элементов")
	case policy.Condition.Any != nil &&  len(policy.Condition.Any) == 0:
		return alert, false, fmt.Errorf("группа any существует, но не содержит элементов")
	case policy.Condition.Field != "" && ((policy.Condition.Equals != "" && len(policy.Condition.In) != 0) || (policy.Condition.Equals != "" && policy.Condition.Contains != "") || (policy.Condition.Equals != "" && policy.Condition.Exists != nil) || (policy.Condition.Contains != "" && len(policy.Condition.In) != 0) || (policy.Condition.Contains != "" && policy.Condition.Exists != nil) || (len(policy.Condition.In) != 0 && policy.Condition.Exists != nil)):
		return alert, false, fmt.Errorf("заполнено более одного оператора")
	case policy.Condition.Field != "" && policy.Condition.Equals != "":
		ok, reason, localErr = CheckIfEquals(event, policy.Condition)
	case policy.Condition.Field != "" && policy.Condition.Contains != "":
		ok, reason, localErr = CheckIfContains(event, policy.Condition)
	case policy.Condition.Field != "" && len(policy.Condition.In) != 0:
		ok, reason, localErr = CheckIfIn(event, policy.Condition)
	case policy.Condition.Field != "" && policy.Condition.Exists != nil:
		ok, reason, localErr = CheckIfExists(event, policy.Condition)
	case policy.Condition.All != nil:
		ok, reasons, localErr = AllConditions(event, policy.Condition)
	case policy.Condition.Any != nil:
		ok, reasons, localErr = AnyConditions(event, policy.Condition)
	default:
		return alert, false, fmt.Errorf("политика %v: у простого условия отсутствует поддерживаемый оператор", policy.PolicyID)
	}

	if localErr != nil {
		return alert, false, fmt.Errorf("политика %v: %w", policy.PolicyID, localErr)
	}
	if ok == true{
		if reason != "" {
			localReasons = append(localReasons, reason)
		} else if len(reasons) != 0 {
			for i := range reasons {
				localReasons = append(localReasons, reasons[i])
			}
		}
	} else {
		return alert, false, nil
	}

	alert = Alert{
		PolicyID: policy.PolicyID,
		PolicyName: policy.Name,
		Severity: policy.Severity,
		EventID: event.EventID,
		UserID: event.UserID,
		Matched: true,
		Reasons: localReasons,
	}
	
	return alert, true, nil
}