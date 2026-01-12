package ticket

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotFound indicates that a ticket was not found
type ErrNotFound struct {
	ID string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("ticket '%s' not found", e.ID)
}

// ErrAmbiguous indicates that a partial ID matches multiple tickets
type ErrAmbiguous struct {
	ID      string
	Matches []string
}

func (e ErrAmbiguous) Error() string {
	return fmt.Sprintf("ambiguous ID '%s' matches multiple tickets", e.ID)
}

// ResolveID resolves a partial ID to a full ticket ID
// It first tries exact match, then partial match
func ResolveID(ticketsDir, partial string) (string, error) {
	// Try exact match first
	exactPath := filepath.Join(ticketsDir, partial+".md")
	if _, err := os.Stat(exactPath); err == nil {
		return partial, nil
	}

	// Try partial match
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound{ID: partial}
		}
		return "", fmt.Errorf("reading tickets directory: %w", err)
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		id := strings.TrimSuffix(name, ".md")
		if strings.Contains(id, partial) {
			matches = append(matches, id)
		}
	}

	switch len(matches) {
	case 0:
		return "", ErrNotFound{ID: partial}
	case 1:
		return matches[0], nil
	default:
		return "", ErrAmbiguous{ID: partial, Matches: matches}
	}
}

// ResolveIDs resolves multiple partial IDs
func ResolveIDs(ticketsDir string, partials []string) ([]string, error) {
	result := make([]string, len(partials))
	for i, partial := range partials {
		id, err := ResolveID(ticketsDir, partial)
		if err != nil {
			return nil, err
		}
		result[i] = id
	}
	return result, nil
}
