package cmd

import (
	"fmt"

	"github.com/lo5/tk/internal/ticket"
	"github.com/spf13/cobra"
)

var readyCmd = &cobra.Command{
	Use:   "ready",
	Short: "List ready tickets",
	Long:  `List open/in-progress tickets with all dependencies resolved.`,
	RunE:  runReady,
}

var readySort string

func init() {
	rootCmd.AddCommand(readyCmd)
	readyCmd.Flags().StringVar(&readySort, "sort", "priority", "Sort by field (date|priority|status)")
}

func runReady(cmd *cobra.Command, args []string) error {
	tickets, err := store.List()
	if err != nil {
		return err
	}

	if err := ticket.SortBy(tickets, readySort); err != nil {
		return err
	}

	// Build status map
	statusMap := make(map[string]ticket.Status)
	for _, t := range tickets {
		statusMap[t.ID] = t.Status
	}

	// Filter ready tickets (preserves sort order)
	var ready []*ticket.Ticket
	for _, t := range tickets {
		// Must be open or in_progress
		if t.Status != ticket.StatusOpen && t.Status != ticket.StatusInProgress {
			continue
		}

		// All deps must be closed
		allDepsResolved := true
		for _, dep := range t.Deps {
			if statusMap[dep] != ticket.StatusClosed {
				allDepsResolved = false
				break
			}
		}

		if allDepsResolved {
			ready = append(ready, t)
		}
	}

	// Print
	for _, t := range ready {
		fmt.Printf("%-8s [P%d][%s] - %s\n", t.ID, t.Priority, t.Status, t.Title)
	}

	return nil
}
