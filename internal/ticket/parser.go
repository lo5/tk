package ticket

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// frontmatter holds the YAML frontmatter fields
// We use a separate struct to control serialization order
type frontmatter struct {
	ID          string   `yaml:"id"`
	Status      Status   `yaml:"status"`
	Deps        []string `yaml:"deps,flow"`
	Links       []string `yaml:"links,flow"`
	Created     string   `yaml:"created"`
	Type        Type     `yaml:"type"`
	Priority    int      `yaml:"priority"`
	Assignee    string   `yaml:"assignee,omitempty"`
	ExternalRef string   `yaml:"external-ref,omitempty"`
	Parent      string   `yaml:"parent,omitempty"`
}

// Parse reads a ticket from a reader and returns the parsed Ticket
func Parse(r io.Reader) (*Ticket, error) {
	scanner := bufio.NewScanner(r)
	var frontmatterLines []string
	var bodyLines []string
	inFrontmatter := false
	frontmatterDone := false
	foundFirstDelim := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" {
			if !foundFirstDelim {
				foundFirstDelim = true
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				inFrontmatter = false
				frontmatterDone = true
				continue
			}
		}

		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		} else if frontmatterDone {
			bodyLines = append(bodyLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading ticket: %w", err)
	}

	// Parse YAML frontmatter
	var fm frontmatter
	yamlContent := strings.Join(frontmatterLines, "\n")
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	// Parse created time
	var created time.Time
	if fm.Created != "" {
		var err error
		created, err = time.Parse(time.RFC3339, fm.Created)
		if err != nil {
			// Try alternate formats
			created, err = time.Parse("2006-01-02T15:04:05Z", fm.Created)
			if err != nil {
				created = time.Time{}
			}
		}
	}

	// Extract title from first # heading
	title := ""
	bodyStart := 0
	for i, line := range bodyLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			title = strings.TrimPrefix(trimmed, "# ")
			bodyStart = i + 1
			break
		} else if trimmed != "" {
			// Non-empty, non-heading line means no title
			break
		}
	}

	// Rest is body
	body := ""
	if bodyStart < len(bodyLines) {
		body = strings.TrimSpace(strings.Join(bodyLines[bodyStart:], "\n"))
	}

	// Ensure deps and links are non-nil
	deps := fm.Deps
	if deps == nil {
		deps = []string{}
	}
	links := fm.Links
	if links == nil {
		links = []string{}
	}

	return &Ticket{
		ID:          fm.ID,
		Status:      fm.Status,
		Deps:        deps,
		Links:       links,
		Created:     created,
		Type:        fm.Type,
		Priority:    fm.Priority,
		Assignee:    fm.Assignee,
		ExternalRef: fm.ExternalRef,
		Parent:      fm.Parent,
		Title:       title,
		Body:        body,
	}, nil
}

// Format writes a ticket to a writer in the standard markdown format
func Format(w io.Writer, t *Ticket) error {
	var buf bytes.Buffer

	buf.WriteString("---\n")

	// Write frontmatter fields in specific order to match bash output
	buf.WriteString(fmt.Sprintf("id: %s\n", t.ID))
	buf.WriteString(fmt.Sprintf("status: %s\n", t.Status))
	buf.WriteString(fmt.Sprintf("deps: %s\n", formatArray(t.Deps)))
	buf.WriteString(fmt.Sprintf("links: %s\n", formatArray(t.Links)))
	buf.WriteString(fmt.Sprintf("created: %s\n", t.Created.UTC().Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("type: %s\n", t.Type))
	buf.WriteString(fmt.Sprintf("priority: %d\n", t.Priority))
	if t.Assignee != "" {
		buf.WriteString(fmt.Sprintf("assignee: %s\n", t.Assignee))
	}
	if t.ExternalRef != "" {
		buf.WriteString(fmt.Sprintf("external-ref: %s\n", t.ExternalRef))
	}
	if t.Parent != "" {
		buf.WriteString(fmt.Sprintf("parent: %s\n", t.Parent))
	}

	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("# %s\n", t.Title))

	if t.Body != "" {
		buf.WriteString("\n")
		buf.WriteString(t.Body)
		if !strings.HasSuffix(t.Body, "\n") {
			buf.WriteString("\n")
		}
	}

	_, err := w.Write(buf.Bytes())
	return err
}

// formatArray formats a string slice as a YAML flow-style array
func formatArray(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	return "[" + strings.Join(arr, ", ") + "]"
}

// UpdateField updates a specific field in a ticket file content
// This preserves the original formatting as much as possible
func UpdateField(content, field, value string) string {
	// Pattern to match the field line
	pattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(field) + `:.*$`)
	newLine := fmt.Sprintf("%s: %s", field, value)

	if pattern.MatchString(content) {
		return pattern.ReplaceAllString(content, newLine)
	}

	// Field doesn't exist, insert after first ---
	lines := strings.SplitN(content, "\n", -1)
	var result []string
	inserted := false
	for _, line := range lines {
		result = append(result, line)
		if !inserted && line == "---" {
			result = append(result, newLine)
			inserted = true
		}
	}
	return strings.Join(result, "\n")
}
