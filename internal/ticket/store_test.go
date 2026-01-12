package ticket

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// newTestStore creates a FileStore with a temporary directory for testing
func newTestStore(t *testing.T) (*FileStore, string) {
	t.Helper()
	tempDir := t.TempDir()
	ticketsDir := filepath.Join(tempDir, ".tickets")
	return NewFileStore(ticketsDir), ticketsDir
}

// createTestTicket creates a test ticket with the given ID and returns it
func createTestTicket(id string) *Ticket {
	return &Ticket{
		ID:       id,
		Status:   StatusOpen,
		Type:     TypeTask,
		Priority: 2,
		Created:  time.Now(),
		Assignee: "testuser",
		Title:    "Test Ticket",
		Body:     "This is a test ticket body.",
		Deps:     []string{},
		Links:    []string{},
	}
}

// TestFileStore_Create tests the Create method
func TestFileStore_Create(t *testing.T) {
	t.Run("create valid ticket", func(t *testing.T) {
		store, dir := newTestStore(t)
		ticket := createTestTicket("test-1234")

		err := store.Create(ticket)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Verify file was written
		path := filepath.Join(dir, "test-1234.md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("ticket file was not created at %s", path)
		}

		// Verify file contains correct content
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading ticket file: %v", err)
		}

		contentStr := string(content)
		if len(contentStr) == 0 {
			t.Error("ticket file is empty")
		}

		// Check for frontmatter delimiters
		if !contains(contentStr, "---") {
			t.Error("ticket file missing YAML frontmatter delimiters")
		}

		// Check for ID in frontmatter
		if !contains(contentStr, "id: test-1234") {
			t.Error("ticket file missing ID in frontmatter")
		}
	})

	t.Run("create ensures directory exists", func(t *testing.T) {
		store, dir := newTestStore(t)

		// Remove the directory
		os.RemoveAll(dir)

		ticket := createTestTicket("test-5678")
		err := store.Create(ticket)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Verify directory was created
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("tickets directory was not created")
		}
	})

	t.Run("file permissions readable", func(t *testing.T) {
		store, dir := newTestStore(t)
		ticket := createTestTicket("test-9012")

		err := store.Create(ticket)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		path := filepath.Join(dir, "test-9012.md")
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat ticket file: %v", err)
		}

		// Check file is readable (at least by owner)
		mode := info.Mode()
		if mode&0400 == 0 {
			t.Errorf("ticket file is not readable by owner: %v", mode)
		}
	})
}

// TestFileStore_Get tests the Get method
func TestFileStore_Get(t *testing.T) {
	t.Run("get existing ticket by full ID", func(t *testing.T) {
		store, _ := newTestStore(t)
		original := createTestTicket("test-1234")
		if err := store.Create(original); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		ticket, err := store.Get("test-1234")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if ticket.ID != "test-1234" {
			t.Errorf("Get() ID = %v, want %v", ticket.ID, "test-1234")
		}
		if ticket.Status != StatusOpen {
			t.Errorf("Get() Status = %v, want %v", ticket.Status, StatusOpen)
		}
	})

	t.Run("get existing ticket by partial ID", func(t *testing.T) {
		store, _ := newTestStore(t)
		original := createTestTicket("abc-1234")
		if err := store.Create(original); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Test partial match with hash only
		ticket, err := store.Get("1234")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if ticket.ID != "abc-1234" {
			t.Errorf("Get() ID = %v, want %v", ticket.ID, "abc-1234")
		}

		// Test partial match with prefix
		ticket, err = store.Get("abc")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if ticket.ID != "abc-1234" {
			t.Errorf("Get() ID = %v, want %v", ticket.ID, "abc-1234")
		}
	})

	t.Run("get non-existent ticket", func(t *testing.T) {
		store, _ := newTestStore(t)

		_, err := store.Get("nonexistent")
		if err == nil {
			t.Fatal("Get() expected error for non-existent ticket, got nil")
		}
	})

	t.Run("returned ticket data matches file", func(t *testing.T) {
		store, _ := newTestStore(t)
		original := &Ticket{
			ID:          "match-1234",
			Status:      StatusInProgress,
			Type:        TypeBug,
			Priority:    1,
			Created:     time.Date(2025, 1, 11, 10, 0, 0, 0, time.UTC),
			Assignee:    "alice",
			ExternalRef: "BUG-123",
			Parent:      "epic-5678",
			Title:       "Bug Fix",
			Body:        "Fix the critical bug",
			Deps:        []string{"dep-1"},
			Links:       []string{"link-1"},
		}
		if err := store.Create(original); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		ticket, err := store.Get("match-1234")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if ticket.ID != original.ID {
			t.Errorf("ID = %v, want %v", ticket.ID, original.ID)
		}
		if ticket.Status != original.Status {
			t.Errorf("Status = %v, want %v", ticket.Status, original.Status)
		}
		if ticket.Type != original.Type {
			t.Errorf("Type = %v, want %v", ticket.Type, original.Type)
		}
		if ticket.Priority != original.Priority {
			t.Errorf("Priority = %v, want %v", ticket.Priority, original.Priority)
		}
		if ticket.Assignee != original.Assignee {
			t.Errorf("Assignee = %v, want %v", ticket.Assignee, original.Assignee)
		}
		if ticket.ExternalRef != original.ExternalRef {
			t.Errorf("ExternalRef = %v, want %v", ticket.ExternalRef, original.ExternalRef)
		}
		if ticket.Parent != original.Parent {
			t.Errorf("Parent = %v, want %v", ticket.Parent, original.Parent)
		}
		if ticket.Title != original.Title {
			t.Errorf("Title = %v, want %v", ticket.Title, original.Title)
		}
		if len(ticket.Deps) != len(original.Deps) {
			t.Errorf("Deps length = %v, want %v", len(ticket.Deps), len(original.Deps))
		}
		if len(ticket.Links) != len(original.Links) {
			t.Errorf("Links length = %v, want %v", len(ticket.Links), len(original.Links))
		}
	})

	t.Run("handles missing .tickets directory", func(t *testing.T) {
		store, dir := newTestStore(t)

		// Remove the directory
		os.RemoveAll(dir)

		_, err := store.Get("test-1234")
		if err == nil {
			t.Fatal("Get() expected error for missing directory, got nil")
		}
	})
}

// TestFileStore_List tests the List method
func TestFileStore_List(t *testing.T) {
	t.Run("list all tickets", func(t *testing.T) {
		store, _ := newTestStore(t)

		// Create multiple tickets
		for i := 1; i <= 3; i++ {
			ticket := createTestTicket(sprintf("test-%d", i))
			if err := store.Create(ticket); err != nil {
				t.Fatalf("Create() error = %v", err)
			}
		}

		tickets, err := store.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(tickets) != 3 {
			t.Errorf("List() returned %d tickets, want 3", len(tickets))
		}
	})

	t.Run("empty directory returns empty list", func(t *testing.T) {
		store, _ := newTestStore(t)

		// Ensure directory exists but is empty
		if err := store.EnsureDir(); err != nil {
			t.Fatalf("EnsureDir() error = %v", err)
		}

		tickets, err := store.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(tickets) != 0 {
			t.Errorf("List() returned %d tickets, want 0", len(tickets))
		}
	})

	t.Run("only .md files included", func(t *testing.T) {
		store, dir := newTestStore(t)

		// Create a valid ticket
		ticket := createTestTicket("valid-1234")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Create non-.md files
		os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*"), 0644)

		tickets, err := store.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(tickets) != 1 {
			t.Errorf("List() returned %d tickets, want 1", len(tickets))
		}
	})

	t.Run("missing directory returns nil", func(t *testing.T) {
		store, dir := newTestStore(t)

		// Remove the directory
		os.RemoveAll(dir)

		tickets, err := store.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if tickets != nil && len(tickets) != 0 {
			t.Errorf("List() returned %d tickets, want 0 or nil", len(tickets))
		}
	})
}

// TestFileStore_Update tests the Update method
func TestFileStore_Update(t *testing.T) {
	t.Run("update existing ticket", func(t *testing.T) {
		store, _ := newTestStore(t)

		// Create original ticket
		ticket := createTestTicket("update-1234")
		ticket.Status = StatusOpen
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Update the ticket
		ticket.Status = StatusInProgress
		ticket.Title = "Updated Title"
		if err := store.Update(ticket); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		// Verify update
		updated, err := store.Get("update-1234")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if updated.Status != StatusInProgress {
			t.Errorf("Status = %v, want %v", updated.Status, StatusInProgress)
		}
		if updated.Title != "Updated Title" {
			t.Errorf("Title = %v, want %v", updated.Title, "Updated Title")
		}
	})

	t.Run("atomic write uses temp file", func(t *testing.T) {
		store, dir := newTestStore(t)

		ticket := createTestTicket("atomic-1234")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// During update, temp file should be created and then renamed
		ticket.Status = StatusClosed
		if err := store.Update(ticket); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		// Verify no .tmp file remains
		tmpPath := filepath.Join(dir, "atomic-1234.md.tmp")
		if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
			t.Errorf("temp file still exists at %s", tmpPath)
		}

		// Verify final file exists
		finalPath := filepath.Join(dir, "atomic-1234.md")
		if _, err := os.Stat(finalPath); os.IsNotExist(err) {
			t.Errorf("final file does not exist at %s", finalPath)
		}
	})
}

// TestFileStore_Delete tests the Delete method
func TestFileStore_Delete(t *testing.T) {
	t.Run("delete existing ticket", func(t *testing.T) {
		store, dir := newTestStore(t)

		ticket := createTestTicket("delete-1234")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Verify file exists
		path := filepath.Join(dir, "delete-1234.md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatal("ticket file does not exist before delete")
		}

		// Delete
		if err := store.Delete("delete-1234"); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify file removed
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("ticket file still exists after delete")
		}
	})

	t.Run("delete with partial ID", func(t *testing.T) {
		store, dir := newTestStore(t)

		ticket := createTestTicket("xyz-5678")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Delete using partial ID
		if err := store.Delete("5678"); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify file removed
		path := filepath.Join(dir, "xyz-5678.md")
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("ticket file still exists after delete")
		}
	})

	t.Run("delete non-existent ticket", func(t *testing.T) {
		store, _ := newTestStore(t)

		err := store.Delete("nonexistent")
		if err == nil {
			t.Fatal("Delete() expected error for non-existent ticket, got nil")
		}
	})
}

// TestFileStore_ListByModTime tests the ListByModTime method
func TestFileStore_ListByModTime(t *testing.T) {
	t.Run("sorted by modification time descending", func(t *testing.T) {
		store, dir := newTestStore(t)

		// Create tickets with different mod times
		ticket1 := createTestTicket("old-1111")
		if err := store.Create(ticket1); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Sleep to ensure different mod times
		time.Sleep(10 * time.Millisecond)

		ticket2 := createTestTicket("new-2222")
		if err := store.Create(ticket2); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Manually set older mod time on first ticket
		oldTime := time.Now().Add(-1 * time.Hour)
		os.Chtimes(filepath.Join(dir, "old-1111.md"), oldTime, oldTime)

		tickets, err := store.ListByModTime(10)
		if err != nil {
			t.Fatalf("ListByModTime() error = %v", err)
		}

		if len(tickets) != 2 {
			t.Fatalf("ListByModTime() returned %d tickets, want 2", len(tickets))
		}

		// Most recent should be first
		if tickets[0].ID != "new-2222" {
			t.Errorf("First ticket ID = %v, want %v", tickets[0].ID, "new-2222")
		}
		if tickets[1].ID != "old-1111" {
			t.Errorf("Second ticket ID = %v, want %v", tickets[1].ID, "old-1111")
		}
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		store, _ := newTestStore(t)

		// Create 5 tickets
		for i := 1; i <= 5; i++ {
			ticket := createTestTicket(sprintf("limit-%d", i))
			if err := store.Create(ticket); err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			time.Sleep(2 * time.Millisecond)
		}

		// Request only 3
		tickets, err := store.ListByModTime(3)
		if err != nil {
			t.Fatalf("ListByModTime() error = %v", err)
		}

		if len(tickets) != 3 {
			t.Errorf("ListByModTime(3) returned %d tickets, want 3", len(tickets))
		}
	})

	t.Run("zero limit returns all", func(t *testing.T) {
		store, _ := newTestStore(t)

		// Create 3 tickets
		for i := 1; i <= 3; i++ {
			ticket := createTestTicket(sprintf("zero-%d", i))
			if err := store.Create(ticket); err != nil {
				t.Fatalf("Create() error = %v", err)
			}
		}

		tickets, err := store.ListByModTime(0)
		if err != nil {
			t.Fatalf("ListByModTime() error = %v", err)
		}

		if len(tickets) != 3 {
			t.Errorf("ListByModTime(0) returned %d tickets, want 3", len(tickets))
		}
	})

	t.Run("missing directory returns nil", func(t *testing.T) {
		store, dir := newTestStore(t)

		// Remove the directory
		os.RemoveAll(dir)

		tickets, err := store.ListByModTime(10)
		if err != nil {
			t.Fatalf("ListByModTime() error = %v", err)
		}

		if tickets != nil && len(tickets) != 0 {
			t.Errorf("ListByModTime() returned %d tickets, want 0 or nil", len(tickets))
		}
	})
}

// TestFileStore_Path tests the Path method
func TestFileStore_Path(t *testing.T) {
	t.Run("returns correct path for existing ticket", func(t *testing.T) {
		store, dir := newTestStore(t)

		ticket := createTestTicket("path-1234")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		path, err := store.Path("path-1234")
		if err != nil {
			t.Fatalf("Path() error = %v", err)
		}

		expected := filepath.Join(dir, "path-1234.md")
		if path != expected {
			t.Errorf("Path() = %v, want %v", path, expected)
		}
	})

	t.Run("resolves partial ID", func(t *testing.T) {
		store, dir := newTestStore(t)

		ticket := createTestTicket("partial-5678")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		path, err := store.Path("5678")
		if err != nil {
			t.Fatalf("Path() error = %v", err)
		}

		expected := filepath.Join(dir, "partial-5678.md")
		if path != expected {
			t.Errorf("Path() = %v, want %v", path, expected)
		}
	})
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func sprintf(format string, a ...interface{}) string {
	// Simple sprintf implementation for test IDs
	result := format
	for i, arg := range a {
		if i == 0 {
			switch v := arg.(type) {
			case int:
				result = replaceFirst(result, "%d", intToString(v))
			case string:
				result = replaceFirst(result, "%s", v)
			}
		}
	}
	return result
}

func replaceFirst(s, old, new string) string {
	idx := indexOf(s, old)
	if idx < 0 {
		return s
	}
	return s[:idx] + new + s[idx+len(old):]
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

// TestFileStore_UpdateField tests the UpdateField method
func TestFileStore_UpdateField(t *testing.T) {
	t.Run("update single field preserves formatting", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("field-1234")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Update status field
		id, err := store.UpdateField("field-1234", "status", "in_progress")
		if err != nil {
			t.Fatalf("UpdateField() error = %v", err)
		}

		if id != "field-1234" {
			t.Errorf("UpdateField() returned id = %v, want %v", id, "field-1234")
		}

		// Verify update
		updated, err := store.Get("field-1234")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if updated.Status != StatusInProgress {
			t.Errorf("Status = %v, want %v", updated.Status, StatusInProgress)
		}
	})

	t.Run("update with partial ID", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("abc-9999")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Update using partial ID
		id, err := store.UpdateField("9999", "status", "closed")
		if err != nil {
			t.Fatalf("UpdateField() error = %v", err)
		}

		if id != "abc-9999" {
			t.Errorf("UpdateField() returned id = %v, want %v", id, "abc-9999")
		}
	})

	t.Run("update non-existent ticket", func(t *testing.T) {
		store, _ := newTestStore(t)

		_, err := store.UpdateField("nonexistent", "status", "closed")
		if err == nil {
			t.Fatal("UpdateField() expected error for non-existent ticket, got nil")
		}
	})

	t.Run("atomic write with temp file", func(t *testing.T) {
		store, dir := newTestStore(t)

		ticket := createTestTicket("atomic-field")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Update field
		_, err := store.UpdateField("atomic-field", "priority", "1")
		if err != nil {
			t.Fatalf("UpdateField() error = %v", err)
		}

		// Verify no .tmp file remains
		tmpPath := filepath.Join(dir, "atomic-field.md.tmp")
		if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
			t.Errorf("temp file still exists at %s", tmpPath)
		}
	})
}

// TestFileStore_ReadRaw tests the ReadRaw method
func TestFileStore_ReadRaw(t *testing.T) {
	t.Run("read raw content", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("raw-1234")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		id, content, err := store.ReadRaw("raw-1234")
		if err != nil {
			t.Fatalf("ReadRaw() error = %v", err)
		}

		if id != "raw-1234" {
			t.Errorf("ReadRaw() id = %v, want %v", id, "raw-1234")
		}

		if !contains(content, "---") {
			t.Error("ReadRaw() content missing YAML frontmatter")
		}

		if !contains(content, "id: raw-1234") {
			t.Error("ReadRaw() content missing ticket ID")
		}
	})

	t.Run("read with partial ID", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("xyz-5555")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		id, _, err := store.ReadRaw("5555")
		if err != nil {
			t.Fatalf("ReadRaw() error = %v", err)
		}

		if id != "xyz-5555" {
			t.Errorf("ReadRaw() id = %v, want %v", id, "xyz-5555")
		}
	})
}

// TestFileStore_WriteRaw tests the WriteRaw method
func TestFileStore_WriteRaw(t *testing.T) {
	t.Run("write raw content", func(t *testing.T) {
		store, dir := newTestStore(t)

		// Ensure directory exists
		if err := store.EnsureDir(); err != nil {
			t.Fatalf("EnsureDir() error = %v", err)
		}

		content := `---
id: raw-write
status: open
deps: []
links: []
created: 2025-01-11T10:00:00Z
type: task
priority: 2
---
# Raw Write Test

This is raw content.`

		err := store.WriteRaw("raw-write", content)
		if err != nil {
			t.Fatalf("WriteRaw() error = %v", err)
		}

		// Verify file was written
		path := filepath.Join(dir, "raw-write.md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("ticket file was not created at %s", path)
		}

		// Read back and verify
		readContent, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading file: %v", err)
		}

		if string(readContent) != content {
			t.Errorf("WriteRaw() content mismatch")
		}
	})

	t.Run("atomic write with temp file", func(t *testing.T) {
		store, dir := newTestStore(t)

		if err := store.EnsureDir(); err != nil {
			t.Fatalf("EnsureDir() error = %v", err)
		}

		content := "---\nid: atomic\n---\n# Test"
		err := store.WriteRaw("atomic", content)
		if err != nil {
			t.Fatalf("WriteRaw() error = %v", err)
		}

		// Verify no .tmp file remains
		tmpPath := filepath.Join(dir, "atomic.md.tmp")
		if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
			t.Errorf("temp file still exists at %s", tmpPath)
		}
	})
}

// TestFileStore_AppendToFile tests the AppendToFile method
func TestFileStore_AppendToFile(t *testing.T) {
	t.Run("append content to existing file", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("append-1234")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Append content
		appendText := "\n\n## Notes\n\nThis is a note."
		id, err := store.AppendToFile("append-1234", appendText)
		if err != nil {
			t.Fatalf("AppendToFile() error = %v", err)
		}

		if id != "append-1234" {
			t.Errorf("AppendToFile() returned id = %v, want %v", id, "append-1234")
		}

		// Verify content was appended
		_, content, err := store.ReadRaw("append-1234")
		if err != nil {
			t.Fatalf("ReadRaw() error = %v", err)
		}

		if !contains(content, "This is a note.") {
			t.Error("AppendToFile() content was not appended")
		}
	})

	t.Run("append with partial ID", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("xyz-append")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		id, err := store.AppendToFile("append", "\n\nAppended text")
		if err != nil {
			t.Fatalf("AppendToFile() error = %v", err)
		}

		if id != "xyz-append" {
			t.Errorf("AppendToFile() returned id = %v, want %v", id, "xyz-append")
		}
	})
}

// TestFileStore_FileContains tests the FileContains method
func TestFileStore_FileContains(t *testing.T) {
	t.Run("file contains string", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("contains-1234")
		ticket.Body = "This ticket contains special text."
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		contains, id, err := store.FileContains("contains-1234", "special text")
		if err != nil {
			t.Fatalf("FileContains() error = %v", err)
		}

		if !contains {
			t.Error("FileContains() = false, want true")
		}

		if id != "contains-1234" {
			t.Errorf("FileContains() returned id = %v, want %v", id, "contains-1234")
		}
	})

	t.Run("file does not contain string", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("nocontain-1234")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		contains, _, err := store.FileContains("nocontain-1234", "nonexistent text")
		if err != nil {
			t.Fatalf("FileContains() error = %v", err)
		}

		if contains {
			t.Error("FileContains() = true, want false")
		}
	})
}

// TestFileStore_GetRawContent tests the GetRawContent method
func TestFileStore_GetRawContent(t *testing.T) {
	t.Run("get raw content with path", func(t *testing.T) {
		store, dir := newTestStore(t)

		ticket := createTestTicket("rawcontent-1234")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		id, path, content, err := store.GetRawContent("rawcontent-1234")
		if err != nil {
			t.Fatalf("GetRawContent() error = %v", err)
		}

		if id != "rawcontent-1234" {
			t.Errorf("GetRawContent() id = %v, want %v", id, "rawcontent-1234")
		}

		expectedPath := filepath.Join(dir, "rawcontent-1234.md")
		if path != expectedPath {
			t.Errorf("GetRawContent() path = %v, want %v", path, expectedPath)
		}

		if !contains(content, "id: rawcontent-1234") {
			t.Error("GetRawContent() content missing ticket ID")
		}
	})
}

// TestFileStore_ConcurrentAccess tests concurrent access to the store
func TestFileStore_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent reads", func(t *testing.T) {
		store, _ := newTestStore(t)

		// Create a ticket
		ticket := createTestTicket("concurrent-1234")
		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Perform concurrent reads
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				_, err := store.Get("concurrent-1234")
				if err != nil {
					t.Errorf("Get() error = %v", err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent list operations", func(t *testing.T) {
		store, _ := newTestStore(t)

		// Create multiple tickets
		for i := 1; i <= 5; i++ {
			ticket := createTestTicket(sprintf("list-%d", i))
			if err := store.Create(ticket); err != nil {
				t.Fatalf("Create() error = %v", err)
			}
		}

		// Perform concurrent list operations
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				_, err := store.List()
				if err != nil {
					t.Errorf("List() error = %v", err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("create while reading", func(t *testing.T) {
		store, _ := newTestStore(t)

		// Create initial tickets
		for i := 1; i <= 3; i++ {
			ticket := createTestTicket(sprintf("init-%d", i))
			if err := store.Create(ticket); err != nil {
				t.Fatalf("Create() error = %v", err)
			}
		}

		done := make(chan bool)
		errors := make(chan error, 20)

		// Start readers
		for i := 0; i < 10; i++ {
			go func(index int) {
				for j := 0; j < 5; j++ {
					_, err := store.List()
					if err != nil {
						errors <- err
					}
				}
				done <- true
			}(i)
		}

		// Start writers
		for i := 0; i < 10; i++ {
			go func(index int) {
				ticket := createTestTicket(sprintf("new-%d", index))
				if err := store.Create(ticket); err != nil {
					// It's ok if some creates fail due to duplicate IDs
					// Just check that we don't get unexpected errors
					if !contains(err.Error(), "file exists") {
						errors <- err
					}
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 20; i++ {
			<-done
		}

		close(errors)
		for err := range errors {
			t.Errorf("Concurrent operation error: %v", err)
		}
	})
}

// TestFileStore_EdgeCases tests edge cases
func TestFileStore_EdgeCases(t *testing.T) {
	t.Run("ticket with empty deps and links", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("empty-arrays")
		ticket.Deps = []string{}
		ticket.Links = []string{}

		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		retrieved, err := store.Get("empty-arrays")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if retrieved.Deps == nil {
			t.Error("Deps should be empty array, not nil")
		}
		if retrieved.Links == nil {
			t.Error("Links should be empty array, not nil")
		}
	})

	t.Run("ticket with special characters in title", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("special-1234")
		ticket.Title = "Test: Special & \"Chars\" <>"

		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		retrieved, err := store.Get("special-1234")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if retrieved.Title != ticket.Title {
			t.Errorf("Title = %v, want %v", retrieved.Title, ticket.Title)
		}
	})

	t.Run("ticket with very long body", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("long-body")
		longBody := ""
		for i := 0; i < 1000; i++ {
			longBody += "This is a very long body text. "
		}
		ticket.Body = longBody

		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		retrieved, err := store.Get("long-body")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		// Body might have trailing whitespace trimmed, so check it's close
		if len(retrieved.Body) < len(ticket.Body)-10 || len(retrieved.Body) > len(ticket.Body) {
			t.Errorf("Body length = %v, want approximately %v", len(retrieved.Body), len(ticket.Body))
		}

		// Verify content is preserved
		if !contains(retrieved.Body, "This is a very long body text.") {
			t.Error("Body content was not preserved")
		}
	})

	t.Run("ticket with multiline body", func(t *testing.T) {
		store, _ := newTestStore(t)

		ticket := createTestTicket("multiline")
		ticket.Body = "Line 1\n\nLine 2\n\n## Section\n\nMore text"

		if err := store.Create(ticket); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		retrieved, err := store.Get("multiline")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if retrieved.Body != ticket.Body {
			t.Errorf("Body mismatch:\ngot: %q\nwant: %q", retrieved.Body, ticket.Body)
		}
	})
}
