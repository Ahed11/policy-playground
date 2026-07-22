package policy

import (
	"slices"
	"fmt"
	"strconv"
)

func intToString(num int) string {
    return strconv.Itoa(num)
}

func getValueForField(event Event, condition Condition) (value string, exist bool, err error) {

	field := condition.Field

	if field == "" {
		return "", false, fmt.Errorf("поле пусто")
	}

	switch field {
	case "event_id":
		return event.EventID, true, nil
	case "time":
		return event.Time, true, nil
	case "user_id":
		return event.UserID, true, nil
	case "action":
		return event.Action, true, nil
	case "object_type":
		return event.ObjectType, true, nil
	case "file_name":
		if event.FileName == nil {return "", false, nil}
		return *event.FileName, true, nil
	case "file_ext":
		if event.FileExt == nil {return "", false, nil}
		return *event.FileExt, true, nil
	case "channel":
		return event.Channel, true, nil
	case "size_bytes":
		if event.SizeBytes == nil {return "", false, nil}
		return intToString(*event.SizeBytes), true, nil
	case "destination_type":
		return event.DestinationType, true, nil
	}
	return "", false, fmt.Errorf("поле не поддерживается")
}

func CheckIfEquals(event Event, condition Condition) (result bool, reason string, err error) {	
	if condition.Equals == "" {
		return false, "", fmt.Errorf("нет значения для equals")
	}

	value, fieldExist, valueErr := getValueForField(event, condition)
	
	if valueErr != nil {
		return false, "", valueErr
	}

	if fieldExist == false {
		return false, "", nil
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

	if event.ContentClasses == nil {
        return false, "", nil
    }

	classes := *event.ContentClasses

	for i := range classes  {
		if classes[i] == condition.Contains {
			return true, fmt.Sprintf("content_classes contains %s", classes[i]), nil
		}
	}

	return false, "", nil
}

func CheckIfIn(event Event, condition Condition) (result bool, reason string, err error) {	
	countOfEmptyElements := 0
	for i := range condition.In {
		if condition.In[i] == "" {
			countOfEmptyElements ++
		}
	}

	if countOfEmptyElements == len(condition.In) {
		return false, "", fmt.Errorf("нет значения для in")
	}

	value, fieldExist, valueErr := getValueForField(event, condition)

	if valueErr != nil {
		return false, "", valueErr
	}

	if fieldExist == false {
		return false, "", nil
	}

	if slices.Contains(condition.In, value) {
		return true, fmt.Sprintf("%v in %v", condition.Field, condition.In), nil
	}
	
	return false, "", nil
}

func CheckIfExists(event Event, condition Condition) (result bool, reason string, err error) {
	if condition.Field == "" {
		return false, "", fmt.Errorf("поле пусто")
	}

	if condition.Exists == nil {
		return false, "", fmt.Errorf("поле exists не существует")
	}

	var fieldExists bool

	switch condition.Field {
	case "file_name":
		fieldExists = event.FileName != nil
	case "file_ext":
		fieldExists = event.FileExt != nil
	case "content_classes":
		fieldExists = event.ContentClasses != nil
	case "size_bytes":
		fieldExists = event.SizeBytes != nil
	default:
		return false, "", fmt.Errorf("поле не поддерживается")
	}

	if fieldExists != *condition.Exists {
		return false, "", nil
	}

	if *condition.Exists {
		return true, fmt.Sprintf("%v exists", condition.Field), nil
	}

	return true, fmt.Sprintf("%v does not exist", condition.Field), nil
}

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
		case (cond.Equals != "" && cond.Contains != "") || (cond.Equals != "" && len(cond.In) != 0) || (cond.Equals != "" && cond.Exists != nil) || (cond.Contains != "" && len(cond.In) != 0) || (cond.Contains != "" && cond.Exists != nil) || (len(cond.In) != 0 && cond.Exists != nil):
			return false, nil, fmt.Errorf("заполнено более одного оператора")
		case cond.Equals == "" && cond.Contains == "" && len(cond.In) == 0 && cond.Exists == nil:
			return false, nil, fmt.Errorf("нет заполненного оператора")
		case cond.Equals != "":
			ok, reason, checkErr = CheckIfEquals(event, cond)
		case cond.Contains != "":
			ok, reason, checkErr = CheckIfContains(event, cond)
		case len(cond.In) != 0:
			ok, reason, checkErr = CheckIfIn(event, cond)
		case cond.Exists != nil:
			ok, reason, checkErr = CheckIfExists(event, cond)
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
		case (cond.Equals != "" && cond.Contains != "") || (cond.Equals != "" && len(cond.In) != 0) || (cond.Equals != "" && cond.Exists != nil) || (cond.Contains != "" && len(cond.In) != 0) || (cond.Contains != "" && cond.Exists != nil) || (len(cond.In) != 0 && cond.Exists != nil):
			return false, nil, fmt.Errorf("заполнено более одного оператора")
		case cond.Equals == "" && cond.Contains == "" && len(cond.In) == 0 && cond.Exists == nil:
			return false, nil, fmt.Errorf("нет заполненного оператора")
		case cond.Equals != "":
			ok, reason, checkErr = CheckIfEquals(event, cond)
		case cond.Contains !="":
			ok, reason, checkErr = CheckIfContains(event, cond)
		case len(cond.In) != 0:
			ok, reason, checkErr = CheckIfIn(event, cond)
		case cond.Exists != nil:
			ok, reason, checkErr = CheckIfExists(event, cond)
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