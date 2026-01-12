package ticket

import (
	"time"
)

// Status represents the status of a ticket
type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusClosed     Status = "closed"
)

// ValidStatuses lists all valid status values
var ValidStatuses = []Status{StatusOpen, StatusInProgress, StatusClosed}

// IsValid checks if a status is valid
func (s Status) IsValid() bool {
	for _, v := range ValidStatuses {
		if s == v {
			return true
		}
	}
	return false
}

// Type represents the type of a ticket
type Type string

const (
	TypeBug     Type = "bug"
	TypeFeature Type = "feature"
	TypeTask    Type = "task"
	TypeEpic    Type = "epic"
	TypeChore   Type = "chore"
)

// ValidTypes lists all valid type values
var ValidTypes = []Type{TypeBug, TypeFeature, TypeTask, TypeEpic, TypeChore}

// IsValid checks if a type is valid
func (t Type) IsValid() bool {
	for _, v := range ValidTypes {
		if t == v {
			return true
		}
	}
	return false
}

// Ticket represents a ticket with all its metadata and content
type Ticket struct {
	ID          string    `yaml:"id"`
	Status      Status    `yaml:"status"`
	Deps        []string  `yaml:"deps,flow"`
	Links       []string  `yaml:"links,flow"`
	Created     time.Time `yaml:"created"`
	Type        Type      `yaml:"type"`
	Priority    int       `yaml:"priority"`
	Assignee    string    `yaml:"assignee,omitempty"`
	ExternalRef string    `yaml:"external-ref,omitempty"`
	Parent      string    `yaml:"parent,omitempty"`
	Title       string    `yaml:"-"` // From # heading
	Body        string    `yaml:"-"` // Markdown content after title
}

// DefaultTicketsDir is the default directory for storing tickets
const DefaultTicketsDir = ".tickets"
