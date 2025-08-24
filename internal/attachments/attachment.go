// Package attachments provides attachment management functionality
// for storing and retrieving files in S3 for the Ralts-CMS application.
package attachments

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	pkgs3 "ralts-cms/pkg/s3"
)

// AttachmentType represents the type of attachment
type AttachmentType string

const (
	// AttachmentTypeMachine represents an attachment for a machine
	AttachmentTypeMachine AttachmentType = "machine"
	// AttachmentTypeMaintenance represents an attachment for a maintenance record
	AttachmentTypeMaintenance AttachmentType = "maintenance"
)

// Attachment represents an attachment in the system
type Attachment struct {
	Name        string         `json:"name"`
	Size        int64          `json:"size"`
	Object      []byte         `json:"object"`
	ContentType string         `json:"content_type"`
	Type        AttachmentType `json:"type"`
}

// AttachmentService defines the interface for attachment operations
//
//go:generate mockgen -destination=./mock_attachment_service.go -package=attachments -source=attachment.go AttachmentService
type AttachmentService interface {
	CreateMachineAttachment(ctx context.Context, machineSerialNumber, attachmentName string, data []byte, contentType string) (*Attachment, error)
	CreateMaintenanceAttachment(ctx context.Context, machineSerialNumber, workOrderNumber, attachmentName string, data []byte, contentType string) (*Attachment, error)
	GetMachineAttachment(ctx context.Context, machineSerialNumber, attachmentName string) (*Attachment, error)
	GetMaintenanceAttachment(ctx context.Context, machineSerialNumber, workOrderNumber, attachmentName string) (*Attachment, error)
	DeleteMachineAttachment(ctx context.Context, machineSerialNumber, attachmentName string) error
	DeleteMaintenanceAttachment(ctx context.Context, machineSerialNumber, workOrderNumber, attachmentName string) error
}

// service implements the AttachmentService interface
type service struct {
	s3Client   pkgs3.Client
	bucketName string
}

// NewService creates a new attachment service instance
func NewService(s3Client pkgs3.Client, bucketName string) AttachmentService {
	return &service{
		s3Client:   s3Client,
		bucketName: bucketName,
	}
}

// CreateMachineAttachment creates a new attachment for a machine
func (s *service) CreateMachineAttachment(ctx context.Context, machineSerialNumber, attachmentName string, data []byte, contentType string) (*Attachment, error) {
	if err := validateMachineInput(machineSerialNumber, attachmentName); err != nil {
		return nil, err
	}

	key := buildMachineAttachmentKey(machineSerialNumber, attachmentName)

	// Check if attachment already exists
	exists, err := s.attachmentExists(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to check if attachment exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("attachment with name '%s' already exists for machine '%s'", attachmentName, machineSerialNumber)
	}

	// Upload to S3
	if err := s.uploadToS3(ctx, key, data, contentType); err != nil {
		return nil, fmt.Errorf("failed to upload attachment: %w", err)
	}

	return &Attachment{
		Name:        attachmentName,
		Size:        int64(len(data)),
		Object:      data,
		ContentType: contentType,
		Type:        AttachmentTypeMachine,
	}, nil
}

// CreateMaintenanceAttachment creates a new attachment for a maintenance record
func (s *service) CreateMaintenanceAttachment(ctx context.Context, machineSerialNumber, workOrderNumber, attachmentName string, data []byte, contentType string) (*Attachment, error) {
	if err := validateMaintenanceInput(machineSerialNumber, workOrderNumber, attachmentName); err != nil {
		return nil, err
	}

	key := buildMaintenanceAttachmentKey(machineSerialNumber, workOrderNumber, attachmentName)

	// Check if attachment already exists
	exists, err := s.attachmentExists(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to check if attachment exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("attachment with name '%s' already exists for maintenance '%s' of machine '%s'", attachmentName, workOrderNumber, machineSerialNumber)
	}

	// Upload to S3
	if err := s.uploadToS3(ctx, key, data, contentType); err != nil {
		return nil, fmt.Errorf("failed to upload attachment: %w", err)
	}

	return &Attachment{
		Name:        attachmentName,
		Size:        int64(len(data)),
		Object:      data,
		ContentType: contentType,
		Type:        AttachmentTypeMaintenance,
	}, nil
}

// GetMachineAttachment retrieves an attachment for a machine
func (s *service) GetMachineAttachment(ctx context.Context, machineSerialNumber, attachmentName string) (*Attachment, error) {
	if err := validateMachineInput(machineSerialNumber, attachmentName); err != nil {
		return nil, err
	}

	key := buildMachineAttachmentKey(machineSerialNumber, attachmentName)
	return s.getAttachment(ctx, key, attachmentName, AttachmentTypeMachine)
}

// GetMaintenanceAttachment retrieves an attachment for a maintenance record
func (s *service) GetMaintenanceAttachment(ctx context.Context, machineSerialNumber, workOrderNumber, attachmentName string) (*Attachment, error) {
	if err := validateMaintenanceInput(machineSerialNumber, workOrderNumber, attachmentName); err != nil {
		return nil, err
	}

	key := buildMaintenanceAttachmentKey(machineSerialNumber, workOrderNumber, attachmentName)
	return s.getAttachment(ctx, key, attachmentName, AttachmentTypeMaintenance)
}

// DeleteMachineAttachment deletes an attachment for a machine
func (s *service) DeleteMachineAttachment(ctx context.Context, machineSerialNumber, attachmentName string) error {
	if err := validateMachineInput(machineSerialNumber, attachmentName); err != nil {
		return err
	}

	key := buildMachineAttachmentKey(machineSerialNumber, attachmentName)
	return s.deleteFromS3(ctx, key)
}

// DeleteMaintenanceAttachment deletes an attachment for a maintenance record
func (s *service) DeleteMaintenanceAttachment(ctx context.Context, machineSerialNumber, workOrderNumber, attachmentName string) error {
	if err := validateMaintenanceInput(machineSerialNumber, workOrderNumber, attachmentName); err != nil {
		return err
	}

	key := buildMaintenanceAttachmentKey(machineSerialNumber, workOrderNumber, attachmentName)
	return s.deleteFromS3(ctx, key)
}

// attachmentExists checks if an attachment exists in S3
func (s *service) attachmentExists(ctx context.Context, key string) (bool, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	_, err := s.s3Client.GetObject(ctx, input)
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// uploadToS3 uploads data to S3
func (s *service) uploadToS3(ctx context.Context, key string, data []byte, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	_, err := s.s3Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	return nil
}

// getAttachment retrieves an attachment from S3
func (s *service) getAttachment(ctx context.Context, key, attachmentName string, attachmentType AttachmentType) (*Attachment, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	result, err := s.s3Client.GetObject(ctx, input)
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, fmt.Errorf("attachment not found")
		}
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return &Attachment{
		Name:        attachmentName,
		Size:        int64(len(data)),
		Object:      data,
		ContentType: contentType,
		Type:        attachmentType,
	}, nil
}

// deleteFromS3 deletes an object from S3
func (s *service) deleteFromS3(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	_, err := s.s3Client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// buildMachineAttachmentKey builds the S3 key for a machine attachment
func buildMachineAttachmentKey(machineSerialNumber, attachmentName string) string {
	return fmt.Sprintf("%s/%s", machineSerialNumber, attachmentName)
}

// buildMaintenanceAttachmentKey builds the S3 key for a maintenance attachment
func buildMaintenanceAttachmentKey(machineSerialNumber, workOrderNumber, attachmentName string) string {
	return fmt.Sprintf("%s_%s/%s", machineSerialNumber, workOrderNumber, attachmentName)
}

// validateMachineInput validates input for machine attachment operations
func validateMachineInput(machineSerialNumber, attachmentName string) error {
	if strings.TrimSpace(machineSerialNumber) == "" {
		return fmt.Errorf("machine serial number cannot be empty")
	}
	if strings.TrimSpace(attachmentName) == "" {
		return fmt.Errorf("attachment name cannot be empty")
	}
	if strings.Contains(attachmentName, "/") {
		return fmt.Errorf("attachment name cannot contain forward slashes")
	}
	return nil
}

// validateMaintenanceInput validates input for maintenance attachment operations
func validateMaintenanceInput(machineSerialNumber, workOrderNumber, attachmentName string) error {
	if strings.TrimSpace(machineSerialNumber) == "" {
		return fmt.Errorf("machine serial number cannot be empty")
	}
	if strings.TrimSpace(workOrderNumber) == "" {
		return fmt.Errorf("work order number cannot be empty")
	}
	if strings.TrimSpace(attachmentName) == "" {
		return fmt.Errorf("attachment name cannot be empty")
	}
	if strings.Contains(attachmentName, "/") {
		return fmt.Errorf("attachment name cannot contain forward slashes")
	}
	return nil
}
