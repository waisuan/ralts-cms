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
	SortOrderWorkOrderDateDesc SortOrder = "work_order_date_desc"
	// SortOrderWorkOrderDateAsc sorts maintenance records by work order date in ascending order (oldest first)
	SortOrderWorkOrderDateAsc SortOrder = "work_order_date_asc"
	// SortOrderUpdatedAtDesc sorts maintenance records by update date in descending order (most recently updated first)
	SortOrderUpdatedAtDesc SortOrder = "updated_at_desc" // Most recently updated (default)
	// SortOrderUpdatedAtAsc sorts maintenance records by update date in ascending order (least recently updated first)
	SortOrderUpdatedAtAsc SortOrder = "updated_at_asc" // Least recently updated
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
		Sort:   SortOrderUpdatedAtDesc,
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
		SELECT id, "serialNumber", "workOrderNumber", COALESCE("workOrderDate", '0001-01-01'::date), 
		       COALESCE("actionTaken", ''), COALESCE("reportedBy", ''), COALESCE("workOrderType", ''), 
		       attachment, "createdAt", "updatedAt"
		FROM maintenance 
		WHERE "serialNumber" = $1 AND "workOrderNumber" = $2
	`

	var maintenance Maintenance
	err := r.client.QueryRow(ctx, query, machineSerialNumber, workOrderNumber).Scan(
		&maintenance.ID, &maintenance.MachineSerialNumber, &maintenance.WorkOrderNumber,
		&maintenance.WorkOrderDate, &maintenance.ActionTaken, &maintenance.ReportedBy,
		&maintenance.WorkOrderType, &maintenance.Attachment, &maintenance.CreatedAt, &maintenance.UpdatedAt,
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
		orderByClause = `ORDER BY "workOrderDate" ASC, "updatedAt" ASC`
	case SortOrderWorkOrderDateDesc:
		orderByClause = `ORDER BY "workOrderDate" DESC, "updatedAt" DESC`
	case SortOrderUpdatedAtAsc:
		orderByClause = `ORDER BY "updatedAt" ASC`
	case SortOrderUpdatedAtDesc:
		orderByClause = `ORDER BY "updatedAt" DESC`
	default:
		orderByClause = `ORDER BY "updatedAt" DESC` // Default to most recently updated first
	}

	query := fmt.Sprintf(`
		SELECT id, "serialNumber", "workOrderNumber", COALESCE("workOrderDate", '0001-01-01'::date), 
		       COALESCE("actionTaken", ''), COALESCE("reportedBy", ''), COALESCE("workOrderType", ''), 
		       attachment, "createdAt", "updatedAt"
		FROM maintenance 
		WHERE "serialNumber" = $1
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
			&maintenance.WorkOrderType, &maintenance.Attachment, &maintenance.CreatedAt, &maintenance.UpdatedAt,
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
			"serialNumber", "workOrderNumber", "workOrderDate", "actionTaken",
			"reportedBy", "workOrderType", attachment, "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id
	`

	err := r.client.QueryRow(ctx, query,
		maintenance.MachineSerialNumber, maintenance.WorkOrderNumber, maintenance.WorkOrderDate,
		maintenance.ActionTaken, maintenance.ReportedBy, maintenance.WorkOrderType,
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
			"workOrderDate" = $1, "actionTaken" = $2, "reportedBy" = $3,
			"workOrderType" = $4, attachment = $5, "updatedAt" = $6
		WHERE "serialNumber" = $7 AND "workOrderNumber" = $8
	`

	result, err := r.client.Exec(ctx, query,
		maintenance.WorkOrderDate, maintenance.ActionTaken, maintenance.ReportedBy,
		maintenance.WorkOrderType, maintenance.Attachment, maintenance.UpdatedAt,
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
	query := `DELETE FROM maintenance WHERE "serialNumber" = $1 AND "workOrderNumber" = $2`

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
	query := `SELECT COUNT(*) FROM maintenance WHERE "serialNumber" = $1`

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
			COUNT(CASE WHEN "workOrderType" = 'Preventive' THEN 1 END) as preventative_count,
			COUNT(CASE WHEN "workOrderType" = 'Corrective' THEN 1 END) as corrective_count,
			COUNT(CASE WHEN "workOrderType" = 'Emergency' THEN 1 END) as emergency_count,
			COUNT(CASE WHEN "workOrderType" = 'Inspection' THEN 1 END) as inspection_count
		FROM maintenance 
		WHERE "serialNumber" = $1
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
		SELECT id, "serialNumber", "workOrderNumber", COALESCE("workOrderDate", '0001-01-01'::date), 
		       COALESCE("actionTaken", ''), COALESCE("reportedBy", ''), COALESCE("workOrderType", ''), 
		       attachment, "createdAt", "updatedAt",
		       ts_rank(search_vector, plainto_tsquery('english', $2)) as rank
		FROM maintenance 
		WHERE "serialNumber" = $1 AND search_vector @@ plainto_tsquery('english', $2)
	`

	// Build ORDER BY clause based on sort option
	var orderByClause string
	switch options.Sort {
	case SortOrderWorkOrderDateAsc:
		orderByClause = `ORDER BY rank DESC, "workOrderDate" ASC, "updatedAt" ASC`
	case SortOrderWorkOrderDateDesc:
		orderByClause = `ORDER BY rank DESC, "workOrderDate" DESC, "updatedAt" DESC`
	case SortOrderUpdatedAtAsc:
		orderByClause = `ORDER BY rank DESC, "updatedAt" ASC`
	case SortOrderUpdatedAtDesc:
		orderByClause = `ORDER BY rank DESC, "updatedAt" DESC`
	default:
		orderByClause = `ORDER BY rank DESC, "updatedAt" DESC`
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
			&maintenance.WorkOrderType, &maintenance.Attachment, &maintenance.CreatedAt, &maintenance.UpdatedAt,
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
		WHERE "serialNumber" = $1 AND search_vector @@ plainto_tsquery('english', $2)
	`

	var count int
	err := r.client.QueryRow(ctx, sqlQuery, machineSerialNumber, searchQuery).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count search results by machine: %w", err)
	}

	return count, nil
}
