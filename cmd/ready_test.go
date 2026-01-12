package cmd

import (
	"strings"
	"testing"
)

// TestReadyCommand tests the ready command
func TestReadyCommand(t *testing.T) {
	t.Run("filter open tickets", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create open ticket with no deps
		id, _ := ctx.exec("new", "Open Ticket")
		id = strings.TrimSpace(id)

		// Create closed ticket (should not appear)
		closedID, _ := ctx.exec("new", "Closed Ticket")
		closedID = strings.TrimSpace(closedID)
		ctx.exec("close", closedID)

		// Run ready
		output, err := ctx.exec("ready")
		if err != nil {
			t.Fatalf("ready command error: %v", err)
		}

		// Open ticket should appear
		if !strings.Contains(output, id) {
			t.Errorf("ready should include open ticket %s", id)
		}

		// Closed ticket should not appear
		if strings.Contains(output, closedID) {
			t.Errorf("ready should not include closed ticket %s", closedID)
		}
	})

	t.Run("filter in_progress tickets", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create ticket and start it
		id, _ := ctx.exec("new", "InProgress Ticket")
		id = strings.TrimSpace(id)
		ctx.exec("start", id)

		// Run ready
		output, err := ctx.exec("ready")
		if err != nil {
			t.Fatalf("ready command error: %v", err)
		}

		// in_progress ticket should appear
		if !strings.Contains(output, id) {
			t.Errorf("ready should include in_progress ticket %s", id)
		}
	})

	t.Run("exclude closed tickets", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create and close ticket
		id, _ := ctx.exec("new", "Closed Ticket")
		id = strings.TrimSpace(id)
		ctx.exec("close", id)

		// Run ready
		output, _ := ctx.exec("ready")

		// Closed ticket should not appear
		if strings.Contains(output, id) {
			t.Errorf("ready should not include closed ticket %s", id)
		}
	})
}

// TestReadyAllDependenciesClosed tests filtering by resolved dependencies
func TestReadyAllDependenciesClosed(t *testing.T) {
	t.Run("ticket with no deps is ready", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		id, _ := ctx.exec("new", "No Deps Ticket")
		id = strings.TrimSpace(id)

		output, _ := ctx.exec("ready")

		if !strings.Contains(output, id) {
			t.Errorf("ticket with no deps should be ready")
		}
	})

	t.Run("ticket with all deps closed is ready", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create dep and close it
		dep, _ := ctx.exec("new", "Dependency")
		dep = strings.TrimSpace(dep)
		ctx.exec("close", dep)

		// Create ticket depending on closed dep
		parent, _ := ctx.exec("new", "Parent")
		parent = strings.TrimSpace(parent)
		ctx.exec("dep", parent, dep)

		output, _ := ctx.exec("ready")

		if !strings.Contains(output, parent) {
			t.Errorf("ticket with closed deps should be ready")
		}
	})

	t.Run("ticket with any open dep is not ready", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create open dep
		dep, _ := ctx.exec("new", "Open Dependency")
		dep = strings.TrimSpace(dep)

		// Create parent depending on open dep
		parent, _ := ctx.exec("new", "Parent")
		parent = strings.TrimSpace(parent)
		ctx.exec("dep", parent, dep)

		output, _ := ctx.exec("ready")

		if strings.Contains(output, parent) {
			t.Errorf("ticket with open deps should not be ready")
		}
	})

	t.Run("ticket with multiple deps all closed is ready", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create two deps and close both
		dep1, _ := ctx.exec("new", "Dep 1")
		dep1 = strings.TrimSpace(dep1)
		ctx.exec("close", dep1)

		dep2, _ := ctx.exec("new", "Dep 2")
		dep2 = strings.TrimSpace(dep2)
		ctx.exec("close", dep2)

		// Create parent depending on both
		parent, _ := ctx.exec("new", "Parent")
		parent = strings.TrimSpace(parent)
		ctx.exec("dep", parent, dep1)
		ctx.exec("dep", parent, dep2)

		output, _ := ctx.exec("ready")

		if !strings.Contains(output, parent) {
			t.Errorf("ticket with all deps closed should be ready")
		}
	})

	t.Run("ticket with one open dep is not ready", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create two deps: one closed, one open
		dep1, _ := ctx.exec("new", "Closed Dep")
		dep1 = strings.TrimSpace(dep1)
		ctx.exec("close", dep1)

		dep2, _ := ctx.exec("new", "Open Dep")
		dep2 = strings.TrimSpace(dep2)

		// Create parent depending on both
		parent, _ := ctx.exec("new", "Parent")
		parent = strings.TrimSpace(parent)
		ctx.exec("dep", parent, dep1)
		ctx.exec("dep", parent, dep2)

		output, _ := ctx.exec("ready")

		if strings.Contains(output, parent) {
			t.Errorf("ticket with any open dep should not be ready")
		}
	})
}

// TestReadySortingByPriority tests priority sorting
func TestReadySortingByPriority(t *testing.T) {
	t.Run("sort by priority ascending", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets with different priorities
		id0, _ := ctx.exec("new", "Priority 0", "--priority", "0")
		id0 = strings.TrimSpace(id0)

		id2, _ := ctx.exec("new", "Priority 2", "--priority", "2")
		id2 = strings.TrimSpace(id2)

		id4, _ := ctx.exec("new", "Priority 4", "--priority", "4")
		id4 = strings.TrimSpace(id4)

		output, _ := ctx.exec("ready")

		// Find positions
		pos0 := strings.Index(output, id0)
		pos2 := strings.Index(output, id2)
		pos4 := strings.Index(output, id4)

		if pos0 == -1 || pos2 == -1 || pos4 == -1 {
			t.Fatal("all tickets should be in output")
		}

		// Priority 0 should come first (highest priority)
		if pos0 > pos2 || pos0 > pos4 {
			t.Errorf("priority 0 should come first, got positions: 0=%d, 2=%d, 4=%d", pos0, pos2, pos4)
		}

		// Priority 4 should come last (lowest priority)
		if pos4 < pos0 || pos4 < pos2 {
			t.Errorf("priority 4 should come last, got positions: 0=%d, 2=%d, 4=%d", pos0, pos2, pos4)
		}
	})

	t.Run("secondary sort by ID", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create two tickets with same priority
		id1, _ := ctx.exec("new", "First", "--priority", "2")
		id1 = strings.TrimSpace(id1)

		id2, _ := ctx.exec("new", "Second", "--priority", "2")
		id2 = strings.TrimSpace(id2)

		output, _ := ctx.exec("ready")

		// Both should be in output and sorted by ID
		if !strings.Contains(output, id1) || !strings.Contains(output, id2) {
			t.Error("both tickets should be in output")
		}

		// Verify they're sorted by ID
		pos1 := strings.Index(output, id1)
		pos2 := strings.Index(output, id2)

		if id1 < id2 {
			if pos1 > pos2 {
				t.Error("tickets with same priority should be sorted by ID")
			}
		} else {
			if pos2 > pos1 {
				t.Error("tickets with same priority should be sorted by ID")
			}
		}
	})
}

// TestReadyOutputFormat tests the output format
func TestReadyOutputFormat(t *testing.T) {
	t.Run("output format", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		id, _ := ctx.exec("new", "Test Ticket", "--priority", "1")
		id = strings.TrimSpace(id)

		output, _ := ctx.exec("ready")

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) == 0 {
			t.Fatal("expected output")
		}

		line := lines[0]

		// Format: "%-8s [P%d][%s] - %s"
		// Should contain ID
		if !strings.Contains(line, id) {
			t.Error("line should contain ID")
		}

		// Should contain priority with P prefix
		if !strings.Contains(line, "[P1]") {
			t.Errorf("line should contain [P1], got: %s", line)
		}

		// Should contain status
		if !strings.Contains(line, "[open]") {
			t.Errorf("line should contain [open], got: %s", line)
		}

		// Should contain title
		if !strings.Contains(line, "Test Ticket") {
			t.Errorf("line should contain title, got: %s", line)
		}

		// Should contain dash separator
		if !strings.Contains(line, " - ") {
			t.Errorf("line should contain ' - ' separator, got: %s", line)
		}
	})

	t.Run("shows ID, priority, status, and title", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		id, _ := ctx.exec("new", "My Ticket", "--priority", "3")
		id = strings.TrimSpace(id)

		output, _ := ctx.exec("ready")

		// All components should be present
		if !strings.Contains(output, id) {
			t.Error("output should contain ID")
		}
		if !strings.Contains(output, "[P3]") {
			t.Error("output should contain [P3]")
		}
		if !strings.Contains(output, "[open]") {
			t.Error("output should contain [open]")
		}
		if !strings.Contains(output, "My Ticket") {
			t.Error("output should contain title")
		}
	})
}

// TestReadyEmptyResult tests empty output when no ready tickets
func TestReadyEmptyResult(t *testing.T) {
	t.Run("no ready tickets returns empty", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create only closed tickets
		id, _ := ctx.exec("new", "Closed Ticket")
		id = strings.TrimSpace(id)
		ctx.exec("close", id)

		output, _ := ctx.exec("ready")

		// Output should be empty
		trimmed := strings.TrimSpace(output)
		if trimmed != "" {
			t.Errorf("expected empty output, got: %s", output)
		}
	})

	t.Run("only blocked tickets returns empty", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create open dependency
		dep, _ := ctx.exec("new", "Open Dep")
		dep = strings.TrimSpace(dep)

		// Create ticket depending on open dep
		parent, _ := ctx.exec("new", "Blocked Ticket")
		parent = strings.TrimSpace(parent)
		ctx.exec("dep", parent, dep)

		output, _ := ctx.exec("ready")

		// Neither should be ready (both have each other as deps or blocker)
		// Actually only parent is blocked, dep is ready
		// Let me fix this test
		if !strings.Contains(output, dep) {
			t.Error("dep with no deps should be ready")
		}
		if strings.Contains(output, parent) {
			t.Error("blocked ticket should not be ready")
		}
	})
}

// TestReadyMissingTicketsDirectory tests handling of missing .tickets directory
func TestReadyMissingTicketsDirectory(t *testing.T) {
	t.Run("missing directory returns empty silently", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Don't create any tickets (directory may not exist)

		output, err := ctx.exec("ready")
		if err != nil {
			t.Fatalf("ready should not error on missing directory: %v", err)
		}

		// Should return empty output
		trimmed := strings.TrimSpace(output)
		if trimmed != "" {
			t.Errorf("expected empty output for missing directory, got: %s", output)
		}
	})
}

// TestReadyDefaultPriority tests handling of default priority
func TestReadyDefaultPriority(t *testing.T) {
	t.Run("default priority is 2", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create ticket without specifying priority
		id, _ := ctx.exec("new", "Default Priority")
		id = strings.TrimSpace(id)

		output, _ := ctx.exec("ready")

		// Should show [P2] for default priority
		if !strings.Contains(output, "[P2]") {
			t.Errorf("should show default priority [P2], got: %s", output)
		}
	})

	t.Run("mixed priorities sorted correctly", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets with various priorities
		idDefault, _ := ctx.exec("new", "Default") // priority 2
		idDefault = strings.TrimSpace(idDefault)

		idHigh, _ := ctx.exec("new", "High", "--priority", "0")
		idHigh = strings.TrimSpace(idHigh)

		idLow, _ := ctx.exec("new", "Low", "--priority", "4")
		idLow = strings.TrimSpace(idLow)

		output, _ := ctx.exec("ready")

		// Find positions
		posHigh := strings.Index(output, idHigh)
		posDefault := strings.Index(output, idDefault)
		posLow := strings.Index(output, idLow)

		// Verify order: 0 < 2 < 4
		if posHigh > posDefault || posDefault > posLow {
			t.Errorf("priorities not sorted correctly: high=%d, default=%d, low=%d", posHigh, posDefault, posLow)
		}
	})
}

// TestReadyInProgressStatus tests in_progress tickets appear in ready
func TestReadyInProgressStatus(t *testing.T) {
	t.Run("in_progress with no deps is ready", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		id, _ := ctx.exec("new", "In Progress")
		id = strings.TrimSpace(id)
		ctx.exec("start", id)

		output, _ := ctx.exec("ready")

		if !strings.Contains(output, id) {
			t.Error("in_progress ticket with no deps should be ready")
		}

		if !strings.Contains(output, "[in_progress]") {
			t.Error("should show [in_progress] status")
		}
	})

	t.Run("in_progress with closed deps is ready", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		dep, _ := ctx.exec("new", "Dep")
		dep = strings.TrimSpace(dep)
		ctx.exec("close", dep)

		parent, _ := ctx.exec("new", "Parent")
		parent = strings.TrimSpace(parent)
		ctx.exec("dep", parent, dep)
		ctx.exec("start", parent)

		output, _ := ctx.exec("ready")

		if !strings.Contains(output, parent) {
			t.Error("in_progress ticket with closed deps should be ready")
		}
	})
}
