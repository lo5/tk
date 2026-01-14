package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPruneNoDanglingRefs - Clean dataset with no dangling refs
func TestPruneNoDanglingRefs(t *testing.T) {
	t.Run("no dangling references", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets with valid references
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)
		idC, _ := ctx.exec("new", "Ticket C")
		idC = strings.TrimSpace(idC)

		// Add valid dependencies and links
		ctx.exec("dep", idA, idB)
		ctx.exec("link", idB, idC)

		// Run prune
		output, err := ctx.exec("prune")
		if err != nil {
			t.Fatalf("prune failed: %v", err)
		}

		if !strings.Contains(output, "No dangling references found") {
			t.Errorf("expected no dangling refs message, got: %s", output)
		}
	})

	t.Run("empty store", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Run prune on empty store
		output, err := ctx.exec("prune")
		if err != nil {
			t.Fatalf("prune failed on empty store: %v", err)
		}

		if !strings.Contains(output, "No tickets found") {
			t.Errorf("expected 'No tickets found' message, got: %s", output)
		}
	})
}

// TestPruneDanglingDeps - Deps pointing to deleted tickets
func TestPruneDanglingDeps(t *testing.T) {
	t.Run("single dangling dep detected", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create A depending on B
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)

		ctx.exec("dep", idA, idB)

		// Manually delete B to create dangling ref
		os.Remove(filepath.Join(ctx.ticketsDir, idB+".md"))

		// Dry-run should detect it
		output, err := ctx.exec("prune")
		if err != nil {
			t.Fatalf("prune dry-run failed: %v", err)
		}

		if !strings.Contains(output, "Found dangling references in 1 ticket") {
			t.Errorf("expected 1 ticket with dangling refs, got: %s", output)
		}
		if !strings.Contains(output, idB) {
			t.Errorf("expected dangling ref %s in output, got: %s", idB, output)
		}
		if !strings.Contains(output, "do not exist") && !strings.Contains(output, "does not exist") {
			t.Errorf("expected 'do(es) not exist' in output, got: %s", output)
		}
		if !strings.Contains(output, "1 dangling deps") {
			t.Errorf("expected summary with '1 dangling deps', got: %s", output)
		}
		if !strings.Contains(output, "Run with --fix") {
			t.Errorf("expected --fix hint, got: %s", output)
		}

		// Verify ticket A unchanged
		ticketA, _ := ctx.store().Get(idA)
		if len(ticketA.Deps) != 1 || ticketA.Deps[0] != idB {
			t.Error("dry-run should not modify ticket")
		}
	})

	t.Run("fix removes dangling dep", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create A depending on B
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)

		ctx.exec("dep", idA, idB)

		// Delete B
		os.Remove(filepath.Join(ctx.ticketsDir, idB+".md"))

		// Fix it
		output, err := ctx.exec("prune", "--fix")
		if err != nil {
			t.Fatalf("prune --fix failed: %v", err)
		}

		if !strings.Contains(output, "Pruning dangling references") {
			t.Errorf("expected 'Pruning' message, got: %s", output)
		}
		if !strings.Contains(output, "Removed deps") {
			t.Errorf("expected 'Removed deps' message, got: %s", output)
		}
		if !strings.Contains(output, idB) {
			t.Errorf("expected dangling ref %s in output, got: %s", idB, output)
		}
		if !strings.Contains(output, "Pruned 1 dangling reference(s)") {
			t.Errorf("expected prune summary, got: %s", output)
		}

		// Verify deps removed
		ticketA, _ := ctx.store().Get(idA)
		if len(ticketA.Deps) != 0 {
			t.Errorf("expected empty deps after prune, got: %v", ticketA.Deps)
		}
	})

	t.Run("multiple dangling deps", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create A depending on B and C
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)
		idC, _ := ctx.exec("new", "Ticket C")
		idC = strings.TrimSpace(idC)

		ctx.exec("dep", idA, idB)
		ctx.exec("dep", idA, idC)

		// Delete B and C
		os.Remove(filepath.Join(ctx.ticketsDir, idB+".md"))
		os.Remove(filepath.Join(ctx.ticketsDir, idC+".md"))

		// Fix it
		output, err := ctx.exec("prune", "--fix")
		if err != nil {
			t.Fatalf("prune --fix failed: %v", err)
		}

		// Verify both removed
		ticketA, _ := ctx.store().Get(idA)
		if len(ticketA.Deps) != 0 {
			t.Errorf("expected empty deps, got: %v", ticketA.Deps)
		}

		if !strings.Contains(output, "Pruned 2 dangling reference(s)") {
			t.Errorf("expected 2 refs pruned, got: %s", output)
		}
	})

	t.Run("partial deps - keep valid, remove dangling", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create A depending on B, C, D
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)
		idC, _ := ctx.exec("new", "Ticket C")
		idC = strings.TrimSpace(idC)
		idD, _ := ctx.exec("new", "Ticket D")
		idD = strings.TrimSpace(idD)

		ctx.exec("dep", idA, idB)
		ctx.exec("dep", idA, idC)
		ctx.exec("dep", idA, idD)

		// Delete only C
		os.Remove(filepath.Join(ctx.ticketsDir, idC+".md"))

		// Fix it
		ctx.exec("prune", "--fix")

		// Verify only C removed, B and D remain
		ticketA, _ := ctx.store().Get(idA)
		if len(ticketA.Deps) != 2 {
			t.Errorf("expected 2 deps remaining, got: %v", ticketA.Deps)
		}
		if ticketA.Deps[0] != idB || ticketA.Deps[1] != idD {
			t.Errorf("expected deps [%s, %s], got: %v", idB, idD, ticketA.Deps)
		}
	})
}

// TestPruneDanglingLinks - Links to deleted tickets
func TestPruneDanglingLinks(t *testing.T) {
	t.Run("single dangling link", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create A and B with link
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)

		ctx.exec("link", idA, idB)

		// Delete B
		os.Remove(filepath.Join(ctx.ticketsDir, idB+".md"))

		// Dry-run
		output, err := ctx.exec("prune")
		if err != nil {
			t.Fatalf("prune failed: %v", err)
		}

		if !strings.Contains(output, "1 dangling links") {
			t.Errorf("expected 1 dangling link in summary, got: %s", output)
		}

		// Fix it
		ctx.exec("prune", "--fix")

		// Verify link removed
		ticketA, _ := ctx.store().Get(idA)
		if len(ticketA.Links) != 0 {
			t.Errorf("expected empty links, got: %v", ticketA.Links)
		}
	})

	t.Run("multiple dangling links", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create A linked to B and C
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)
		idC, _ := ctx.exec("new", "Ticket C")
		idC = strings.TrimSpace(idC)

		ctx.exec("link", idA, idB, idC)

		// Delete B and C
		os.Remove(filepath.Join(ctx.ticketsDir, idB+".md"))
		os.Remove(filepath.Join(ctx.ticketsDir, idC+".md"))

		// Fix it
		ctx.exec("prune", "--fix")

		// Verify all links removed
		ticketA, _ := ctx.store().Get(idA)
		if len(ticketA.Links) != 0 {
			t.Errorf("expected empty links, got: %v", ticketA.Links)
		}
	})

	t.Run("partial links - keep valid", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create A linked to B and C
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)
		idC, _ := ctx.exec("new", "Ticket C")
		idC = strings.TrimSpace(idC)

		ctx.exec("link", idA, idB, idC)

		// Delete only C
		os.Remove(filepath.Join(ctx.ticketsDir, idC+".md"))

		// Fix it
		ctx.exec("prune", "--fix")

		// Verify only C removed, B remains
		ticketA, _ := ctx.store().Get(idA)
		if len(ticketA.Links) != 1 {
			t.Errorf("expected 1 link remaining, got: %v", ticketA.Links)
		}
		if ticketA.Links[0] != idB {
			t.Errorf("expected link to %s, got: %v", idB, ticketA.Links)
		}
	})
}

// TestPruneDanglingParent - Parent references to deleted tickets
func TestPruneDanglingParent(t *testing.T) {
	t.Run("dangling parent detected and removed", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create parent and child
		idP, _ := ctx.exec("new", "Parent ticket")
		idP = strings.TrimSpace(idP)
		idC, _ := ctx.exec("new", "Child ticket", "--parent", idP)
		idC = strings.TrimSpace(idC)

		// Verify parent set
		child, _ := ctx.store().Get(idC)
		if child.Parent != idP {
			t.Fatalf("parent not set correctly: got %s, want %s", child.Parent, idP)
		}

		// Delete parent
		os.Remove(filepath.Join(ctx.ticketsDir, idP+".md"))

		// Dry-run
		output, err := ctx.exec("prune")
		if err != nil {
			t.Fatalf("prune failed: %v", err)
		}

		if !strings.Contains(output, "1 dangling parent") {
			t.Errorf("expected 1 dangling parent in summary, got: %s", output)
		}

		// Fix it
		output, err = ctx.exec("prune", "--fix")
		if err != nil {
			t.Fatalf("prune --fix failed: %v", err)
		}

		if !strings.Contains(output, "Removed parent") {
			t.Errorf("expected 'Removed parent' message, got: %s", output)
		}

		// Verify parent cleared
		child, _ = ctx.store().Get(idC)
		if child.Parent != "" {
			t.Errorf("expected empty parent, got: %s", child.Parent)
		}
	})
}

// TestPruneMixedDanglingRefs - Ticket with multiple types of dangling refs
func TestPruneMixedDanglingRefs(t *testing.T) {
	t.Run("ticket with deps, links, and parent all dangling", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create tickets
		idP, _ := ctx.exec("new", "Parent")
		idP = strings.TrimSpace(idP)
		idD, _ := ctx.exec("new", "Dep")
		idD = strings.TrimSpace(idD)
		idL, _ := ctx.exec("new", "Link")
		idL = strings.TrimSpace(idL)
		idA, _ := ctx.exec("new", "Target", "--parent", idP)
		idA = strings.TrimSpace(idA)

		// Add deps and links
		ctx.exec("dep", idA, idD)
		ctx.exec("link", idA, idL)

		// Delete all referenced tickets
		os.Remove(filepath.Join(ctx.ticketsDir, idP+".md"))
		os.Remove(filepath.Join(ctx.ticketsDir, idD+".md"))
		os.Remove(filepath.Join(ctx.ticketsDir, idL+".md"))

		// Dry-run
		output, err := ctx.exec("prune")
		if err != nil {
			t.Fatalf("prune failed: %v", err)
		}

		if !strings.Contains(output, "3 total dangling references") {
			t.Errorf("expected 3 total dangling refs, got: %s", output)
		}

		// Fix it
		ctx.exec("prune", "--fix")

		// Verify all cleared
		ticket, _ := ctx.store().Get(idA)
		if len(ticket.Deps) != 0 {
			t.Errorf("expected empty deps, got: %v", ticket.Deps)
		}
		if len(ticket.Links) != 0 {
			t.Errorf("expected empty links, got: %v", ticket.Links)
		}
		if ticket.Parent != "" {
			t.Errorf("expected empty parent, got: %s", ticket.Parent)
		}
	})
}

// TestPruneMultipleTickets - Multiple tickets with dangling refs
func TestPruneMultipleTickets(t *testing.T) {
	t.Run("multiple tickets affected", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create shared dependency
		idShared, _ := ctx.exec("new", "Shared")
		idShared = strings.TrimSpace(idShared)

		// Create tickets depending on shared
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)
		idC, _ := ctx.exec("new", "Ticket C")
		idC = strings.TrimSpace(idC)

		ctx.exec("dep", idA, idShared)
		ctx.exec("dep", idB, idShared)
		ctx.exec("dep", idC, idShared)

		// Delete shared
		os.Remove(filepath.Join(ctx.ticketsDir, idShared+".md"))

		// Dry-run
		output, err := ctx.exec("prune")
		if err != nil {
			t.Fatalf("prune failed: %v", err)
		}

		if !strings.Contains(output, "Found dangling references in 3 ticket") {
			t.Errorf("expected 3 tickets affected, got: %s", output)
		}
		if !strings.Contains(output, "3 total dangling references") {
			t.Errorf("expected 3 total refs, got: %s", output)
		}

		// Fix it
		output, err = ctx.exec("prune", "--fix")
		if err != nil {
			t.Fatalf("prune --fix failed: %v", err)
		}

		if !strings.Contains(output, "Pruned 3 dangling reference(s) from 3 ticket(s)") {
			t.Errorf("expected summary of 3 refs from 3 tickets, got: %s", output)
		}

		// Verify all cleaned
		for _, id := range []string{idA, idB, idC} {
			ticket, _ := ctx.store().Get(id)
			if len(ticket.Deps) != 0 {
				t.Errorf("ticket %s should have empty deps, got: %v", id, ticket.Deps)
			}
		}
	})
}

// TestPrunePersistence - Verify changes persist to disk
func TestPrunePersistence(t *testing.T) {
	t.Run("changes persist after reload", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// Create A depending on B
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)

		ctx.exec("dep", idA, idB)

		// Delete B
		os.Remove(filepath.Join(ctx.ticketsDir, idB+".md"))

		// Fix it
		ctx.exec("prune", "--fix")

		// Create new store instance and verify
		newStore := ctx.store()
		ticket, err := newStore.Get(idA)
		if err != nil {
			t.Fatalf("failed to reload ticket: %v", err)
		}

		if len(ticket.Deps) != 0 {
			t.Errorf("changes should persist after reload, got deps: %v", ticket.Deps)
		}

		// Also verify the file content
		content, err := os.ReadFile(filepath.Join(ctx.ticketsDir, idA+".md"))
		if err != nil {
			t.Fatalf("failed to read ticket file: %v", err)
		}

		if strings.Contains(string(content), idB) {
			t.Error("dangling ref should not appear in file after prune")
		}
	})
}

// TestPruneFlagReset - Ensure flag resets between tests
func TestPruneFlagReset(t *testing.T) {
	t.Run("flag resets correctly", func(t *testing.T) {
		ctx, cleanup := setupTestCmd(t)
		defer cleanup()

		// pruneFix should be false by default
		if pruneFix {
			t.Error("pruneFix should be false at test start")
		}

		// Create scenario
		idA, _ := ctx.exec("new", "Ticket A")
		idA = strings.TrimSpace(idA)
		idB, _ := ctx.exec("new", "Ticket B")
		idB = strings.TrimSpace(idB)

		ctx.exec("dep", idA, idB)
		os.Remove(filepath.Join(ctx.ticketsDir, idB+".md"))

		// Run without --fix
		ctx.exec("prune")

		// Verify not modified
		ticket, _ := ctx.store().Get(idA)
		if len(ticket.Deps) != 1 {
			t.Error("dry-run should not modify ticket")
		}
	})
}
