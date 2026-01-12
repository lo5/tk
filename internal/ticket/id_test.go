package ticket

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// 1.2 ID Generation Tests

// TestPrefixExtraction tests prefix extraction from various directory names
func TestPrefixExtraction(t *testing.T) {
	tests := []struct {
		name           string
		dirName        string
		expectedPrefix string // prefix part only, before the hash
	}{
		{"single segment", "/path/to/myproject", "m"},
		{"hyphenated", "/path/to/my-ticket-keeper", "mtk"},
		{"underscored", "/path/to/my_ticket_keeper", "mtk"},
		{"mixed hyphen and underscore", "/path/to/my-ticket_keeper", "mtk"},
		{"single char", "/path/to/a", "a"},
		{"numeric start", "/path/to/123project", "1"},
		{"three segments", "/path/to/go-tk", "gt"},
		{"gotk", "/path/to/gotk", "g"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := GenerateID(tt.dirName)
			parts := strings.Split(id, "-")
			if len(parts) < 2 {
				t.Fatalf("Expected ID format {prefix}-{hash}, got %q", id)
			}
			prefix := parts[0]
			// All but last part is prefix (in case prefix itself contains hyphens)
			if len(parts) > 2 {
				prefix = strings.Join(parts[:len(parts)-1], "-")
			}
			if prefix != tt.expectedPrefix {
				t.Errorf("GenerateID(%q) prefix = %q, expected %q (full ID: %s)",
					filepath.Base(tt.dirName), prefix, tt.expectedPrefix, id)
			}
		})
	}
}

// TestHashGeneration tests that the hash portion of the ID is correctly formatted
func TestHashGeneration(t *testing.T) {
	tests := []struct {
		name    string
		dirName string
	}{
		{"simple", "/path/to/project"},
		{"complex", "/path/to/my-complex-project_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := GenerateID(tt.dirName)
			parts := strings.Split(id, "-")
			if len(parts) < 2 {
				t.Fatalf("Expected ID format {prefix}-{hash}, got %q", id)
			}

			// Hash is the last part
			hash := parts[len(parts)-1]

			// Hash should be exactly 4 characters
			if len(hash) != 4 {
				t.Errorf("Hash should be 4 characters, got %d: %q", len(hash), hash)
			}

			// Hash should be lowercase alphanumeric (a-z, 0-9)
			alphanumPattern := regexp.MustCompile(`^[a-z0-9]{4}$`)
			if !alphanumPattern.MatchString(hash) {
				t.Errorf("Hash should be 4 lowercase alphanumeric characters, got %q", hash)
			}
		})
	}
}

// TestFullIDFormat tests that the complete ID follows the expected format
func TestFullIDFormat(t *testing.T) {
	tests := []struct {
		dirName string
	}{
		{"/path/to/myproject"},
		{"/path/to/my-ticket-keeper"},
		{"/path/to/gotk"},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.dirName), func(t *testing.T) {
			id := GenerateID(tt.dirName)

			// Should match pattern: {prefix}-{4-alphanumeric-chars}
			pattern := regexp.MustCompile(`^[a-z0-9]+-[a-z0-9]{4}$`)
			if !pattern.MatchString(id) {
				t.Errorf("ID %q does not match expected format {prefix}-{4-char-alphanumeric}", id)
			}

			// Should contain exactly one or more hyphens (at least one separating prefix from hash)
			if !strings.Contains(id, "-") {
				t.Errorf("ID %q should contain hyphen separator", id)
			}
		})
	}
}

// TestPrefixExtractedCorrectly verifies prefix extraction logic
func TestPrefixExtractedCorrectly(t *testing.T) {
	tests := []struct {
		name           string
		dirName        string
		expectedPrefix string
	}{
		{"single word lowercase", "/path/to/myproject", "m"},
		{"single word uppercase", "/path/to/MyProject", "m"}, // Expecting lowercase
		{"two words hyphen", "/path/to/my-project", "mp"},
		{"three words hyphen", "/path/to/my-new-project", "mnp"},
		{"two words underscore", "/path/to/my_project", "mp"},
		{"three words underscore", "/path/to/my_new_project", "mnp"},
		{"mixed separators", "/path/to/my-new_project", "mnp"},
		{"single letter", "/path/to/x", "x"},
		{"numbers", "/path/to/123", "1"},
		{"word with numbers", "/path/to/proj123", "p"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := GenerateID(tt.dirName)
			parts := strings.Split(id, "-")

			// Get prefix (all parts except last which is hash)
			var prefix string
			if len(parts) == 2 {
				prefix = parts[0]
			} else {
				prefix = strings.Join(parts[:len(parts)-1], "-")
			}

			if prefix != tt.expectedPrefix {
				t.Errorf("GenerateID(%q) prefix = %q, expected %q",
					filepath.Base(tt.dirName), prefix, tt.expectedPrefix)
			}
		})
	}
}

// TestHashAppendedCorrectly verifies hash is properly appended
func TestHashAppendedCorrectly(t *testing.T) {
	dirName := "/path/to/myproject"
	id := GenerateID(dirName)

	if !strings.Contains(id, "-") {
		t.Fatalf("ID should contain hyphen separator: %q", id)
	}

	parts := strings.Split(id, "-")
	if len(parts) < 2 {
		t.Fatalf("ID should have at least prefix and hash parts: %q", id)
	}

	hash := parts[len(parts)-1]
	if len(hash) != 4 {
		t.Errorf("Hash portion should be 4 characters, got %d: %q", len(hash), hash)
	}

	// Verify hash is at the end
	if !strings.HasSuffix(id, hash) {
		t.Errorf("ID should end with hash, ID=%q, hash=%q", id, hash)
	}
}

// TestEmptyDirectoryNameFallback tests handling of edge cases
func TestEmptyDirectoryNameFallback(t *testing.T) {
	tests := []struct {
		name    string
		dirName string
	}{
		{"root directory", "/"},
		{"dot", "/."},
		{"empty path", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			id := GenerateID(tt.dirName)

			// Should still produce valid ID format
			if id == "" {
				t.Error("GenerateID should not return empty string")
			}

			if !strings.Contains(id, "-") {
				t.Errorf("ID should contain hyphen separator even for edge cases: %q", id)
			}
		})
	}
}

// TestIDFormatConsistency verifies the format is consistent across calls
func TestIDFormatConsistency(t *testing.T) {
	dirName := "/path/to/my-test-project"

	// Generate multiple IDs
	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		ids[i] = GenerateID(dirName)
	}

	// All should have the same prefix
	var commonPrefix string
	for i, id := range ids {
		parts := strings.Split(id, "-")
		if len(parts) < 2 {
			t.Fatalf("Invalid ID format: %q", id)
		}

		// Prefix is all parts except the last (hash)
		prefix := strings.Join(parts[:len(parts)-1], "-")

		if i == 0 {
			commonPrefix = prefix
		} else {
			if prefix != commonPrefix {
				t.Errorf("Inconsistent prefix: got %q, expected %q", prefix, commonPrefix)
			}
		}

		// All should have 4-char alphanumeric hash at end
		hash := parts[len(parts)-1]
		if len(hash) != 4 {
			t.Errorf("Hash should be 4 characters: %q", hash)
		}
	}
}

// TestPrefixLowercase verifies that prefix is lowercased
func TestPrefixLowercase(t *testing.T) {
	tests := []struct {
		dirName string
	}{
		{"/path/to/MyProject"},
		{"/path/to/MY-PROJECT"},
		{"/path/to/My-Ticket-Keeper"},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.dirName), func(t *testing.T) {
			id := GenerateID(tt.dirName)
			parts := strings.Split(id, "-")

			// Get prefix (all but last part)
			prefix := strings.Join(parts[:len(parts)-1], "-")

			if prefix != strings.ToLower(prefix) {
				t.Errorf("Prefix should be lowercase: got %q", prefix)
			}
		})
	}
}
