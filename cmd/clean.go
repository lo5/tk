package cmd

import (
	"fmt"

	"github.com/lo5/tk/internal/ticket"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Delete all closed tickets",
	Long: `Delete all closed tickets that are safe to remove.

By default, performs a dry-run showing what would be deleted.
Use --fix to actually delete the tickets.

Refuses deletion if a closed ticket:
  - Has dependants (other tickets depend on it, regardless of status)
  - Has non-closed children (other tickets have it as parent and are open/in_progress)
  - Has bidirectional links

This ensures that only truly obsolete closed tickets are removed.`,
	Args: cobra.NoArgs,
	RunE: runClean,
}

var cleanFix bool

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolVar(&cleanFix, "fix", false,
		"Actually delete closed tickets (default is dry-run)")
}

type cleanableTicket struct {
	ticket  *ticket.Ticket
	blocked bool
	reason  string
}

func runClean(cmd *cobra.Command, args []string) error {
	// 1. Load all tickets
	allTickets, err := store.List()
	if err != nil {
		return err
	}

	// 2. Filter for closed tickets and check if they're safe to delete
	var cleanable []cleanableTicket
	for _, t := range allTickets {
		if t.Status != ticket.StatusClosed {
			continue
		}

		ct := cleanableTicket{ticket: t}

		// Check for dependants
		dependants := findDependants(allTickets, t.ID)
		if len(dependants) > 0 {
			ct.blocked = true
			ct.reason = "has dependants"
			cleanable = append(cleanable, ct)
			continue
		}

		// Check for children (only non-closed children block deletion)
		children := findChildren(allTickets, t.ID)
		var nonClosedChildren []*ticket.Ticket
		for _, child := range children {
			if child.Status != ticket.StatusClosed {
				nonClosedChildren = append(nonClosedChildren, child)
			}
		}
		if len(nonClosedChildren) > 0 {
			ct.blocked = true
			ct.reason = "has non-closed children"
			cleanable = append(cleanable, ct)
			continue
		}

		// Check for links
		if len(t.Links) > 0 {
			ct.blocked = true
			ct.reason = "has links"
			cleanable = append(cleanable, ct)
			continue
		}

		// If we get here, ticket is safe to delete
		cleanable = append(cleanable, ct)
	}

	// 3. Separate into deletable and blocked lists
	var deletable []cleanableTicket
	var blocked []cleanableTicket
	for _, ct := range cleanable {
		if ct.blocked {
			blocked = append(blocked, ct)
		} else {
			deletable = append(deletable, ct)
		}
	}

	// 4. Handle dry-run (default)
	if !cleanFix {
		totalClosed := len(cleanable)
		numDeletable := len(deletable)
		numBlocked := len(blocked)

		if totalClosed == 0 {
			fmt.Println("No closed tickets found.")
			return nil
		}

		fmt.Printf("Found %d closed ticket(s):\n", totalClosed)
		fmt.Printf("  %d deletable\n", numDeletable)
		fmt.Printf("  %d blocked\n", numBlocked)

		if numBlocked > 0 {
			fmt.Println("\nBlocked tickets:")
			for _, ct := range blocked {
				fmt.Printf("  %s [%s] %s - %s\n", ct.ticket.ID, ct.ticket.Status, ct.ticket.Title, ct.reason)
			}
		}

		if numDeletable > 0 {
			fmt.Printf("\nRun with --fix to delete %d deletable ticket(s).\n", numDeletable)
		}

		return nil
	}

	// 5. Handle --fix mode (actual deletion)
	if len(deletable) == 0 {
		if len(blocked) > 0 {
			fmt.Printf("No deletable tickets. All %d closed ticket(s) are blocked.\n", len(blocked))
		} else {
			fmt.Println("No closed tickets found.")
		}
		return nil
	}

	fmt.Println("Deleting closed tickets...")
	fmt.Println()

	successCount := 0
	errorCount := 0

	for _, ct := range deletable {
		if err := store.Delete(ct.ticket.ID); err != nil {
			fmt.Printf("Warning: failed to delete %s: %v\n", ct.ticket.ID, err)
			errorCount++
			continue
		}
		fmt.Printf("Deleted: %s\n", ct.ticket.ID)
		successCount++
	}

	fmt.Printf("\nDeleted %d ticket(s)", successCount)
	if len(blocked) > 0 {
		fmt.Printf(", skipped %d blocked ticket(s)", len(blocked))
	}
	if errorCount > 0 {
		fmt.Printf(", %d error(s)", errorCount)
	}
	fmt.Println(".")

	return nil
}
