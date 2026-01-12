package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestQueryCommand tests the query command
func TestQueryCommand(t *testing.T) {
	t.Run("JSON output with no filter", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create test tickets
		id1, _ := ctx.exec("new", "First Ticket", "--priority", "0")
		id1 = strings.TrimSpace(id1)
		id2, _ := ctx.exec("new", "Second Ticket", "--priority", "2", "--type", "bug")
		id2 = strings.TrimSpace(id2)

		// Query all tickets
		output, err := ctx.exec("query")
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Should have 2 lines of JSON
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 2 {
			t.Errorf("expected 2 JSON lines, got %d", len(lines))
		}

		// Each line should be valid JSON
		for i, line := range lines {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(line), &result); err != nil {
				t.Errorf("line %d is not valid JSON: %v", i, err)
			}
		}
	})

	t.Run("frontmatter to JSON mapping", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create ticket with all fields
		id, _ := ctx.exec("new", "Full Ticket",
			"--priority", "1",
			"--type", "feature",
			"--assignee", "alice",
			"--external-ref", "FEAT-123",
		)
		id = strings.TrimSpace(id)

		// Query
		output, err := ctx.exec("query")
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Parse JSON
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		// Verify all fields are present
		if result["id"] != id {
			t.Errorf("id = %v, want %v", result["id"], id)
		}
		if result["status"] != "open" {
			t.Errorf("status = %v, want open", result["status"])
		}
		if result["type"] != "feature" {
			t.Errorf("type = %v, want feature", result["type"])
		}
		if result["priority"] != "1" {
			t.Errorf("priority = %v, want 1 (as string)", result["priority"])
		}
		if result["assignee"] != "alice" {
			t.Errorf("assignee = %v, want alice", result["assignee"])
		}
		if result["external-ref"] != "FEAT-123" {
			t.Errorf("external-ref = %v, want FEAT-123", result["external-ref"])
		}

		// Verify arrays are present (even if empty)
		if _, ok := result["deps"]; !ok {
			t.Error("deps field missing")
		}
		if _, ok := result["links"]; !ok {
			t.Error("links field missing")
		}

		// Verify created timestamp is present
		if _, ok := result["created"]; !ok {
			t.Error("created field missing")
		}
	})

	t.Run("JSON field types", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create ticket
		id, _ := ctx.exec("new", "Type Test", "--priority", "2")
		id = strings.TrimSpace(id)

		// Query
		output, err := ctx.exec("query")
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Parse JSON
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		// Check types
		if _, ok := result["id"].(string); !ok {
			t.Error("id should be string")
		}
		if _, ok := result["status"].(string); !ok {
			t.Error("status should be string")
		}
		if _, ok := result["type"].(string); !ok {
			t.Error("type should be string")
		}
		if _, ok := result["priority"].(string); !ok {
			t.Error("priority should be string")
		}
		if _, ok := result["created"].(string); !ok {
			t.Error("created should be string (ISO format)")
		}
		if _, ok := result["deps"].([]interface{}); !ok {
			t.Error("deps should be array")
		}
		if _, ok := result["links"].([]interface{}); !ok {
			t.Error("links should be array")
		}
	})

	t.Run("filter by priority", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets with different priorities
		id1, _ := ctx.exec("new", "High Priority", "--priority", "0")
		id1 = strings.TrimSpace(id1)
		_, _ = ctx.exec("new", "Normal Priority", "--priority", "2")
		_, _ = ctx.exec("new", "Low Priority", "--priority", "4")

		// Query for priority 0
		output, err := ctx.exec("query", `.priority == "0"`)
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Should only have 1 result
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Errorf("expected 1 result, got %d", len(lines))
		}

		// Parse and verify it's the right ticket
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(lines[0]), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if result["id"] != id1 {
			t.Errorf("wrong ticket returned: got %v, want %v", result["id"], id1)
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets
		id1, _ := ctx.exec("new", "Open Ticket")
		id1 = strings.TrimSpace(id1)
		id2, _ := ctx.exec("new", "To Close")
		id2 = strings.TrimSpace(id2)

		// Close one ticket
		_, _ = ctx.exec("close", id2)

		// Query for open tickets
		output, err := ctx.exec("query", `.status == "open"`)
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Should only have 1 result
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Errorf("expected 1 result, got %d", len(lines))
		}

		// Parse and verify
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(lines[0]), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if result["id"] != id1 {
			t.Errorf("wrong ticket returned: got %v, want %v", result["id"], id1)
		}
	})

	t.Run("filter by type", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets with different types
		id1, _ := ctx.exec("new", "Bug Ticket", "--type", "bug")
		id1 = strings.TrimSpace(id1)
		_, _ = ctx.exec("new", "Feature Ticket", "--type", "feature")

		// Query for bugs
		output, err := ctx.exec("query", `.type == "bug"`)
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Should only have 1 result
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Errorf("expected 1 result, got %d", len(lines))
		}

		// Parse and verify
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(lines[0]), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if result["id"] != id1 {
			t.Errorf("wrong ticket returned: got %v, want %v", result["id"], id1)
		}
	})

	t.Run("filter by deps array length", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets
		id1, _ := ctx.exec("new", "Has Deps")
		id1 = strings.TrimSpace(id1)
		id2, _ := ctx.exec("new", "No Deps")
		id2 = strings.TrimSpace(id2)

		// Add dependency to first ticket
		_, _ = ctx.exec("dep", id1, id2)

		// Query for tickets with dependencies
		output, err := ctx.exec("query", `(.deps | length) > 0`)
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Should only have 1 result
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Errorf("expected 1 result, got %d", len(lines))
		}

		// Parse and verify
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(lines[0]), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if result["id"] != id1 {
			t.Errorf("wrong ticket returned: got %v, want %v", result["id"], id1)
		}
	})

	t.Run("complex filter", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets
		id1, _ := ctx.exec("new", "Match", "--priority", "0", "--type", "bug")
		id1 = strings.TrimSpace(id1)
		_, _ = ctx.exec("new", "No Match 1", "--priority", "0", "--type", "feature")
		_, _ = ctx.exec("new", "No Match 2", "--priority", "2", "--type", "bug")

		// Query for priority 0 AND type bug
		output, err := ctx.exec("query", `.priority == "0" and .type == "bug"`)
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Should only have 1 result
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Errorf("expected 1 result, got %d", len(lines))
		}

		// Parse and verify
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(lines[0]), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if result["id"] != id1 {
			t.Errorf("wrong ticket returned: got %v, want %v", result["id"], id1)
		}
	})

	t.Run("filter output is compact JSON", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create ticket
		_, _ = ctx.exec("new", "Compact Test")

		// Query
		output, err := ctx.exec("query")
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Output should be compact (no indentation)
		if strings.Contains(output, "  ") {
			t.Error("JSON output should be compact (no indentation)")
		}

		// Should be on a single line
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Log("Each ticket should be on its own line")
		}
	})

	t.Run("invalid filter returns error", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create ticket
		_, _ = ctx.exec("new", "Test")

		// Query with invalid filter
		_, err := ctx.exec("query", "invalid filter syntax !@#")
		if err == nil {
			t.Error("expected error for invalid filter, got nil")
		}
	})

	t.Run("empty result for no matches", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create ticket
		_, _ = ctx.exec("new", "Test", "--priority", "2")

		// Query for priority that doesn't exist
		output, err := ctx.exec("query", `.priority == "0"`)
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Output should be empty
		if strings.TrimSpace(output) != "" {
			t.Errorf("expected empty output, got: %s", output)
		}
	})

	t.Run("missing tickets directory", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Don't create any tickets (directory won't exist)

		// Query should return empty (not error)
		output, err := ctx.exec("query")
		if err != nil {
			t.Fatalf("query command should not error on missing directory: %v", err)
		}

		// Output should be empty
		if strings.TrimSpace(output) != "" {
			t.Errorf("expected empty output, got: %s", output)
		}
	})

	t.Run("arrays in JSON", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets with dependencies and links
		id1, _ := ctx.exec("new", "Parent")
		id1 = strings.TrimSpace(id1)
		id2, _ := ctx.exec("new", "Child")
		id2 = strings.TrimSpace(id2)
		id3, _ := ctx.exec("new", "Linked")
		id3 = strings.TrimSpace(id3)

		// Add dependency and link
		_, _ = ctx.exec("dep", id1, id2)
		_, _ = ctx.exec("link", id1, id3)

		// Query for the parent
		output, err := ctx.exec("query", `.id == "`+id1+`"`)
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Parse JSON
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		// Check deps array
		deps, ok := result["deps"].([]interface{})
		if !ok {
			t.Fatal("deps is not an array")
		}
		if len(deps) != 1 {
			t.Errorf("expected 1 dep, got %d", len(deps))
		}
		if deps[0] != id2 {
			t.Errorf("dep = %v, want %v", deps[0], id2)
		}

		// Check links array
		links, ok := result["links"].([]interface{})
		if !ok {
			t.Fatal("links is not an array")
		}
		if len(links) != 1 {
			t.Errorf("expected 1 link, got %d", len(links))
		}
		if links[0] != id3 {
			t.Errorf("link = %v, want %v", links[0], id3)
		}
	})

	t.Run("empty arrays serialized correctly", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create ticket without deps or links
		id, _ := ctx.exec("new", "Empty Arrays")
		id = strings.TrimSpace(id)

		// Query
		output, err := ctx.exec("query", `.id == "`+id+`"`)
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Parse JSON
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		// Check empty arrays
		deps, ok := result["deps"].([]interface{})
		if !ok {
			t.Fatal("deps is not an array")
		}
		if len(deps) != 0 {
			t.Errorf("expected empty deps array, got %d items", len(deps))
		}

		links, ok := result["links"].([]interface{})
		if !ok {
			t.Fatal("links is not an array")
		}
		if len(links) != 0 {
			t.Errorf("expected empty links array, got %d items", len(links))
		}

		// Verify JSON format shows [] not null
		jsonStr := strings.TrimSpace(output)
		if !strings.Contains(jsonStr, `"deps":[]`) {
			t.Error("deps should be serialized as []")
		}
		if !strings.Contains(jsonStr, `"links":[]`) {
			t.Error("links should be serialized as []")
		}
	})

	t.Run("filter with select wrapper", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets
		id1, _ := ctx.exec("new", "High Priority", "--priority", "0")
		id1 = strings.TrimSpace(id1)
		_, _ = ctx.exec("new", "Normal Priority", "--priority", "2")

		// Query with explicit select()
		output, err := ctx.exec("query", `select(.priority == "0")`)
		if err != nil {
			t.Fatalf("query command error: %v", err)
		}

		// Should only have 1 result
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Errorf("expected 1 result, got %d", len(lines))
		}

		// Parse and verify
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(lines[0]), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if result["id"] != id1 {
			t.Errorf("wrong ticket returned: got %v, want %v", result["id"], id1)
		}
	})
}
