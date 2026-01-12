package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note <id> [note text]",
	Short: "Append timestamped note to ticket",
	Long: `Append a timestamped note to a ticket.
Note text can be provided as arguments or piped via stdin.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runNote,
}

func init() {
	rootCmd.AddCommand(noteCmd)
}

func runNote(cmd *cobra.Command, args []string) error {
	ticketID := args[0]

	// Get note text
	var note string
	if len(args) > 1 {
		note = strings.Join(args[1:], " ")
	} else if !isTerminal() {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		note = strings.Join(lines, "\n")
	} else {
		return fmt.Errorf("no note provided")
	}

	// Check if Notes section exists
	contains, id, err := store.FileContains(ticketID, "## Notes")
	if err != nil {
		return err
	}

	// Build content to append
	timestamp := time.Now().UTC().Format(time.RFC3339)
	var content strings.Builder

	if !contains {
		content.WriteString("\n## Notes\n")
	}
	content.WriteString(fmt.Sprintf("\n**%s**\n\n%s\n", timestamp, note))

	// Append to file
	_, err = store.AppendToFile(id, content.String())
	if err != nil {
		return err
	}

	fmt.Printf("Note added to %s\n", id)
	return nil
}
