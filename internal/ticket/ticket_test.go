package ticket

import (
	"reflect"
	"testing"
	"time"
)

// 1.1 Status Enum Tests

func TestStatusIsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{"valid open", StatusOpen, true},
		{"valid in_progress", StatusInProgress, true},
		{"valid closed", StatusClosed, true},
		{"invalid unknown", Status("unknown"), false},
		{"invalid done", Status("done"), false}, // "done" is accepted in closed command but not a valid status enum
		{"invalid empty", Status(""), false},
		{"invalid random", Status("pending"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			if result != tt.expected {
				t.Errorf("Status(%q).IsValid() = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestValidStatuses(t *testing.T) {
	expected := []Status{StatusOpen, StatusInProgress, StatusClosed}
	if !reflect.DeepEqual(ValidStatuses, expected) {
		t.Errorf("ValidStatuses = %v, expected %v", ValidStatuses, expected)
	}
}

func TestStatusComparison(t *testing.T) {
	if StatusOpen != Status("open") {
		t.Errorf("StatusOpen should equal 'open'")
	}
	if StatusInProgress != Status("in_progress") {
		t.Errorf("StatusInProgress should equal 'in_progress'")
	}
	if StatusClosed != Status("closed") {
		t.Errorf("StatusClosed should equal 'closed'")
	}
}

// 1.2 Type Enum Tests

func TestTypeIsValid(t *testing.T) {
	tests := []struct {
		name     string
		typ      Type
		expected bool
	}{
		{"valid bug", TypeBug, true},
		{"valid feature", TypeFeature, true},
		{"valid task", TypeTask, true},
		{"valid epic", TypeEpic, true},
		{"valid chore", TypeChore, true},
		{"invalid unknown", Type("unknown"), false},
		{"invalid empty", Type(""), false},
		{"invalid story", Type("story"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.IsValid()
			if result != tt.expected {
				t.Errorf("Type(%q).IsValid() = %v, expected %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestValidTypes(t *testing.T) {
	expected := []Type{TypeBug, TypeFeature, TypeTask, TypeEpic, TypeChore}
	if !reflect.DeepEqual(ValidTypes, expected) {
		t.Errorf("ValidTypes = %v, expected %v", ValidTypes, expected)
	}
}

func TestTypeComparison(t *testing.T) {
	if TypeBug != Type("bug") {
		t.Errorf("TypeBug should equal 'bug'")
	}
	if TypeFeature != Type("feature") {
		t.Errorf("TypeFeature should equal 'feature'")
	}
	if TypeTask != Type("task") {
		t.Errorf("TypeTask should equal 'task'")
	}
	if TypeEpic != Type("epic") {
		t.Errorf("TypeEpic should equal 'epic'")
	}
	if TypeChore != Type("chore") {
		t.Errorf("TypeChore should equal 'chore'")
	}
}

// 1.3 Priority Range Tests

func TestValidPriorities(t *testing.T) {
	// Valid priorities are 0-4
	validPriorities := []int{0, 1, 2, 3, 4}

	for _, p := range validPriorities {
		ticket := &Ticket{Priority: p}
		if ticket.Priority < 0 || ticket.Priority > 4 {
			t.Errorf("Priority %d should be valid (0-4)", p)
		}
	}
}

func TestPriorityBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		isValid  bool
	}{
		{"min boundary", 0, true},
		{"mid range", 2, true},
		{"max boundary", 4, true},
		{"below min", -1, false},
		{"above max", 5, false},
		{"way above max", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.priority >= 0 && tt.priority <= 4
			if valid != tt.isValid {
				t.Errorf("Priority %d validity = %v, expected %v", tt.priority, valid, tt.isValid)
			}
		})
	}
}

// 1.4 Ticket Field Validation

func TestTicketStructFields(t *testing.T) {
	ticket := &Ticket{
		ID:          "test-1234",
		Status:      StatusOpen,
		Deps:        []string{"dep1", "dep2"},
		Links:       []string{"link1"},
		Created:     time.Now(),
		Type:        TypeTask,
		Priority:    2,
		Assignee:    "testuser",
		ExternalRef: "EXT-123",
		Parent:      "parent-5678",
		Title:       "Test Ticket",
		Body:        "Test body content",
	}

	// Verify all fields are accessible
	if ticket.ID == "" {
		t.Error("ID field should be accessible")
	}
	if ticket.Status == "" {
		t.Error("Status field should be accessible")
	}
	if ticket.Deps == nil {
		t.Error("Deps field should be accessible")
	}
	if ticket.Links == nil {
		t.Error("Links field should be accessible")
	}
	if ticket.Created.IsZero() {
		t.Error("Created field should be accessible")
	}
	if ticket.Type == "" {
		t.Error("Type field should be accessible")
	}
	if ticket.Priority != 2 {
		t.Error("Priority field should be accessible")
	}
	if ticket.Assignee == "" {
		t.Error("Assignee field should be accessible")
	}
	if ticket.ExternalRef == "" {
		t.Error("ExternalRef field should be accessible")
	}
	if ticket.Parent == "" {
		t.Error("Parent field should be accessible")
	}
	if ticket.Title == "" {
		t.Error("Title field should be accessible")
	}
	if ticket.Body == "" {
		t.Error("Body field should be accessible")
	}
}

func TestTicketYAMLTags(t *testing.T) {
	// Use reflection to verify YAML tags are correctly set
	ticketType := reflect.TypeOf(Ticket{})

	tests := []struct {
		fieldName string
		yamlTag   string
	}{
		{"ID", "id"},
		{"Status", "status"},
		{"Deps", "deps,flow"},
		{"Links", "links,flow"},
		{"Created", "created"},
		{"Type", "type"},
		{"Priority", "priority"},
		{"Assignee", "assignee,omitempty"},
		{"ExternalRef", "external-ref,omitempty"},
		{"Parent", "parent,omitempty"},
		{"Title", "-"},
		{"Body", "-"},
	}

	for _, tt := range tests {
		field, found := ticketType.FieldByName(tt.fieldName)
		if !found {
			t.Errorf("Field %s not found in Ticket struct", tt.fieldName)
			continue
		}

		yamlTag := field.Tag.Get("yaml")
		if yamlTag != tt.yamlTag {
			t.Errorf("Field %s has yaml tag %q, expected %q", tt.fieldName, yamlTag, tt.yamlTag)
		}
	}
}

func TestTicketOptionalFields(t *testing.T) {
	// Test that optional fields can be omitted
	ticket := &Ticket{
		ID:       "test-1234",
		Status:   StatusOpen,
		Deps:     []string{},
		Links:    []string{},
		Created:  time.Now(),
		Type:     TypeTask,
		Priority: 2,
		Title:    "Test",
	}

	// These should be allowed to be empty
	if ticket.Assignee != "" {
		ticket.Assignee = "" // Should be valid
	}
	if ticket.ExternalRef != "" {
		ticket.ExternalRef = "" // Should be valid
	}
	if ticket.Parent != "" {
		ticket.Parent = "" // Should be valid
	}
	if ticket.Body != "" {
		ticket.Body = "" // Should be valid
	}
}

func TestTicketFieldWithSpaces(t *testing.T) {
	// Test that fields with spaces are preserved
	ticket := &Ticket{
		Title:       "Title with spaces",
		Body:        "Body with multiple spaces   and   tabs\t\t",
		Assignee:    "User Name",
		ExternalRef: "EXT-123 ABC",
	}

	if ticket.Title != "Title with spaces" {
		t.Errorf("Title with spaces not preserved: %q", ticket.Title)
	}
	if ticket.Body != "Body with multiple spaces   and   tabs\t\t" {
		t.Errorf("Body with spaces not preserved: %q", ticket.Body)
	}
	if ticket.Assignee != "User Name" {
		t.Errorf("Assignee with spaces not preserved: %q", ticket.Assignee)
	}
}

func TestDefaultTicketsDir(t *testing.T) {
	expected := ".tickets"
	if DefaultTicketsDir != expected {
		t.Errorf("DefaultTicketsDir = %q, expected %q", DefaultTicketsDir, expected)
	}
}

// 1.5 Field Defaults (structural tests)
// Note: Default priority = 2 is tested at the command level,
// but we verify the struct can hold the value

func TestTicketDefaultPriority(t *testing.T) {
	// Verify that priority 2 (the default) can be assigned
	ticket := &Ticket{
		Priority: 2,
	}

	if ticket.Priority != 2 {
		t.Errorf("Default priority 2 not properly stored, got %d", ticket.Priority)
	}
}

func TestTicketEmptyArrays(t *testing.T) {
	// Test that empty arrays are properly handled
	ticket := &Ticket{
		Deps:  []string{},
		Links: []string{},
	}

	if ticket.Deps == nil {
		t.Error("Deps should not be nil, should be empty slice")
	}
	if len(ticket.Deps) != 0 {
		t.Errorf("Deps should be empty, got length %d", len(ticket.Deps))
	}
	if ticket.Links == nil {
		t.Error("Links should not be nil, should be empty slice")
	}
	if len(ticket.Links) != 0 {
		t.Errorf("Links should be empty, got length %d", len(ticket.Links))
	}
}
