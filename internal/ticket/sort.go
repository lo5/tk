package ticket

import (
	"fmt"
	"sort"
)

// SortBy sorts tickets in-place by the given field.
// Valid fields: "date" (newest first), "priority" (ascending, 0=highest), "status" (alphabetical).
func SortBy(tickets []*Ticket, field string) error {
	switch field {
	case "date":
		sort.Slice(tickets, func(i, j int) bool {
			if !tickets[i].Created.Equal(tickets[j].Created) {
				return tickets[i].Created.After(tickets[j].Created)
			}
			return tickets[i].ID < tickets[j].ID
		})
	case "priority":
		sort.Slice(tickets, func(i, j int) bool {
			if tickets[i].Priority != tickets[j].Priority {
				return tickets[i].Priority < tickets[j].Priority
			}
			return tickets[i].ID < tickets[j].ID
		})
	case "status":
		sort.Slice(tickets, func(i, j int) bool {
			if tickets[i].Status != tickets[j].Status {
				return tickets[i].Status < tickets[j].Status
			}
			return tickets[i].ID < tickets[j].ID
		})
	default:
		return fmt.Errorf("unknown sort field %q: must be one of date, priority, status", field)
	}
	return nil
}
