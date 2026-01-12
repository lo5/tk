package ticket

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// 1.3 YAML Parsing Tests

// TestParseMinimalTicket tests parsing a minimal valid ticket
func TestParseMinimalTicket(t *testing.T) {
	content := `---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# Test Ticket
`
	ticket, err := Parse(strings.NewReader(content))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if ticket.ID != "test-1234" {
		t.Errorf("ID = %q, expected %q", ticket.ID, "test-1234")
	}
	if ticket.Status != StatusOpen {
		t.Errorf("Status = %q, expected %q", ticket.Status, StatusOpen)
	}
	if ticket.Title != "Test Ticket" {
		t.Errorf("Title = %q, expected %q", ticket.Title, "Test Ticket")
	}
}

// TestParseFullTicket tests parsing a ticket with all fields
func TestParseFullTicket(t *testing.T) {
	content := `---
id: test-1234
status: in_progress
deps: [dep-1, dep-2]
links: [link-1]
created: 2025-01-11T10:00:00Z
type: feature
priority: 1
assignee: testuser
external-ref: EXT-123
parent: parent-5678
---
# Full Test Ticket

This is the body content.

Multiple paragraphs are supported.
`
	ticket, err := Parse(strings.NewReader(content))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify all fields
	if ticket.ID != "test-1234" {
		t.Errorf("ID = %q, expected %q", ticket.ID, "test-1234")
	}
	if ticket.Status != StatusInProgress {
		t.Errorf("Status = %q, expected %q", ticket.Status, StatusInProgress)
	}
	if len(ticket.Deps) != 2 || ticket.Deps[0] != "dep-1" || ticket.Deps[1] != "dep-2" {
		t.Errorf("Deps = %v, expected [dep-1, dep-2]", ticket.Deps)
	}
	if len(ticket.Links) != 1 || ticket.Links[0] != "link-1" {
		t.Errorf("Links = %v, expected [link-1]", ticket.Links)
	}
	if ticket.Type != TypeFeature {
		t.Errorf("Type = %q, expected %q", ticket.Type, TypeFeature)
	}
	if ticket.Priority != 1 {
		t.Errorf("Priority = %d, expected 1", ticket.Priority)
	}
	if ticket.Assignee != "testuser" {
		t.Errorf("Assignee = %q, expected %q", ticket.Assignee, "testuser")
	}
	if ticket.ExternalRef != "EXT-123" {
		t.Errorf("ExternalRef = %q, expected %q", ticket.ExternalRef, "EXT-123")
	}
	if ticket.Parent != "parent-5678" {
		t.Errorf("Parent = %q, expected %q", ticket.Parent, "parent-5678")
	}
	if ticket.Title != "Full Test Ticket" {
		t.Errorf("Title = %q, expected %q", ticket.Title, "Full Test Ticket")
	}
	if !strings.Contains(ticket.Body, "Multiple paragraphs") {
		t.Errorf("Body does not contain expected content: %q", ticket.Body)
	}
}

// TestParseTypesPreserved tests that field types are correctly preserved
func TestParseTypesPreserved(t *testing.T) {
	content := `---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:30:45Z
type: bug
priority: 3
---
# Test
`
	ticket, err := Parse(strings.NewReader(content))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify types
	if _, ok := interface{}(ticket.Priority).(int); !ok {
		t.Error("Priority should be int type")
	}
	if ticket.Priority != 3 {
		t.Errorf("Priority = %d, expected 3", ticket.Priority)
	}

	expectedTime := time.Date(2025, 1, 11, 10, 30, 45, 0, time.UTC)
	if !ticket.Created.Equal(expectedTime) {
		t.Errorf("Created = %v, expected %v", ticket.Created, expectedTime)
	}
}

// TestFrontmatterFlowStyle tests that deps and links use flow style
func TestFrontmatterFlowStyle(t *testing.T) {
	ticket := &Ticket{
		ID:       "test-1234",
		Status:   StatusOpen,
		Deps:     []string{"dep1", "dep2"},
		Links:    []string{"link1", "link2"},
		Created:  time.Date(2025, 1, 11, 10, 0, 0, 0, time.UTC),
		Type:     TypeTask,
		Priority: 2,
		Title:    "Test",
	}

	var buf bytes.Buffer
	if err := Format(&buf, ticket); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()

	// Should use flow style: [item1, item2]
	if !strings.Contains(output, "deps: [dep1, dep2]") {
		t.Errorf("Deps should use flow style [item1, item2], got:\n%s", output)
	}
	if !strings.Contains(output, "links: [link1, link2]") {
		t.Errorf("Links should use flow style [item1, item2], got:\n%s", output)
	}

	// Should NOT use block style
	if strings.Contains(output, "deps:\n  - dep1") {
		t.Error("Deps should not use block style")
	}
	if strings.Contains(output, "links:\n  - link1") {
		t.Error("Links should not use block style")
	}
}

// TestTitleExtraction tests title extraction from markdown
func TestTitleExtraction(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedTitle string
	}{
		{
			"single heading",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# My Title
Body content`,
			"My Title",
		},
		{
			"no title",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
Body without heading`,
			"",
		},
		{
			"title with special chars",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# Title with @#$% special chars!`,
			"Title with @#$% special chars!",
		},
		{
			"multiple headings",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# First Title
Body
## Second Heading`,
			"First Title",
		},
		{
			"heading with leading spaces",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
   # Title With Spaces
Body`,
			"Title With Spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket, err := Parse(strings.NewReader(tt.content))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if ticket.Title != tt.expectedTitle {
				t.Errorf("Title = %q, expected %q", ticket.Title, tt.expectedTitle)
			}
		})
	}
}

// TestBodyContentPreservation tests that body content is preserved
func TestBodyContentPreservation(t *testing.T) {
	tests := []struct {
		name         string
		bodyContent  string
		expectedBody string
	}{
		{
			"multi-paragraph",
			`Paragraph one.

Paragraph two.

Paragraph three.`,
			"Paragraph one.\n\nParagraph two.\n\nParagraph three.",
		},
		{
			"code block",
			"Some text\n```go\nfunc main() {\n}\n```\nMore text",
			"Some text\n```go\nfunc main() {\n}\n```\nMore text",
		},
		{
			"markdown formatting",
			"**Bold** and *italic* and `code`",
			"**Bold** and *italic* and `code`",
		},
		{
			"empty body",
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := `---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# Test Title
` + tt.bodyContent

			ticket, err := Parse(strings.NewReader(content))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if ticket.Body != tt.expectedBody {
				t.Errorf("Body = %q, expected %q", ticket.Body, tt.expectedBody)
			}
		})
	}
}

// TestFrontmatterFormat tests frontmatter delimiter handling
func TestFrontmatterFormat(t *testing.T) {
	content := `---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# Test
`
	ticket, err := Parse(strings.NewReader(content))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if ticket.ID != "test-1234" {
		t.Errorf("Failed to parse with correct delimiters")
	}
}

// TestISODateFormat tests RFC3339 date parsing
func TestISODateFormat(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		expected time.Time
	}{
		{
			"RFC3339 with Z",
			"2025-01-11T10:30:45Z",
			time.Date(2025, 1, 11, 10, 30, 45, 0, time.UTC),
		},
		{
			"RFC3339 with timezone",
			"2025-01-11T10:30:45-05:00",
			time.Date(2025, 1, 11, 15, 30, 45, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := `---
id: test-1234
status: open
deps: []
links: []
created: ` + tt.dateStr + `
type: task
priority: 2
---
# Test
`
			ticket, err := Parse(strings.NewReader(content))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if !ticket.Created.Equal(tt.expected) {
				t.Errorf("Created = %v, expected %v", ticket.Created, tt.expected)
			}
		})
	}
}

// TestDateRoundTrip tests parse -> format -> parse preserves date
func TestDateRoundTrip(t *testing.T) {
	originalTime := time.Date(2025, 1, 11, 10, 30, 45, 0, time.UTC)

	ticket := &Ticket{
		ID:       "test-1234",
		Status:   StatusOpen,
		Deps:     []string{},
		Links:    []string{},
		Created:  originalTime,
		Type:     TypeTask,
		Priority: 2,
		Title:    "Test",
	}

	// Format
	var buf bytes.Buffer
	if err := Format(&buf, ticket); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Parse
	parsed, err := Parse(&buf)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !parsed.Created.Equal(originalTime) {
		t.Errorf("Round-trip failed: got %v, expected %v", parsed.Created, originalTime)
	}
}

// TestOptionalFields tests parsing without optional fields
func TestOptionalFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			"without assignee",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# Test
`,
		},
		{
			"without external-ref",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
assignee: testuser
---
# Test
`,
		},
		{
			"without parent",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
assignee: testuser
external-ref: EXT-123
---
# Test
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket, err := Parse(strings.NewReader(tt.content))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			// Should parse successfully with zero values for missing fields
			if ticket.ID != "test-1234" {
				t.Errorf("Parse failed for ticket with optional fields omitted")
			}
		})
	}
}

// TestInvalidFrontmatter tests error handling for invalid input
func TestInvalidFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			"invalid YAML syntax",
			`---
id: test-1234
status: [open: invalid
---
# Test`,
			true, // YAML parsing should fail
		},
		{
			"malformed frontmatter",
			`---
this is not: valid: yaml: syntax:::
---
# Test`,
			true, // YAML parsing should fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tt.content))
			if tt.wantErr && err == nil {
				t.Error("Expected error for invalid frontmatter, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			"no body just frontmatter",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
`,
			false,
		},
		{
			"frontmatter with trailing spaces",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# Test
`,
			false,
		},
		{
			"multiple # headings",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# First Title
Content
## Second Heading
More content`,
			false,
		},
		{
			"very long title",
			`---
id: test-1234
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# ` + strings.Repeat("Very Long Title ", 20),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFormatPreservesFields tests that Format writes all fields correctly
func TestFormatPreservesFields(t *testing.T) {
	ticket := &Ticket{
		ID:          "test-1234",
		Status:      StatusOpen,
		Deps:        []string{"dep1", "dep2"},
		Links:       []string{"link1"},
		Created:     time.Date(2025, 1, 11, 10, 0, 0, 0, time.UTC),
		Type:        TypeTask,
		Priority:    2,
		Assignee:    "testuser",
		ExternalRef: "EXT-123",
		Parent:      "parent-5678",
		Title:       "Test Ticket",
		Body:        "Test body",
	}

	var buf bytes.Buffer
	if err := Format(&buf, ticket); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()

	// Verify all fields are present
	expectedFields := []string{
		"id: test-1234",
		"status: open",
		"deps: [dep1, dep2]",
		"links: [link1]",
		"type: task",
		"priority: 2",
		"assignee: testuser",
		"external-ref: EXT-123",
		"parent: parent-5678",
		"# Test Ticket",
		"Test body",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Output missing field %q:\n%s", field, output)
		}
	}
}

// TestEmptyArraysFormat tests formatting of empty arrays
func TestEmptyArraysFormat(t *testing.T) {
	ticket := &Ticket{
		ID:       "test-1234",
		Status:   StatusOpen,
		Deps:     []string{},
		Links:    []string{},
		Created:  time.Date(2025, 1, 11, 10, 0, 0, 0, time.UTC),
		Type:     TypeTask,
		Priority: 2,
		Title:    "Test",
	}

	var buf bytes.Buffer
	if err := Format(&buf, ticket); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "deps: []") {
		t.Errorf("Empty deps should format as [], got:\n%s", output)
	}
	if !strings.Contains(output, "links: []") {
		t.Errorf("Empty links should format as [], got:\n%s", output)
	}
}

// TestParseAndFormatRoundTrip tests full round-trip
func TestParseAndFormatRoundTrip(t *testing.T) {
	original := `---
id: test-1234
status: open
deps: [dep1, dep2]
links: [link1]
created: 2025-01-11T10:00:00Z
type: task
priority: 2
assignee: testuser
---
# Test Ticket

This is the body.
`

	// Parse
	ticket, err := Parse(strings.NewReader(original))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Format
	var buf bytes.Buffer
	if err := Format(&buf, ticket); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Parse again
	ticket2, err := Parse(&buf)
	if err != nil {
		t.Fatalf("Second parse failed: %v", err)
	}

	// Compare
	if ticket.ID != ticket2.ID {
		t.Errorf("Round-trip ID mismatch: %q != %q", ticket.ID, ticket2.ID)
	}
	if ticket.Status != ticket2.Status {
		t.Errorf("Round-trip Status mismatch")
	}
	if ticket.Title != ticket2.Title {
		t.Errorf("Round-trip Title mismatch")
	}
}
