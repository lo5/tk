package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lo5/tk/internal/ticket"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "ls [--status=X]",
	Aliases: []string{"list"},
	Short:   "List tickets",
	Long:    "List all tickets, optionally filtered by status.",
	RunE:    runList,
}

var listStatus string

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (open|in_progress|closed)")
}

func runList(cmd *cobra.Command, args []string) error {
	tickets, err := store.List()
	if err != nil {
		return err
	}

	// Filter by status if specified
	if listStatus != "" {
		status := ticket.Status(listStatus)
		var filtered []*ticket.Ticket
		for _, t := range tickets {
			if t.Status == status {
				filtered = append(filtered, t)
			}
		}
		tickets = filtered
	}

	// Sort by ID for consistent output
	sort.Slice(tickets, func(i, j int) bool {
		return tickets[i].ID < tickets[j].ID
	})

	for _, t := range tickets {
		depStr := ""
		if len(t.Deps) > 0 {
			depStr = " <- [" + strings.Join(t.Deps, ", ") + "]"
		}
		fmt.Printf("%-8s [%s] - %s%s\n", t.ID, t.Status, t.Title, depStr)
	}

	return nil
}
