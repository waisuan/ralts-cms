package machines

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PPMStatus defines the PPM status values
type PPMStatus string

const (
	PPMStatusOverdue   PPMStatus = "overdue"
	PPMStatusDue       PPMStatus = "due"
	PPMStatusAlmostDue PPMStatus = "almost_due"
)

// SortOrder defines the sorting order for machine listing
type SortOrder string

const (
	SortOrderCreatedAtDesc SortOrder = "created_at_desc" // Most recently created (default)
	SortOrderCreatedAtAsc  SortOrder = "created_at_asc"  // Least recently created
)

// ListOptions defines the options for listing machines
type ListOptions struct {
	Limit  int32     `json:"limit"`
	Offset int32     `json:"offset"`
	Sort   SortOrder `json:"sort"`
}

// DefaultListOptions returns default list options
func DefaultListOptions() *ListOptions {
	return &ListOptions{
		Limit:  50,
		Offset: 0,
		Sort:   SortOrderCreatedAtDesc,
	}
}

//go:generate mockgen -destination=../machines/mock_machines_repository.go -package=machines -source=repository.go
type Repository interface {
	GetBySerialNumber(ctx context.Context, serialNumber string) (*Machine, error)
	List(ctx context.Context, options *ListOptions) ([]*Machine, error)
	Create(ctx context.Context, machine *Machine) error
	Update(ctx context.Context, machine *Machine) error
	Delete(ctx context.Context, serialNumber string) error
	Count(ctx context.Context) (int, error)
	DuePPM(ctx context.Context) ([]*Machine, error)
}

type db struct {
	client *pgxpool.Pool
}

func NewRepository(client *pgxpool.Pool) Repository {
	return &db{
		client: client,
	}
}

func (r *db) GetBySerialNumber(ctx context.Context, serialNumber string) (*Machine, error) {
	query := `
		SELECT id, serial_number, customer, state, account_type, model, status, brand, 
		       district, person_in_charge, reported_by, additional_notes, attachment, 
		       ppm_status, tnc_date, ppm_date, created_at, updated_at
		FROM machines 
		WHERE serial_number = $1
	`

	var machine Machine
	err := r.client.QueryRow(ctx, query, serialNumber).Scan(
		&machine.ID, &machine.SerialNumber, &machine.Customer, &machine.State,
		&machine.AccountType, &machine.Model, &machine.Status, &machine.Brand,
		&machine.District, &machine.PersonInCharge, &machine.ReportedBy,
		&machine.AdditionalNotes, &machine.Attachment, &machine.PpmStatus,
		&machine.TncDate, &machine.PpmDate, &machine.CreatedAt, &machine.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("machine not found")
		}
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	return &machine, nil
}

func (r *db) List(ctx context.Context, options *ListOptions) ([]*Machine, error) {
	// Use default options if none provided
	if options == nil {
		options = DefaultListOptions()
	}

	// Build the ORDER BY clause based on sort option
	var orderByClause string
	switch options.Sort {
	case SortOrderCreatedAtAsc:
		orderByClause = "ORDER BY created_at ASC"
	case SortOrderCreatedAtDesc:
		orderByClause = "ORDER BY created_at DESC"
	default:
		orderByClause = "ORDER BY created_at DESC" // Default to most recent first
	}

	query := fmt.Sprintf(`
		SELECT id, serial_number, customer, state, account_type, model, status, brand, 
		       district, person_in_charge, reported_by, additional_notes, attachment, 
		       ppm_status, tnc_date, ppm_date, created_at, updated_at
		FROM machines 
		%s
		LIMIT $%d OFFSET $%d
	`, orderByClause, 1, 2)

	rows, err := r.client.Query(ctx, query, options.Limit, options.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query machines: %w", err)
	}
	defer rows.Close()

	var machines []*Machine
	for rows.Next() {
		var machine Machine
		err := rows.Scan(
			&machine.ID, &machine.SerialNumber, &machine.Customer, &machine.State,
			&machine.AccountType, &machine.Model, &machine.Status, &machine.Brand,
			&machine.District, &machine.PersonInCharge, &machine.ReportedBy,
			&machine.AdditionalNotes, &machine.Attachment, &machine.PpmStatus,
			&machine.TncDate, &machine.PpmDate, &machine.CreatedAt, &machine.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}
		machines = append(machines, &machine)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating machines: %w", err)
	}

	return machines, nil
}

func (r *db) Create(ctx context.Context, machine *Machine) error {
	machine.SetTimestamps()

	query := `
		INSERT INTO machines (
			serial_number, customer, state, account_type, model, status, brand,
			district, person_in_charge, reported_by, additional_notes, attachment,
			ppm_status, tnc_date, ppm_date, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		) RETURNING id
	`

	err := r.client.QueryRow(ctx, query,
		machine.SerialNumber, machine.Customer, machine.State, machine.AccountType,
		machine.Model, machine.Status, machine.Brand, machine.District,
		machine.PersonInCharge, machine.ReportedBy, machine.AdditionalNotes,
		machine.Attachment, machine.PpmStatus, machine.TncDate, machine.PpmDate,
		machine.CreatedAt, machine.UpdatedAt,
	).Scan(&machine.ID)

	if err != nil {
		return fmt.Errorf("failed to create machine: %w", err)
	}

	return nil
}

func (r *db) Update(ctx context.Context, machine *Machine) error {
	machine.SetTimestamps()

	query := `
		UPDATE machines SET
			customer = $1, state = $2, account_type = $3, model = $4, status = $5,
			brand = $6, district = $7, person_in_charge = $8, reported_by = $9,
			additional_notes = $10, attachment = $11, ppm_status = $12,
			tnc_date = $13, ppm_date = $14, updated_at = $15
		WHERE serial_number = $16
	`

	result, err := r.client.Exec(ctx, query,
		machine.Customer, machine.State, machine.AccountType, machine.Model,
		machine.Status, machine.Brand, machine.District, machine.PersonInCharge,
		machine.ReportedBy, machine.AdditionalNotes, machine.Attachment,
		machine.PpmStatus, machine.TncDate, machine.PpmDate, machine.UpdatedAt,
		machine.SerialNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to update machine: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("machine not found")
	}

	return nil
}

func (r *db) Delete(ctx context.Context, serialNumber string) error {
	query := `DELETE FROM machines WHERE serial_number = $1`

	result, err := r.client.Exec(ctx, query, serialNumber)
	if err != nil {
		return fmt.Errorf("failed to delete machine: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("machine not found")
	}

	return nil
}

func (r *db) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM machines`

	var count int
	err := r.client.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count machines: %w", err)
	}

	return count, nil
}

func (r *db) DuePPM(ctx context.Context) ([]*Machine, error) {
	query := `SELECT * FROM machines WHERE ppm_date <= CURRENT_DATE OR ppm_date <= CURRENT_DATE + INTERVAL '2 weeks'`

	rows, err := r.client.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query machines: %w", err)
	}

	var machines []*Machine
	for rows.Next() {
		var machine Machine
		err := rows.Scan(&machine.ID, &machine.SerialNumber, &machine.Customer, &machine.State,
			&machine.AccountType, &machine.Model, &machine.Status, &machine.Brand,
			&machine.District, &machine.PersonInCharge, &machine.ReportedBy,
			&machine.AdditionalNotes, &machine.Attachment, &machine.PpmStatus,
			&machine.TncDate, &machine.PpmDate, &machine.CreatedAt, &machine.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}

		// Set PPM status based on the PPM date
		machine.PpmStatus = string(r.calculatePPMStatus(machine.PpmDate))

		machines = append(machines, &machine)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating machines: %w", err)
	}

	return machines, nil
}

// calculatePPMStatus determines the PPM status based on the PPM date
func (r *db) calculatePPMStatus(ppmDate time.Time) PPMStatus {
	now := time.Now().UTC()

	// Remove time components for accurate day comparison
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	ppmDate = time.Date(ppmDate.Year(), ppmDate.Month(), ppmDate.Day(), 0, 0, 0, 0, time.UTC)

	diffDays := int(ppmDate.Sub(now).Hours() / 24)

	if diffDays < 0 {
		return PPMStatusOverdue
	} else if diffDays == 0 {
		return PPMStatusDue
	} else if diffDays <= 14 { // 2 weeks = 14 days
		return PPMStatusAlmostDue
	} else {
		// More than 2 weeks in the future - return empty status
		return ""
	}
}
