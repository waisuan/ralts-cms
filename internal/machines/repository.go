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
	// PPMStatusOverdue indicates that the PPM (Planned Preventive Maintenance) is overdue
	PPMStatusOverdue PPMStatus = "overdue"
	// PPMStatusDue indicates that the PPM (Planned Preventive Maintenance) is due today
	PPMStatusDue PPMStatus = "due"
	// PPMStatusAlmostDue indicates that the PPM (Planned Preventive Maintenance) is due within 2 weeks
	PPMStatusAlmostDue PPMStatus = "almost_due"
)

// SortOrder defines the sorting order for machine listing
type SortOrder string

const (
	// SortOrderUpdatedAtDesc sorts machines by update date in descending order (most recently updated first)
	SortOrderUpdatedAtDesc SortOrder = "updated_at_desc" // Most recently updated (default)
	// SortOrderUpdatedAtAsc sorts machines by update date in ascending order (least recently updated first)
	SortOrderUpdatedAtAsc SortOrder = "updated_at_asc" // Least recently updated
)

// ListOptions defines the options for listing machines
type ListOptions struct {
	Limit           int32     `json:"limit"`
	Offset          int32     `json:"offset"`
	Sort            SortOrder `json:"sort"`
	PpmStatusFilter PPMStatus `json:"ppm_status_filter"`
}

// DefaultListOptions returns default list options
func DefaultListOptions() *ListOptions {
	return &ListOptions{
		Limit:           50,
		Offset:          0,
		Sort:            SortOrderUpdatedAtDesc,
		PpmStatusFilter: "",
	}
}

// Repository defines the interface for machine data access operations
//
//go:generate mockgen -destination=../machines/mock_machines_repository.go -package=machines -source=repository.go
type Repository interface {
	GetBySerialNumber(ctx context.Context, serialNumber string) (*Machine, error)
	List(ctx context.Context, options *ListOptions) ([]*Machine, error)
	Create(ctx context.Context, machine *Machine) error
	Update(ctx context.Context, machine *Machine) error
	Delete(ctx context.Context, serialNumber string) error
	Count(ctx context.Context) (int, error)
	CountByStatus(ctx context.Context) (int32, int32, int32, error)
	Search(ctx context.Context, query string, options *ListOptions) ([]*Machine, error)
	CountSearch(ctx context.Context, query string, options *ListOptions) (int, error)
}

type db struct {
	client *pgxpool.Pool
}

// NewRepository creates a new machine repository instance with the given database connection
func NewRepository(client *pgxpool.Pool) Repository {
	return &db{
		client: client,
	}
}

func (r *db) GetBySerialNumber(ctx context.Context, serialNumber string) (*Machine, error) {
	query := `
		SELECT id, "serialNumber", COALESCE(customer, ''), COALESCE(state, ''), COALESCE("accountType", ''), 
		       COALESCE(model, ''), COALESCE(status, ''), COALESCE(brand, ''), 
		       COALESCE(district, ''), COALESCE("personInCharge", ''), COALESCE("reportedBy", ''), 
		       COALESCE("additionalNotes", ''), COALESCE(attachment, ''), 
		       COALESCE("tncDate", '0001-01-01'::date), COALESCE("ppmDate", '0001-01-01'::date), "createdAt", "updatedAt"
		FROM machines 
		WHERE "serialNumber" = $1
	`

	var machine Machine
	err := r.client.QueryRow(ctx, query, serialNumber).Scan(
		&machine.ID, &machine.SerialNumber, &machine.Customer, &machine.State,
		&machine.AccountType, &machine.Model, &machine.Status, &machine.Brand,
		&machine.District, &machine.PersonInCharge, &machine.ReportedBy,
		&machine.AdditionalNotes, &machine.Attachment,
		&machine.TncDate, &machine.PpmDate, &machine.CreatedAt, &machine.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("machine not found")
		}

		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	// Set PPM status based on the PPM date
	machine.PpmStatus = string(r.calculatePPMStatus(machine.PpmDate))

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
	case SortOrderUpdatedAtAsc:
		orderByClause = `ORDER BY "updatedAt" ASC`
	case SortOrderUpdatedAtDesc:
		orderByClause = `ORDER BY "updatedAt" DESC`
	default:
		orderByClause = `ORDER BY "updatedAt" DESC` // Default to most recently updated first
	}

	// Build WHERE clause for PPM status filtering
	var whereClause string
	var args []interface{}
	argIndex := 1

	if options.PpmStatusFilter != "" {
		// Filter by specific PPM status
		switch options.PpmStatusFilter {
		case PPMStatusOverdue:
			whereClause = `WHERE "ppmDate" < CURRENT_DATE`
		case PPMStatusDue:
			whereClause = `WHERE "ppmDate" = CURRENT_DATE`
		case PPMStatusAlmostDue:
			whereClause = `WHERE "ppmDate" > CURRENT_DATE AND "ppmDate" <= CURRENT_DATE + INTERVAL '2 weeks'`
		}
	}

	query := fmt.Sprintf(`
		SELECT id, "serialNumber", COALESCE(customer, ''), COALESCE(state, ''), COALESCE("accountType", ''), 
		       COALESCE(model, ''), COALESCE(status, ''), COALESCE(brand, ''), 
		       COALESCE(district, ''), COALESCE("personInCharge", ''), COALESCE("reportedBy", ''), 
		       COALESCE("additionalNotes", ''), COALESCE(attachment, ''), 
		       COALESCE("tncDate", '0001-01-01'::date), COALESCE("ppmDate", '0001-01-01'::date), "createdAt", "updatedAt"
		FROM machines 
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderByClause, argIndex, argIndex+1)

	// Add limit and offset to args
	args = append(args, options.Limit, options.Offset)

	rows, err := r.client.Query(ctx, query, args...)
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
			&machine.AdditionalNotes, &machine.Attachment,
			&machine.TncDate, &machine.PpmDate, &machine.CreatedAt, &machine.UpdatedAt,
		)
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

func (r *db) Create(ctx context.Context, machine *Machine) error {
	machine.SetTimestamps()

	query := `
		INSERT INTO machines (
			"serialNumber", customer, state, "accountType", model, status, brand,
			district, "personInCharge", "reportedBy", "additionalNotes", attachment,
			"tncDate", "ppmDate", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) RETURNING id
	`

	err := r.client.QueryRow(ctx, query,
		machine.SerialNumber, machine.Customer, machine.State, machine.AccountType,
		machine.Model, machine.Status, machine.Brand, machine.District,
		machine.PersonInCharge, machine.ReportedBy, machine.AdditionalNotes,
		machine.Attachment, machine.TncDate, machine.PpmDate,
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
			customer = $1, state = $2, "accountType" = $3, model = $4, status = $5,
			brand = $6, district = $7, "personInCharge" = $8, "reportedBy" = $9,
			"additionalNotes" = $10, attachment = $11,
			"tncDate" = $12, "ppmDate" = $13, "updatedAt" = $14
		WHERE "serialNumber" = $15
	`

	result, err := r.client.Exec(ctx, query,
		machine.Customer, machine.State, machine.AccountType, machine.Model,
		machine.Status, machine.Brand, machine.District, machine.PersonInCharge,
		machine.ReportedBy, machine.AdditionalNotes, machine.Attachment,
		machine.TncDate, machine.PpmDate, machine.UpdatedAt,
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
	query := `DELETE FROM machines WHERE "serialNumber" = $1`

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

func (r *db) CountByStatus(ctx context.Context) (int32, int32, int32, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN "ppmDate" < CURRENT_DATE THEN 1 END) as overdue_count,
			COUNT(CASE WHEN "ppmDate" = CURRENT_DATE THEN 1 END) as due_count,
			COUNT(CASE WHEN "ppmDate" > CURRENT_DATE AND "ppmDate" <= CURRENT_DATE + INTERVAL '2 weeks' THEN 1 END) as almost_due_count
		FROM machines
	`

	var overdueCount, dueCount, almostDueCount int32
	err := r.client.QueryRow(ctx, query).Scan(&overdueCount, &dueCount, &almostDueCount)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to count machines by status: %w", err)
	}

	return overdueCount, dueCount, almostDueCount, nil
}

func (r *db) Search(ctx context.Context, query string, options *ListOptions) ([]*Machine, error) {
	// Use default options if none provided
	if options == nil {
		options = DefaultListOptions()
	}

	// Build the base query with full-text search
	baseQuery := `
		SELECT id, "serialNumber", COALESCE(customer, ''), COALESCE(state, ''), COALESCE("accountType", ''), 
		       COALESCE(model, ''), COALESCE(status, ''), COALESCE(brand, ''), 
		       COALESCE(district, ''), COALESCE("personInCharge", ''), COALESCE("reportedBy", ''), 
		       COALESCE("additionalNotes", ''), COALESCE(attachment, ''), 
		       COALESCE("tncDate", '0001-01-01'::date), COALESCE("ppmDate", '0001-01-01'::date), "createdAt", "updatedAt",
		       ts_rank(search_vector, plainto_tsquery('english', $1)) as rank
		FROM machines 
		WHERE search_vector @@ plainto_tsquery('english', $1)
	`

	// Build WHERE clause for PPM status filtering
	var args []interface{}
	args = append(args, query)
	argIndex := 2

	if options.PpmStatusFilter != "" {
		// Add PPM status filter condition
		switch options.PpmStatusFilter {
		case PPMStatusOverdue:
			baseQuery += ` AND "ppmDate" < CURRENT_DATE`
		case PPMStatusDue:
			baseQuery += ` AND "ppmDate" = CURRENT_DATE`
		case PPMStatusAlmostDue:
			baseQuery += ` AND "ppmDate" > CURRENT_DATE AND "ppmDate" <= CURRENT_DATE + INTERVAL '2 weeks'`
		}
	}

	// Build ORDER BY clause based on sort option
	var orderByClause string
	switch options.Sort {
	case SortOrderUpdatedAtAsc:
		orderByClause = `ORDER BY rank DESC, "updatedAt" ASC`
	case SortOrderUpdatedAtDesc:
		orderByClause = `ORDER BY rank DESC, "updatedAt" DESC`
	default:
		orderByClause = `ORDER BY rank DESC, "updatedAt" DESC`
	}

	// Add pagination
	finalQuery := fmt.Sprintf("%s %s LIMIT $%d OFFSET $%d", baseQuery, orderByClause, argIndex, argIndex+1)
	args = append(args, options.Limit, options.Offset)

	rows, err := r.client.Query(ctx, finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search machines: %w", err)
	}
	defer rows.Close()

	var machines []*Machine
	for rows.Next() {
		var machine Machine
		var rank float32
		err := rows.Scan(
			&machine.ID, &machine.SerialNumber, &machine.Customer, &machine.State,
			&machine.AccountType, &machine.Model, &machine.Status, &machine.Brand,
			&machine.District, &machine.PersonInCharge, &machine.ReportedBy,
			&machine.AdditionalNotes, &machine.Attachment,
			&machine.TncDate, &machine.PpmDate, &machine.CreatedAt, &machine.UpdatedAt,
			&rank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}

		// Set PPM status based on the PPM date
		machine.PpmStatus = string(r.calculatePPMStatus(machine.PpmDate))
		machines = append(machines, &machine)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return machines, nil
}

func (r *db) CountSearch(ctx context.Context, query string, options *ListOptions) (int, error) {
	// Use default options if none provided
	if options == nil {
		options = DefaultListOptions()
	}

	// Build the base query for counting search results
	baseQuery := `
		SELECT COUNT(*)
		FROM machines 
		WHERE search_vector @@ plainto_tsquery('english', $1)
	`

	// Build WHERE clause for PPM status filtering
	var args []interface{}
	args = append(args, query)

	if options.PpmStatusFilter != "" {
		// Add PPM status filter condition
		switch options.PpmStatusFilter {
		case PPMStatusOverdue:
			baseQuery += ` AND "ppmDate" < CURRENT_DATE`
		case PPMStatusDue:
			baseQuery += ` AND "ppmDate" = CURRENT_DATE`
		case PPMStatusAlmostDue:
			baseQuery += ` AND "ppmDate" > CURRENT_DATE AND "ppmDate" <= CURRENT_DATE + INTERVAL '2 weeks'`
		}
	}

	var count int
	err := r.client.QueryRow(ctx, baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count search results: %w", err)
	}

	return count, nil
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
	}
	if diffDays == 0 {
		return PPMStatusDue
	}
	if diffDays <= 14 { // 2 weeks = 14 days
		return PPMStatusAlmostDue
	}
	// More than 2 weeks in the future - return empty status
	return ""
}
