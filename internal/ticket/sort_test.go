package ticket

import (
	"testing"
	"time"
)

func TestSortBy(t *testing.T) {
	now := time.Now()

	t.Run("sort by date newest first", func(t *testing.T) {
		older := &Ticket{ID: "a-aaaa", Created: now.Add(-2 * time.Hour)}
		newer := &Ticket{ID: "b-bbbb", Created: now.Add(-1 * time.Hour)}
		newest := &Ticket{ID: "c-cccc", Created: now}

		tickets := []*Ticket{older, newest, newer}
		if err := SortBy(tickets, "date"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tickets[0] != newest || tickets[1] != newer || tickets[2] != older {
			t.Errorf("expected newest first, got: %v %v %v", tickets[0].ID, tickets[1].ID, tickets[2].ID)
		}
	})

	t.Run("sort by date ID tie-break", func(t *testing.T) {
		a := &Ticket{ID: "a-aaaa", Created: now}
		b := &Ticket{ID: "b-bbbb", Created: now}

		tickets := []*Ticket{b, a}
		if err := SortBy(tickets, "date"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tickets[0].ID != "a-aaaa" {
			t.Errorf("expected a-aaaa first on tie, got %s", tickets[0].ID)
		}
	})

	t.Run("sort by priority ascending", func(t *testing.T) {
		p0 := &Ticket{ID: "a-aaaa", Priority: 0}
		p2 := &Ticket{ID: "b-bbbb", Priority: 2}
		p4 := &Ticket{ID: "c-cccc", Priority: 4}

		tickets := []*Ticket{p4, p0, p2}
		if err := SortBy(tickets, "priority"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tickets[0] != p0 || tickets[1] != p2 || tickets[2] != p4 {
			t.Errorf("expected ascending priority order, got: %v %v %v", tickets[0].ID, tickets[1].ID, tickets[2].ID)
		}
	})

	t.Run("sort by priority ID tie-break", func(t *testing.T) {
		a := &Ticket{ID: "a-aaaa", Priority: 1}
		b := &Ticket{ID: "b-bbbb", Priority: 1}

		tickets := []*Ticket{b, a}
		if err := SortBy(tickets, "priority"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tickets[0].ID != "a-aaaa" {
			t.Errorf("expected a-aaaa first on tie, got %s", tickets[0].ID)
		}
	})

	t.Run("sort by status alphabetical", func(t *testing.T) {
		closed := &Ticket{ID: "a-aaaa", Status: StatusClosed}
		inProgress := &Ticket{ID: "b-bbbb", Status: StatusInProgress}
		open := &Ticket{ID: "c-cccc", Status: StatusOpen}

		tickets := []*Ticket{open, closed, inProgress}
		if err := SortBy(tickets, "status"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// "closed" < "in_progress" < "open" alphabetically
		if tickets[0] != closed || tickets[1] != inProgress || tickets[2] != open {
			t.Errorf("expected alphabetical status order, got: %v %v %v", tickets[0].Status, tickets[1].Status, tickets[2].Status)
		}
	})

	t.Run("sort by status ID tie-break", func(t *testing.T) {
		a := &Ticket{ID: "a-aaaa", Status: StatusOpen}
		b := &Ticket{ID: "b-bbbb", Status: StatusOpen}

		tickets := []*Ticket{b, a}
		if err := SortBy(tickets, "status"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tickets[0].ID != "a-aaaa" {
			t.Errorf("expected a-aaaa first on tie, got %s", tickets[0].ID)
		}
	})

	t.Run("invalid field returns error", func(t *testing.T) {
		tickets := []*Ticket{{ID: "a-aaaa"}}
		err := SortBy(tickets, "id")
		if err == nil {
			t.Fatal("expected error for invalid field")
		}
		if err.Error() != `unknown sort field "id": must be one of date, priority, status` {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
