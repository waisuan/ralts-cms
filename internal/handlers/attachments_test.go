package handlers_test

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/attachments"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

// AttachmentHandlerTestSuite defines the test suite for attachment handler
type AttachmentHandlerTestSuite struct {
	suite.Suite

	handler               *handlers.AttachmentHandler
	mockAttachmentService *attachments.MockAttachmentService
	ctrl                  *gomock.Controller
}

// SetupTest sets up each test
func (suite *AttachmentHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockAttachmentService = attachments.NewMockAttachmentService(suite.ctrl)

	deps := &deps.Dependencies{
		AttachmentService: suite.mockAttachmentService,
	}
	suite.handler = handlers.NewAttachmentHandler(deps)
}

// TearDownTest cleans up after each test
func (suite *AttachmentHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// TestCreateMachineAttachment tests the CreateMachineAttachment handler
func (suite *AttachmentHandlerTestSuite) TestCreateMachineAttachment() {
	testCases := []struct {
		name                 string
		machineSerialNumber  string
		fileName             string
		fileContent          []byte
		fileSize             int64
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                "successful creation",
			machineSerialNumber: "SN123456",
			fileName:            "manual.pdf",
			fileContent:         []byte("test pdf content"),
			fileSize:            16,
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					CreateMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
						[]byte("test pdf content"),
						"application/pdf",
					).
					Return(&attachments.Attachment{
						Name:        "manual.pdf",
						Size:        16,
						Object:      []byte("test pdf content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMachine,
					}, nil)
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:                 "file too large",
			machineSerialNumber:  "SN123456",
			fileName:             "large.pdf",
			fileContent:          make([]byte, 6*1024*1024), // 6MB
			fileSize:             6 * 1024 * 1024,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: file size 6291456 exceeds maximum allowed size 5242880",
		},
		{
			name:                 "empty file",
			machineSerialNumber:  "SN123456",
			fileName:             "empty.pdf",
			fileContent:          []byte{},
			fileSize:             0,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: file cannot be empty",
		},

		{
			name:                 "invalid file extension",
			machineSerialNumber:  "SN123456",
			fileName:             "document.txt",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: only PDF files are allowed",
		},
		{
			name:                 "filename too long",
			machineSerialNumber:  "SN123456",
			fileName:             strings.Repeat("a", 101) + ".pdf",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: filename " + strings.Repeat("a", 101) + ".pdf exceeds maximum allowed length of 100 characters",
		},
		{
			name:                "attachment already exists",
			machineSerialNumber: "SN123456",
			fileName:            "manual.pdf",
			fileContent:         []byte("test pdf content"),
			fileSize:            16,
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					CreateMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
						[]byte("test pdf content"),
						"application/pdf",
					).
					Return(nil, errors.New("attachment with name 'manual.pdf' already exists for machine 'SN123456'"))
			},
			expectedStatusCode:   http.StatusConflict,
			expectedResponseBody: "Attachment already exists: attachment with name 'manual.pdf' already exists for machine 'SN123456'",
		},
		{
			name:                "service error",
			machineSerialNumber: "SN123456",
			fileName:            "manual.pdf",
			fileContent:         []byte("test pdf content"),
			fileSize:            16,
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					CreateMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
						[]byte("test pdf content"),
						"application/pdf",
					).
					Return(nil, errors.New("S3 upload failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to create attachment: S3 upload failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup mocks
			tc.setupMocks()

			// Create request with multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", tc.fileName)
			suite.Require().NoError(err)

			_, err = part.Write(tc.fileContent)
			suite.Require().NoError(err)

			err = writer.Close()
			suite.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/machines/"+tc.machineSerialNumber+"/attachments", body)

			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Set up router with vars
			router := mux.NewRouter()
			router.HandleFunc("/machines/{serial_number}/attachments", suite.handler.CreateMachineAttachment)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assertions
			suite.Equal(tc.expectedStatusCode, w.Code)

			if tc.expectedResponseBody != "" {
				suite.Contains(w.Body.String(), tc.expectedResponseBody)
			}
		})
	}
}

// TestCreateMachineAttachmentMultipartFormErrors tests multipart form parsing errors
func (suite *AttachmentHandlerTestSuite) TestCreateMachineAttachmentMultipartFormErrors() {
	testCases := []struct {
		name                 string
		machineSerialNumber  string
		requestBody          io.Reader
		contentType          string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                 "invalid multipart form",
			machineSerialNumber:  "SN123456",
			requestBody:          strings.NewReader("invalid multipart data"),
			contentType:          "multipart/form-data; boundary=invalid",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Failed to parse multipart form:",
		},
		{
			name:                 "missing file field",
			machineSerialNumber:  "SN123456",
			requestBody:          strings.NewReader(""),
			contentType:          "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Failed to parse multipart form:",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/machines/"+tc.machineSerialNumber+"/attachments", tc.requestBody)
			req.Header.Set("Content-Type", tc.contentType)

			// Set up router with vars
			router := mux.NewRouter()
			router.HandleFunc("/machines/{serial_number}/attachments", suite.handler.CreateMachineAttachment)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assertions
			suite.Equal(tc.expectedStatusCode, w.Code)
			suite.Contains(w.Body.String(), tc.expectedResponseBody)
		})
	}
}

// TestDeleteMachineAttachment tests the DeleteMachineAttachment handler
func (suite *AttachmentHandlerTestSuite) TestDeleteMachineAttachment() {
	testCases := []struct {
		name                 string
		machineSerialNumber  string
		attachmentName       string
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                "successful deletion",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			setupMocks: func() {
				// Mock GetMachineAttachment to return success (attachment exists)
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "manual.pdf",
						Size:        1024,
						Object:      []byte("test content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMachine,
					}, nil)

				// Mock DeleteMachineAttachment to return success
				suite.mockAttachmentService.EXPECT().
					DeleteMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:                 "missing machine serial number",
			machineSerialNumber:  "",
			attachmentName:       "manual.pdf",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number and attachment name are required",
		},
		{
			name:                 "missing attachment name",
			machineSerialNumber:  "SN123456",
			attachmentName:       "",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number and attachment name are required",
		},
		{
			name:                 "both fields missing",
			machineSerialNumber:  "",
			attachmentName:       "",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number and attachment name are required",
		},
		{
			name:                "attachment not found",
			machineSerialNumber: "SN123456",
			attachmentName:      "nonexistent.pdf",
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"nonexistent.pdf",
					).
					Return(nil, errors.New("attachment 'nonexistent.pdf' not found for machine 'SN123456'"))
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "Attachment not found",
		},
		{
			name:                "service error when checking attachment existence",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(nil, errors.New("database connection failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to check attachment existence: database connection failed",
		},
		{
			name:                "service error when deleting attachment",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			setupMocks: func() {
				// Mock GetMachineAttachment to return success (attachment exists)
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "manual.pdf",
						Size:        1024,
						Object:      []byte("test content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMachine,
					}, nil)

				// Mock DeleteMachineAttachment to return error
				suite.mockAttachmentService.EXPECT().
					DeleteMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(errors.New("S3 deletion failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to delete attachment: S3 deletion failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup mocks
			tc.setupMocks()

			// Create HTTP request
			req := httptest.NewRequest(http.MethodDelete, "/machines/"+tc.machineSerialNumber+"/attachments/"+tc.attachmentName, nil)

			// Set URL vars manually to simulate router behavior
			req = mux.SetURLVars(req, map[string]string{
				"serial_number":   tc.machineSerialNumber,
				"attachment_name": tc.attachmentName,
			})

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request directly through handler
			suite.handler.DeleteMachineAttachment(w, req)

			// Assertions
			suite.Equal(tc.expectedStatusCode, w.Code)

			if tc.expectedResponseBody != "" {
				suite.Contains(w.Body.String(), tc.expectedResponseBody)
			}
		})
	}
}

// TestReplaceMachineAttachment tests the ReplaceMachineAttachment handler
func (suite *AttachmentHandlerTestSuite) TestReplaceMachineAttachment() {
	testCases := []struct {
		name                 string
		machineSerialNumber  string
		oldAttachmentName    string
		fileName             string
		fileContent          []byte
		fileSize             int64
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                "successful replacement",
			machineSerialNumber: "SN123456",
			oldAttachmentName:   "old_manual.pdf",
			fileName:            "new_manual.pdf",
			fileContent:         []byte("new pdf content"),
			fileSize:            15,
			setupMocks: func() {
				// Mock GetMachineAttachment to return success (attachment exists)
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"old_manual.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "old_manual.pdf",
						Size:        1024,
						Object:      []byte("old content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMachine,
					}, nil)

				// Mock DeleteMachineAttachment to return success
				suite.mockAttachmentService.EXPECT().
					DeleteMachineAttachment(
						gomock.Any(),
						"SN123456",
						"old_manual.pdf",
					).
					Return(nil)

				// Mock CreateMachineAttachment to return success
				suite.mockAttachmentService.EXPECT().
					CreateMachineAttachment(
						gomock.Any(),
						"SN123456",
						"new_manual.pdf",
						[]byte("new pdf content"),
						"application/pdf",
					).
					Return(&attachments.Attachment{
						Name:        "new_manual.pdf",
						Size:        15,
						Object:      []byte("new pdf content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMachine,
					}, nil)
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:                 "missing machine serial number",
			machineSerialNumber:  "",
			oldAttachmentName:    "manual.pdf",
			fileName:             "new_manual.pdf",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Missing required fields",
		},
		{
			name:                 "missing old attachment name",
			machineSerialNumber:  "SN123456",
			oldAttachmentName:    "",
			fileName:             "new_manual.pdf",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Missing required fields",
		},
		{
			name:                 "both fields missing",
			machineSerialNumber:  "",
			oldAttachmentName:    "",
			fileName:             "new_manual.pdf",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Missing required fields",
		},
		{
			name:                 "file too large",
			machineSerialNumber:  "SN123456",
			oldAttachmentName:    "manual.pdf",
			fileName:             "large.pdf",
			fileContent:          make([]byte, 6*1024*1024), // 6MB
			fileSize:             6 * 1024 * 1024,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: file size 6291456 exceeds maximum allowed size 5242880",
		},
		{
			name:                 "empty file",
			machineSerialNumber:  "SN123456",
			oldAttachmentName:    "manual.pdf",
			fileName:             "empty.pdf",
			fileContent:          []byte{},
			fileSize:             0,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: file cannot be empty",
		},
		{
			name:                 "invalid file extension",
			machineSerialNumber:  "SN123456",
			oldAttachmentName:    "manual.pdf",
			fileName:             "document.txt",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: only PDF files are allowed",
		},
		{
			name:                 "filename too long",
			machineSerialNumber:  "SN123456",
			oldAttachmentName:    "manual.pdf",
			fileName:             strings.Repeat("a", 101) + ".pdf",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: filename " + strings.Repeat("a", 101) + ".pdf exceeds maximum allowed length of 100 characters",
		},
		{
			name:                "old attachment not found",
			machineSerialNumber: "SN123456",
			oldAttachmentName:   "nonexistent.pdf",
			fileName:            "new_manual.pdf",
			fileContent:         []byte("test content"),
			fileSize:            12,
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"nonexistent.pdf",
					).
					Return(nil, errors.New("attachment 'nonexistent.pdf' not found for machine 'SN123456'"))
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "Attachment not found",
		},
		{
			name:                "service error when checking attachment existence",
			machineSerialNumber: "SN123456",
			oldAttachmentName:   "manual.pdf",
			fileName:            "new_manual.pdf",
			fileContent:         []byte("test content"),
			fileSize:            12,
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(nil, errors.New("database connection failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to check attachment existence: database connection failed",
		},
		{
			name:                "service error when deleting old attachment",
			machineSerialNumber: "SN123456",
			oldAttachmentName:   "manual.pdf",
			fileName:            "new_manual.pdf",
			fileContent:         []byte("test content"),
			fileSize:            12,
			setupMocks: func() {
				// Mock GetMachineAttachment to return success (attachment exists)
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "manual.pdf",
						Size:        1024,
						Object:      []byte("old content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMachine,
					}, nil)

				// Mock DeleteMachineAttachment to return error
				suite.mockAttachmentService.EXPECT().
					DeleteMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(errors.New("S3 deletion failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to delete existing attachment: S3 deletion failed",
		},
		{
			name:                "service error when creating new attachment",
			machineSerialNumber: "SN123456",
			oldAttachmentName:   "manual.pdf",
			fileName:            "new_manual.pdf",
			fileContent:         []byte("test content"),
			fileSize:            12,
			setupMocks: func() {
				// Mock GetMachineAttachment to return success (attachment exists)
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "manual.pdf",
						Size:        1024,
						Object:      []byte("old content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMachine,
					}, nil)

				// Mock DeleteMachineAttachment to return success
				suite.mockAttachmentService.EXPECT().
					DeleteMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(nil)

				// Mock CreateMachineAttachment to return error
				suite.mockAttachmentService.EXPECT().
					CreateMachineAttachment(
						gomock.Any(),
						"SN123456",
						"new_manual.pdf",
						[]byte("test content"),
						"application/pdf",
					).
					Return(nil, errors.New("S3 upload failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to create new attachment 'new_manual.pdf' after deleting old one 'manual.pdf': S3 upload failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup mocks
			tc.setupMocks()

			// Create request with multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", tc.fileName)
			suite.Require().NoError(err)

			_, err = part.Write(tc.fileContent)
			suite.Require().NoError(err)

			err = writer.Close()
			suite.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPut, "/machines/"+tc.machineSerialNumber+"/attachments/"+tc.oldAttachmentName, body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Set URL vars manually to simulate router behavior
			req = mux.SetURLVars(req, map[string]string{
				"serial_number":   tc.machineSerialNumber,
				"attachment_name": tc.oldAttachmentName,
			})

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request directly through handler
			suite.handler.ReplaceMachineAttachment(w, req)

			// Assertions
			suite.Equal(tc.expectedStatusCode, w.Code)

			if tc.expectedResponseBody != "" {
				suite.Contains(w.Body.String(), tc.expectedResponseBody)
			}
		})
	}
}

// TestReplaceMachineAttachmentMultipartFormErrors tests multipart form parsing errors for replace
func (suite *AttachmentHandlerTestSuite) TestReplaceMachineAttachmentMultipartFormErrors() {
	testCases := []struct {
		name                 string
		machineSerialNumber  string
		oldAttachmentName    string
		requestBody          io.Reader
		contentType          string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                 "invalid multipart form",
			machineSerialNumber:  "SN123456",
			oldAttachmentName:    "manual.pdf",
			requestBody:          strings.NewReader("invalid multipart data"),
			contentType:          "multipart/form-data; boundary=invalid",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Failed to parse multipart form:",
		},
		{
			name:                 "missing file field",
			machineSerialNumber:  "SN123456",
			oldAttachmentName:    "manual.pdf",
			requestBody:          strings.NewReader(""),
			contentType:          "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Failed to parse multipart form:",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create HTTP request
			req := httptest.NewRequest(http.MethodPut, "/machines/"+tc.machineSerialNumber+"/attachments/"+tc.oldAttachmentName, tc.requestBody)
			req.Header.Set("Content-Type", tc.contentType)

			// Set URL vars manually to simulate router behavior
			req = mux.SetURLVars(req, map[string]string{
				"serial_number":   tc.machineSerialNumber,
				"attachment_name": tc.oldAttachmentName,
			})

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request directly through handler
			suite.handler.ReplaceMachineAttachment(w, req)

			// Assertions
			suite.Equal(tc.expectedStatusCode, w.Code)
			suite.Contains(w.Body.String(), tc.expectedResponseBody)
		})
	}
}

// TestGetMachineAttachment tests the GetMachineAttachment handler
func (suite *AttachmentHandlerTestSuite) TestGetMachineAttachment() {
	testCases := []struct {
		name                 string
		machineSerialNumber  string
		attachmentName       string
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
		expectedHeaders      map[string]string
	}{
		{
			name:                "successful get attachment",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "manual.pdf",
						Size:        16,
						Object:      []byte("test pdf content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMachine,
					}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: "test pdf content",
			expectedHeaders: map[string]string{
				"Content-Type":        "application/pdf",
				"Content-Disposition": "attachment; filename=\"manual.pdf\"",
				"Content-Length":      "16",
			},
		},
		{
			name:                "attachment not found",
			machineSerialNumber: "SN123456",
			attachmentName:      "nonexistent.pdf",
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"nonexistent.pdf",
					).
					Return(nil, errors.New("attachment not found"))
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "Attachment not found",
		},
		{
			name:                "service error",
			machineSerialNumber: "SN123456",
			attachmentName:      "manual.pdf",
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMachineAttachment(
						gomock.Any(),
						"SN123456",
						"manual.pdf",
					).
					Return(nil, errors.New("S3 connection failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to get attachment: S3 connection failed",
		},
		{
			name:                 "missing machine serial number",
			machineSerialNumber:  "",
			attachmentName:       "manual.pdf",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number and attachment name are required",
		},
		{
			name:                 "missing attachment name",
			machineSerialNumber:  "SN123456",
			attachmentName:       "",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number and attachment name are required",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup mocks
			tc.setupMocks()

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/machines/"+tc.machineSerialNumber+"/attachments/"+tc.attachmentName, nil)

			// Set URL vars manually to simulate router behavior
			req = mux.SetURLVars(req, map[string]string{
				"serial_number":   tc.machineSerialNumber,
				"attachment_name": tc.attachmentName,
			})

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request directly through handler
			suite.handler.GetMachineAttachment(w, req)

			// Assertions
			suite.Equal(tc.expectedStatusCode, w.Code)

			if tc.expectedResponseBody != "" {
				suite.Contains(w.Body.String(), tc.expectedResponseBody)
			}

			// Check expected headers if provided
			if tc.expectedHeaders != nil {
				for headerName, expectedValue := range tc.expectedHeaders {
					suite.Equal(expectedValue, w.Header().Get(headerName))
				}
			}
		})
	}
}

// TestCreateMaintenanceAttachment tests the CreateMaintenanceAttachment handler
func (suite *AttachmentHandlerTestSuite) TestCreateMaintenanceAttachment() {
	testCases := []struct {
		name                 string
		machineSerialNumber  string
		workOrderNumber      string
		fileName             string
		fileContent          []byte
		fileSize             int64
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                "successful creation",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			fileName:            "maintenance.pdf",
			fileContent:         []byte("test pdf content"),
			fileSize:            16,
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					CreateMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
						[]byte("test pdf content"),
						"application/pdf",
					).
					Return(&attachments.Attachment{
						Name:        "maintenance.pdf",
						Size:        16,
						Object:      []byte("test pdf content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMaintenance,
					}, nil)
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:                 "missing machine serial number",
			machineSerialNumber:  "",
			workOrderNumber:      "WO789",
			fileName:             "maintenance.pdf",
			fileContent:          []byte("test pdf content"),
			fileSize:             16,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number is required",
		},
		{
			name:                 "missing work order number",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "",
			fileName:             "maintenance.pdf",
			fileContent:          []byte("test pdf content"),
			fileSize:             16,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Work order number is required",
		},
		{
			name:                 "file too large",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			fileName:             "large.pdf",
			fileContent:          make([]byte, 6*1024*1024), // 6MB
			fileSize:             6 * 1024 * 1024,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: file size 6291456 exceeds maximum allowed size 5242880",
		},
		{
			name:                 "empty file",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			fileName:             "empty.pdf",
			fileContent:          []byte{},
			fileSize:             0,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: file cannot be empty",
		},
		{
			name:                 "invalid file extension",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			fileName:             "document.txt",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: only PDF files are allowed",
		},
		{
			name:                 "filename too long",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			fileName:             strings.Repeat("a", 101) + ".pdf",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: filename " + strings.Repeat("a", 101) + ".pdf exceeds maximum allowed length of 100 characters",
		},
		{
			name:                "attachment already exists",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			fileName:            "maintenance.pdf",
			fileContent:         []byte("test pdf content"),
			fileSize:            16,
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					CreateMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
						[]byte("test pdf content"),
						"application/pdf",
					).
					Return(nil, errors.New("attachment with name 'maintenance.pdf' already exists for maintenance 'WO789' of machine 'SN123456'"))
			},
			expectedStatusCode:   http.StatusConflict,
			expectedResponseBody: "Attachment already exists: attachment with name 'maintenance.pdf' already exists for maintenance 'WO789' of machine 'SN123456'",
		},
		{
			name:                "service error",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			fileName:            "maintenance.pdf",
			fileContent:         []byte("test pdf content"),
			fileSize:            16,
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					CreateMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
						[]byte("test pdf content"),
						"application/pdf",
					).
					Return(nil, errors.New("S3 upload failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to create attachment: S3 upload failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup mocks
			tc.setupMocks()

			// Create request with multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", tc.fileName)
			suite.Require().NoError(err)

			_, err = part.Write(tc.fileContent)
			suite.Require().NoError(err)

			err = writer.Close()
			suite.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/machines/"+tc.machineSerialNumber+"/maintenance/"+tc.workOrderNumber+"/attachments", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Set URL vars manually to simulate router behavior
			req = mux.SetURLVars(req, map[string]string{
				"serial_number":     tc.machineSerialNumber,
				"work_order_number": tc.workOrderNumber,
			})

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request directly through handler
			suite.handler.CreateMaintenanceAttachment(w, req)

			// Assertions
			suite.Equal(tc.expectedStatusCode, w.Code)

			if tc.expectedResponseBody != "" {
				suite.Contains(w.Body.String(), tc.expectedResponseBody)
			}
		})
	}
}

// TestGetMaintenanceAttachment tests the GetMaintenanceAttachment handler
func (suite *AttachmentHandlerTestSuite) TestGetMaintenanceAttachment() {
	testCases := []struct {
		name                 string
		machineSerialNumber  string
		workOrderNumber      string
		attachmentName       string
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
		expectedHeaders      map[string]string
	}{
		{
			name:                "successful get attachment",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			attachmentName:      "maintenance.pdf",
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "maintenance.pdf",
						Size:        16,
						Object:      []byte("test pdf content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMaintenance,
					}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: "test pdf content",
			expectedHeaders: map[string]string{
				"Content-Type":        "application/pdf",
				"Content-Disposition": "attachment; filename=\"maintenance.pdf\"",
				"Content-Length":      "16",
			},
		},
		{
			name:                 "missing machine serial number",
			machineSerialNumber:  "",
			workOrderNumber:      "WO789",
			attachmentName:       "maintenance.pdf",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number, work order number, and attachment name are required",
		},
		{
			name:                 "missing work order number",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "",
			attachmentName:       "maintenance.pdf",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number, work order number, and attachment name are required",
		},
		{
			name:                 "missing attachment name",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			attachmentName:       "",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number, work order number, and attachment name are required",
		},
		{
			name:                "attachment not found",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			attachmentName:      "nonexistent.pdf",
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"nonexistent.pdf",
					).
					Return(nil, errors.New("attachment not found"))
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "Attachment not found",
		},
		{
			name:                "service error",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			attachmentName:      "maintenance.pdf",
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(nil, errors.New("S3 connection failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to get attachment: S3 connection failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup mocks
			tc.setupMocks()

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/machines/"+tc.machineSerialNumber+"/maintenance/"+tc.workOrderNumber+"/attachments/"+tc.attachmentName, nil)

			// Set URL vars manually to simulate router behavior
			req = mux.SetURLVars(req, map[string]string{
				"serial_number":     tc.machineSerialNumber,
				"work_order_number": tc.workOrderNumber,
				"attachment_name":   tc.attachmentName,
			})

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request directly through handler
			suite.handler.GetMaintenanceAttachment(w, req)

			// Assertions
			suite.Equal(tc.expectedStatusCode, w.Code)

			if tc.expectedResponseBody != "" {
				suite.Contains(w.Body.String(), tc.expectedResponseBody)
			}

			// Check expected headers if provided
			if tc.expectedHeaders != nil {
				for headerName, expectedValue := range tc.expectedHeaders {
					suite.Equal(expectedValue, w.Header().Get(headerName))
				}
			}
		})
	}
}

// TestReplaceMaintenanceAttachment tests the ReplaceMaintenanceAttachment handler
func (suite *AttachmentHandlerTestSuite) TestReplaceMaintenanceAttachment() {
	testCases := []struct {
		name                 string
		machineSerialNumber  string
		workOrderNumber      string
		oldAttachmentName    string
		fileName             string
		fileContent          []byte
		fileSize             int64
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                "successful replacement",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			oldAttachmentName:   "old_maintenance.pdf",
			fileName:            "new_maintenance.pdf",
			fileContent:         []byte("new pdf content"),
			fileSize:            15,
			setupMocks: func() {
				// Mock GetMaintenanceAttachment to return success (attachment exists)
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"old_maintenance.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "old_maintenance.pdf",
						Size:        1024,
						Object:      []byte("old content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMaintenance,
					}, nil)

				// Mock DeleteMaintenanceAttachment to return success
				suite.mockAttachmentService.EXPECT().
					DeleteMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"old_maintenance.pdf",
					).
					Return(nil)

				// Mock CreateMaintenanceAttachment to return success
				suite.mockAttachmentService.EXPECT().
					CreateMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"new_maintenance.pdf",
						[]byte("new pdf content"),
						"application/pdf",
					).
					Return(&attachments.Attachment{
						Name:        "new_maintenance.pdf",
						Size:        15,
						Object:      []byte("new pdf content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMaintenance,
					}, nil)
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:                 "missing machine serial number",
			machineSerialNumber:  "",
			workOrderNumber:      "WO789",
			oldAttachmentName:    "maintenance.pdf",
			fileName:             "new_maintenance.pdf",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Missing required fields",
		},
		{
			name:                 "missing work order number",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "",
			oldAttachmentName:    "maintenance.pdf",
			fileName:             "new_maintenance.pdf",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Missing required fields",
		},
		{
			name:                 "missing old attachment name",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			oldAttachmentName:    "",
			fileName:             "new_maintenance.pdf",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Missing required fields",
		},
		{
			name:                 "file too large",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			oldAttachmentName:    "maintenance.pdf",
			fileName:             "large.pdf",
			fileContent:          make([]byte, 6*1024*1024), // 6MB
			fileSize:             6 * 1024 * 1024,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: file size 6291456 exceeds maximum allowed size 5242880",
		},
		{
			name:                 "empty file",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			oldAttachmentName:    "maintenance.pdf",
			fileName:             "empty.pdf",
			fileContent:          []byte{},
			fileSize:             0,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: file cannot be empty",
		},
		{
			name:                 "invalid file extension",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			oldAttachmentName:    "maintenance.pdf",
			fileName:             "document.txt",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: only PDF files are allowed",
		},
		{
			name:                 "filename too long",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			oldAttachmentName:    "maintenance.pdf",
			fileName:             strings.Repeat("a", 101) + ".pdf",
			fileContent:          []byte("test content"),
			fileSize:             12,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "File validation failed: filename " + strings.Repeat("a", 101) + ".pdf exceeds maximum allowed length of 100 characters",
		},
		{
			name:                "old attachment not found",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			oldAttachmentName:   "nonexistent.pdf",
			fileName:            "new_maintenance.pdf",
			fileContent:         []byte("test content"),
			fileSize:            12,
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"nonexistent.pdf",
					).
					Return(nil, errors.New("attachment 'nonexistent.pdf' not found for maintenance 'WO789' of machine 'SN123456'"))
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "Attachment not found",
		},
		{
			name:                "service error when checking attachment existence",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			oldAttachmentName:   "maintenance.pdf",
			fileName:            "new_maintenance.pdf",
			fileContent:         []byte("test content"),
			fileSize:            12,
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(nil, errors.New("database connection failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to check attachment existence: database connection failed",
		},
		{
			name:                "service error when deleting old attachment",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			oldAttachmentName:   "maintenance.pdf",
			fileName:            "new_maintenance.pdf",
			fileContent:         []byte("test content"),
			fileSize:            12,
			setupMocks: func() {
				// Mock GetMaintenanceAttachment to return success (attachment exists)
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "maintenance.pdf",
						Size:        1024,
						Object:      []byte("old content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMaintenance,
					}, nil)

				// Mock DeleteMaintenanceAttachment to return error
				suite.mockAttachmentService.EXPECT().
					DeleteMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(errors.New("S3 deletion failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to delete existing attachment: S3 deletion failed",
		},
		{
			name:                "service error when creating new attachment",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			oldAttachmentName:   "maintenance.pdf",
			fileName:            "new_maintenance.pdf",
			fileContent:         []byte("test content"),
			fileSize:            12,
			setupMocks: func() {
				// Mock GetMaintenanceAttachment to return success (attachment exists)
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "maintenance.pdf",
						Size:        1024,
						Object:      []byte("old content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMaintenance,
					}, nil)

				// Mock DeleteMaintenanceAttachment to return success
				suite.mockAttachmentService.EXPECT().
					DeleteMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(nil)

				// Mock CreateMaintenanceAttachment to return error
				suite.mockAttachmentService.EXPECT().
					CreateMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"new_maintenance.pdf",
						[]byte("test content"),
						"application/pdf",
					).
					Return(nil, errors.New("S3 upload failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to create new attachment 'new_maintenance.pdf' after deleting old one 'maintenance.pdf': S3 upload failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup mocks
			tc.setupMocks()

			// Create request with multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", tc.fileName)
			suite.Require().NoError(err)

			_, err = part.Write(tc.fileContent)
			suite.Require().NoError(err)

			err = writer.Close()
			suite.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPut, "/machines/"+tc.machineSerialNumber+"/maintenance/"+tc.workOrderNumber+"/attachments/"+tc.oldAttachmentName, body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Set URL vars manually to simulate router behavior
			req = mux.SetURLVars(req, map[string]string{
				"serial_number":     tc.machineSerialNumber,
				"work_order_number": tc.workOrderNumber,
				"attachment_name":   tc.oldAttachmentName,
			})

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request directly through handler
			suite.handler.ReplaceMaintenanceAttachment(w, req)

			// Assertions
			suite.Equal(tc.expectedStatusCode, w.Code)

			if tc.expectedResponseBody != "" {
				suite.Contains(w.Body.String(), tc.expectedResponseBody)
			}
		})
	}
}

// TestDeleteMaintenanceAttachment tests the DeleteMaintenanceAttachment handler
func (suite *AttachmentHandlerTestSuite) TestDeleteMaintenanceAttachment() {
	testCases := []struct {
		name                 string
		machineSerialNumber  string
		workOrderNumber      string
		attachmentName       string
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                "successful deletion",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			attachmentName:      "maintenance.pdf",
			setupMocks: func() {
				// Mock GetMaintenanceAttachment to return success (attachment exists)
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "maintenance.pdf",
						Size:        1024,
						Object:      []byte("test content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMaintenance,
					}, nil)

				// Mock DeleteMaintenanceAttachment to return success
				suite.mockAttachmentService.EXPECT().
					DeleteMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:                 "missing machine serial number",
			machineSerialNumber:  "",
			workOrderNumber:      "WO789",
			attachmentName:       "maintenance.pdf",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number, work order number, and attachment name are required",
		},
		{
			name:                 "missing work order number",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "",
			attachmentName:       "maintenance.pdf",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number, work order number, and attachment name are required",
		},
		{
			name:                 "missing attachment name",
			machineSerialNumber:  "SN123456",
			workOrderNumber:      "WO789",
			attachmentName:       "",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number, work order number, and attachment name are required",
		},
		{
			name:                 "all fields missing",
			machineSerialNumber:  "",
			workOrderNumber:      "",
			attachmentName:       "",
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Machine serial number, work order number, and attachment name are required",
		},
		{
			name:                "attachment not found",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			attachmentName:      "nonexistent.pdf",
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"nonexistent.pdf",
					).
					Return(nil, errors.New("attachment 'nonexistent.pdf' not found for maintenance 'WO789' of machine 'SN123456'"))
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "Attachment not found",
		},
		{
			name:                "service error when checking attachment existence",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			attachmentName:      "maintenance.pdf",
			setupMocks: func() {
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(nil, errors.New("database connection failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to check attachment existence: database connection failed",
		},
		{
			name:                "service error when deleting attachment",
			machineSerialNumber: "SN123456",
			workOrderNumber:     "WO789",
			attachmentName:      "maintenance.pdf",
			setupMocks: func() {
				// Mock GetMaintenanceAttachment to return success (attachment exists)
				suite.mockAttachmentService.EXPECT().
					GetMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(&attachments.Attachment{
						Name:        "maintenance.pdf",
						Size:        1024,
						Object:      []byte("test content"),
						ContentType: "application/pdf",
						Type:        attachments.AttachmentTypeMaintenance,
					}, nil)

				// Mock DeleteMaintenanceAttachment to return error
				suite.mockAttachmentService.EXPECT().
					DeleteMaintenanceAttachment(
						gomock.Any(),
						"SN123456",
						"WO789",
						"maintenance.pdf",
					).
					Return(errors.New("S3 deletion failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Failed to delete attachment: S3 deletion failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup mocks
			tc.setupMocks()

			// Create HTTP request
			req := httptest.NewRequest(http.MethodDelete, "/machines/"+tc.machineSerialNumber+"/maintenance/"+tc.workOrderNumber+"/attachments/"+tc.attachmentName, nil)

			// Set URL vars manually to simulate router behavior
			req = mux.SetURLVars(req, map[string]string{
				"serial_number":     tc.machineSerialNumber,
				"work_order_number": tc.workOrderNumber,
				"attachment_name":   tc.attachmentName,
			})

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request directly through handler
			suite.handler.DeleteMaintenanceAttachment(w, req)

			// Assertions
			suite.Equal(tc.expectedStatusCode, w.Code)

			if tc.expectedResponseBody != "" {
				suite.Contains(w.Body.String(), tc.expectedResponseBody)
			}
		})
	}
}

// TestAttachmentHandlerTestSuite runs the attachment handler test suite
func TestAttachmentHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AttachmentHandlerTestSuite))
}
