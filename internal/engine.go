package policy

import (
	"fmt"
)

func CreateAlert(policy Policy, event Event) (alert Alert, created bool, err error) {
	localReasons := []string{}
	
	if policy.Condition.Field == "" && policy.Condition.All == nil {
		return alert, false, fmt.Errorf("нет ни простого, ни составного условия")
	}
	
	if policy.Condition.Field != "" && policy.Condition.All != nil {
		return alert, false, fmt.Errorf("одновременно заполнены простое и all")
	}

	if policy.Condition.All != nil && len(policy.Condition.All) == 0 {
			return alert, false, fmt.Errorf("группа all существует, но не содержит элементов")
	}

	if policy.Condition.Field != "" && policy.Condition.Equals != "" {
		result, reason, err := CheckIfEquals(event, policy.Condition)
		
		if err != nil {
			return alert, false, fmt.Errorf("политика %v: %w", policy.PolicyID, err)
		}
		if result == true{
			localReasons = append(localReasons, reason)
		} else {
			return alert, false, nil
		}
	} else if policy.Condition.All != nil {
		result, reasons, err := AllConditions(event, policy.Condition)

		if err != nil {
			return alert, false, fmt.Errorf("политика %v: %w", policy.PolicyID, err)
		}

		if result == true {
			for i := range reasons {
				localReasons = append(localReasons, reasons[i])
			}
		} else {
			return alert, false, nil
		}
	} else {
		return alert, false, fmt.Errorf("политика %v: у простого условия отсутствует поддерживаемый оператор", policy.PolicyID)
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