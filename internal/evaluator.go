package policy

import (
	"slices"
	"fmt"
)

func getValueForField(event Event, condition Condition) (value string, err error) {

	field := condition.Field

	if field == "" {
		return "", fmt.Errorf("поле пусто")
	}

	switch field {
	case "event_id":
		return event.EventID, nil
	case "time":
		return event.Time, nil
	case "user_id":
		return event.UserID, nil
	case "action":
		return event.Action, nil
	case "object_type":
		return event.ObjectType, nil
	case "file_name":
		return *event.FileName, nil
	case "file_ext":
		return event.FileExt, nil
	case "channel":
		return event.Channel, nil
	case "destination_type":
		return event.DestinationType, nil
	}
	return "", fmt.Errorf("поле не поддерживается")
}

func CheckIfEquals(event Event, condition Condition) (result bool, reason string, err error) {
	value, valueErr := getValueForField(event, condition)

	if valueErr != nil {
		return false, "", valueErr
	}

	if condition.Equals == "" {
		return false, "", fmt.Errorf("нет значения для equals")
	}

	if condition.Equals == value {
		return true, fmt.Sprintf("%s equals %s", condition.Field, value), nil
	}

	return false, "", nil
}

func CheckIfContains(event Event, condition Condition) (result bool, reason string, err error) {
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
		if event.ContentClasses[i] == condition.Contains {
			return true, fmt.Sprintf("content_classes contains %s", event.ContentClasses[i]), nil
		}
	}

	return false, "", nil
}

func CheckIfIn(event Event, condition Condition) (result bool, reason string, err error) {
	value, valueErr := getValueForField(event, condition)

	if valueErr != nil {
		return false, "", valueErr
	}
	
	countOfEmptyElements := 0
	for i := range condition.In {
		if condition.In[i] == "" {
			countOfEmptyElements ++
		}
	}

	if countOfEmptyElements == len(condition.In) {
		return false, "", fmt.Errorf("нет значения для in")
	}

	if slices.Contains(condition.In, value) {
		return true, fmt.Sprintf("%v in %v", condition.Field, condition.In), nil
	}
	
	return false, "", nil
}

// Дописать функцию для exists

// func CheckIfExists(event Event, condition Condition) (result bool, reason string, err error) {

// 	if condition.Field == "" {
// 		return false, "", fmt.Errorf("поле пусто")
// 	}

// 	if condition.Field != "file_name"{
// 		return false, "", fmt.Errorf("поле не поддерживается")
// 	}

// 	if condition.Exists == nil {
// 		return false, "", fmt.Errorf("поле exists не существует")
// 	}

// 	if *condition.Exists == true && event.FileName != nil {
// 		return true, fmt.Sprintf("%v exists", condition.Field), nil
// 	}
	
// 	return false, "", nil
// }

func AllConditions(event Event, condition Condition) (result bool, reasons []string, err error) {
	if len(condition.All) == 0 {
		return false, nil, fmt.Errorf("группа all пуста")
	}

	for i := range condition.All {
		cond := condition.All[i]

		var ok bool
		var reason string
		var checkErr error

		switch {
		case cond.Field == "":
			return false, nil, fmt.Errorf("поле не заполнено") 
		case (cond.Equals != "" && cond.Contains != "") || (cond.Equals != "" && len(cond.In) != 0) || (cond.Contains != "" && len(cond.In) != 0):
			return false, nil, fmt.Errorf("заполнено более одного оператора")
		case cond.Equals == "" && cond.Contains == "" && len(cond.In) == 0:
			return false, nil, fmt.Errorf("нет заполненного оператора")
		case cond.Equals != "":
			ok, reason, checkErr = CheckIfEquals(event, cond)
		case cond.Contains != "":
			ok, reason, checkErr = CheckIfContains(event, cond)
		case len(cond.In) != 0:
			ok, reason, checkErr = CheckIfIn(event, cond)
		}

		if checkErr != nil {
			return false, nil, fmt.Errorf("ошибка %v: %w", i+1, checkErr)
		}
		if !ok {
			return false, nil, nil
		}

		reasons = append(reasons, reason)
	}

	return true, reasons, nil
}

func AnyConditions(event Event, condition Condition) (result bool, reasons []string, err error) {
	if len(condition.Any) == 0 {
		return false, nil, fmt.Errorf("группа any пуста")
	}

	for i := range condition.Any {
		cond := condition.Any[i]

		var ok bool
		var reason string
		var checkErr error

		switch {
		case cond.Field == "":
			return false, nil, fmt.Errorf("поле не заполнено") 
		case (cond.Equals != "" && cond.Contains != "") || (cond.Equals != "" && len(cond.In) != 0) || (cond.Contains != "" && len(cond.In) != 0):
			return false, nil, fmt.Errorf("заполнено более одного оператора")
		case cond.Equals == "" && cond.Contains == "" && len(cond.In) == 0:
			return false, nil, fmt.Errorf("нет заполненного оператора")
		case cond.Equals != "":
			ok, reason, checkErr = CheckIfEquals(event, cond)
		case cond.Contains !="":
			ok, reason, checkErr = CheckIfContains(event, cond)
		case len(cond.In) != 0:
			ok, reason, checkErr = CheckIfIn(event, cond)
		}

		if checkErr != nil {
			return false, nil, fmt.Errorf("ошибка %v: %w", i+1, checkErr)
		}

		if !ok {
			continue
		}

		reasons = append(reasons, reason)
	}

	if len(reasons) == 0 {
		return false, nil, nil
	}

	return true, reasons, nil
}