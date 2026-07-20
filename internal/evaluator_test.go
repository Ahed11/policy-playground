package policy

import (
	"testing"
)

func TestCheckIfEqualsErrors(t *testing.T) {
	tests := []struct {
		name string
		condition Condition
		expectedError string
	}{
		{
			"поле пусто",
			Condition{
				Field: "",
				Equals: "email_send",
			},
			"поле пусто",
		},
		{
			"поле не поддерживается",
			Condition{
				Field: "path",
				Equals: "C:/directoryA/directoryB",
			},
			"поле не поддерживается",
		},
		{
			"нет значения для equals",
			Condition{
				Field: "action",
				Equals: "",
			},
			"нет значения для equals",
		},
	}

	event := Event{
		Action: "email_send",
	}

	for _, test := range tests {
		t.Run(test.name, func (t *testing.T){
			_, _, err := CheckIfEquals(event, test.condition)
			
			if err == nil {
				t.Fatalf("ожидалась ошибка: %v, но ошибка не получена", test.expectedError)
			}

			if err.Error() != test.expectedError {
				t.Errorf("ожидалась ошибка: %v, но получена ошибка: %v", test.expectedError, err)
			}
		})
	}
}

func generalCheckIfEquals(t *testing.T, event Event, condition Condition, expectedResult bool, expectedReason string) {

	t.Helper()

	result, reason, err := CheckIfEquals(event, condition)

	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if result != expectedResult {
		t.Errorf("ожидался %v, получен %v", expectedResult, result)
	}

	if reason != expectedReason {
		t.Errorf("ожидалась причина: %q, получена: %q", expectedReason, reason)
	}
}

func TestCheckIfEquals(t *testing.T) {
	type testCases struct {
		name string
		event Event
		condition Condition
		expectedResult bool
		expectedReason string
	}

	

	arrayOfForms := []testCases{
		{
			"successEquals",
			Event{
				Action: "email_send",
			},
			Condition{
				Field: "action",
				Equals: "email_send",
			},
			true,
			"action equals email_send",
		},
		{
			"NotSuccessEquals",
			Event{
				Action: "open_file",
			},
			Condition{
				Field: "action",
				Equals: "email_send",
			},
			false,
			"",
		},
	}

	for _, test := range arrayOfForms {
		t.Run(test.name, func(t *testing.T){
			generalCheckIfEquals(t, test.event, test.condition, test.expectedResult, test.expectedReason)
		})
	}
}

func TestCheckIfContainsErrors(t *testing.T) {
	tests := []struct {
		name string
		condition Condition
		expectedError string
	}{
		{
			"поле пусто",
			Condition{
				Field: "",
				Contains: "client_data",
			},
			"поле пусто",
		},
		{
			"поле не поддерживается",
			Condition{
				Field: "path",
				Contains: "C:/directoryA/directoryB",
			},
			"поле не поддерживается",
		},
		{
			"нет значения для contains",
			Condition{
				Field: "content_classes",
				Contains: "",
			},
			"нет значения для contains",
		},
	}

	event := Event{
		ContentClasses: []string{
			"client_data",
			"personal_data",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func (t *testing.T){
			_, _, err := CheckIfContains(event, test.condition)

			if err == nil {
				t.Fatalf("ожидалась ошибка: %v, но ошибка не получена", test.expectedError)
			}

			if err.Error() != test.expectedError {
				t.Errorf("ожидалась ошибка: %v, но получена ошибка: %v", test.expectedError, err)
			}
		})
	}
}

func generalCheckIfContains(t *testing.T, event Event, condition Condition, expectedResult bool, expectedReason string) {

	t.Helper()

	result, reason, err := CheckIfContains(event, condition)

	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if result != expectedResult {
		t.Errorf("ожидался %v, получен %v", expectedResult, result)
	}

	if reason != expectedReason {
		t.Errorf("ожидалась причина: %q, получена: %q", expectedReason, reason)
	}
}

func TestCheckIfContains(t *testing.T) {
	type testCases struct {
		name string
		event Event
		condition Condition
		expectedResult bool
		expectedReason string
	}

	

	arrayOfForms := []testCases{
		{
			"successContains",
			Event{
				ContentClasses: []string{
					"client_data",
					"personal_data",
				},
			},
			Condition{
				Field: "content_classes",
				Contains: "client_data",
			},
			true,
			"content_classes contains client_data",
		},
		{
			"NotSuccessContains",
			Event{
				ContentClasses: []string{
					"other_data",
					"personal_data",
				},
			},
			Condition{
				Field: "content_classes",
				Contains: "client_data",
			},
			false,
			"",
		},
	}

	for _, test := range arrayOfForms {
		t.Run(test.name, func(t *testing.T){
			generalCheckIfContains(t, test.event, test.condition, test.expectedResult, test.expectedReason)
		})
	}
}

func TestCheckIfInErrors(t *testing.T) {
	tests := []struct {
		name string
		condition Condition
		expectedError string
	}{
		{
			"поле пусто",
			Condition{
				Field: "",
				In: []string{
					"xlsx",
					"docx",
					"pdf",
				},
			},
			"поле пусто",
		},
		{
			"поле не поддерживается",
			Condition{
				Field: "path",
				In: []string{"C:/directoryA/directoryB"},
			},
			"поле не поддерживается",
		},
		{
			"нет значения для in",
			Condition{
				Field: "file_ext",
				In: []string{},
			},
			"нет значения для in",
		},
	}

	var event Event = Event{
		FileExt: "xlsx",
	}

	for _, test := range tests {
		t.Run(test.name, func (t *testing.T){
			_, _, err := CheckIfIn(event, test.condition)

			if err == nil {
				t.Fatalf("ожидалась ошибка: %v, но ошибка не получена", test.expectedError)
			}

			if err.Error() != test.expectedError {
				t.Errorf("ожидалась ошибка: %v, но получена ошибка: %v", test.expectedError, err)
			}
		})
	}
}

func generalCheckIfIn(t *testing.T, event Event, condition Condition, expectedResult bool, expectedReason string) {

	t.Helper()

	result, reason, err := CheckIfIn(event, condition)

	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if result != expectedResult {
		t.Errorf("ожидался %v, получен %v", expectedResult, result)
	}

	if reason != expectedReason {
		t.Errorf("ожидалась причина: %q, получена: %q", expectedReason, reason)
	}
}

func TestCheckIfIn(t *testing.T) {
	type testCases struct {
		name string
		event Event
		condition Condition
		expectedResult bool
		expectedReason string
	}

	

	arrayOfForms := []testCases{
		{
			"successIn",
			Event{
				FileExt: "xlsx",
			},
			Condition{
				Field: "file_ext",
				In: []string{
					"xlsx",
					"docx",
					"pdf",
				},
			},
			true,
			"file_ext in [xlsx docx pdf]",
		},
		{
			"NotSuccessIn",
			Event{
				FileExt: "go",
			},
			Condition{
				Field: "file_ext",
				In: []string{
					"xlsx",
					"docx",
					"pdf",
				},
			},
			false,
			"",
		},
	}

	for _, test := range arrayOfForms {
		t.Run(test.name, func(t *testing.T){
			generalCheckIfIn(t, test.event, test.condition, test.expectedResult, test.expectedReason)
		})
	}
}

func TestCheckIfAllConditionsErrors(t *testing.T) {
	tests := []struct {
		name string
		condition Condition
		expectedError string
	}{
		{
			"группа all пуста",
			Condition{
				All: []Condition{},
			},
			"группа all пуста",
		},
		{
			"поле не заполнено",
			Condition{
				All: []Condition{
					{Field: "", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			"поле не заполнено",
		},
		{
			"заполнено более одного оператора",
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send", Contains: "personal_data"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"нет заполненного оператора",
			Condition{
				All: []Condition{
					{Field: "action", Equals: ""},
					{Field: "destination_type", Equals: ""},
					{Field: "content_classes", Contains: ""},
					{Field: "file_ext", In: []string{}},
				},
			},
			"нет заполненного оператора",
		},
	}

	var event Event = Event{
		Action: "email_send",
		DestinationType: "external",
		ContentClasses: []string{"client_data","personal_data",},
		FileExt: "xlsx",
	}

	for _, test := range tests {
		t.Run(test.name, func (t *testing.T){
			_, _, err := AllConditions(event, test.condition)

			if err == nil {
				t.Fatalf("ожидалась ошибка: %v, но ошибка не получена", test.expectedError)
			}

			if err.Error() != test.expectedError {
				t.Errorf("ожидалась ошибка: %v, но получена ошибка: %v", test.expectedError, err)
			}
		})
	}
}

func generalCheckIfAllConditions(t *testing.T, event Event, condition Condition, expectedResult bool, expectedReasons []string) {

	t.Helper()

	result, reasons, err := AllConditions(event, condition)

	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if result != expectedResult {
			t.Errorf("ожидался %v, получен %v", expectedResult, result)
	}

	if expectedResult {
		if len(reasons) != len(expectedReasons) {
			t.Fatalf("количество причин неверно")
		} else {
			for i := range reasons {
				if reasons[i] != expectedReasons[i] {
					t.Errorf("ожидалась причина: %q, получена: %q", expectedReasons[i], reasons[i])
				}
			}
		}
	} else {
		if len(reasons) != 0 { 
			t.Fatalf("количетсво причин неверно")
		}
	}
}

func TestCheckIfAllConditions(t *testing.T) {
	type testCases struct {
		name string
		event Event
		condition Condition
		expectedResult bool
		expectedReasons []string
	}

	arrayOfForms := []testCases{
		{
			"successAllConditions",
			Event{
				Action: "email_send",
				DestinationType: "external",
				ContentClasses: []string{"client_data","personal_data",},
				FileExt: "xlsx",
			},
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			true,
			[]string{
				"action equals email_send",
				"destination_type equals external",
				"content_classes contains personal_data",
				"file_ext in [xlsx docx pdf]",
			},
		},
		{
			"NotSuccessAllConditions",
			Event{
				Action: "email_send",
				DestinationType: "external",
				ContentClasses: []string{"client_data","personal_data",},
				FileExt: "xlsx",
			},
			Condition{
				All: []Condition{
					{Field: "action", Equals: "open_file"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "other_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			false,
			[]string{},
		},
	}

	for _, test := range arrayOfForms {
		t.Run(test.name, func(t *testing.T){
			generalCheckIfAllConditions(t, test.event, test.condition, test.expectedResult, test.expectedReasons)
		})
	}
}

func TestCheckIfAnyConditionsErrors(t *testing.T) {
	tests := []struct {
		name string
		condition Condition
		expectedError string
	}{
		{
			"группа any пуста",
			Condition{
				Any: []Condition{},
			},
			"группа any пуста",
		},
		{
			"поле не заполнено",
			Condition{
				Any: []Condition{
					{Field: "", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			"поле не заполнено",
		},
		{
			"заполнено более одного оператора",
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send", Contains: "personal_data"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"нет заполненного оператора",
			Condition{
				Any: []Condition{
					{Field: "action", Equals: ""},
					{Field: "destination_type", Equals: ""},
					{Field: "content_classes", Contains: ""},
					{Field: "file_ext", In: []string{}},
				},
			},
			"нет заполненного оператора",
		},
	}

	var event Event = Event{
		Action: "email_send",
		DestinationType: "external",
		ContentClasses: []string{"client_data","personal_data",},
		FileExt: "xlsx",
	}

	for _, test := range tests {
		t.Run(test.name, func (t *testing.T){
			_, _, err := AnyConditions(event, test.condition)

			if err == nil {
				t.Fatalf("ожидалась ошибка: %v, но ошибка не получена", test.expectedError)
			}

			if err.Error() != test.expectedError {
				t.Errorf("ожидалась ошибка: %v, но получена ошибка: %v", test.expectedError, err)
			}
		})
	}
}

func generalCheckIfAnyConditions(t *testing.T, event Event, condition Condition, expectedResult bool, expectedReasons []string) {

	t.Helper()

	result, reasons, err := AnyConditions(event, condition)

	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if result != expectedResult {
			t.Errorf("ожидался %v, получен %v", expectedResult, result)
	}

	if expectedResult {
		if len(reasons) != len(expectedReasons) {
			t.Fatalf("количество причин неверно")
		} else {
			for i := range reasons {
				if reasons[i] != expectedReasons[i] {
					t.Errorf("ожидалась причина: %q, получена: %q", expectedReasons[i], reasons[i])
				}
			}
		}
	} else {
		if len(reasons) != 0 { 
			t.Fatalf("количетсво причин неверно")
		}
	}
}

func TestCheckIfAnyConditions(t *testing.T) {
	type testCases struct {
		name string
		event Event
		condition Condition
		expectedResult bool
		expectedReasons []string
	}

	arrayOfForms := []testCases{
		{
			"successAnyConditions_1",
			Event{
				Action: "open_file",
				DestinationType: "internal",
				ContentClasses: []string{"client_data","personal_data",},
				FileExt: "xlsx",
			},
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			true,
			[]string{
				"content_classes contains personal_data",
				"file_ext in [xlsx docx pdf]",
			},
		},
		{
			"successAnyConditions_2",
			Event{
				Action: "email_send",
				DestinationType: "internal",
				ContentClasses: []string{"client_data"},
				FileExt: "go",
			},
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			true,
			[]string{
				"action equals email_send",
			},
		},
		{
			"successAnyConditions_3",
			Event{
				Action: "open_file",
				DestinationType: "external",
				ContentClasses: []string{"client_data"},
				FileExt: "go",
			},
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			true,
			[]string{
				"destination_type equals external",
			},
		},
		{
			"successAnyConditions_4",
			Event{
				Action: "email_send",
				DestinationType: "external",
				ContentClasses: []string{"client_data","personal_data",},
				FileExt: "xlsx",
			},
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			true,
			[]string{
				"action equals email_send",
				"destination_type equals external",
				"content_classes contains personal_data",
				"file_ext in [xlsx docx pdf]",
			},
		},
		{
			"NotSuccessAnyConditions",
			Event{
				Action: "open_file",
				DestinationType: "internal",
				ContentClasses: []string{
					"client_data",
				},
				FileExt: "go",
			},
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
				},
			},
			false,
			[]string{},
		},
	}

	for _, test := range arrayOfForms {
		t.Run(test.name, func(t *testing.T){
			generalCheckIfAnyConditions(t, test.event, test.condition, test.expectedResult, test.expectedReasons)
		})
	}
}