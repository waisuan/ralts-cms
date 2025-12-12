package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"ralts-cms/internal/audit"
	"ralts-cms/internal/deps"
	"strings"

	"github.com/gorilla/mux"
)

const (
	// MaxFileSize represents the maximum allowed file size in bytes (5MB)
	MaxFileSize = 5 * 1024 * 1024
	// MaxFilenameLength represents the maximum allowed filename length in characters
	MaxFilenameLength = 100
	// AllowedFileExtension represents the only allowed file extension for uploads
	AllowedFileExtension = "pdf"
	// DefaultContentType represents the default content type for uploaded files
	DefaultContentType = "application/pdf"
)

// AttachmentHandler handles HTTP requests for attachment-related operations
type AttachmentHandler struct {
	deps *deps.Dependencies
}

// NewAttachmentHandler creates a new attachment handler instance with the given dependencies
func NewAttachmentHandler(deps *deps.Dependencies) *AttachmentHandler {
	return &AttachmentHandler{
		deps: deps,
	}
}

// CreateMachineAttachment handles POST /machines/{serial_number}/attachments
// Creates a new attachment for a machine
func (h *AttachmentHandler) CreateMachineAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	// Check if machine serial number is provided
	if strings.TrimSpace(machineSerialNumber) == "" {
		http.Error(w, "Machine serial number is required", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse multipart form: %v", err), http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get uploaded file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file
	if err := h.validateAttachmentFile(header); err != nil {
		http.Error(w, fmt.Sprintf("File validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file data: %v", err), http.StatusInternalServerError)
		return
	}

	// Create attachment using service
	_, err = h.deps.AttachmentService.CreateMachineAttachment(
		r.Context(),
		machineSerialNumber,
		header.Filename,
		fileData,
		DefaultContentType,
	)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, fmt.Sprintf("Attachment already exists: %v", err), http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to create attachment: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log attachment creation
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionCreated, audit.ResourceAttachment, header.Filename, map[string]any{
		"machine_serial_number": machineSerialNumber,
		"filename":              header.Filename,
		"size":                  header.Size,
	}))

	w.WriteHeader(http.StatusCreated)
}

// GetMachineAttachment handles GET /machines/{serial_number}/attachments/{attachment_name}
// Retrieves and returns a specific machine attachment
func (h *AttachmentHandler) GetMachineAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	attachmentName := vars["attachment_name"]

	if machineSerialNumber == "" || attachmentName == "" {
		http.Error(w, "Machine serial number and attachment name are required", http.StatusBadRequest)
		return
	}

	// Get attachment from service
	attachment, err := h.deps.AttachmentService.GetMachineAttachment(r.Context(), machineSerialNumber, attachmentName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get attachment: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log attachment view
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionViewed, audit.ResourceAttachment, attachmentName, map[string]any{
		"machine_serial_number": machineSerialNumber,
	}))

	// Set headers for file download
	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.Name))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", attachment.Size))

	// Write file content
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(attachment.Object); err != nil {
		// Log error but don't send HTTP error as headers are already written
		h.deps.Logger.Error("Failed to write attachment content",
			"error", err,
			"machine_serial_number", machineSerialNumber,
			"attachment_name", attachmentName)
	}
}

// ReplaceMachineAttachment handles PUT /machines/{serial_number}/attachments/{attachment_name}
// Replaces an existing machine attachment with a new one.
//
// Form fields:
// - file: The new attachment file (required)
//
// The old attachment is deleted before the new one is created.
func (h *AttachmentHandler) ReplaceMachineAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	oldAttachmentName := vars["attachment_name"]

	if machineSerialNumber == "" || oldAttachmentName == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse multipart form: %v", err), http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get uploaded file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file
	if err := h.validateAttachmentFile(header); err != nil {
		http.Error(w, fmt.Sprintf("File validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file data: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if attachment exists before replacing
	_, err = h.deps.AttachmentService.GetMachineAttachment(r.Context(), machineSerialNumber, oldAttachmentName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check attachment existence: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete existing attachment first
	if err := h.deps.AttachmentService.DeleteMachineAttachment(r.Context(), machineSerialNumber, oldAttachmentName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete existing attachment: %v", err), http.StatusInternalServerError)
		return
	}

	// Create new attachment with the new name (which might be the same or different)
	_, err = h.deps.AttachmentService.CreateMachineAttachment(
		r.Context(),
		machineSerialNumber,
		header.Filename,
		fileData,
		DefaultContentType,
	)
	if err != nil {
		// If creation fails after deletion, log the error but don't attempt rollback
		// as the old attachment is already deleted
		http.Error(w, fmt.Sprintf("Failed to create new attachment '%s' after deleting old one '%s': %v", header.Filename, oldAttachmentName, err), http.StatusInternalServerError)
		return
	}

	// Audit: Log attachment replacement
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionUpdated, audit.ResourceAttachment, header.Filename, map[string]any{
		"machine_serial_number": machineSerialNumber,
		"old_filename":          oldAttachmentName,
		"new_filename":          header.Filename,
		"size":                  header.Size,
	}))

	w.WriteHeader(http.StatusCreated)
}

// DeleteMachineAttachment handles DELETE /machines/{serial_number}/attachments/{attachment_name}
// Deletes a specific machine attachment
func (h *AttachmentHandler) DeleteMachineAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	attachmentName := vars["attachment_name"]

	if machineSerialNumber == "" || attachmentName == "" {
		http.Error(w, "Machine serial number and attachment name are required", http.StatusBadRequest)
		return
	}

	// Check if attachment exists before deleting
	_, err := h.deps.AttachmentService.GetMachineAttachment(r.Context(), machineSerialNumber, attachmentName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check attachment existence: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete attachment
	if err := h.deps.AttachmentService.DeleteMachineAttachment(r.Context(), machineSerialNumber, attachmentName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete attachment: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log attachment deletion
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionDeleted, audit.ResourceAttachment, attachmentName, map[string]any{
		"machine_serial_number": machineSerialNumber,
	}))

	w.WriteHeader(http.StatusNoContent)
}

// validateAttachmentFile validates the uploaded file
func (h *AttachmentHandler) validateAttachmentFile(header *multipart.FileHeader) error {
	// Check file size
	if header.Size > MaxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", header.Size, MaxFileSize)
	}

	// Check if file size is 0
	if header.Size == 0 {
		return fmt.Errorf("file cannot be empty")
	}

	if header.Filename == "" {
		return fmt.Errorf("filename is required")
	}

	// Validate file extension - only PDF files are allowed
	if !h.isAllowedFileExtension(header.Filename) {
		return fmt.Errorf("only PDF files are allowed")
	}

	if len(header.Filename) > MaxFilenameLength {
		return fmt.Errorf("filename %s exceeds maximum allowed length of %d characters", header.Filename, MaxFilenameLength)
	}

	return nil
}

// isAllowedFileExtension checks if the file extension is allowed
func (h *AttachmentHandler) isAllowedFileExtension(filename string) bool {
	// Extract file extension
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return false
	}
	extension := strings.ToLower(parts[len(parts)-1])

	// Only PDF files are allowed
	return extension == AllowedFileExtension
}

// CreateMaintenanceAttachment handles POST /machines/{serial_number}/maintenance/{work_order_number}/attachments
// Creates a new attachment for a maintenance record
func (h *AttachmentHandler) CreateMaintenanceAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	workOrderNumber := vars["work_order_number"]

	// Check if required parameters are provided
	if strings.TrimSpace(machineSerialNumber) == "" {
		http.Error(w, "Machine serial number is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(workOrderNumber) == "" {
		http.Error(w, "Work order number is required", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse multipart form: %v", err), http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get uploaded file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file
	if err := h.validateAttachmentFile(header); err != nil {
		http.Error(w, fmt.Sprintf("File validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file data: %v", err), http.StatusInternalServerError)
		return
	}

	// Create attachment using service
	_, err = h.deps.AttachmentService.CreateMaintenanceAttachment(
		r.Context(),
		machineSerialNumber,
		workOrderNumber,
		header.Filename,
		fileData,
		DefaultContentType,
	)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, fmt.Sprintf("Attachment already exists: %v", err), http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to create attachment: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log maintenance attachment creation
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionCreated, audit.ResourceAttachment, header.Filename, map[string]any{
		"machine_serial_number": machineSerialNumber,
		"work_order_number":     workOrderNumber,
		"filename":              header.Filename,
		"size":                  header.Size,
	}))

	w.WriteHeader(http.StatusCreated)
}

// GetMaintenanceAttachment handles GET /machines/{serial_number}/maintenance/{work_order_number}/attachments/{attachment_name}
// Retrieves and returns a specific maintenance attachment
func (h *AttachmentHandler) GetMaintenanceAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	workOrderNumber := vars["work_order_number"]
	attachmentName := vars["attachment_name"]

	if machineSerialNumber == "" || workOrderNumber == "" || attachmentName == "" {
		http.Error(w, "Machine serial number, work order number, and attachment name are required", http.StatusBadRequest)
		return
	}

	// Get attachment from service
	attachment, err := h.deps.AttachmentService.GetMaintenanceAttachment(r.Context(), machineSerialNumber, workOrderNumber, attachmentName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get attachment: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log maintenance attachment view
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionViewed, audit.ResourceAttachment, attachmentName, map[string]any{
		"machine_serial_number": machineSerialNumber,
		"work_order_number":     workOrderNumber,
	}))

	// Set headers for file download
	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.Name))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", attachment.Size))

	// Write file content
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(attachment.Object); err != nil {
		// Log error but don't send HTTP error as headers are already written
		h.deps.Logger.Error("Failed to write attachment content",
			"error", err,
			"machine_serial_number", machineSerialNumber,
			"attachment_name", attachmentName)
	}
}

// ReplaceMaintenanceAttachment handles PUT /machines/{serial_number}/maintenance/{work_order_number}/attachments/{attachment_name}
// Replaces an existing maintenance attachment with a new one.
//
// Form fields:
// - file: The new attachment file (required)
//
// The old attachment is deleted before the new one is created.
func (h *AttachmentHandler) ReplaceMaintenanceAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	workOrderNumber := vars["work_order_number"]
	oldAttachmentName := vars["attachment_name"]

	if machineSerialNumber == "" || workOrderNumber == "" || oldAttachmentName == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse multipart form: %v", err), http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get uploaded file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file
	if err := h.validateAttachmentFile(header); err != nil {
		http.Error(w, fmt.Sprintf("File validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file data: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if attachment exists before replacing
	_, err = h.deps.AttachmentService.GetMaintenanceAttachment(r.Context(), machineSerialNumber, workOrderNumber, oldAttachmentName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check attachment existence: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete existing attachment first
	if err := h.deps.AttachmentService.DeleteMaintenanceAttachment(r.Context(), machineSerialNumber, workOrderNumber, oldAttachmentName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete existing attachment: %v", err), http.StatusInternalServerError)
		return
	}

	// Create new attachment with the new name (which might be the same or different)
	_, err = h.deps.AttachmentService.CreateMaintenanceAttachment(
		r.Context(),
		machineSerialNumber,
		workOrderNumber,
		header.Filename,
		fileData,
		DefaultContentType,
	)
	if err != nil {
		// If creation fails after deletion, log the error but don't attempt rollback
		// as the old attachment is already deleted
		http.Error(w, fmt.Sprintf("Failed to create new attachment '%s' after deleting old one '%s': %v", header.Filename, oldAttachmentName, err), http.StatusInternalServerError)
		return
	}

	// Audit: Log maintenance attachment replacement
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionUpdated, audit.ResourceAttachment, header.Filename, map[string]any{
		"machine_serial_number": machineSerialNumber,
		"work_order_number":     workOrderNumber,
		"old_filename":          oldAttachmentName,
		"new_filename":          header.Filename,
		"size":                  header.Size,
	}))

	w.WriteHeader(http.StatusCreated)
}

// DeleteMaintenanceAttachment handles DELETE /machines/{serial_number}/maintenance/{work_order_number}/attachments/{attachment_name}
// Deletes a specific maintenance attachment
func (h *AttachmentHandler) DeleteMaintenanceAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	workOrderNumber := vars["work_order_number"]
	attachmentName := vars["attachment_name"]

	if machineSerialNumber == "" || workOrderNumber == "" || attachmentName == "" {
		http.Error(w, "Machine serial number, work order number, and attachment name are required", http.StatusBadRequest)
		return
	}

	// Check if attachment exists before deleting
	_, err := h.deps.AttachmentService.GetMaintenanceAttachment(r.Context(), machineSerialNumber, workOrderNumber, attachmentName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check attachment existence: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete attachment
	if err := h.deps.AttachmentService.DeleteMaintenanceAttachment(r.Context(), machineSerialNumber, workOrderNumber, attachmentName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete attachment: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log maintenance attachment deletion
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionDeleted, audit.ResourceAttachment, attachmentName, map[string]any{
		"machine_serial_number": machineSerialNumber,
		"work_order_number":     workOrderNumber,
	}))

	w.WriteHeader(http.StatusNoContent)
}
