package ticket

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 1.4 ID Resolver Tests

// setupTestTickets creates a temporary directory with test ticket files
func setupTestTickets(t *testing.T, ticketIDs []string) string {
	t.Helper()
	tmpDir := t.TempDir()
	ticketsDir := filepath.Join(tmpDir, ".tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatalf("Failed to create tickets dir: %v", err)
	}

	for _, id := range ticketIDs {
		filename := filepath.Join(ticketsDir, id+".md")
		content := `---
id: ` + id + `
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# Test
`
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write ticket file: %v", err)
		}
	}

	return ticketsDir
}

// TestResolveID_ExactMatch tests exact ID matching
func TestResolveID_ExactMatch(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"mtk-5c46", "mtk-5c47", "abc-1234"})

	tests := []struct {
		name     string
		partial  string
		expected string
	}{
		{"exact match 1", "mtk-5c46", "mtk-5c46"},
		{"exact match 2", "mtk-5c47", "mtk-5c47"},
		{"exact match 3", "abc-1234", "abc-1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveID(ticketsDir, tt.partial)
			if err != nil {
				t.Fatalf("ResolveID failed: %v", err)
			}
			if resolved != tt.expected {
				t.Errorf("ResolveID(%q) = %q, expected %q", tt.partial, resolved, tt.expected)
			}
		})
	}
}

// TestResolveID_PartialMatch_SingleMatch tests partial ID matching with single result
func TestResolveID_PartialMatch_SingleMatch(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"mtk-5c46", "abc-1234", "xyz-9999"})

	tests := []struct {
		name     string
		partial  string
		expected string
	}{
		{"hash only", "5c46", "mtk-5c46"},
		{"hash partial", "5c4", "mtk-5c46"},
		{"hash start", "5c", "mtk-5c46"},
		{"prefix only", "mtk", "mtk-5c46"},
		{"prefix and hash start", "abc-12", "abc-1234"},
		{"middle of hash", "234", "abc-1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveID(ticketsDir, tt.partial)
			if err != nil {
				t.Fatalf("ResolveID failed: %v", err)
			}
			if resolved != tt.expected {
				t.Errorf("ResolveID(%q) = %q, expected %q", tt.partial, resolved, tt.expected)
			}
		})
	}
}

// TestResolveID_PartialMatch_NoMatch tests partial ID with no matches
func TestResolveID_PartialMatch_NoMatch(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"mtk-5c46", "abc-1234"})

	tests := []struct {
		name    string
		partial string
	}{
		{"no match 1", "xyz"},
		{"no match 2", "9999"},
		{"no match 3", "nonexistent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveID(ticketsDir, tt.partial)
			if err == nil {
				t.Error("Expected error for non-existent ID, got nil")
			}

			var notFoundErr ErrNotFound
			if !errors.As(err, &notFoundErr) {
				t.Errorf("Expected ErrNotFound, got %T: %v", err, err)
			}

			if notFoundErr.ID != tt.partial {
				t.Errorf("Error ID = %q, expected %q", notFoundErr.ID, tt.partial)
			}

			// Error message should include the partial ID
			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.partial) {
				t.Errorf("Error message should include partial ID %q: %s", tt.partial, errMsg)
			}
		})
	}
}

// TestResolveID_PartialMatch_Ambiguous tests partial ID matching multiple tickets
func TestResolveID_PartialMatch_Ambiguous(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"mtk-5c46", "mtk-5c47", "abc-5c48"})

	tests := []struct {
		name    string
		partial string
	}{
		{"ambiguous prefix", "mtk"},    // Matches mtk-5c46 and mtk-5c47
		{"ambiguous hash start", "5c"}, // Matches all three
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveID(ticketsDir, tt.partial)
			if err == nil {
				t.Error("Expected error for ambiguous ID, got nil")
			}

			var ambigErr ErrAmbiguous
			if !errors.As(err, &ambigErr) {
				t.Errorf("Expected ErrAmbiguous, got %T: %v", err, err)
			}

			if ambigErr.ID != tt.partial {
				t.Errorf("Error ID = %q, expected %q", ambigErr.ID, tt.partial)
			}

			if len(ambigErr.Matches) < 2 {
				t.Errorf("Expected at least 2 matches, got %d: %v", len(ambigErr.Matches), ambigErr.Matches)
			}

			// Error message should mention it's ambiguous
			errMsg := err.Error()
			if !strings.Contains(errMsg, "ambiguous") {
				t.Errorf("Error message should contain 'ambiguous': %s", errMsg)
			}
		})
	}
}

// TestResolveID_CaseSensitivity tests case-sensitive matching
func TestResolveID_CaseSensitivity(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"mtk-5c46", "abc-1234"})

	tests := []struct {
		name      string
		partial   string
		shouldErr bool
	}{
		{"lowercase matches", "mtk", false},
		{"uppercase no match", "MTK", true},
		{"mixed case no match", "Mtk", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveID(ticketsDir, tt.partial)
			if tt.shouldErr && err == nil {
				t.Error("Expected error for case mismatch, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
		})
	}
}

// TestResolveID_SpecialCharacters tests IDs with hyphens
func TestResolveID_SpecialCharacters(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"my-tk-5c46", "x-y-z-1234"})

	tests := []struct {
		name     string
		partial  string
		expected string
	}{
		{"hyphenated ID exact", "my-tk-5c46", "my-tk-5c46"},
		{"hyphenated ID partial", "my-tk", "my-tk-5c46"},
		{"multiple hyphens", "x-y-z", "x-y-z-1234"},
		{"hash from hyphenated", "5c46", "my-tk-5c46"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveID(ticketsDir, tt.partial)
			if err != nil {
				t.Fatalf("ResolveID failed: %v", err)
			}
			if resolved != tt.expected {
				t.Errorf("ResolveID(%q) = %q, expected %q", tt.partial, resolved, tt.expected)
			}
		})
	}
}

// TestResolveID_EmptyDirectory tests handling of empty tickets directory
func TestResolveID_EmptyDirectory(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{}) // Empty directory

	_, err := ResolveID(ticketsDir, "nonexistent")
	if err == nil {
		t.Error("Expected error for empty directory, got nil")
	}

	var notFoundErr ErrNotFound
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected ErrNotFound, got %T: %v", err, err)
	}
}

// TestResolveID_MissingDirectory tests handling of missing tickets directory
func TestResolveID_MissingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	ticketsDir := filepath.Join(tmpDir, "nonexistent")

	_, err := ResolveID(ticketsDir, "test")
	if err == nil {
		t.Error("Expected error for missing directory, got nil")
	}

	var notFoundErr ErrNotFound
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected ErrNotFound for missing directory, got %T: %v", err, err)
	}
}

// TestResolveID_WildcardMatching tests that partial matching works anywhere in filename
func TestResolveID_WildcardMatching(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"mtk-5c46", "abc-1234"})

	tests := []struct {
		name     string
		partial  string
		expected string
	}{
		{"match at start", "mtk", "mtk-5c46"},
		{"match in middle", "tk-5", "mtk-5c46"},
		{"match at end", "c46", "mtk-5c46"},
		{"match whole hash", "5c46", "mtk-5c46"},
		{"match partial hash", "5c4", "mtk-5c46"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveID(ticketsDir, tt.partial)
			if err != nil {
				t.Fatalf("ResolveID failed: %v", err)
			}
			if resolved != tt.expected {
				t.Errorf("ResolveID(%q) = %q, expected %q", tt.partial, resolved, tt.expected)
			}
		})
	}
}

// TestResolveID_OnlyMarkdownFiles tests that only .md files are considered
func TestResolveID_OnlyMarkdownFiles(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"mtk-5c46"})

	// Create a non-.md file
	nonMdFile := filepath.Join(ticketsDir, "mtk-9999.txt")
	if err := os.WriteFile(nonMdFile, []byte("not a ticket"), 0644); err != nil {
		t.Fatalf("Failed to write non-.md file: %v", err)
	}

	// Should only find the .md file
	resolved, err := ResolveID(ticketsDir, "mtk")
	if err != nil {
		t.Fatalf("ResolveID failed: %v", err)
	}
	if resolved != "mtk-5c46" {
		t.Errorf("Should match only .md file, got %q", resolved)
	}

	// Non-.md file should not be matched
	_, err = ResolveID(ticketsDir, "9999")
	if err == nil {
		t.Error("Should not match non-.md file")
	}
}

// TestResolveIDs tests resolving multiple IDs
func TestResolveIDs(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"mtk-5c46", "abc-1234", "xyz-9999"})

	partials := []string{"mtk", "abc-1234", "999"}
	expected := []string{"mtk-5c46", "abc-1234", "xyz-9999"}

	resolved, err := ResolveIDs(ticketsDir, partials)
	if err != nil {
		t.Fatalf("ResolveIDs failed: %v", err)
	}

	if len(resolved) != len(expected) {
		t.Fatalf("ResolveIDs returned %d results, expected %d", len(resolved), len(expected))
	}

	for i, exp := range expected {
		if resolved[i] != exp {
			t.Errorf("ResolveIDs[%d] = %q, expected %q", i, resolved[i], exp)
		}
	}
}

// TestResolveIDs_Error tests that ResolveIDs fails if any ID fails
func TestResolveIDs_Error(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"mtk-5c46", "abc-1234"})

	partials := []string{"mtk", "nonexistent", "abc"}

	_, err := ResolveIDs(ticketsDir, partials)
	if err == nil {
		t.Error("Expected error when one ID fails to resolve, got nil")
	}
}

// TestResolveIDs_Order tests that order is preserved
func TestResolveIDs_Order(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"aaa-1111", "bbb-2222", "ccc-3333"})

	partials := []string{"ccc", "aaa", "bbb"}
	expected := []string{"ccc-3333", "aaa-1111", "bbb-2222"}

	resolved, err := ResolveIDs(ticketsDir, partials)
	if err != nil {
		t.Fatalf("ResolveIDs failed: %v", err)
	}

	for i, exp := range expected {
		if resolved[i] != exp {
			t.Errorf("ResolveIDs[%d] = %q, expected %q (order should be preserved)", i, resolved[i], exp)
		}
	}
}

// TestErrNotFound_Error tests the error message format
func TestErrNotFound_Error(t *testing.T) {
	err := ErrNotFound{ID: "test-123"}
	msg := err.Error()

	if !strings.Contains(msg, "test-123") {
		t.Errorf("Error message should contain ID: %s", msg)
	}
	if !strings.Contains(msg, "not found") {
		t.Errorf("Error message should contain 'not found': %s", msg)
	}
}

// TestErrAmbiguous_Error tests the ambiguous error message format
func TestErrAmbiguous_Error(t *testing.T) {
	err := ErrAmbiguous{
		ID:      "test",
		Matches: []string{"test-123", "test-456"},
	}
	msg := err.Error()

	if !strings.Contains(msg, "test") {
		t.Errorf("Error message should contain partial ID: %s", msg)
	}
	if !strings.Contains(msg, "ambiguous") {
		t.Errorf("Error message should contain 'ambiguous': %s", msg)
	}
}

// TestResolveID_SubstringMatch tests that substring matching works correctly
func TestResolveID_SubstringMatch(t *testing.T) {
	ticketsDir := setupTestTickets(t, []string{"prefix-abc123-suffix"})

	tests := []struct {
		partial  string
		expected string
	}{
		{"prefix", "prefix-abc123-suffix"},
		{"abc", "prefix-abc123-suffix"},
		{"123", "prefix-abc123-suffix"},
		{"suffix", "prefix-abc123-suffix"},
		{"abc123", "prefix-abc123-suffix"},
	}

	for _, tt := range tests {
		t.Run(tt.partial, func(t *testing.T) {
			resolved, err := ResolveID(ticketsDir, tt.partial)
			if err != nil {
				t.Fatalf("ResolveID(%q) failed: %v", tt.partial, err)
			}
			if resolved != tt.expected {
				t.Errorf("ResolveID(%q) = %q, expected %q", tt.partial, resolved, tt.expected)
			}
		})
	}
}
