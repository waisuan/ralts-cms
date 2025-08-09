package attachments_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ralts-cms/internal/attachments"
	pkgs3 "ralts-cms/pkg/s3"
)

func TestAttachmentService_CreateMachineAttachment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		machineSerialNumber string
		attachmentName      string
		data                []byte
		contentType         string
		setupMocks          func(*pkgs3.MockClient)
		expectedError       string
		expectedAttachment  *attachments.Attachment
	}{
		{
			name:                "successful creation",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			data:                []byte("test pdf content"),
			contentType:         "application/pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				// Check if file exists (returns not found)
				mockClient.EXPECT().
					GetObject(gomock.Any(), &s3.GetObjectInput{
						Bucket: aws.String("test-bucket"),
						Key:    aws.String("SN123456/manual.pdf"),
					}).
					Return(nil, &types.NoSuchKey{})

				// Upload file
				mockClient.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						assert.Equal(t, "test-bucket", *input.Bucket)
						assert.Equal(t, "SN123456/manual.pdf", *input.Key)
						assert.Equal(t, "application/pdf", *input.ContentType)

						// Verify body content
						body, err := io.ReadAll(input.Body)
						require.NoError(t, err)
						assert.Equal(t, []byte("test pdf content"), body)

						return &s3.PutObjectOutput{}, nil
					})
			},
			expectedAttachment: &attachments.Attachment{
				Name:        "manual.pdf",
				Size:        16,
				Object:      []byte("test pdf content"),
				ContentType: "application/pdf",
				Type:        attachments.AttachmentTypeMachine,
			},
		},
		{
			name:                "attachment already exists",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			data:                []byte("test pdf content"),
			contentType:         "application/pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				// Check if file exists (returns found)
				mockClient.EXPECT().
					GetObject(gomock.Any(), &s3.GetObjectInput{
						Bucket: aws.String("test-bucket"),
						Key:    aws.String("SN123456/manual.pdf"),
					}).
					Return(&s3.GetObjectOutput{}, nil)
			},
			expectedError: "attachment with name 'manual.pdf' already exists for machine 'SN123456'",
		},
		{
			name:                "empty machine serial number",
			machineSerialNumber: "",
			attachmentName:      "manual.pdf",
			data:                []byte("test pdf content"),
			contentType:         "application/pdf",
			setupMocks:          func(_ *pkgs3.MockClient) {},
			expectedError:       "machine serial number cannot be empty",
		},
		{
			name:                "empty attachment name",
			machineSerialNumber: "SN123456",
			attachmentName:      "",
			data:                []byte("test pdf content"),
			contentType:         "application/pdf",
			setupMocks:          func(_ *pkgs3.MockClient) {},
			expectedError:       "attachment name cannot be empty",
		},
		{
			name:                "attachment name with forward slash",
			machineSerialNumber: "SN123456",
			attachmentName:      "folder/manual.pdf",
			data:                []byte("test pdf content"),
			contentType:         "application/pdf",
			setupMocks:          func(_ *pkgs3.MockClient) {},
			expectedError:       "attachment name cannot contain forward slashes",
		},
		{
			name:                "S3 upload failure",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			data:                []byte("test pdf content"),
			contentType:         "application/pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				// Check if file exists (returns not found)
				mockClient.EXPECT().
					GetObject(gomock.Any(), gomock.Any()).
					Return(nil, &types.NoSuchKey{})

				// Upload file fails
				mockClient.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("upload failed"))
			},
			expectedError: "failed to upload attachment: failed to put object: upload failed",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockS3Client := pkgs3.NewMockClient(ctrl)
			tc.setupMocks(mockS3Client)

			service := attachments.NewService(mockS3Client, "test-bucket")

			attachment, err := service.CreateMachineAttachment(
				context.Background(),
				tc.machineSerialNumber,
				tc.attachmentName,
				tc.data,
				tc.contentType,
			)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAttachment, attachment)
			}
		})
	}
}

func TestAttachmentService_CreateMaintenanceAttachment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		machineSerialNumber string
		workOrderNumber     string
		attachmentName      string
		data                []byte
		contentType         string
		setupMocks          func(*pkgs3.MockClient)
		expectedError       string
		expectedAttachment  *attachments.Attachment
	}{
		{
			name:                "successful creation",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789012",
			attachmentName:      "report.pdf",
			data:                []byte("test report content"),
			contentType:         "application/pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				// Check if file exists (returns not found)
				mockClient.EXPECT().
					GetObject(gomock.Any(), &s3.GetObjectInput{
						Bucket: aws.String("test-bucket"),
						Key:    aws.String("SN123456/WO789012/report.pdf"),
					}).
					Return(nil, &types.NoSuchKey{})

				// Upload file
				mockClient.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						assert.Equal(t, "test-bucket", *input.Bucket)
						assert.Equal(t, "SN123456/WO789012/report.pdf", *input.Key)
						assert.Equal(t, "application/pdf", *input.ContentType)

						// Verify body content
						body, err := io.ReadAll(input.Body)
						require.NoError(t, err)
						assert.Equal(t, []byte("test report content"), body)

						return &s3.PutObjectOutput{}, nil
					})
			},
			expectedAttachment: &attachments.Attachment{
				Name:        "report.pdf",
				Size:        19,
				Object:      []byte("test report content"),
				ContentType: "application/pdf",
				Type:        attachments.AttachmentTypeMaintenance,
			},
		},
		{
			name:                "empty work order number",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "",
			attachmentName:      "report.pdf",
			data:                []byte("test report content"),
			contentType:         "application/pdf",
			setupMocks:          func(_ *pkgs3.MockClient) {},
			expectedError:       "work order number cannot be empty",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockS3Client := pkgs3.NewMockClient(ctrl)
			tc.setupMocks(mockS3Client)

			service := attachments.NewService(mockS3Client, "test-bucket")

			attachment, err := service.CreateMaintenanceAttachment(
				context.Background(),
				tc.machineSerialNumber,
				tc.workOrderNumber,
				tc.attachmentName,
				tc.data,
				tc.contentType,
			)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAttachment, attachment)
			}
		})
	}
}

func TestAttachmentService_GetMachineAttachment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		machineSerialNumber string
		attachmentName      string
		setupMocks          func(*pkgs3.MockClient)
		expectedError       string
		expectedAttachment  *attachments.Attachment
	}{
		{
			name:                "successful retrieval",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				mockClient.EXPECT().
					GetObject(gomock.Any(), &s3.GetObjectInput{
						Bucket: aws.String("test-bucket"),
						Key:    aws.String("SN123456/manual.pdf"),
					}).
					Return(&s3.GetObjectOutput{
						Body:        io.NopCloser(bytes.NewReader([]byte("test pdf content"))),
						ContentType: aws.String("application/pdf"),
					}, nil)
			},
			expectedAttachment: &attachments.Attachment{
				Name:        "manual.pdf",
				Size:        16,
				Object:      []byte("test pdf content"),
				ContentType: "application/pdf",
				Type:        attachments.AttachmentTypeMachine,
			},
		},
		{
			name:                "attachment not found",
			machineSerialNumber: "SN123456",
			attachmentName:      "nonexistent.pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				mockClient.EXPECT().
					GetObject(gomock.Any(), &s3.GetObjectInput{
						Bucket: aws.String("test-bucket"),
						Key:    aws.String("SN123456/nonexistent.pdf"),
					}).
					Return(nil, &types.NoSuchKey{})
			},
			expectedError: "attachment not found",
		},
		{
			name:                "S3 get failure",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				mockClient.EXPECT().
					GetObject(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("S3 error"))
			},
			expectedError: "failed to get object: S3 error",
		},
		{
			name:                "empty machine serial number",
			machineSerialNumber: "",
			attachmentName:      "manual.pdf",
			setupMocks:          func(_ *pkgs3.MockClient) {},
			expectedError:       "machine serial number cannot be empty",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockS3Client := pkgs3.NewMockClient(ctrl)
			tc.setupMocks(mockS3Client)

			service := attachments.NewService(mockS3Client, "test-bucket")

			attachment, err := service.GetMachineAttachment(
				context.Background(),
				tc.machineSerialNumber,
				tc.attachmentName,
			)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAttachment, attachment)
			}
		})
	}
}

func TestAttachmentService_GetMaintenanceAttachment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		machineSerialNumber string
		workOrderNumber     string
		attachmentName      string
		setupMocks          func(*pkgs3.MockClient)
		expectedError       string
		expectedAttachment  *attachments.Attachment
	}{
		{
			name:                "successful retrieval",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789012",
			attachmentName:      "report.pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				mockClient.EXPECT().
					GetObject(gomock.Any(), &s3.GetObjectInput{
						Bucket: aws.String("test-bucket"),
						Key:    aws.String("SN123456/WO789012/report.pdf"),
					}).
					Return(&s3.GetObjectOutput{
						Body:        io.NopCloser(bytes.NewReader([]byte("test report content"))),
						ContentType: aws.String("application/pdf"),
					}, nil)
			},
			expectedAttachment: &attachments.Attachment{
				Name:        "report.pdf",
				Size:        19,
				Object:      []byte("test report content"),
				ContentType: "application/pdf",
				Type:        attachments.AttachmentTypeMaintenance,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockS3Client := pkgs3.NewMockClient(ctrl)
			tc.setupMocks(mockS3Client)

			service := attachments.NewService(mockS3Client, "test-bucket")

			attachment, err := service.GetMaintenanceAttachment(
				context.Background(),
				tc.machineSerialNumber,
				tc.workOrderNumber,
				tc.attachmentName,
			)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAttachment, attachment)
			}
		})
	}
}

func TestAttachmentService_DeleteMachineAttachment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		machineSerialNumber string
		attachmentName      string
		setupMocks          func(*pkgs3.MockClient)
		expectedError       string
	}{
		{
			name:                "successful deletion",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				mockClient.EXPECT().
					DeleteObject(gomock.Any(), &s3.DeleteObjectInput{
						Bucket: aws.String("test-bucket"),
						Key:    aws.String("SN123456/manual.pdf"),
					}).
					Return(&s3.DeleteObjectOutput{}, nil)
			},
		},
		{
			name:                "S3 delete failure",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				mockClient.EXPECT().
					DeleteObject(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("S3 delete error"))
			},
			expectedError: "failed to delete object: S3 delete error",
		},
		{
			name:                "empty machine serial number",
			machineSerialNumber: "",
			attachmentName:      "manual.pdf",
			setupMocks:          func(_ *pkgs3.MockClient) {},
			expectedError:       "machine serial number cannot be empty",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockS3Client := pkgs3.NewMockClient(ctrl)
			tc.setupMocks(mockS3Client)

			service := attachments.NewService(mockS3Client, "test-bucket")

			err := service.DeleteMachineAttachment(
				context.Background(),
				tc.machineSerialNumber,
				tc.attachmentName,
			)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAttachmentService_DeleteMaintenanceAttachment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		machineSerialNumber string
		workOrderNumber     string
		attachmentName      string
		setupMocks          func(*pkgs3.MockClient)
		expectedError       string
	}{
		{
			name:                "successful deletion",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789012",
			attachmentName:      "report.pdf",
			setupMocks: func(mockClient *pkgs3.MockClient) {
				mockClient.EXPECT().
					DeleteObject(gomock.Any(), &s3.DeleteObjectInput{
						Bucket: aws.String("test-bucket"),
						Key:    aws.String("SN123456/WO789012/report.pdf"),
					}).
					Return(&s3.DeleteObjectOutput{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockS3Client := pkgs3.NewMockClient(ctrl)
			tc.setupMocks(mockS3Client)

			service := attachments.NewService(mockS3Client, "test-bucket")

			err := service.DeleteMaintenanceAttachment(
				context.Background(),
				tc.machineSerialNumber,
				tc.workOrderNumber,
				tc.attachmentName,
			)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
