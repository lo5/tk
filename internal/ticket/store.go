package ticket

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Store defines the interface for ticket storage operations
type Store interface {
	Create(t *Ticket) error
	Get(id string) (*Ticket, error)
	List() ([]*Ticket, error)
	Update(t *Ticket) error
	Delete(id string) error
	Path(id string) (string, error)
	Dir() string
}

// FileStore implements Store using the filesystem
type FileStore struct {
	dir string
}

// NewFileStore creates a new FileStore with the given directory
func NewFileStore(dir string) *FileStore {
	return &FileStore{dir: dir}
}

// Dir returns the tickets directory
func (s *FileStore) Dir() string {
	return s.dir
}

// EnsureDir creates the tickets directory if it doesn't exist
func (s *FileStore) EnsureDir() error {
	return os.MkdirAll(s.dir, 0755)
}

// Create creates a new ticket file
func (s *FileStore) Create(t *Ticket) error {
	if err := s.EnsureDir(); err != nil {
		return fmt.Errorf("creating tickets directory: %w", err)
	}

	path := filepath.Join(s.dir, t.ID+".md")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating ticket file: %w", err)
	}
	defer f.Close()

	if err := Format(f, t); err != nil {
		return fmt.Errorf("writing ticket: %w", err)
	}

	return nil
}

// Get retrieves a ticket by ID (supports partial matching)
func (s *FileStore) Get(partial string) (*Ticket, error) {
	id, err := ResolveID(s.dir, partial)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(s.dir, id+".md")
	return s.readTicket(path)
}

// Path returns the file path for a ticket (supports partial matching)
func (s *FileStore) Path(partial string) (string, error) {
	id, err := ResolveID(s.dir, partial)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.dir, id+".md"), nil
}

// List returns all tickets
func (s *FileStore) List() ([]*Ticket, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading tickets directory: %w", err)
	}

	var tickets []*Ticket
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(s.dir, entry.Name())
		t, err := s.readTicket(path)
		if err != nil {
			// Skip malformed tickets
			continue
		}
		tickets = append(tickets, t)
	}

	return tickets, nil
}

// ListByModTime returns tickets sorted by modification time (most recent first)
func (s *FileStore) ListByModTime(limit int) ([]*Ticket, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading tickets directory: %w", err)
	}

	type fileInfo struct {
		path    string
		modTime int64
	}

	var files []fileInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(s.dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{path: path, modTime: info.ModTime().UnixNano()})
	}

	// Sort by modTime descending
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime > files[j].modTime
	})

	var tickets []*Ticket
	for _, f := range files {
		if limit > 0 && len(tickets) >= limit {
			break
		}
		t, err := s.readTicket(f.path)
		if err != nil {
			continue
		}
		tickets = append(tickets, t)
	}

	return tickets, nil
}

// Update updates an existing ticket
func (s *FileStore) Update(t *Ticket) error {
	path := filepath.Join(s.dir, t.ID+".md")

	// Write to temp file first for atomicity
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	if err := Format(f, t); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing ticket: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// UpdateField updates a single field in a ticket file, preserving original formatting
func (s *FileStore) UpdateField(partial, field, value string) (string, error) {
	id, err := ResolveID(s.dir, partial)
	if err != nil {
		return "", err
	}

	path := filepath.Join(s.dir, id+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading ticket: %w", err)
	}

	newContent := UpdateField(string(content), field, value)

	// Write to temp file first for atomicity
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("writing temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("renaming temp file: %w", err)
	}

	return id, nil
}

// Delete removes a ticket
func (s *FileStore) Delete(partial string) error {
	id, err := ResolveID(s.dir, partial)
	if err != nil {
		return err
	}

	path := filepath.Join(s.dir, id+".md")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("deleting ticket: %w", err)
	}

	return nil
}

// readTicket reads and parses a ticket from a file path
func (s *FileStore) readTicket(path string) (*Ticket, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening ticket: %w", err)
	}
	defer f.Close()

	return Parse(f)
}

// ReadRaw reads the raw content of a ticket file
func (s *FileStore) ReadRaw(partial string) (string, string, error) {
	id, err := ResolveID(s.dir, partial)
	if err != nil {
		return "", "", err
	}

	path := filepath.Join(s.dir, id+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("reading ticket: %w", err)
	}

	return id, string(content), nil
}

// WriteRaw writes raw content to a ticket file
func (s *FileStore) WriteRaw(id, content string) error {
	path := filepath.Join(s.dir, id+".md")

	// Write to temp file first for atomicity
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// AppendToFile appends content to a ticket file
func (s *FileStore) AppendToFile(partial, content string) (string, error) {
	id, err := ResolveID(s.dir, partial)
	if err != nil {
		return "", err
	}

	path := filepath.Join(s.dir, id+".md")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("opening ticket: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return "", fmt.Errorf("appending to ticket: %w", err)
	}

	return id, nil
}

// FileContains checks if a ticket file contains a string
func (s *FileStore) FileContains(partial, search string) (bool, string, error) {
	id, _, content, err := s.GetRawContent(partial)
	if err != nil {
		return false, "", err
	}
	return strings.Contains(content, search), id, nil
}

// GetRawContent gets the raw content and ID of a ticket
func (s *FileStore) GetRawContent(partial string) (string, string, string, error) {
	id, err := ResolveID(s.dir, partial)
	if err != nil {
		return "", "", "", err
	}

	path := filepath.Join(s.dir, id+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", "", fmt.Errorf("reading ticket: %w", err)
	}

	return id, path, string(content), nil
}
