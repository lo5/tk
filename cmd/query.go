package cmd

import (
	"fmt"

	"github.com/lo5/tk/internal/query"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query [jq-filter]",
	Short: "Output tickets as JSON",
	Long: `Output tickets as JSON, one object per line.
Optionally apply a jq-style filter.

Examples:
  tk query                          # All tickets as JSON
  tk query '.priority == "0"'       # High priority tickets
  tk query '.status == "open"'      # Open tickets`,
	RunE: runQuery,
}

func init() {
	rootCmd.AddCommand(queryCmd)
}

func runQuery(cmd *cobra.Command, args []string) error {
	tickets, err := store.List()
	if err != nil {
		return err
	}

	// Convert all tickets to JSON
	var jsonLines []string
	for _, t := range tickets {
		line, err := query.ToJSON(t)
		if err != nil {
			continue
		}
		jsonLines = append(jsonLines, line)
	}

	// Apply filter if provided
	if len(args) > 0 {
		filtered, err := query.Filter(jsonLines, args[0])
		if err != nil {
			return err
		}
		jsonLines = filtered
	}

	// Print results
	for _, line := range jsonLines {
		fmt.Println(line)
	}

	return nil
}
