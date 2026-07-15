package policy

import (
	"fmt"
)

func CreateAlert(policy Policy, event Event) (alert Alert, created bool, err error) {
	localReasons := []string{}
	
	if policy.Condition.Field == "" && policy.Condition.All == nil {
		return alert, false, fmt.Errorf("нет ни просто условия ни составного")
	}

	if policy.Condition.Field != "" && policy.Condition.Equals == "" {
		return alert, false, fmt.Errorf("указано поле, но нет поддерживаемого оператора")
	}

	if policy.Condition.Field != "" && policy.Condition.All != nil {
		return alert, false, fmt.Errorf("одновременно заполнены простое и составное условие")
	}

	if policy.Condition.All != nil && len(policy.Condition.All) == 0 {
			return alert, false, fmt.Errorf("составное условие существует но содержит все пустые условия")
	} 

	// if policy.Condition.All != nil && len(policy.Condition.All) != 0 {
	// 	for i := range policy.Condition.All {
	// 		if policy.Condition.All[i].Field == "" {
	// 			if policy.Condition.All[i].Equals == "" {
	// 				return alert, false, fmt.Errorf("составное условие существует но содержит пустое/ые условие/я")
	// 			} else {
	// 				return alert, false, fmt.Errorf("составное условие существует но содержит пустое/ые условие/я")
	// 			}
	// 		}
	// 	}
	// }

	if policy.Condition.Field != "" {
		if policy.Condition.Equals != "" {
			result, reason, err := CheckIfEquals(event, policy.Condition)
			if err != nil {
				return alert, false, fmt.Errorf("политика %v: %w", policy.PolicyID, err)
			}
			if result == true{
				localReasons = append(localReasons, reason)
			} else {
				return alert, false, nil
			}
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