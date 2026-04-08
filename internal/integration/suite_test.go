//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"ralts-cms/internal/attachments"
	"ralts-cms/internal/audit"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
	"ralts-cms/internal/router"
	"ralts-cms/internal/testutils"
	"ralts-cms/internal/users"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	integrationJWTSecret = "integration-test-jwt-secret-key-32chars-min"
	integrationS3Bucket  = "integration-attachments"
)

type IntegrationSuite struct {
	suite.Suite
	db        *testutils.TestDatabase
	ls        testcontainers.Container
	deps      *deps.Dependencies
	server    *httptest.Server
	jwtSecret string
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("integration tests skipped with -short")
	}
	suite.Run(t, new(IntegrationSuite))
}

func (s *IntegrationSuite) SetupSuite() {
	ctx := context.Background()

	db, err := testutils.NewTestDatabase(ctx)
	s.Require().NoError(err)
	s.db = db

	lsContainer, endpoint, err := startLocalStack(ctx)
	s.Require().NoError(err)
	s.ls = lsContainer

	s.jwtSecret = integrationJWTSecret
	cfg := &deps.Config{
		Env:                     "test",
		DatabaseURL:             db.ConnStr,
		JWTSecret:               s.jwtSecret,
		AWSS3BucketName:            integrationS3Bucket,
		AWSAccessKeyID:          "test",
		AWSSecretAccessKey:      "test",
		AWSDefaultRegion:        "us-east-1",
		AWSEndpointURL:          endpoint,
		AWSS3ForcePathStyle:     true,
		DefaultMachinesLimit:    50,
		MaxMachinesLimit:        100,
		DefaultMaintenanceLimit: 50,
		MaxMaintenanceLimit:     100,
		AuditRetentionDays:      7,
		AuditCleanupInterval:    time.Hour,
	}

	s3Client, err := deps.NewS3Client(ctx, cfg)
	s.Require().NoError(err)

	_, err = s3Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(cfg.AWSS3BucketName)})
	if err != nil && !bucketExistsErr(err) {
		s.Require().NoError(err)
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))

	machinesRepo := machines.NewRepository(db.PostgresClient)
	maintenanceRepo := maintenance.NewRepository(db.PostgresClient)
	usersRepo := users.NewRepository(db.PostgresClient)
	auditRepo := audit.NewRepository(db.PostgresClient)

	auditSvc := audit.NewServiceWithConfig(auditRepo, logger, audit.ServiceConfig{
		BufferSize:      128,
		RetentionDays:   7,
		CleanupInterval: 24 * time.Hour,
	})
	auditSvc.Start()

	// Avoid global slog -> stdout from middleware during tests (keeps output small and deterministic).
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError})))

	s.deps = &deps.Dependencies{
		Config:                cfg,
		Logger:                logger,
		PostgresClient:        nil,
		S3Client:              s3Client,
		MachinesRepository:    machinesRepo,
		MaintenanceRepository: maintenanceRepo,
		UsersRepository:       usersRepo,
		AuditRepository:       auditRepo,
		AttachmentService:     attachments.NewService(s3Client, cfg.AWSS3BucketName),
		AuditService:          auditSvc,
	}

	s.server = httptest.NewServer(router.NewRouter(s.deps))
}

func (s *IntegrationSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if s.server != nil {
		s.server.Close()
	}
	if s.deps != nil {
		s.deps.Shutdown(ctx)
	}
	if s.ls != nil {
		_ = s.ls.Terminate(ctx)
	}
	if s.db != nil {
		s.db.Close()
	}
}

func (s *IntegrationSuite) apiURL(path string) string {
	return strings.TrimSuffix(s.server.URL, "/") + path
}

func (s *IntegrationSuite) postJSON(path string, body any, token string) *http.Response {
	b, err := json.Marshal(body)
	s.Require().NoError(err)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.apiURL(path), bytes.NewReader(b))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *IntegrationSuite) get(path string, token string) *http.Response {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, s.apiURL(path), nil)
	s.Require().NoError(err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *IntegrationSuite) putJSON(path string, body any, token string) *http.Response {
	b, err := json.Marshal(body)
	s.Require().NoError(err)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, s.apiURL(path), bytes.NewReader(b))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *IntegrationSuite) delete(path string, token string) *http.Response {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, s.apiURL(path), nil)
	s.Require().NoError(err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *IntegrationSuite) postMultipart(path, formField, filename string, content []byte, token string) *http.Response {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile(formField, filename)
	s.Require().NoError(err)
	_, err = fw.Write(content)
	s.Require().NoError(err)
	s.Require().NoError(w.Close())
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.apiURL(path), &buf)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *IntegrationSuite) createApprovedUser(username, password, email, role string) {
	s.T().Helper()
	reqBody := map[string]any{
		"username": username,
		"email":    email,
		"password": password,
		"role":     role,
		"approved": true,
		"status":   users.StatusApproved,
	}
	resp := s.postJSON("/api/v1/users", reqBody, "")
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, resp.StatusCode, string(raw))
}

func (s *IntegrationSuite) login(username, password string) string {
	s.T().Helper()
	resp := s.postJSON("/api/v1/users/login", map[string]string{
		"username": username,
		"password": password,
	}, "")
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode, string(raw))
	var out struct {
		Token string `json:"token"`
	}
	s.Require().NoError(json.Unmarshal(raw, &out))
	s.Require().NotEmpty(out.Token)
	return out.Token
}

func (s *IntegrationSuite) machineJSON(serial, customer string) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]any{
		"serial_number":    serial,
		"customer":         customer,
		"state":            "Active",
		"account_type":     "Standard",
		"model":            "M1",
		"status":           "Operational",
		"brand":            "B",
		"district":         "D1",
		"person_in_charge": "PIC",
		"reported_by":      "RB",
		"additional_notes": "",
		"attachment":       "x.pdf",
		"tnc_date":         now,
		"ppm_date":         now,
	}
}

func startLocalStack(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "localstack/localstack:3.8",
		ExposedPorts: []string{"4566/tcp"},
		Env: map[string]string{
			"SERVICES":              "s3",
			"EAGER_SERVICE_LOADING": "1",
			"DEBUG":                 "0",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("4566/tcp"),
			wait.ForHTTP("/_localstack/health").WithPort("4566/tcp").WithStartupTimeout(2*time.Minute),
		),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("localstack: %w", err)
	}
	host, err := c.Host(ctx)
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, "", err
	}
	port, err := c.MappedPort(ctx, "4566")
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, "", err
	}
	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())
	return c, endpoint, nil
}

func bucketExistsErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "BucketAlreadyExists") || strings.Contains(msg, "BucketAlreadyOwnedByYou")
}
