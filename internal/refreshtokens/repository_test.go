package refreshtokens_test

import (
	"context"
	"testing"
	"time"

	"ralts-cms/internal/refreshtokens"
	"ralts-cms/internal/testutils"
	"ralts-cms/internal/users"
	"ralts-cms/pkg/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RepositoryTestSuite struct {
	suite.Suite
	db        *testutils.TestDatabase
	usersRepo users.Repository
	repo      refreshtokens.Repository
	pepper    string
}

func (s *RepositoryTestSuite) SetupSuite() {
	db := testutils.SetupTestDatabase(s.T())
	s.db = db
	s.usersRepo = users.NewRepository(db.PostgresClient)
	s.repo = refreshtokens.NewRepository(db.PostgresClient)
	s.pepper = "repo-test-pepper"
}

func (s *RepositoryTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *RepositoryTestSuite) TearDownSubTest() {
	ctx := context.Background()
	err := s.db.CleanupAllTables(ctx)
	require.NoError(s.T(), err)
}

func (s *RepositoryTestSuite) createApprovedUser(username string) *users.User {
	ctx := context.Background()
	st := users.StatusApproved
	u := testutils.CreateUser(username, "password123")
	u.Approved = true
	u.Status = &st
	err := s.usersRepo.Create(ctx, u)
	s.Require().NoError(err)
	return u
}

func (s *RepositoryTestSuite) TestCreateGetRotateRevoke() {
	ctx := context.Background()
	u := s.createApprovedUser("tokuser")
	raw := "opaque-refresh-client-value"
	hash := auth.HashRefreshToken(raw, s.pepper)
	exp := time.Now().UTC().Add(2 * time.Hour)

	s.Require().NoError(s.repo.Create(ctx, u.ID, hash, exp))

	tok, err := s.repo.GetActiveByHash(ctx, hash)
	s.Require().NoError(err)
	s.Require().NotNil(tok)
	s.Equal(u.ID, tok.UserID)

	newRaw := "rotated-client-value"
	newHash := auth.HashRefreshToken(newRaw, s.pepper)
	newExp := time.Now().UTC().Add(48 * time.Hour)
	newID, err := s.repo.Rotate(ctx, tok.ID, u.ID, newHash, newExp)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, newID)

	_, err = s.repo.GetActiveByHash(ctx, hash)
	s.ErrorIs(err, refreshtokens.ErrNotFound)

	got, err := s.repo.GetActiveByHash(ctx, newHash)
	s.Require().NoError(err)
	s.Equal(u.ID, got.UserID)

	s.Require().NoError(s.repo.RevokeByID(ctx, got.ID))
	_, err = s.repo.GetActiveByHash(ctx, newHash)
	s.ErrorIs(err, refreshtokens.ErrNotFound)
}

func (s *RepositoryTestSuite) TestRevokeAllForUser() {
	ctx := context.Background()
	u := s.createApprovedUser("multi")
	for i, raw := range []string{"r1", "r2"} {
		h := auth.HashRefreshToken(raw, s.pepper)
		s.Require().NoError(s.repo.Create(ctx, u.ID, h, time.Now().UTC().Add(time.Hour)), "iteration %d", i)
	}
	s.Require().NoError(s.repo.RevokeAllForUser(ctx, u.ID))
	_, err := s.repo.GetActiveByHash(ctx, auth.HashRefreshToken("r1", s.pepper))
	s.ErrorIs(err, refreshtokens.ErrNotFound)
	_, err = s.repo.GetActiveByHash(ctx, auth.HashRefreshToken("r2", s.pepper))
	s.ErrorIs(err, refreshtokens.ErrNotFound)
}

func (s *RepositoryTestSuite) TestGetActiveByHash_Expired() {
	ctx := context.Background()
	u := s.createApprovedUser("expuser")
	raw := "will-expire"
	h := auth.HashRefreshToken(raw, s.pepper)
	past := time.Now().UTC().Add(-1 * time.Hour)
	s.Require().NoError(s.repo.Create(ctx, u.ID, h, past))
	_, err := s.repo.GetActiveByHash(ctx, h)
	s.ErrorIs(err, refreshtokens.ErrNotFound)
}

func (s *RepositoryTestSuite) TestGetActiveByHash_Unknown() {
	ctx := context.Background()
	_ = s.createApprovedUser("unknown")
	_, err := s.repo.GetActiveByHash(ctx, auth.HashRefreshToken("nope", s.pepper))
	s.ErrorIs(err, refreshtokens.ErrNotFound)
}

func TestRepositoryTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres container tests")
	}
	suite.Run(t, new(RepositoryTestSuite))
}
