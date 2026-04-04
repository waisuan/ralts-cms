package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"ralts-cms/internal/audit"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
	"time"

	"github.com/gorilla/mux"
)

const csvExportMaxRows = 10000

// CSVExportHandler handles CSV export requests
type CSVExportHandler struct {
	deps *deps.Dependencies
}

// NewCSVExportHandler creates a new CSV export handler instance
func NewCSVExportHandler(deps *deps.Dependencies) *CSVExportHandler {
	return &CSVExportHandler{deps: deps}
}

// ExportMachinesCSV handles GET /machines/export/csv
func (h *CSVExportHandler) ExportMachinesCSV(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	ppmStatusFilterStr := r.URL.Query().Get("ppm_status_filter")
	sortStr := r.URL.Query().Get("sort")

	sort := machines.SortOrderUpdatedAtDesc
	switch sortStr {
	case "updated_at_asc":
		sort = machines.SortOrderUpdatedAtAsc
	case "ppm_date_asc":
		sort = machines.SortOrderPpmDateAsc
	case "ppm_date_desc":
		sort = machines.SortOrderPpmDateDesc
	case "tnc_date_asc":
		sort = machines.SortOrderTncDateAsc
	case "tnc_date_desc":
		sort = machines.SortOrderTncDateDesc
	}

	var ppmStatusFilter machines.PPMStatus
	switch ppmStatusFilterStr {
	case "overdue":
		ppmStatusFilter = machines.PPMStatusOverdue
	case "due":
		ppmStatusFilter = machines.PPMStatusDue
	case "almost_due":
		ppmStatusFilter = machines.PPMStatusAlmostDue
	}

	ppmDateFrom := parseOptionalDate(r, "ppm_date_from")
	ppmDateTo := parseOptionalDate(r, "ppm_date_to")
	tncDateFrom := parseOptionalDate(r, "tnc_date_from")
	tncDateTo := parseOptionalDate(r, "tnc_date_to")

	options := &machines.ListOptions{
		Limit:           csvExportMaxRows,
		Offset:          0,
		Sort:            sort,
		PpmStatusFilter: ppmStatusFilter,
		PpmDateFrom:     ppmDateFrom,
		PpmDateTo:       ppmDateTo,
		TncDateFrom:     tncDateFrom,
		TncDateTo:       tncDateTo,
	}

	var records []*machines.Machine
	var err error
	if query != "" {
		records, err = h.deps.MachinesRepository.Search(r.Context(), query, options)
	} else {
		records, err = h.deps.MachinesRepository.List(r.Context(), options)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch machines: %v", err), http.StatusInternalServerError)
		return
	}

	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionExported, audit.ResourceMachine, "", map[string]any{
		"format": "csv",
		"count":  len(records),
		"query":  query,
	}))

	if len(records) >= csvExportMaxRows {
		w.Header().Set("X-CSV-Truncated", "true")
		w.Header().Set("X-CSV-Max-Rows", fmt.Sprintf("%d", csvExportMaxRows))
	}

	filename := fmt.Sprintf("machines_%s.csv", time.Now().UTC().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"Serial Number", "Customer", "State", "District", "Model", "Brand",
		"Status", "Account Type", "Person In Charge", "Reported By",
		"TNC Date", "PPM Date", "PPM Status", "Additional Notes",
		"Created At", "Updated At",
	}
	if err := writer.Write(header); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write CSV header: %v", err), http.StatusInternalServerError)
		return
	}

	for _, m := range records {
		row := []string{
			sanitizeCSVField(m.SerialNumber),
			sanitizeCSVField(m.Customer),
			sanitizeCSVField(m.State),
			sanitizeCSVField(m.District),
			sanitizeCSVField(m.Model),
			sanitizeCSVField(m.Brand),
			sanitizeCSVField(m.Status),
			sanitizeCSVField(m.AccountType),
			sanitizeCSVField(m.PersonInCharge),
			sanitizeCSVField(m.ReportedBy),
			formatCSVDate(m.TncDate),
			formatCSVDate(m.PpmDate),
			m.PpmStatus,
			sanitizeCSVField(m.AdditionalNotes),
			formatCSVTimestamp(m.CreatedAt),
			formatCSVTimestamp(m.UpdatedAt),
		}
		if err := writer.Write(row); err != nil {
			return
		}
	}
}

// ExportMaintenanceCSV handles GET /machines/{serial_number}/maintenance/export/csv
func (h *CSVExportHandler) ExportMaintenanceCSV(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serialNumber := vars["serial_number"]
	if serialNumber == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	query := r.URL.Query().Get("q")

	options := &maintenance.ListOptions{
		Limit:  csvExportMaxRows,
		Offset: 0,
		Sort:   maintenance.SortOrderWorkOrderDateDesc,
	}

	var records []*maintenance.Maintenance
	var err error
	if query != "" {
		records, err = h.deps.MaintenanceRepository.SearchByMachine(r.Context(), serialNumber, query, options)
	} else {
		records, err = h.deps.MaintenanceRepository.ListByMachine(r.Context(), serialNumber, options)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch maintenance records: %v", err), http.StatusInternalServerError)
		return
	}

	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionExported, audit.ResourceMaintenance, serialNumber, map[string]any{
		"format": "csv",
		"count":  len(records),
		"query":  query,
	}))

	if len(records) >= csvExportMaxRows {
		w.Header().Set("X-CSV-Truncated", "true")
		w.Header().Set("X-CSV-Max-Rows", fmt.Sprintf("%d", csvExportMaxRows))
	}

	filename := fmt.Sprintf("maintenance_%s_%s.csv", serialNumber, time.Now().UTC().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"Work Order Number", "Machine Serial Number", "Work Order Date",
		"Work Order Type", "Action Taken", "Reported By",
		"Created At", "Updated At",
	}
	if err := writer.Write(header); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write CSV header: %v", err), http.StatusInternalServerError)
		return
	}

	for _, m := range records {
		row := []string{
			sanitizeCSVField(m.WorkOrderNumber),
			sanitizeCSVField(m.MachineSerialNumber),
			formatCSVDate(m.WorkOrderDate),
			sanitizeCSVField(m.WorkOrderType),
			sanitizeCSVField(m.ActionTaken),
			sanitizeCSVField(m.ReportedBy),
			formatCSVTimestamp(m.CreatedAt),
			formatCSVTimestamp(m.UpdatedAt),
		}
		if err := writer.Write(row); err != nil {
			return
		}
	}
}

func sanitizeCSVField(s string) string {
	if len(s) == 0 {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + s
	}
	return s
}

var zeroTime time.Time

func formatCSVDate(t time.Time) string {
	if t.IsZero() || t.Equal(zeroTime) {
		return ""
	}
	return t.Format("2006-01-02")
}

func formatCSVTimestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func parseOptionalDate(r *http.Request, param string) *time.Time {
	s := r.URL.Query().Get(param)
	if s == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &parsed
}
