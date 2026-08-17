package flags

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"ralts-cms/internal/machines"
	"ralts-cms/internal/notifications"
)

// Service is the interface exposed to handlers and the deps container.
//
//go:generate mockgen -destination=mock_service.go -package=flags -source=service.go
type Service interface {
	// CreateManualFlag records an admin-raised flag (missing_values or other)
	// and notifies the machine's current assignee (if any). A machine holds one
	// open flag at a time, so flagging an already-flagged machine replaces its
	// reason and note; the returned bool reports whether the flag is new.
	CreateManualFlag(ctx context.Context, actorUserID int64, serialNumber string, reason Reason, note string) (*Flag, bool, error)
	// Resolve marks a flag as resolved, recording an optional note about what
	// was done. Authorisation is the caller's responsibility — see CanResolve.
	// isAdmin decides whether the admin who raised the flag hears about it:
	// resolutions by assignees are reported back, admin ones are not.
	Resolve(ctx context.Context, actorUserID int64, isAdmin bool, id string, resolutionNote string) error
	// CanResolve reports whether the given user may resolve the given flag.
	// Admins may resolve any flag; other users may only resolve flags on
	// machines currently assigned to them (the flags they were notified about).
	CanResolve(ctx context.Context, userID int64, isAdmin bool, flagID string) (bool, error)
}

type service struct {
	repo            Repository
	machinesRepo    machines.Repository
	notificationSvc notifications.Service
	logger          *slog.Logger
}

// NewService constructs a flags service.
func NewService(
	repo Repository,
	machinesRepo machines.Repository,
	notificationSvc notifications.Service,
	logger *slog.Logger,
) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &service{
		repo:            repo,
		machinesRepo:    machinesRepo,
		notificationSvc: notificationSvc,
		logger:          logger,
	}
}

func (s *service) CreateManualFlag(ctx context.Context, actorUserID int64, serialNumber string, reason Reason, note string) (*Flag, bool, error) {
	if err := ValidateManualReason(reason); err != nil {
		return nil, false, err
	}
	// Trim before validating so whitespace can neither stand in for a required
	// note nor push a note over the limit.
	note = strings.TrimSpace(note)
	if err := ValidateNote(reason, note); err != nil {
		return nil, false, err
	}

	// Confirm the machine exists and capture its current assignee for notification.
	machine, err := s.machinesRepo.GetBySerialNumber(ctx, serialNumber)
	if err != nil {
		return nil, false, fmt.Errorf("failed to load machine: %w", err)
	}

	created := &Flag{
		MachineSerialNumber: serialNumber,
		Reason:              reason,
		Note:                note,
		Status:              StatusOpen,
		CreatedBy:           &actorUserID,
	}
	isNew, err := s.repo.Create(ctx, created)
	if err != nil {
		return nil, false, err
	}

	// Rewriting an open flag is a fresh request for attention, so the assignee
	// hears about it just as they would about a first flag.
	s.notifyAssigneeOfFlag(ctx, machine, created, &actorUserID)

	return created, isNew, nil
}

func (s *service) Resolve(ctx context.Context, actorUserID int64, isAdmin bool, id string, resolutionNote string) error {
	resolutionNote = strings.TrimSpace(resolutionNote)
	if err := ValidateResolutionNote(resolutionNote); err != nil {
		return err
	}

	// Whoever raised the flag is only knowable while it is still readable as the
	// flag being closed, since resolving reports nothing back. Admins clearing
	// their own queue need no follow-up, so the read is skipped for them.
	var raised *Flag
	if !isAdmin {
		loaded, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		raised = loaded
	}

	if err := s.repo.Resolve(ctx, id, actorUserID, resolutionNote); err != nil {
		return err
	}

	if raised != nil {
		s.notifyRaiserOfResolution(ctx, raised, actorUserID, resolutionNote)
	}
	return nil
}

func (s *service) CanResolve(ctx context.Context, userID int64, isAdmin bool, flagID string) (bool, error) {
	if isAdmin {
		return true, nil
	}

	flag, err := s.repo.GetByID(ctx, flagID)
	if err != nil {
		return false, err
	}

	machine, err := s.machinesRepo.GetBySerialNumber(ctx, flag.MachineSerialNumber)
	if err != nil {
		return false, fmt.Errorf("failed to load machine for flag: %w", err)
	}

	// Free-text ("ghost") assignees have no user id, so nobody but an admin can
	// resolve flags on those machines.
	if machine.AssignedUserID == nil {
		return false, nil
	}
	return *machine.AssignedUserID == userID, nil
}

// notifyAssigneeOfFlag sends a notification to the machine's assignee describing
// the flag. Best effort: the notification service itself never blocks/errors.
// Machines with a free-text assignee have no user id and so are skipped.
func (s *service) notifyAssigneeOfFlag(ctx context.Context, machine *machines.Machine, flag *Flag, actorID *int64) {
	if machine == nil || machine.AssignedUserID == nil {
		return
	}
	if s.notificationSvc == nil {
		return
	}

	title := titleForFlag(machine, flag)
	body := flag.Note
	serial := machine.SerialNumber
	id := flag.ID

	s.notificationSvc.Notify(ctx, notifications.Notification{
		UserID:              *machine.AssignedUserID,
		Type:                notifications.TypeFlagged,
		MachineSerialNumber: &serial,
		FlagID:              &id,
		Title:               title,
		Body:                body,
		ActorUserID:         actorID,
	})
}

// notifyRaiserOfResolution tells the admin who raised a flag that it has been
// dealt with, carrying whatever the resolver said they did. Flags raised before
// notifications existed, or by a since-deleted account, have no recipient and
// are skipped; a raiser resolving their own flag is dropped downstream.
func (s *service) notifyRaiserOfResolution(ctx context.Context, flag *Flag, actorUserID int64, resolutionNote string) {
	if s.notificationSvc == nil || flag.CreatedBy == nil {
		return
	}

	serial := flag.MachineSerialNumber
	id := flag.ID

	s.notificationSvc.Notify(ctx, notifications.Notification{
		UserID:              *flag.CreatedBy,
		Type:                notifications.TypeFlagResolved,
		MachineSerialNumber: &serial,
		FlagID:              &id,
		Title:               fmt.Sprintf("Machine %s: flag resolved", serial),
		Body:                resolutionNote,
		ActorUserID:         &actorUserID,
	})
}

// titleForFlag composes a short human-readable notification title for the
// given flag+machine combination.
func titleForFlag(machine *machines.Machine, flag *Flag) string {
	switch flag.Reason {
	case ReasonRequiresAttention:
		if flag.PpmStatus != "" {
			return fmt.Sprintf("Machine %s: PPM %s", machine.SerialNumber, humanPPMStatus(flag.PpmStatus))
		}
		return fmt.Sprintf("Machine %s requires attention", machine.SerialNumber)
	case ReasonMissingValues:
		return fmt.Sprintf("Machine %s flagged: missing values", machine.SerialNumber)
	case ReasonOther:
		return fmt.Sprintf("Machine %s flagged", machine.SerialNumber)
	default:
		return fmt.Sprintf("Machine %s flagged", machine.SerialNumber)
	}
}

func humanPPMStatus(s string) string {
	switch s {
	case string(machines.PPMStatusOverdue):
		return "overdue"
	case string(machines.PPMStatusDue):
		return "due today"
	case string(machines.PPMStatusAlmostDue):
		return "due soon"
	default:
		return s
	}
}
