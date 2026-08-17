package machines

import (
	"context"
	"fmt"
	"strings"
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
	// SortOrderPpmDateAsc sorts machines by PPM date in ascending order (earliest PPM date first)
	SortOrderPpmDateAsc SortOrder = "ppm_date_asc"
	// SortOrderPpmDateDesc sorts machines by PPM date in descending order (latest PPM date first)
	SortOrderPpmDateDesc SortOrder = "ppm_date_desc"
	// SortOrderTncDateAsc sorts machines by TNC date in ascending order (earliest TNC date first)
	SortOrderTncDateAsc SortOrder = "tnc_date_asc"
	// SortOrderTncDateDesc sorts machines by TNC date in descending order (latest TNC date first)
	SortOrderTncDateDesc SortOrder = "tnc_date_desc"
)

// ListOptions defines the options for listing machines
type ListOptions struct {
	Limit           int32     `json:"limit"`
	Offset          int32     `json:"offset"`
	Sort            SortOrder `json:"sort"`
	PpmStatusFilter PPMStatus `json:"ppm_status_filter"`
	// Date range filters (inclusive, nil means no filter)
	PpmDateFrom *time.Time `json:"ppm_date_from"`
	PpmDateTo   *time.Time `json:"ppm_date_to"`
	TncDateFrom *time.Time `json:"tnc_date_from"`
	TncDateTo   *time.Time `json:"tnc_date_to"`
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

// filterResult holds the result of building filter conditions
type filterResult struct {
	conditions []string
	args       []interface{}
	nextIndex  int
}

// buildFilterConditions builds WHERE clause conditions from ListOptions
// startIndex is the starting parameter index (e.g., 1 for List, 2 for Search which uses $1 for query)
// Returns conditions slice, args slice, and the next available parameter index
func buildFilterConditions(options *ListOptions, startIndex int) filterResult {
	return buildFilterConditionsWithAlias(options, startIndex, "m")
}

// buildFilterConditionsWithAlias mirrors buildFilterConditions but prefixes
// column references with the given table alias. Pass an empty string to omit
// the prefix (e.g. for the un-aliased Count query).
func buildFilterConditionsWithAlias(options *ListOptions, startIndex int, alias string) filterResult {
	var conditions []string
	var args []interface{}
	argIndex := startIndex

	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}

	if options.PpmStatusFilter != "" {
		if cond := getPpmStatusConditionWithAlias(options.PpmStatusFilter, alias); cond != "" {
			conditions = append(conditions, cond)
		}
	}

	if options.PpmDateFrom != nil {
		conditions = append(conditions, fmt.Sprintf(`%s"ppmDate" >= $%d`, prefix, argIndex))
		args = append(args, *options.PpmDateFrom)
		argIndex++
	}
	if options.PpmDateTo != nil {
		conditions = append(conditions, fmt.Sprintf(`%s"ppmDate" <= $%d`, prefix, argIndex))
		args = append(args, *options.PpmDateTo)
		argIndex++
	}

	if options.TncDateFrom != nil {
		conditions = append(conditions, fmt.Sprintf(`%s"tncDate" >= $%d`, prefix, argIndex))
		args = append(args, *options.TncDateFrom)
		argIndex++
	}
	if options.TncDateTo != nil {
		conditions = append(conditions, fmt.Sprintf(`%s"tncDate" <= $%d`, prefix, argIndex))
		args = append(args, *options.TncDateTo)
		argIndex++
	}

	return filterResult{
		conditions: conditions,
		args:       args,
		nextIndex:  argIndex,
	}
}

// Sentinel PPM dates (NULL or calendar year ≤ 1) must match IsPpmDateUnset / calculatePPMStatus.
const sqlPpmDateEligibleForStatus = `"ppmDate" IS NOT NULL AND (EXTRACT(YEAR FROM "ppmDate"))::int > 1`

var sqlPpmStatusExtra = map[PPMStatus]string{
	PPMStatusOverdue:   `"ppmDate" < CURRENT_DATE`,
	PPMStatusDue:       `"ppmDate" = CURRENT_DATE`,
	PPMStatusAlmostDue: `"ppmDate" > CURRENT_DATE AND "ppmDate" <= CURRENT_DATE + INTERVAL '2 weeks'`,
}

// getPpmStatusCondition returns the SQL WHERE fragment for filtering by computed PPM bucket.
func getPpmStatusCondition(status PPMStatus) string {
	return getPpmStatusConditionWithAlias(status, "")
}

// getPpmStatusConditionWithAlias mirrors getPpmStatusCondition but prefixes the
// column reference with the given table alias when non-empty.
func getPpmStatusConditionWithAlias(status PPMStatus, alias string) string {
	extra, ok := sqlPpmStatusExtra[status]
	if !ok {
		return ""
	}
	base := sqlPpmDateEligibleForStatus + ` AND ` + extra
	if alias == "" {
		return base
	}
	prefix := alias + "."
	return strings.ReplaceAll(base, `"ppmDate"`, prefix+`"ppmDate"`)
}

// buildOrderByClause returns the ORDER BY clause for the given sort option
// If includeRank is true, rank DESC is prepended (used for search results)
func buildOrderByClause(sort SortOrder, includeRank bool) string {
	var column string
	var direction string

	switch sort {
	case SortOrderUpdatedAtAsc:
		column, direction = `m."updatedAt"`, "ASC"
	case SortOrderUpdatedAtDesc:
		column, direction = `m."updatedAt"`, "DESC"
	case SortOrderPpmDateAsc:
		column, direction = `m."ppmDate"`, "ASC"
	case SortOrderPpmDateDesc:
		column, direction = `m."ppmDate"`, "DESC"
	case SortOrderTncDateAsc:
		column, direction = `m."tncDate"`, "ASC"
	case SortOrderTncDateDesc:
		column, direction = `m."tncDate"`, "DESC"
	default:
		column, direction = `m."updatedAt"`, "DESC"
	}

	if includeRank {
		return fmt.Sprintf("ORDER BY rank DESC, %s %s", column, direction)
	}
	return fmt.Sprintf("ORDER BY %s %s", column, direction)
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
	// Count returns the number of machines matching the same filters as List (omit limit/offset).
	Count(ctx context.Context, options *ListOptions) (int, error)
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

// selectMachineColumns is the shared column list for machine reads. It joins
// users to project a slim assignee summary so callers don't need a second
// query. The order here must match the Scan calls below and in scanMachineRow.
const selectMachineColumns = `
	m.id, m."serialNumber", COALESCE(m.customer, ''), COALESCE(m.state, ''), COALESCE(m."accountType", ''),
	COALESCE(m.model, ''), COALESCE(m.status, ''), COALESCE(m.brand, ''),
	COALESCE(m.district, ''), COALESCE(m."personInCharge", ''), COALESCE(m."reportedBy", ''),
	COALESCE(m."additionalNotes", ''), COALESCE(m.attachment, ''),
	COALESCE(m."tncDate", '0001-01-01'::date), COALESCE(m."ppmDate", '0001-01-01'::date), m."createdAt", m."updatedAt",
	COALESCE(m."updatedBy", ''),
	m."assignedUserId", u.username, u.email
`

// scanMachineRow scans a row that projects selectMachineColumns into a Machine
// and populates the embedded AssignedUser when the FK is set.
func scanMachineRow(row pgx.Row, machine *Machine, extra ...any) error {
	var assignedUserID *int64
	var assigneeUsername, assigneeEmail *string
	dest := []any{
		&machine.ID, &machine.SerialNumber, &machine.Customer, &machine.State,
		&machine.AccountType, &machine.Model, &machine.Status, &machine.Brand,
		&machine.District, &machine.PersonInCharge, &machine.ReportedBy,
		&machine.AdditionalNotes, &machine.Attachment,
		&machine.TncDate, &machine.PpmDate, &machine.CreatedAt, &machine.UpdatedAt,
		&machine.UpdatedBy,
		&assignedUserID, &assigneeUsername, &assigneeEmail,
	}
	dest = append(dest, extra...)
	if err := row.Scan(dest...); err != nil {
		return err
	}
	machine.AssignedUserID = assignedUserID
	if assignedUserID != nil {
		au := &AssignedUser{ID: *assignedUserID}
		if assigneeUsername != nil {
			au.Username = *assigneeUsername
		}
		if assigneeEmail != nil {
			au.Email = *assigneeEmail
		}
		machine.AssignedUser = au
	}
	return nil
}

func (r *db) GetBySerialNumber(ctx context.Context, serialNumber string) (*Machine, error) {
	query := `
		SELECT ` + selectMachineColumns + `
		FROM machines m
		LEFT JOIN users u ON u.id = m."assignedUserId"
		WHERE m."serialNumber" = $1
	`

	var machine Machine
	err := scanMachineRow(r.client.QueryRow(ctx, query, serialNumber), &machine)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("%w", ErrNotFound)
		}

		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	machine.PpmStatus = string(r.calculatePPMStatus(machine.PpmDate))

	return &machine, nil
}

func (r *db) List(ctx context.Context, options *ListOptions) ([]*Machine, error) {
	if options == nil {
		options = DefaultListOptions()
	}

	orderByClause := buildOrderByClause(options.Sort, false)
	filter := buildFilterConditions(options, 1)

	var whereClause string
	if len(filter.conditions) > 0 {
		whereClause = "WHERE " + strings.Join(filter.conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT `+selectMachineColumns+`
		FROM machines m
		LEFT JOIN users u ON u.id = m."assignedUserId"
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderByClause, filter.nextIndex, filter.nextIndex+1)

	args := append(filter.args, options.Limit, options.Offset)

	rows, err := r.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query machines: %w", err)
	}
	defer rows.Close()

	var machines []*Machine
	for rows.Next() {
		var machine Machine
		if err := scanMachineRow(rows, &machine); err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}
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
			"tncDate", "ppmDate", "createdAt", "updatedAt", "updatedBy", "assignedUserId"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		) RETURNING id
	`

	err := r.client.QueryRow(ctx, query,
		machine.SerialNumber, machine.Customer, machine.State, machine.AccountType,
		machine.Model, machine.Status, machine.Brand, machine.District,
		machine.PersonInCharge, machine.ReportedBy, machine.AdditionalNotes,
		machine.Attachment, machine.TncDate, machine.PpmDate,
		machine.CreatedAt, machine.UpdatedAt, machine.UpdatedBy, machine.AssignedUserID,
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
			"tncDate" = $12, "ppmDate" = $13, "updatedAt" = $14, "updatedBy" = $15,
			"assignedUserId" = $16
		WHERE "serialNumber" = $17
	`

	result, err := r.client.Exec(ctx, query,
		machine.Customer, machine.State, machine.AccountType, machine.Model,
		machine.Status, machine.Brand, machine.District, machine.PersonInCharge,
		machine.ReportedBy, machine.AdditionalNotes, machine.Attachment,
		machine.TncDate, machine.PpmDate, machine.UpdatedAt, machine.UpdatedBy,
		machine.AssignedUserID, machine.SerialNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to update machine: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w", ErrNotFound)
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
		return fmt.Errorf("%w", ErrNotFound)
	}

	return nil
}

func (r *db) Count(ctx context.Context, options *ListOptions) (int, error) {
	if options == nil {
		options = DefaultListOptions()
	}

	filter := buildFilterConditions(options, 1)
	var whereClause string
	if len(filter.conditions) > 0 {
		whereClause = "WHERE " + strings.Join(filter.conditions, " AND ")
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM machines m %s`, whereClause)

	var count int
	err := r.client.QueryRow(ctx, query, filter.args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count machines: %w", err)
	}

	return count, nil
}

func (r *db) CountByStatus(ctx context.Context) (int32, int32, int32, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN ` + sqlPpmDateEligibleForStatus + ` AND "ppmDate" < CURRENT_DATE THEN 1 END) as overdue_count,
			COUNT(CASE WHEN ` + sqlPpmDateEligibleForStatus + ` AND "ppmDate" = CURRENT_DATE THEN 1 END) as due_count,
			COUNT(CASE WHEN ` + sqlPpmDateEligibleForStatus + ` AND "ppmDate" > CURRENT_DATE AND "ppmDate" <= CURRENT_DATE + INTERVAL '2 weeks' THEN 1 END) as almost_due_count
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
	if options == nil {
		options = DefaultListOptions()
	}

	baseQuery := `
		SELECT ` + selectMachineColumns + `,
		       word_similarity($1, m.search_text) as rank
		FROM machines m
		LEFT JOIN users u ON u.id = m."assignedUserId"
		WHERE $1 <% m.search_text
	`

	filter := buildFilterConditions(options, 2)

	for _, cond := range filter.conditions {
		baseQuery += " AND " + cond
	}

	args := append([]interface{}{query}, filter.args...)

	orderByClause := buildOrderByClause(options.Sort, true)

	finalQuery := fmt.Sprintf("%s %s LIMIT $%d OFFSET $%d", baseQuery, orderByClause, filter.nextIndex, filter.nextIndex+1)
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
		if err := scanMachineRow(rows, &machine, &rank); err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}
		machine.PpmStatus = string(r.calculatePPMStatus(machine.PpmDate))
		machines = append(machines, &machine)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return machines, nil
}

func (r *db) CountSearch(ctx context.Context, query string, options *ListOptions) (int, error) {
	if options == nil {
		options = DefaultListOptions()
	}

	baseQuery := `
		SELECT COUNT(*)
		FROM machines m
		WHERE $1 <% m.search_text
	`

	filter := buildFilterConditions(options, 2)

	for _, cond := range filter.conditions {
		baseQuery += " AND " + cond
	}

	args := append([]interface{}{query}, filter.args...)

	var count int
	err := r.client.QueryRow(ctx, baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count search results: %w", err)
	}

	return count, nil
}

// calculatePPMStatus determines the PPM status based on the PPM date
func (r *db) calculatePPMStatus(ppmDate time.Time) PPMStatus {
	if IsPpmDateUnset(ppmDate) {
		return ""
	}

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
