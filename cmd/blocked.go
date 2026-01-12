package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lo5/tk/internal/ticket"
	"github.com/spf13/cobra"
)

var blockedCmd = &cobra.Command{
	Use:   "blocked",
	Short: "List blocked tickets",
	Long: `List open/in-progress tickets with unresolved dependencies.
Shows only the unclosed blockers for each ticket.
Sorted by priority (ascending, 0=highest), then by ID.`,
	RunE: runBlocked,
}

func init() {
	rootCmd.AddCommand(blockedCmd)
}

type blockedTicket struct {
	ticket   *ticket.Ticket
	blockers []string
}

func runBlocked(cmd *cobra.Command, args []string) error {
	tickets, err := store.List()
	if err != nil {
		return err
	}

	// Build status map
	statusMap := make(map[string]ticket.Status)
	for _, t := range tickets {
		statusMap[t.ID] = t.Status
	}

	// Filter blocked tickets
	var blocked []blockedTicket
	for _, t := range tickets {
		// Must be open or in_progress
		if t.Status != ticket.StatusOpen && t.Status != ticket.StatusInProgress {
			continue
		}

		// Must have deps
		if len(t.Deps) == 0 {
			continue
		}

		// Find unclosed blockers
		var blockers []string
		for _, dep := range t.Deps {
			if statusMap[dep] != ticket.StatusClosed {
				blockers = append(blockers, dep)
			}
		}

		if len(blockers) > 0 {
			blocked = append(blocked, blockedTicket{ticket: t, blockers: blockers})
		}
	}

	// Sort by priority, then by ID
	sort.Slice(blocked, func(i, j int) bool {
		if blocked[i].ticket.Priority != blocked[j].ticket.Priority {
			return blocked[i].ticket.Priority < blocked[j].ticket.Priority
		}
		return blocked[i].ticket.ID < blocked[j].ticket.ID
	})

	// Print
	for _, b := range blocked {
		blockersStr := "[" + strings.Join(b.blockers, ", ") + "]"
		fmt.Printf("%-8s [P%d][%s] - %s <- %s\n", b.ticket.ID, b.ticket.Priority, b.ticket.Status, b.ticket.Title, blockersStr)
	}

	return nil
}
