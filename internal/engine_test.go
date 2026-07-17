package policy

import (
	"slices"
	"testing"
)

func TestCreateAlert_matched(t *testing.T) {
	var event Event = Event{
		EventID:         "evt_002",
		Time:            "12:00",
		UserID:          "user_001",
		Action:          "email_send",
		ObjectType:      "file",
		FileName:        "client_base.xlsx",
		FileExt:         "xlsx",
		ContentClasses:  []string{"client_data", "personal_data"},
		Channel:         "local",
		DestinationType: "external",
		SizeBytes:       204800,
	}

	var policy Policy = Policy{
		PolicyID:    "pol_external_client_data",
		Name:        "Client data to external channel",
		Severity:    "high",
		Description: "Detects sending client data to external destination",
		Condition: Condition{
			All: []Condition{
				{Field: "action", Equals: "email_send"},
				{Field: "destination_type", Equals: "external"},
			},
		},
	}

	alert, created, err := CreateAlert(policy, event)

	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if created != true {
		t.Fatalf("ожидался true, получен %v", created)
	}

	var expectedAlert Alert = Alert{
		PolicyID:   policy.PolicyID,
		PolicyName: policy.Name,
		Severity:   policy.Severity,
		EventID:    event.EventID,
		UserID:     event.UserID,
		Matched:    true,
		Reasons: []string{
			"action equals email_send",
			"destination_type equals external",
		},
	}

	var errors []string
	if alert.PolicyID != expectedAlert.PolicyID {
    	errors = append(errors, "PolicyID")
	}
	if alert.PolicyName != expectedAlert.PolicyName {
    	errors = append(errors, "PolicyName")
	}
	if alert.Severity != expectedAlert.Severity {
    	errors = append(errors, "PolicySeverity")
	}
	if alert.EventID != expectedAlert.EventID {
    	errors = append(errors, "EventID")
	}
	if alert.UserID != expectedAlert.UserID {
    	errors = append(errors, "UserID")
	}
	if alert.Matched != expectedAlert.Matched {
    	errors = append(errors, "Matched")
	}
	if !slices.Equal(alert.Reasons, expectedAlert.Reasons) {
    	errors = append(errors, "Reasons")
	}
	if len(errors) > 0 {
    	t.Errorf("Не совпадают поля: %v", errors)
	}
}

func TestCreateAlert_not_matched(t *testing.T) {
	var event Event = Event{
		EventID:         "evt_002",
		Time:            "12:00",
		UserID:          "user_001",
		Action:          "email_send",
		ObjectType:      "file",
		FileName:        "client_base.xlsx",
		FileExt:         "xlsx",
		ContentClasses:  []string{"client_data", "personal_data"},
		Channel:         "local",
		DestinationType: "",
		SizeBytes:       204800,
	}

	var policy Policy = Policy{
		PolicyID:    "pol_external_client_data",
		Name:        "Client data to external channel",
		Severity:    "high",
		Description: "Detects sending client data to external destination",
		Condition: Condition{
			All: []Condition{
				{Field: "action", Equals: "email_send"},
				{Field: "destination_type", Equals: "none"},
			},
		},
	}

	alert, created, err := CreateAlert(policy, event)

	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if created != false {
		t.Fatalf("ожидался false, получен %v", created)
	}

	var expectedAlert Alert = Alert{
		PolicyID:   "",
		PolicyName: "",
		Severity:   "",
		EventID:    "",
		UserID:     "",
		Matched:    false,
		Reasons: []string{},
	}

	var errors []string
	if alert.PolicyID != expectedAlert.PolicyID {
    	errors = append(errors, "PolicyID")
	}
	if alert.PolicyName != expectedAlert.PolicyName {
    	errors = append(errors, "PolicyName")
	}
	if alert.Severity != expectedAlert.Severity {
    	errors = append(errors, "PolicySeverity")
	}
	if alert.EventID != expectedAlert.EventID {
    	errors = append(errors, "EventID")
	}
	if alert.UserID != expectedAlert.UserID {
    	errors = append(errors, "UserID")
	}
	if alert.Matched != expectedAlert.Matched {
    	errors = append(errors, "Matched")
	}
	if !slices.Equal(alert.Reasons, expectedAlert.Reasons) {
    	errors = append(errors, "Reasons")
	}
	if len(errors) > 0 {
    	t.Errorf("Не совпадают поля: %v", errors)
	}
}