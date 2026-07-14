package policy

import "testing"

//Везде где идет разделение проверки функции на событие 1 и 2, для события 1 она должна возвращать результат false , а 2 возвращать true.  

func TestCheckIfEquals_FieldIsEmptyInEvent(t *testing.T) {
	var event Event = Event{
		Action: "email_send",
	}

	var condition Condition = Condition{
		Field: "",
		Equals: "email_send",
	}

	_, _, err := CheckIfEquals(event, condition)

	if err == nil {
		t.Fatalf("ожидалась ошибка, но она не была возвращена")
	}
}

func TestCheckIfEquals_FieldIsNotActionOrDestinationTypeInEvent(t *testing.T) {
	var event Event = Event{
		Action: "email_send",
	}

	var condition Condition = Condition{
		Field: "Example",
		Equals: "email_send",
	}

	_, _, err := CheckIfEquals(event, condition)

	if err == nil {
		t.Fatalf("ожидалась ошибка, но она не была возвращена")
	}
}

func TestCheckIfEquals_ActionInEvent1(t *testing.T) {

	var event_001 Event = Event{
		Action: "open_file",
	}

	var condition Condition = Condition{
		Field: "action",
		Equals: "email_send",
	}

	result, reason, err := CheckIfEquals(event_001, condition)
	
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	} 
	
	if result != false {
		t.Errorf("Ожидался false, получен %v", result)	
	} 

	expectedReason := ""
	
	if reason != expectedReason {
		t.Errorf("Ожидалась причина %q, получена %q", expectedReason, reason)
	}
}

func TestCheckIfEquals_ActionInEvent2(t *testing.T) {

	var event_002 Event = Event{
		Action: "email_send",
	}

	var condition Condition = Condition{
		Field: "action",
		Equals: "email_send",
	}

	result, reason, err := CheckIfEquals(event_002, condition)
	
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	} 
	
	if result != true {
		t.Errorf("Ожидался true, получен %v", result)	
	} 

	expectedReason := "action equals email_send"
	
	if reason != expectedReason {
		t.Errorf("Ожидалась причина %q, получена %q", expectedReason, reason)
	}
}

func TestCheckIfEquals_DestinationTypeInEvent1(t *testing.T) {

	var event_001 Event = Event{
		DestinationType: "none",
	}

	var condition Condition = Condition{
		Field: "destination_type",
		Equals: "external",
	}

	result, reason, err := CheckIfEquals(event_001, condition)
	
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	} 
	
	if result != false {
		t.Errorf("Ожидался false, получен %v", result)	
	} 

	expectedReason := ""
	
	if reason != expectedReason {
		t.Errorf("Ожидалась причина %q, получена %q", expectedReason, reason)
	}
}

func TestCheckIfEquals_DestinationTypeInEvent2(t *testing.T) {

	var event_002 Event = Event{
		DestinationType: "external",
	}

	var condition Condition = Condition{
		Field: "destination_type",
		Equals: "external",
	}

	result, reason, err := CheckIfEquals(event_002, condition)
	
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	} 
	
	if result != true {
		t.Errorf("Ожидался true, получен %v", result)	
	} 

	expectedReason := "destination_type equals external"
	
	if reason != expectedReason {
		t.Errorf("Ожидалась причина %q, получена %q", expectedReason, reason)
	}
}

func TestAllConditionsEvent1(t *testing.T) {
	
	var event Event = Event{
		Action: "email_send",
		DestinationType: "none",
	}

	var condition Condition = Condition{
		All: []Condition{
			{Field: "action", Equals: "email_send"},
			{Field: "destination_type", Equals: "external"},
		},
	}

	result, reasons, err := AllConditions(event, condition)

	if err != nil {
		t.Fatalf("Ошибка: %v", err)
	}

	if result != false {
		t.Errorf("ожидется false, получен %v", result)
	}

	if len(reasons) != 0 {
		t.Fatalf("количетсво причин неверно")
	}
}

func TestAllConditionsEvent2(t *testing.T) {
	
	var event Event = Event{
		Action: "email_send",
		DestinationType: "external",
	}

	var condition Condition = Condition{
		All: []Condition{
			{Field: "action", Equals: "email_send"},
			{Field: "destination_type", Equals: "external"},
		},
	}

	result, reasons, err := AllConditions(event, condition)

	if err != nil {
		t.Fatalf("Ошибка: %v", err)
	}

	if result != true {
		t.Errorf("Ожидался true, получен %v", result)
	}

	if len(reasons) != 2 {
		t.Fatalf("количество причин неверно")
	} else {
		expectedReasons := []string{"action equals email_send" ,"destination_type equals external"}
	
		for i := range reasons {
			if reasons[i] != expectedReasons[i] {
				t.Errorf("Причина неправильно установлена")
			}
		}
	}
}