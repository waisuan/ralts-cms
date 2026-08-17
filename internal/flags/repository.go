package flags

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for machine flags.
//
//go:generate mockgen -destination=mock_repository.go -package=flags -source=repository.go
type Repository interface {
	// Create raises a flag on a machine, replacing that machine's open flag if
	// it already has one, since a machine carries at most one. It reports
	// whether a new flag was inserted (false means an open one was rewritten)
	// and sets f.ID and f.CreatedAt from the stored row either way, so a
	// replacement is dated by when it was re-raised.
	Create(ctx context.Context, f *Flag) (bool, error)
	GetByID(ctx context.Context, id string) (*Flag, error)
	ListForMachine(ctx context.Context, serialNumber string, includeResolved bool) ([]*Flag, error)
	ListOpenBySerials(ctx context.Context, serialNumbers []string) (map[string][]*Flag, error)
	// List returns flags across machines, newest first, for the flagged-records
	// page.
	List(ctx context.Context, opts ListOptions) ([]*Flag, error)
	// Count returns how many flags match the options' filters, ignoring limit
	// and offset, so the page can show a total.
	Count(ctx context.Context, opts ListOptions) (int, error)
	// Resolve closes an open flag, stamping resolved_at from the database clock
	// and recording who resolved it plus an optional note about what was done.
	// An empty note is stored as NULL.
	Resolve(ctx context.Context, id string, resolvedBy int64, resolutionNote string) error
}

// flagColumns is the SELECT list every flag query shares. scanFlagFields returns
// scan targets in exactly this order, so the two must be edited together.
const flagColumns = `f.id, f.machine_serial_number, f.reason, COALESCE(f.ppm_status, ''),
		       COALESCE(f.note, ''), f.status, f.created_by, cu.username, f.created_at,
		       f.resolved_by, ru.username, f.resolved_at, COALESCE(f.resolution_note, '')`

// scanFlagFields lists scan destinations matching flagColumns. reason and status
// are scanned as strings because their Go types are named string types.
func scanFlagFields(f *Flag, reason, status *string) []any {
	return []any{
		&f.ID, &f.MachineSerialNumber, reason, &f.PpmStatus,
		&f.Note, status, &f.CreatedBy, &f.CreatedByUsername, &f.CreatedAt,
		&f.ResolvedBy, &f.ResolvedByUsername, &f.ResolvedAt, &f.ResolutionNote,
	}
}

type db struct {
	client *pgxpool.Pool
}

// NewRepository returns a Postgres-backed flag repository.
func NewRepository(client *pgxpool.Pool) Repository {
	return &db{client: client}
}

func (r *db) Create(ctx context.Context, f *Flag) (bool, error) {
	newID := f.ID
	if newID == "" {
		newID = uuid.New().String()
	}
	if f.Status == "" {
		f.Status = StatusOpen
	}

	// The raise time comes from the database clock, the one resolved_at is also
	// stamped from, so a flag's two timestamps are always comparable. A caller
	// that supplies its own time keeps it, which is how tests backdate flags.
	var createdAt *time.Time
	if !f.CreatedAt.IsZero() {
		createdAt = &f.CreatedAt
	}

	// A machine carries at most one open flag (idx_machine_flags_one_open_per_machine),
	// so flagging an already-flagged machine rewrites the open row in place. The
	// conflicting row is open by definition, which is why resolved_by/resolved_at
	// need no clearing here.
	query := `
		INSERT INTO machine_flags (
			id, machine_serial_number, reason, ppm_status, note,
			status, created_by, created_at
		) VALUES (
			$1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7,
			COALESCE($8::timestamptz, NOW())
		)
		ON CONFLICT (machine_serial_number) WHERE status = 'open'
		DO UPDATE SET
			reason = EXCLUDED.reason,
			ppm_status = EXCLUDED.ppm_status,
			note = EXCLUDED.note,
			created_by = EXCLUDED.created_by,
			created_at = EXCLUDED.created_at
		RETURNING id, created_at
	`

	var storedID string
	err := r.client.QueryRow(ctx, query,
		newID, f.MachineSerialNumber, string(f.Reason), f.PpmStatus, f.Note,
		string(f.Status), f.CreatedBy, createdAt,
	).Scan(&storedID, &f.CreatedAt)
	if err != nil {
		return false, fmt.Errorf("failed to create flag: %w", err)
	}

	// The row keeps its original id when the open flag was rewritten, which is
	// also how we know an insert did not happen.
	f.ID = storedID
	return storedID == newID, nil
}

func (r *db) GetByID(ctx context.Context, id string) (*Flag, error) {
	query := `
		SELECT ` + flagColumns + `
		FROM machine_flags f
		LEFT JOIN users cu ON cu.id = f.created_by
		LEFT JOIN users ru ON ru.id = f.resolved_by
		WHERE f.id = $1
	`
	var f Flag
	var reason, status string
	err := r.client.QueryRow(ctx, query, id).Scan(scanFlagFields(&f, &reason, &status)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get flag: %w", err)
	}
	f.Reason = Reason(reason)
	f.Status = Status(status)
	return &f, nil
}

func (r *db) ListForMachine(ctx context.Context, serialNumber string, includeResolved bool) ([]*Flag, error) {
	query := `
		SELECT ` + flagColumns + `
		FROM machine_flags f
		LEFT JOIN users cu ON cu.id = f.created_by
		LEFT JOIN users ru ON ru.id = f.resolved_by
		WHERE f.machine_serial_number = $1
	`
	if !includeResolved {
		query += " AND f.status = 'open'"
	}
	query += " ORDER BY f.created_at DESC"

	rows, err := r.client.Query(ctx, query, serialNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to list flags: %w", err)
	}
	defer rows.Close()

	return scanFlagRows(rows)
}

func (r *db) ListOpenBySerials(ctx context.Context, serialNumbers []string) (map[string][]*Flag, error) {
	if len(serialNumbers) == 0 {
		return map[string][]*Flag{}, nil
	}

	query := `
		SELECT ` + flagColumns + `
		FROM machine_flags f
		LEFT JOIN users cu ON cu.id = f.created_by
		LEFT JOIN users ru ON ru.id = f.resolved_by
		WHERE f.status = 'open' AND f.machine_serial_number = ANY($1)
		ORDER BY f.created_at DESC
	`

	rows, err := r.client.Query(ctx, query, serialNumbers)
	if err != nil {
		return nil, fmt.Errorf("failed to list flags by serials: %w", err)
	}
	defer rows.Close()

	list, err := scanFlagRows(rows)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*Flag, len(serialNumbers))
	for _, f := range list {
		result[f.MachineSerialNumber] = append(result[f.MachineSerialNumber], f)
	}
	return result, nil
}

// listFilters renders the shared WHERE clause for List and Count, appending the
// bound values to args so both can add their own placeholders afterwards.
func listFilters(opts ListOptions, args *[]any) string {
	var conditions []string
	if opts.Status != "" {
		*args = append(*args, string(opts.Status))
		conditions = append(conditions, fmt.Sprintf("f.status = $%d", len(*args)))
	}
	if opts.AssignedUserID != nil {
		*args = append(*args, *opts.AssignedUserID)
		conditions = append(conditions, fmt.Sprintf(`m."assignedUserId" = $%d`, len(*args)))
	}
	if opts.OwnResolutionsUserID != nil {
		*args = append(*args, *opts.OwnResolutionsUserID)
		conditions = append(conditions, fmt.Sprintf("(f.status = 'open' OR f.resolved_by = $%d)", len(*args)))
	}
	if len(conditions) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(conditions, " AND ")
}

func (r *db) List(ctx context.Context, opts ListOptions) ([]*Flag, error) {
	// The machines join is what allows filtering by assignee; it never drops
	// rows, since a flag cannot outlive its machine.
	query := `
		SELECT ` + flagColumns + `
		FROM machine_flags f
		JOIN machines m ON m."serialNumber" = f.machine_serial_number
		LEFT JOIN users cu ON cu.id = f.created_by
		LEFT JOIN users ru ON ru.id = f.resolved_by
	`
	var args []any
	query += listFilters(opts, &args)
	args = append(args, opts.Limit, opts.Offset)
	query += fmt.Sprintf(" ORDER BY f.created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list flags: %w", err)
	}
	defer rows.Close()

	return scanFlagRows(rows)
}

func (r *db) Count(ctx context.Context, opts ListOptions) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM machine_flags f
		JOIN machines m ON m."serialNumber" = f.machine_serial_number
	`
	var args []any
	query += listFilters(opts, &args)

	var count int
	if err := r.client.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count flags: %w", err)
	}
	return count, nil
}

func (r *db) Resolve(ctx context.Context, id string, resolvedBy int64, resolutionNote string) error {
	query := `
		UPDATE machine_flags
		SET status = 'resolved', resolved_by = $1, resolved_at = NOW(),
		    resolution_note = NULLIF($3, '')
		WHERE id = $2 AND status = 'open'
	`

	result, err := r.client.Exec(ctx, query, resolvedBy, id, resolutionNote)
	if err != nil {
		return fmt.Errorf("failed to resolve flag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanFlagRows(rows pgx.Rows) ([]*Flag, error) {
	var results []*Flag
	for rows.Next() {
		var f Flag
		var reason, status string
		if err := rows.Scan(scanFlagFields(&f, &reason, &status)...); err != nil {
			return nil, fmt.Errorf("failed to scan flag: %w", err)
		}
		f.Reason = Reason(reason)
		f.Status = Status(status)
		results = append(results, &f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating flags: %w", err)
	}
	return results, nil
}
