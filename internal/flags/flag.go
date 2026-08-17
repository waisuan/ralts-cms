// Package flags provides persistent flag/mark records on machines.
//
// A flag captures a reason a machine needs attention. Flags are raised by admin
// users with reason missing_values or other. They have an open/resolved
// lifecycle so a badge can be shown on the machine until the flag is resolved
// by an admin or by the machine's assignee.
package flags

import (
	"errors"
	"fmt"
	"time"
	"unicode/utf8"
)

// Reason enumerates why a flag was raised.
type Reason string

const (
	// ReasonRequiresAttention is reserved for system-raised PPM flags. No code
	// path creates one — ValidateManualReason rejects it — but the value is part
	// of the stored vocabulary, so keep it when reading rows.
	ReasonRequiresAttention Reason = "requires_attention"
	// ReasonMissingValues is raised by admins when required data is missing.
	ReasonMissingValues Reason = "missing_values"
	// ReasonOther is raised by admins with a free-text note.
	ReasonOther Reason = "other"
)

// Status enumerates the lifecycle state of a flag.
type Status string

const (
	// StatusOpen indicates the flag is unresolved.
	StatusOpen Status = "open"
	// StatusResolved indicates an admin has resolved the flag.
	StatusResolved Status = "resolved"
)

// MaxNoteLength bounds a flag note, counted in runes. The note is copied into
// the notification sent to the assignee, so an unbounded note would grow every
// row derived from it. The flag form advertises the same limit.
const MaxNoteLength = 500

// ValidateManualReason returns an error unless the reason is one that admins
// can raise manually.
func ValidateManualReason(reason Reason) error {
	switch reason {
	case ReasonMissingValues, ReasonOther:
		return nil
	case ReasonRequiresAttention:
		return errors.New("requires_attention is reserved for the system and cannot be created manually")
	default:
		return errors.New("unknown flag reason")
	}
}

// ValidateNote checks an already-trimmed note against the given reason: it is
// mandatory when the reason is other, since nothing else explains the flag, and
// may never exceed MaxNoteLength.
func ValidateNote(reason Reason, note string) error {
	if reason == ReasonOther && note == "" {
		return ErrNoteRequired
	}
	return checkNoteLength(note)
}

// ValidateResolutionNote checks an already-trimmed resolution note. It is always
// optional — plenty of flags need no explanation once cleared — but is bounded
// like any other note.
func ValidateResolutionNote(note string) error {
	return checkNoteLength(note)
}

func checkNoteLength(note string) error {
	if utf8.RuneCountInString(note) > MaxNoteLength {
		return fmt.Errorf("%w: limit is %d characters", ErrNoteTooLong, MaxNoteLength)
	}
	return nil
}

// Flag is a persisted flag/mark on a machine record.
type Flag struct {
	ID                  string `json:"id"`
	MachineSerialNumber string `json:"machine_serial_number"`
	Reason              Reason `json:"reason"`
	// PpmStatus is the PPM status snapshot at flag creation (only set for
	// requires_attention flags). Empty for manual flags.
	PpmStatus          string     `json:"ppm_status,omitempty"`
	Note               string     `json:"note,omitempty"`
	Status             Status     `json:"status"`
	CreatedBy          *int64     `json:"created_by,omitempty"`
	CreatedByUsername  *string    `json:"created_by_username,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	ResolvedBy         *int64     `json:"resolved_by,omitempty"`
	ResolvedByUsername *string    `json:"resolved_by_username,omitempty"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
	// ResolutionNote is what the resolver optionally said they did about the
	// flag. Always empty while the flag is open.
	ResolutionNote string `json:"resolution_note,omitempty"`
}

// ListOptions parameterises a cross-machine flag listing.
type ListOptions struct {
	// Status filters by lifecycle state. Empty means every status.
	Status Status
	// AssignedUserID, when set, restricts the listing to flags on machines
	// assigned to that user. This is how non-admins are scoped to the flags they
	// are responsible for, which is also exactly the set they may resolve.
	AssignedUserID *int64
	// OwnResolutionsUserID, when set, hides resolved flags that somebody else
	// closed, keeping the listing to open work plus that user's own history.
	// Without it, being assigned a machine would surface every flag closed on it
	// before the assignee had anything to do with it.
	OwnResolutionsUserID *int64
	Limit                int32
	Offset               int32
}

// ErrNotFound is returned when a flag is not found.
var ErrNotFound = errors.New("flag not found")

// ErrNoteRequired is returned when reason=other arrives without a note.
var ErrNoteRequired = errors.New("note is required for reason=other")

// ErrNoteTooLong is returned when a note exceeds MaxNoteLength.
var ErrNoteTooLong = errors.New("note is too long")
