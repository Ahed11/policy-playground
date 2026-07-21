package policy

import (
	"slices"
	"testing"
)
func TestCreateAlertErrors(t *testing.T){
	tests := []struct {
        name   string
        policy Policy
		expectedError string
    }{        
		{
			"нет ни простого, ни составного условия",
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{},
			},
			"нет ни простого, ни составного условия",
		},
		{
			"у простого условия отсутствует поддерживаемый оператор",
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "action",
				},
			},
			"политика pol_external_client_data: у простого условия отсутствует поддерживаемый оператор",
		},
		{
			"одновременно заполнены простое условие и группы all и any",
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "action",
					Equals: "email_send",
					All: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
					Any: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
				},
			},
			"одновременно заполнены простое условие и группы all и any",
		},
		{
			"одновременно заполнены простое условие и группа all",
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "action",
					Equals: "email_send",
					All: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
				},
			},
			"одновременно заполнены простое условие и группа all",
		},
		{
			"одновременно заполнены простое условие и группа any",
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "action",
					Equals: "email_send",
					Any: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
				},
			},
			"одновременно заполнены простое условие и группа any",
		},
		{
			"одновременно заполнены группы all и any",
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					All: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
					Any: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
				},
			},
			"одновременно заполнены группы all и any",
		},
		{
			"группа all существует, но не содержит элементов",
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					All: []Condition{},
				},
			},
			"группа all существует, но не содержит элементов",
		},
		{
			"группа any существует, но не содержит элементов",
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Any: []Condition{},
				},
			},
			"группа any существует, но не содержит элементов",
		},
		{
			"заполнено более одного оператора",
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "action",
					Equals: "email_send",
					Contains: "client_data",
				},
			},
			"заполнено более одного оператора",
		},
    }

	event := Event{
        EventID:         "evt_002",
		Time:            "12:00",
		UserID:          "user_001",
		Action:          "email_send",
		ObjectType:      "file",
		FileName:        new("client_base.xlsx"),
		FileExt:         new("xlsx"),
		ContentClasses:  new([]string{"client_data", "personal_data"}),
		Channel:         "local",
		DestinationType: "external",
		SizeBytes:       new(204800),
    }

	for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            _, _, err := CreateAlert(test.policy, event)
		
			if err == nil {
				t.Fatalf("ожидалась ошибка: %v, но ошибка не получена", test.expectedError)
			}

			if err.Error() != test.expectedError {
				t.Errorf("ожидалась ошибка: %v, но получена ошибка %v", test.expectedError, err)
			}
        })
    }
}

func generalCheck(t *testing.T, expectedResult bool, event Event, policy Policy, expectedReasons []string) {
	
	t.Helper()
	
	alert, created, err := CreateAlert(policy, event)

	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if expectedResult {
		if !created {
			t.Fatalf("ожидался true, получен %v", created)
		}
	
		var expectedAlert Alert = Alert{
			PolicyID:   policy.PolicyID,
			PolicyName: policy.Name,
			Severity:   policy.Severity,
			EventID:    event.EventID,
			UserID:     event.UserID,
			Matched:    true,
			Reasons: expectedReasons,
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
			t.Errorf("не совпадают поля: %v", errors)
		}
	} else {
		if created {
			t.Errorf("ожидался false, получен %v", created)
		}

		var expectedAlert Alert = Alert{
			PolicyID:   "",
			PolicyName: "",
			Severity:   "",
			EventID:    "",
			UserID:     "",
			Matched:    false,
			Reasons: expectedReasons,
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
			t.Errorf("не совпадают поля: %v", errors)
		}

		
	}
}

func TestCreateAlert(t *testing.T) {
	type formForTestCreateAlert struct {
		name string
		event Event
		policy Policy
		expectedResult bool
		expectedReasons []string
	}

	arrayOfForms := []formForTestCreateAlert{
		{
			"successEquals",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "email_send",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("xlsx"),
				ContentClasses:  new([]string{"client_data", "personal_data"}),
				Channel:         "local",
				DestinationType: "external",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "action",
					Equals: "email_send",
				},
			},
			true,
			[]string{"action equals email_send"},
		},
		{
			"notSuccessEquals",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "open_file",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("xlsx"),
				ContentClasses:  new([]string{"client_data", "personal_data"}),
				Channel:         "local",
				DestinationType: "external",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "action",
					Equals: "email_send",
				},
			},
			false,
			[]string{},
		},
		{
			"successContains",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "email_send",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("xlsx"),
				ContentClasses:  new([]string{"client_data", "personal_data"}),
				Channel:         "local",
				DestinationType: "external",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "content_classes",
					Contains: "client_data",
				},
			},
			true,
			[]string{"content_classes contains client_data"},
		},
		{
			"notSuccessContains",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "email_send",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("xlsx"),
				ContentClasses:  new([]string{"personal_data"}),
				Channel:         "local",
				DestinationType: "external",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "content_classes",
					Contains: "client_data",
				},
			},
			false,
			[]string{},
		},
		{
			"successIn",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "email_send",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("xlsx"),
				ContentClasses:  new([]string{"client_data", "personal_data"}),
				Channel:         "local",
				DestinationType: "external",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "file_ext",
					In: []string{"xlsx", "docx", "pdf"},
				},
			},
			true,
			[]string{"file_ext in [xlsx docx pdf]"},
		},
		{
			"notSuccessIn",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "email_send",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("go"),
				ContentClasses:  new([]string{"client_data", "personal_data"}),
				Channel:         "local",
				DestinationType: "external",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Field: "file_ext",
					In: []string{"xlsx", "docx", "pdf"},
				},
			},
			false,
			[]string{},
		},
		{
			"successAll",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "email_send",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("xlsx"),
				ContentClasses:  new([]string{"client_data", "personal_data"}),
				Channel:         "local",
				DestinationType: "external",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					All: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
				},
			},
			true,
			[]string{"action equals email_send", "destination_type equals external", "content_classes contains personal_data", "file_ext in [xlsx docx pdf]"},
		},
		{
			"notSuccessAll",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "open_file",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("xlsx"),
				ContentClasses:  new([]string{"client_data", "personal_data"}),
				Channel:         "local",
				DestinationType: "external",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					All: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
				},
			},
			false,
			[]string{},
		},
		{
			"successAny_1",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "email_send",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("xlsx"),
				ContentClasses:  new([]string{"client_data"}),
				Channel:         "local",
				DestinationType: "internal",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Any: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
				},
			},
			true,
			[]string{"action equals email_send", "file_ext in [xlsx docx pdf]"},
		},
		{
			"successAny_2",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "email_send",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("go"),
				ContentClasses:  new([]string{"client_data"}),
				Channel:         "local",
				DestinationType: "internal",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Any: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
				},
			},
			true,
			[]string{"action equals email_send"},
		},
		{
			"notSuccessAny",
			Event{
				EventID:         "evt_002",
				Time:            "12:00",
				UserID:          "user_001",
				Action:          "open_file",
				ObjectType:      "file",
				FileName:        new("client_base.xlsx"),
				FileExt:         new("go"),
				ContentClasses:  new([]string{"client_data"}),
				Channel:         "local",
				DestinationType: "internal",
				SizeBytes:       new(204800),
			},
			Policy{
				PolicyID:    "pol_external_client_data",
				Name:        "Client data to external channel",
				Severity:    "high",
				Description: "Detects sending client data to external destination",
				Condition: Condition{
					Any: []Condition{
						{Field: "action", Equals: "email_send"},
						{Field: "destination_type", Equals: "external"},
						{Field: "content_classes", Contains: "personal_data"},
						{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					},
				},
			},
			false,
			[]string{},
		},
	}

	for _, test := range arrayOfForms{
		t.Run(test.name, func(t *testing.T){generalCheck(t, test.expectedResult, test.event, test.policy, test.expectedReasons)})
	}
}