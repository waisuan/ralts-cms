package maintenance

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SortOrder defines the sorting order for maintenance listing
type SortOrder string

const (
	// SortOrderWorkOrderDateDesc sorts maintenance records by work order date in descending order (newest first)
	SortOrderWorkOrderDateDesc SortOrder = "work_order_date_desc" // Most recent work order date (default)
	// SortOrderWorkOrderDateAsc sorts maintenance records by work order date in ascending order (oldest first)
	SortOrderWorkOrderDateAsc SortOrder = "work_order_date_asc" // Oldest work order date
	// SortOrderCreatedAtDesc sorts maintenance records by creation date in descending order (newest first)
	SortOrderCreatedAtDesc SortOrder = "created_at_desc" // Most recently created
	// SortOrderCreatedAtAsc sorts maintenance records by creation date in ascending order (oldest first)
	SortOrderCreatedAtAsc SortOrder = "created_at_asc" // Least recently created
)

// ListOptions defines the options for listing maintenance records
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
		Sort:   SortOrderWorkOrderDateDesc,
	}
}

// Repository defines the interface for maintenance data access operations
//
//go:generate mockgen -destination=../maintenance/mock_maintenance_repository.go -package=maintenance -source=repository.go
type Repository interface {
	GetByWorkOrder(ctx context.Context, machineSerialNumber, workOrderNumber string) (*Maintenance, error)
	ListByMachine(ctx context.Context, machineSerialNumber string, options *ListOptions) ([]*Maintenance, error)
	SearchByMachine(ctx context.Context, machineSerialNumber, query string, options *ListOptions) ([]*Maintenance, error)
	CountSearchByMachine(ctx context.Context, machineSerialNumber, query string) (int, error)
	Create(ctx context.Context, maintenance *Maintenance) error
	Update(ctx context.Context, maintenance *Maintenance) error
	Delete(ctx context.Context, machineSerialNumber, workOrderNumber string) error
	Count(ctx context.Context) (int, error)
	CountByMachine(ctx context.Context, machineSerialNumber string) (int, error)
	CountByWorkOrderType(ctx context.Context, machineSerialNumber string) (int, int, int, int, error)
}

type db struct {
	client *pgxpool.Pool
}

// NewRepository creates a new maintenance repository instance with the given database connection
func NewRepository(client *pgxpool.Pool) Repository {
	return &db{
		client: client,
	}
}

func (r *db) GetByWorkOrder(ctx context.Context, machineSerialNumber, workOrderNumber string) (*Maintenance, error) {
	query := `
		SELECT id, machine_serial_number, work_order_number, work_order_date, action_taken,
		       reported_by, worker_order_type, attachment, created_at, updated_at
		FROM maintenance 
		WHERE machine_serial_number = $1 AND work_order_number = $2
	`

	var maintenance Maintenance
	err := r.client.QueryRow(ctx, query, machineSerialNumber, workOrderNumber).Scan(
		&maintenance.ID, &maintenance.MachineSerialNumber, &maintenance.WorkOrderNumber,
		&maintenance.WorkOrderDate, &maintenance.ActionTaken, &maintenance.ReportedBy,
		&maintenance.WorkerOrderType, &maintenance.Attachment, &maintenance.CreatedAt, &maintenance.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("maintenance not found")
		}
		return nil, fmt.Errorf("failed to get maintenance: %w", err)
	}

	return &maintenance, nil
}

func (r *db) ListByMachine(ctx context.Context, machineSerialNumber string, options *ListOptions) ([]*Maintenance, error) {
	// Use default options if none provided
	if options == nil {
		options = DefaultListOptions()
	}

	// Build the ORDER BY clause based on sort option
	var orderByClause string
	switch options.Sort {
	case SortOrderWorkOrderDateAsc:
		orderByClause = "ORDER BY work_order_date ASC, created_at ASC"
	case SortOrderCreatedAtDesc:
		orderByClause = "ORDER BY created_at DESC, work_order_date DESC"
	case SortOrderCreatedAtAsc:
		orderByClause = "ORDER BY created_at ASC, work_order_date ASC"
	case SortOrderWorkOrderDateDesc:
		orderByClause = "ORDER BY work_order_date DESC, created_at DESC"
	default:
		orderByClause = "ORDER BY work_order_date DESC, created_at DESC" // Default to most recent work order first
	}

	query := fmt.Sprintf(`
		SELECT id, machine_serial_number, work_order_number, work_order_date, action_taken,
		       reported_by, worker_order_type, attachment, created_at, updated_at
		FROM maintenance 
		WHERE machine_serial_number = $1
		%s
		LIMIT $2 OFFSET $3
	`, orderByClause)

	rows, err := r.client.Query(ctx, query, machineSerialNumber, options.Limit, options.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query maintenance records: %w", err)
	}
	defer rows.Close()

	var maintenanceRecords []*Maintenance
	for rows.Next() {
		var maintenance Maintenance
		err := rows.Scan(
			&maintenance.ID, &maintenance.MachineSerialNumber, &maintenance.WorkOrderNumber,
			&maintenance.WorkOrderDate, &maintenance.ActionTaken, &maintenance.ReportedBy,
			&maintenance.WorkerOrderType, &maintenance.Attachment, &maintenance.CreatedAt, &maintenance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance: %w", err)
		}
		maintenanceRecords = append(maintenanceRecords, &maintenance)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating maintenance records: %w", err)
	}

	return maintenanceRecords, nil
}

func (r *db) Create(ctx context.Context, maintenance *Maintenance) error {
	maintenance.SetTimestamps()

	query := `
		INSERT INTO maintenance (
			machine_serial_number, work_order_number, work_order_date, action_taken,
			reported_by, worker_order_type, attachment, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id
	`

	err := r.client.QueryRow(ctx, query,
		maintenance.MachineSerialNumber, maintenance.WorkOrderNumber, maintenance.WorkOrderDate,
		maintenance.ActionTaken, maintenance.ReportedBy, maintenance.WorkerOrderType,
		maintenance.Attachment, maintenance.CreatedAt, maintenance.UpdatedAt,
	).Scan(&maintenance.ID)

	if err != nil {
		return fmt.Errorf("failed to create maintenance: %w", err)
	}

	return nil
}

func (r *db) Update(ctx context.Context, maintenance *Maintenance) error {
	maintenance.SetTimestamps()

	query := `
		UPDATE maintenance SET
			work_order_date = $1, action_taken = $2, reported_by = $3,
			worker_order_type = $4, attachment = $5, updated_at = $6
		WHERE machine_serial_number = $7 AND work_order_number = $8
	`

	result, err := r.client.Exec(ctx, query,
		maintenance.WorkOrderDate, maintenance.ActionTaken, maintenance.ReportedBy,
		maintenance.WorkerOrderType, maintenance.Attachment, maintenance.UpdatedAt,
		maintenance.MachineSerialNumber, maintenance.WorkOrderNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to update maintenance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("maintenance not found")
	}

	return nil
}

func (r *db) Delete(ctx context.Context, machineSerialNumber, workOrderNumber string) error {
	query := `DELETE FROM maintenance WHERE machine_serial_number = $1 AND work_order_number = $2`

	result, err := r.client.Exec(ctx, query, machineSerialNumber, workOrderNumber)
	if err != nil {
		return fmt.Errorf("failed to delete maintenance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("maintenance not found")
	}

	return nil
}

func (r *db) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM maintenance`

	var count int
	err := r.client.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count maintenance records: %w", err)
	}

	return count, nil
}

func (r *db) CountByMachine(ctx context.Context, machineSerialNumber string) (int, error) {
	query := `SELECT COUNT(*) FROM maintenance WHERE machine_serial_number = $1`

	var count int
	err := r.client.QueryRow(ctx, query, machineSerialNumber).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count maintenance records for machine: %w", err)
	}

	return count, nil
}

func (r *db) CountByWorkOrderType(ctx context.Context, machineSerialNumber string) (int, int, int, int, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN worker_order_type = 'Preventive' THEN 1 END) as preventative_count,
			COUNT(CASE WHEN worker_order_type = 'Corrective' THEN 1 END) as corrective_count,
			COUNT(CASE WHEN worker_order_type = 'Emergency' THEN 1 END) as emergency_count,
			COUNT(CASE WHEN worker_order_type = 'Inspection' THEN 1 END) as inspection_count
		FROM maintenance 
		WHERE machine_serial_number = $1
	`

	var preventativeCount, correctiveCount, emergencyCount, inspectionCount int
	err := r.client.QueryRow(ctx, query, machineSerialNumber).Scan(
		&preventativeCount, &correctiveCount, &emergencyCount, &inspectionCount,
	)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to count maintenance records by work order type: %w", err)
	}

	return preventativeCount, correctiveCount, emergencyCount, inspectionCount, nil
}

func (r *db) SearchByMachine(ctx context.Context, machineSerialNumber, query string, options *ListOptions) ([]*Maintenance, error) {
	// Use default options if none provided
	if options == nil {
		options = DefaultListOptions()
	}

	// Build the base query with full-text search filtered by machine
	baseQuery := `
		SELECT id, machine_serial_number, work_order_number, work_order_date, action_taken,
		       reported_by, worker_order_type, attachment, created_at, updated_at,
		       ts_rank(search_vector, plainto_tsquery('english', $2)) as rank
		FROM maintenance 
		WHERE machine_serial_number = $1 AND search_vector @@ plainto_tsquery('english', $2)
	`

	// Build ORDER BY clause based on sort option
	var orderByClause string
	switch options.Sort {
	case SortOrderWorkOrderDateAsc:
		orderByClause = "ORDER BY rank DESC, work_order_date ASC, created_at ASC"
	case SortOrderCreatedAtDesc:
		orderByClause = "ORDER BY rank DESC, created_at DESC, work_order_date DESC"
	case SortOrderCreatedAtAsc:
		orderByClause = "ORDER BY rank DESC, created_at ASC, work_order_date ASC"
	case SortOrderWorkOrderDateDesc:
		orderByClause = "ORDER BY rank DESC, work_order_date DESC, created_at DESC"
	default:
		orderByClause = "ORDER BY rank DESC, work_order_date DESC, created_at DESC"
	}

	// Add pagination
	finalQuery := fmt.Sprintf("%s %s LIMIT $3 OFFSET $4", baseQuery, orderByClause)

	rows, err := r.client.Query(ctx, finalQuery, machineSerialNumber, query, options.Limit, options.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search maintenance records by machine: %w", err)
	}
	defer rows.Close()

	var maintenanceRecords []*Maintenance
	for rows.Next() {
		var maintenance Maintenance
		var rank float32
		err := rows.Scan(
			&maintenance.ID, &maintenance.MachineSerialNumber, &maintenance.WorkOrderNumber,
			&maintenance.WorkOrderDate, &maintenance.ActionTaken, &maintenance.ReportedBy,
			&maintenance.WorkerOrderType, &maintenance.Attachment, &maintenance.CreatedAt, &maintenance.UpdatedAt,
			&rank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance: %w", err)
		}
		maintenanceRecords = append(maintenanceRecords, &maintenance)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating maintenance records: %w", err)
	}

	return maintenanceRecords, nil
}

func (r *db) CountSearchByMachine(ctx context.Context, machineSerialNumber, searchQuery string) (int, error) {
	sqlQuery := `
		SELECT COUNT(*)
		FROM maintenance 
		WHERE machine_serial_number = $1 AND search_vector @@ plainto_tsquery('english', $2)
	`

	var count int
	err := r.client.QueryRow(ctx, sqlQuery, machineSerialNumber, searchQuery).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count search results by machine: %w", err)
	}

	return count, nil
}
