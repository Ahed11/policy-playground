package policy

import (
	"fmt"
)

func CheckIfEquals(event Event, condition Condition) (result bool, reason string, err error) {
	if condition.Field == "" {
		return false, "", fmt.Errorf("поле пусто")
	}
	if condition.Field != "action" && condition.Field != "destination_type" {
		return false, "", fmt.Errorf("поле не поддерживается")
	}
	switch condition.Field {
	case "action":
		if condition.Equals == "" {
			return false, "", fmt.Errorf("нет значения для equals")
		}
		if condition.Equals == event.Action {
			return true, fmt.Sprintf("action equals %s", event.Action), nil
		}
	case "destination_type":
		if condition.Equals == event.DestinationType {
			return true, fmt.Sprintf("destination_type equals %s", event.DestinationType), nil
		}
	}
	return false, "", nil
}

func CheckifContains(event Event, condition Condition) (result bool, reason string, err error) {
	if condition.Field == "" {
		return false, "", fmt.Errorf("поле пусто")
	}
	if condition.Field != "content_classes" {
		return false, "", fmt.Errorf("поле не поддерживается")
	}
	if condition.Contains == "" {
		return false, "", fmt.Errorf("нет значения для contains")
	}
	for i := range event.ContentClasses {
		if condition.Contains == event.ContentClasses[i] {
			return true, fmt.Sprintf("content_classes contains %s", event.ContentClasses[i]), nil
		}
	}
	return false, "", nil
}

func AllConditions(event Event, condition Condition) (result bool, reasons []string, err error) {

	for i := range condition.All {
		result, reason, err := CheckIfEquals(event, condition.All[i])
		if err != nil {
			return result, reasons, fmt.Errorf("Ошибка: %s", err)
		} 
		if result != true {
			if reasons != nil {
				reasons = nil
				return result, reasons, err
			} else {
				return result, reasons, err
			}
		} else {
			reasons = append(reasons, reason)
		}
	}

	return true, reasons, nil
}