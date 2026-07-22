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
			"successEquals_1",
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
			"successEquals_2",
			Event{
				SizeBytes: new(204800),
			},
			Condition{
				Field: "size_bytes",
				Equals: "204800",
			},
			true,
			"size_bytes equals 204800",
		},
		{
			"NotSuccessEquals_1",
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
		{
			"NotSuccessEquals_2",
			Event{
				FileName: nil,
			},
			Condition{
				Field: "file_name",
				Equals: "file",
			},
			false,
			"",
		},
		{
			"NotSuccessEquals_3",
			Event{
				FileExt: nil,
			},
			Condition{
				Field: "file_ext",
				Equals: "xlsx",
			},
			false,
			"",
		},
		{
			"NotSuccessEquals_4",
			Event{
				SizeBytes: nil,
			},
			Condition{
				Field: "size_bytes",
				Equals: "204800",
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
		ContentClasses: new([]string{
			"client_data",
			"personal_data",
		}),
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
				ContentClasses: new([]string{
					"client_data",
					"personal_data",
				}),
			},
			Condition{
				Field: "content_classes",
				Contains: "client_data",
			},
			true,
			"content_classes contains client_data",
		},
		{
			"NotSuccessContains_1",
			Event{
				ContentClasses: new([]string{
					"other_data",
					"personal_data",
				}),
			},
			Condition{
				Field: "content_classes",
				Contains: "client_data",
			},
			false,
			"",
		},
		{
			"NotSuccessContains_2",
			Event{
				ContentClasses: nil,
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
		FileExt: new("xlsx"),
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
			"successIn_1",
			Event{
				FileExt: new("xlsx"),
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
			"successIn_2",
			Event{
				SizeBytes: new(204800),
			},
			Condition{
				Field: "size_bytes",
				In: []string{
					"204800",
					"102400",
					"51200",
				},
			},
			true,
			"size_bytes in [204800 102400 51200]",
		},
		{
			"NotSuccessIn_1",
			Event{
				FileExt: new("go"),
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
		{
			"NotSuccessIn_2",
			Event{
				FileExt: nil,
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
		{
			"NotSuccessIn_3",
			Event{
				FileName: nil,
			},
			Condition{
				Field: "file_name",
				In: []string{
					"document",
					"table",
					"file",
				},
			},
			false,
			"",
		},
		{
			"NotSuccessIn_4",
			Event{
				SizeBytes: nil,
			},
			Condition{
				Field: "size_bytes",
				In: []string{
					"204800",
					"102400",
					"51200",
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

func TestCheckIfExistsErrors(t *testing.T) {
	tests := []struct {
		name string
		condition Condition
		expectedError string
	}{
		{
			"поле пусто",
			Condition{
				Field: "",
				Exists: new(true),
			},
			"поле пусто",
		},
		{
			"поле exists не существует",
			Condition{
				Field: "file_ext",
			},
			"поле exists не существует",
		},
		{
			"поле не поддерживается",
			Condition{
				Field: "path",
				Exists: new(true),
			},
			"поле не поддерживается",
		},
	}

	var event Event = Event{
		FileName:       new("client_base.xlsx"),
		FileExt:        new("xlsx"),
		ContentClasses: new([]string{"client_data", "personal_data"}),
		SizeBytes:      new(204800),
	}

	for _, test := range tests {
		t.Run(test.name, func (t *testing.T){
			_, _, err := CheckIfExists(event, test.condition)

			if err == nil {
				t.Fatalf("ожидалась ошибка: %v, но ошибка не получена", test.expectedError)
			}

			if err.Error() != test.expectedError {
				t.Errorf("ожидалась ошибка: %v, но получена ошибка: %v", test.expectedError, err)
			}
		})
	}
}

func generalCheckIfExists(t *testing.T, event Event, condition Condition, expectedResult bool, expectedReason string) {

	t.Helper()

	result, reason, err := CheckIfExists(event, condition)

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

func TestCheckIfExists(t *testing.T) {
	type testCases struct {
		name string
		event Event
		condition Condition
		expectedResult bool
		expectedReason string
	}

	arrayOfForms := []testCases{
		{
			"successExist_file_ext_1",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "file_ext",
				Exists: new(true),
			},
			true,
			"file_ext exists",
		},
		{
			"successExist_file_ext_2",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        nil,
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "file_ext",
				Exists: new(false),
			},
			true,
			"file_ext does not exist",
		},
		{
			"successExist_file_ext_3",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new(""),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "file_ext",
				Exists: new(true),
			},
			true,
			"file_ext exists",
		},
		{
			"NotSuccessExists_file_ext_1",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "file_ext",
				Exists: new(false),
			},
			false,
			"",
		},
		{
			"NotSuccessExists_file_ext_2",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        nil,
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "file_ext",
				Exists: new(true),
			},
			false,
			"",
		},
		{
			"successExist_file_name_1",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "file_name",
				Exists: new(true),
			},
			true,
			"file_name exists",
		},
		{
			"successExist_file_name_2",
			Event{
				FileName:       nil,
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "file_name",
				Exists: new(false),
			},
			true,
			"file_name does not exist",
		},
		{
			"successExist_file_name_3",
			Event{
				FileName:       new(""),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "file_name",
				Exists: new(true),
			},
			true,
			"file_name exists",
		},
		{
			"NotSuccessExists_file_name_1",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "file_name",
				Exists: new(false),
			},
			false,
			"",
		},
		{
			"NotSuccessExists_file_name_2",
			Event{
				FileName:       nil,
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "file_name",
				Exists: new(true),
			},
			false,
			"",
		},
		{
			"successExist_content_classes_1",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "content_classes",
				Exists: new(true),
			},
			true,
			"content_classes exists",
		},
		{
			"successExist_content_classes_2",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: nil,
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "content_classes",
				Exists: new(false),
			},
			true,
			"content_classes does not exist",
		},
		{
			"successExist_content_classes_3",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "content_classes",
				Exists: new(true),
			},
			true,
			"content_classes exists",
		},
		{
			"NotSuccessExists_content_classes_1",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "content_classes",
				Exists: new(false),
			},
			false,
			"",
		},
		{
			"NotSuccessExists_content_classes_2",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: nil,
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "content_classes",
				Exists: new(true),
			},
			false,
			"",
		},
		{
			"successExist_size_bytes_1",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "size_bytes",
				Exists: new(true),
			},
			true,
			"size_bytes exists",
		},
		{
			"successExist_size_bytes_2",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      nil,
			},
			Condition{
				Field: "size_bytes",
				Exists: new(false),
			},
			true,
			"size_bytes does not exist",
		},
		{
			"successExist_size_bytes_3",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(0),
			},
			Condition{
				Field: "size_bytes",
				Exists: new(true),
			},
			true,
			"size_bytes exists",
		},
		{
			"NotSuccessExists_size_bytes_1",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      new(204800),
			},
			Condition{
				Field: "size_bytes",
				Exists: new(false),
			},
			false,
			"",
		},
		{
			"NotSuccessExists_size_bytes_2",
			Event{
				FileName:       new("client_base.xlsx"),
				FileExt:        new("xlsx"),
				ContentClasses: new([]string{"client_data", "personal_data"}),
				SizeBytes:      nil,
			},
			Condition{
				Field: "size_bytes",
				Exists: new(true),
			},
			false,
			"",
		},
	}

	for _, test := range arrayOfForms {
		t.Run(test.name, func(t *testing.T){
			generalCheckIfExists(t, test.event, test.condition, test.expectedResult, test.expectedReason)
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
					{Field: "file_name", Exists: new(true)},
				},
			},
			"поле не заполнено",
		},
		{
			"заполнено более одного оператора 1",
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send", Contains: "personal_data"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"заполнено более одного оператора 2",
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"заполнено более одного оператора 3",
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send", Exists: new(true)},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"заполнено более одного оператора 4",
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"заполнено более одного оператора 5",
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data", Exists: new(true)},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"заполнено более одного оператора 6",
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}, Exists: new(true)},
					{Field: "file_name", Exists: new(true)},
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
					{Field: "file_name", Exists: nil},
				},
			},
			"нет заполненного оператора",
		},
	}

	var event Event = Event{
		Action: "email_send",
		DestinationType: "external",
		ContentClasses: new([]string{"client_data","personal_data",}),
		FileExt: new("xlsx"),
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
			"successAllConditions_1",
			Event{
				Action: "email_send",
				DestinationType: "external",
				ContentClasses: new([]string{"client_data","personal_data",}),
				FileExt: new("xlsx"),
			},
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_ext", Exists: new(true)},
				},
			},
			true,
			[]string{
				"action equals email_send",
				"destination_type equals external",
				"content_classes contains personal_data",
				"file_ext in [xlsx docx pdf]",
				"file_ext exists",
			},
		},
		{
			"successAllConditions_2",
			Event{
				Action: "email_send",
				DestinationType: "external",
				ContentClasses: new([]string{"client_data","personal_data",}),
				FileExt: new("xlsx"),
			},
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(false)},
				},
			},
			true,
			[]string{
				"action equals email_send",
				"destination_type equals external",
				"content_classes contains personal_data",
				"file_ext in [xlsx docx pdf]",
				"file_name does not exist",
			},
		},
		{
			"successAllConditions_3",
			Event{
				Action: "email_send",
				DestinationType: "external",
				ContentClasses: new([]string{"client_data","personal_data",}),
				FileExt: new(""),
			},
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_ext", Exists: new(true)},
				},
			},
			true,
			[]string{
				"action equals email_send",
				"destination_type equals external",
				"content_classes contains personal_data",
				"file_ext in [xlsx docx pdf]",
				"file_ext exists",
			},
		},
		{
			"NotSuccessAllConditions_1",
			Event{
				Action: "open_file",
				DestinationType: "external",
				ContentClasses: new([]string{"client_data"}),
				FileExt: new("xlsx"),
			},
			Condition{
				All: []Condition{
					{Field: "action", Equals: "emai_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_ext", Exists: new(true)},
				},
			},
			false,
			[]string{},
		},
		{
			"NotSuccessAllConditions_2",
			Event{
				Action: "email_send",
				DestinationType: "external",
				ContentClasses: new([]string{"client_data", "personal_data"}),
				FileExt: new("xlsx"),
			},
			Condition{
				All: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
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
					{Field: "file_ext", Exists: new(true)},
				},
			},
			"поле не заполнено",
		},
		{
			"заполнено более одного оператора 1",
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send", Contains: "personal_data"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"заполнено более одного оператора 2",
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"заполнено более одного оператора 3",
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send", Exists: new(true)},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"заполнено более одного оператора 4",
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"заполнено более одного оператора 5",
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data", Exists: new(true)},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
				},
			},
			"заполнено более одного оператора",
		},
		{
			"заполнено более одного оператора 6",
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}, Exists: new(true)},
					{Field: "file_name", Exists: new(true)},
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
					{Field: "file_ext", Exists: nil},
				},
			},
			"нет заполненного оператора",
		},
	}

	var event Event = Event{
		Action: "email_send",
		DestinationType: "external",
		ContentClasses: new([]string{"client_data","personal_data",}),
		FileExt: new("xlsx"),
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
				ContentClasses: new([]string{"client_data","personal_data",}),
				FileExt: new("xlsx"),
			},
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(false)},
				},
			},
			true,
			[]string{
				"content_classes contains personal_data",
				"file_ext in [xlsx docx pdf]",
				"file_name does not exist",
			},
		},
		{
			"successAnyConditions_2",
			Event{
				Action: "email_send",
				DestinationType: "internal",
				ContentClasses: new([]string{"client_data"}),
				FileExt: new("go"),
			},
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_ext", Exists: new(true)},
				},
			},
			true,
			[]string{
				"action equals email_send",
				"file_ext exists",
			},
		},
		{
			"successAnyConditions_3",
			Event{
				Action: "open_file",
				DestinationType: "external",
				ContentClasses: new([]string{"client_data"}),
				FileExt: new("go"),
			},
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
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
				ContentClasses: new([]string{"client_data","personal_data",}),
				FileExt: new("xlsx"),
			},
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_ext", Exists: new(true)},
				},
			},
			true,
			[]string{
				"action equals email_send",
				"destination_type equals external",
				"content_classes contains personal_data",
				"file_ext in [xlsx docx pdf]",
				"file_ext exists",
			},
		},
		{
			"NotSuccessAnyConditions",
			Event{
				Action: "open_file",
				DestinationType: "internal",
				ContentClasses: new([]string{
					"client_data",
				}),
				FileExt: new("go"),
			},
			Condition{
				Any: []Condition{
					{Field: "action", Equals: "email_send"},
					{Field: "destination_type", Equals: "external"},
					{Field: "content_classes", Contains: "personal_data"},
					{Field: "file_ext", In: []string{"xlsx", "docx", "pdf"}},
					{Field: "file_name", Exists: new(true)},
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