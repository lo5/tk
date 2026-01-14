# tk

`tk` is a minimal ticket tracking system for managing tasks and collaborating with AI agents on long-horizon features.

`tk` is similar to [beads](https://github.com/steveyegge/beads), but stores everything as simple markdown files with YAML frontmatter — no database or daemon to manage. `tk` started out as a Go port of the [ticket](https://github.com/wedow/ticket) single-file bash script, inspired by Joe Armstrong's [Minimal Viable Program](https://joearms.github.io/published/2014-06-25-minimal-viable-program.html)

## Workflow

Use `tk` only for long-horizon tasks that cannot be completed in one shot (say 200K context sans-compaction).

Enter plan mode with your AI agent, make a plan, ask the agent to *save the plan, break it into 5-10 issues, each with a link to the original plan and file it in `tk`*. From this point on, you can simply clear the context and direct the agent to *fix issue x-h42g* or *run `tk ready` and fix the next available issue*.

## Key Features

- **File-based storage**: Tickets are `.md` files in `.tickets/`, editable in any text editor
- **Dependency tracking**: Define dependencies between tickets and visualize them as trees
- **Cross-linking**: Link related tickets together for better context
- **Partial ID matching**: Refer to tickets by any substring of their ID (e.g., `h42` matches `x-h42g`)
- **Status management**: Track tickets as `open`, `in_progress`, or `closed`
- **jq-style queries**: Filter tickets with powerful query expressions
- **Git-friendly**: Store `.tickets/` in git (like `git-bug`) or gitignore it as a local todo list
- **AI-friendly**: Designed to be easily traced by AI agents following dependency graphs and context


## Quick Start

```bash
# Create a new ticket
tk new "Fix login page"

# List all tickets
tk ls

# Show a ticket (partial ID matching)
tk show h42

# Add a dependency
tk dep h42 8a2

# View dependency tree
tk dep tree h42

# Update status
tk start h42      # Mark as in_progress
tk close h42      # Mark as closed

# Append notes
tk note h42 "Made progress on authentication"

# Query tickets
tk ls --status in_progress
tk query '.status == "in_progress"'
```
## Installation

```bash
go install github.com/lo5/tk@latest
```

This installs `tk` to `$GOPATH/bin` (or `$HOME/go/bin` by default). Ensure this directory is in your PATH:

```bash
export PATH=$PATH:$HOME/go/bin
```

## All Commands

```bash

./tk help
tk - minimal ticket system with dependency tracking

Tickets are stored as markdown files with YAML frontmatter in .tickets/
Supports partial ID matching (e.g., 'tk show h42' matches 'nw-h42g')

Usage:
  tk [command]

Available Commands:
  blocked     List blocked tickets
  close       Set ticket status to closed
  closed      List recently closed tickets
  completion  Generate the autocompletion script for the specified shell
  dep         Add a dependency
  edit        Open ticket in $EDITOR
  help        Help about any command
  link        Link tickets together
  ls          List tickets
  new         Create a new ticket
  note        Append timestamped note to ticket
  query       Output tickets as JSON
  ready       List ready tickets
  reopen      Set ticket status to open
  rm          Delete a ticket
  show        Display a ticket
  start       Set ticket status to in_progress
  status      Update ticket status
  undep       Remove a dependency
  unlink      Remove link between tickets

Flags:
      --dir string   tickets directory (default ".tickets")
  -h, --help         help for tk

Use "tk [command] --help" for more information about a command.
```

## Ticket Format

Tickets are markdown files stored in `.tickets/{id}.md` with YAML frontmatter:

```yaml
---
id: gotk-h42g
status: open
created: 2025-01-12T10:30:00Z
deps: []
links: []
priority: 1
type: feature
assignee: Jack B. Nimble
---

# Ticket Title

Description of the ticket goes here. You can use markdown formatting.
More details, context, and notes can be added.
```

