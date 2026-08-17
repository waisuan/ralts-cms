package flags_test

import (
	"context"
	"testing"
	"time"

	"ralts-cms/internal/flags"
	"ralts-cms/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// FlagsRepositoryTestSuite exercises the Postgres-backed flags repository
// against a real database in a testcontainer.
type FlagsRepositoryTestSuite struct {
	suite.Suite

	db   *testutils.TestDatabase
	repo flags.Repository
}

func (suite *FlagsRepositoryTestSuite) SetupSuite() {
	db := testutils.SetupTestDatabase(suite.T())
	suite.db = db
	suite.repo = flags.NewRepository(db.PostgresClient)
}

func (suite *FlagsRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *FlagsRepositoryTestSuite) TearDownSubTest() {
	err := suite.db.CleanupAllTables(context.Background())
	require.NoError(suite.T(), err)
}

// insertUser adds a bare user row so foreign keys are satisfied.
func (suite *FlagsRepositoryTestSuite) insertUser(ctx context.Context, username string) int64 {
	var id int64
	err := suite.db.PostgresClient.QueryRow(ctx, `
		INSERT INTO users (username, email, password, salt, role, approved, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, username+"_"+uuid.NewString()[:8],
		username+"_"+uuid.NewString()[:8]+"@example.com",
		"hashedpassword", "salt", "ADMIN", true, time.Now()).Scan(&id)
	suite.Require().NoError(err)
	return id
}

// insertMachine adds a bare machine row so foreign keys are satisfied.
func (suite *FlagsRepositoryTestSuite) insertMachine(ctx context.Context, serial string) {
	_, err := suite.db.PostgresClient.Exec(ctx, `
		INSERT INTO machines ("serialNumber", customer, "createdAt", "updatedAt")
		VALUES ($1, $2, $3, $3)
	`, serial, "Test Customer", time.Now())
	suite.Require().NoError(err)
}

// insertMachineAssignedTo adds a machine linked to a registered assignee.
func (suite *FlagsRepositoryTestSuite) insertMachineAssignedTo(ctx context.Context, serial string, userID int64) {
	_, err := suite.db.PostgresClient.Exec(ctx, `
		INSERT INTO machines ("serialNumber", customer, "assignedUserId", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, $4, $4)
	`, serial, "Test Customer", userID, time.Now())
	suite.Require().NoError(err)
}

// mustCreate raises a flag, failing the test if the write fails.
func (suite *FlagsRepositoryTestSuite) mustCreate(ctx context.Context, f *flags.Flag) {
	suite.T().Helper()
	_, err := suite.repo.Create(ctx, f)
	suite.Require().NoError(err)
}

func (suite *FlagsRepositoryTestSuite) TestCreateAndListForMachine() {
	ctx := context.Background()

	suite.Run("persists a flag and returns it via ListForMachine", func() {
		suite.insertMachine(ctx, "SN-100")
		adminID := suite.insertUser(ctx, "admin")

		f := &flags.Flag{
			MachineSerialNumber: "SN-100",
			Reason:              flags.ReasonOther,
			Note:                "please check",
			CreatedBy:           &adminID,
		}
		isNew, err := suite.repo.Create(ctx, f)
		suite.Require().NoError(err)
		suite.Assert().True(isNew)
		suite.Assert().NotEmpty(f.ID)
		suite.Assert().Equal(flags.StatusOpen, f.Status)

		list, err := suite.repo.ListForMachine(ctx, "SN-100", false)
		suite.Require().NoError(err)
		suite.Require().Len(list, 1)
		suite.Assert().Equal(f.ID, list[0].ID)
		suite.Assert().Equal(flags.ReasonOther, list[0].Reason)
		suite.Assert().Equal("please check", list[0].Note)
		suite.Require().NotNil(list[0].CreatedByUsername)
	})

	suite.Run("dates the flag from the database clock and reports it back", func() {
		suite.insertMachine(ctx, "SN-101")

		f := &flags.Flag{MachineSerialNumber: "SN-101", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, f)
		suite.Require().False(f.CreatedAt.IsZero(), "the caller learns when the flag was raised")

		list, err := suite.repo.ListForMachine(ctx, "SN-101", false)
		suite.Require().NoError(err)
		suite.Require().Len(list, 1)
		suite.Assert().WithinDuration(f.CreatedAt, list[0].CreatedAt, time.Second,
			"the returned time is the stored one")
		suite.Assert().Nil(list[0].ResolvedAt, "an open flag has no resolution time")
	})
}

func (suite *FlagsRepositoryTestSuite) TestResolveStampsItsOwnTime() {
	ctx := context.Background()

	suite.insertMachine(ctx, "SN-150")
	adminID := suite.insertUser(ctx, "admin_stamp")

	f := &flags.Flag{
		MachineSerialNumber: "SN-150",
		Reason:              flags.ReasonMissingValues,
		CreatedBy:           &adminID,
		CreatedAt:           time.Now().Add(-2 * time.Hour).UTC(),
	}
	suite.mustCreate(ctx, f)

	suite.Require().NoError(suite.repo.Resolve(ctx, f.ID, adminID, "sorted"))

	resolved, err := suite.repo.GetByID(ctx, f.ID)
	suite.Require().NoError(err)
	suite.Require().NotNil(resolved.ResolvedAt)
	// The two events are dated separately: closing a flag leaves its raise time
	// alone rather than reusing it or overwriting it.
	suite.Assert().WithinDuration(f.CreatedAt, resolved.CreatedAt, time.Second)
	suite.Assert().True(resolved.ResolvedAt.After(resolved.CreatedAt),
		"the resolution is dated after the raise")
}

func (suite *FlagsRepositoryTestSuite) TestCreateReplacesAnOpenFlag() {
	ctx := context.Background()

	suite.Run("flagging an already-flagged machine rewrites the open flag", func() {
		suite.insertMachine(ctx, "SN-200")
		first := suite.insertUser(ctx, "first_admin")
		second := suite.insertUser(ctx, "second_admin")

		original := &flags.Flag{
			MachineSerialNumber: "SN-200",
			Reason:              flags.ReasonMissingValues,
			Note:                "district is blank",
			CreatedBy:           &first,
			CreatedAt:           time.Now().Add(-time.Hour).UTC(),
		}
		suite.mustCreate(ctx, original)

		replacement := &flags.Flag{
			MachineSerialNumber: "SN-200",
			Reason:              flags.ReasonOther,
			Note:                "still incomplete",
			CreatedBy:           &second,
		}
		isNew, err := suite.repo.Create(ctx, replacement)
		suite.Require().NoError(err)
		suite.Assert().False(isNew, "no second flag was inserted")
		suite.Assert().Equal(original.ID, replacement.ID, "the open flag keeps its identity")

		list, err := suite.repo.ListForMachine(ctx, "SN-200", true)
		suite.Require().NoError(err)
		suite.Require().Len(list, 1, "the machine still carries exactly one flag")
		suite.Assert().Equal(flags.ReasonOther, list[0].Reason)
		suite.Assert().Equal("still incomplete", list[0].Note)
		suite.Require().NotNil(list[0].CreatedBy)
		suite.Assert().Equal(second, *list[0].CreatedBy, "the latest flagger is recorded")
		suite.Assert().True(list[0].CreatedAt.After(original.CreatedAt), "the raise time moves forward")
	})

	suite.Run("a resolved flag does not block a new one", func() {
		suite.insertMachine(ctx, "SN-201")
		adminID := suite.insertUser(ctx, "admin_resolve")

		f := &flags.Flag{MachineSerialNumber: "SN-201", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, f)
		suite.Require().NoError(suite.repo.Resolve(ctx, f.ID, adminID, ""))

		again := &flags.Flag{MachineSerialNumber: "SN-201", Reason: flags.ReasonMissingValues}
		isNew, err := suite.repo.Create(ctx, again)
		suite.Require().NoError(err)
		suite.Assert().True(isNew)
		suite.Assert().NotEqual(f.ID, again.ID, "history is kept as a separate row")

		list, err := suite.repo.ListForMachine(ctx, "SN-201", true)
		suite.Require().NoError(err)
		suite.Assert().Len(list, 2)
	})

	suite.Run("machines do not interfere with each other", func() {
		suite.insertMachine(ctx, "SN-202")
		suite.insertMachine(ctx, "SN-203")

		a := &flags.Flag{MachineSerialNumber: "SN-202", Reason: flags.ReasonMissingValues}
		b := &flags.Flag{MachineSerialNumber: "SN-203", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, a)

		isNew, err := suite.repo.Create(ctx, b)
		suite.Require().NoError(err)
		suite.Assert().True(isNew)
		suite.Assert().NotEqual(a.ID, b.ID)
	})
}

func (suite *FlagsRepositoryTestSuite) TestListOpenBySerials() {
	ctx := context.Background()

	suite.Run("returns only open flags per serial", func() {
		suite.insertMachine(ctx, "SN-A")
		suite.insertMachine(ctx, "SN-B")
		suite.insertMachine(ctx, "SN-C")
		adminID := suite.insertUser(ctx, "admin")

		// Raised and cleared before the current one, since a machine holds only
		// one open flag at a time.
		resolvedA := &flags.Flag{MachineSerialNumber: "SN-A", Reason: flags.ReasonOther, Note: "already done"}
		suite.mustCreate(ctx, resolvedA)
		suite.Require().NoError(suite.repo.Resolve(ctx, resolvedA.ID, adminID, ""))

		openA := &flags.Flag{MachineSerialNumber: "SN-A", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, openA)

		openB := &flags.Flag{MachineSerialNumber: "SN-B", Reason: flags.ReasonRequiresAttention}
		suite.mustCreate(ctx, openB)

		byserial, err := suite.repo.ListOpenBySerials(ctx, []string{"SN-A", "SN-B", "SN-C"})
		suite.Require().NoError(err)
		suite.Assert().Len(byserial["SN-A"], 1)
		suite.Assert().Equal(openA.ID, byserial["SN-A"][0].ID)
		suite.Assert().Len(byserial["SN-B"], 1)
		suite.Assert().Equal(openB.ID, byserial["SN-B"][0].ID)
		suite.Assert().Empty(byserial["SN-C"])
	})

	suite.Run("empty input returns empty map without querying", func() {
		byserial, err := suite.repo.ListOpenBySerials(ctx, nil)
		suite.Require().NoError(err)
		suite.Assert().Empty(byserial)
	})
}

func (suite *FlagsRepositoryTestSuite) TestListAndCount() {
	ctx := context.Background()

	suite.Run("lists flags across machines newest first and filters by status", func() {
		suite.insertMachine(ctx, "SN-L1")
		suite.insertMachine(ctx, "SN-L2")
		adminID := suite.insertUser(ctx, "admin_list")

		older := &flags.Flag{MachineSerialNumber: "SN-L1", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, older)

		newer := &flags.Flag{MachineSerialNumber: "SN-L2", Reason: flags.ReasonOther, Note: "check cabling"}
		suite.mustCreate(ctx, newer)
		suite.Require().NoError(suite.repo.Resolve(ctx, newer.ID, adminID, ""))

		all, err := suite.repo.List(ctx, flags.ListOptions{Limit: 10})
		suite.Require().NoError(err)
		suite.Require().Len(all, 2)
		suite.Assert().Equal(newer.ID, all[0].ID, "newest flag comes first")
		suite.Require().NotNil(all[0].ResolvedByUsername)

		open, err := suite.repo.List(ctx, flags.ListOptions{Status: flags.StatusOpen, Limit: 10})
		suite.Require().NoError(err)
		suite.Require().Len(open, 1)
		suite.Assert().Equal(older.ID, open[0].ID)

		resolved, err := suite.repo.List(ctx, flags.ListOptions{Status: flags.StatusResolved, Limit: 10})
		suite.Require().NoError(err)
		suite.Require().Len(resolved, 1)
		suite.Assert().Equal(newer.ID, resolved[0].ID)

		total, err := suite.repo.Count(ctx, flags.ListOptions{})
		suite.Require().NoError(err)
		suite.Assert().Equal(2, total)

		openCount, err := suite.repo.Count(ctx, flags.ListOptions{Status: flags.StatusOpen})
		suite.Require().NoError(err)
		suite.Assert().Equal(1, openCount)
	})

	suite.Run("honours limit and offset while count ignores them", func() {
		// One flag each: a machine cannot hold more than one open flag.
		for _, serial := range []string{"SN-L3", "SN-L4", "SN-L5"} {
			suite.insertMachine(ctx, serial)
			suite.mustCreate(ctx, &flags.Flag{
				MachineSerialNumber: serial,
				Reason:              flags.ReasonOther,
				Note:                "note",
			})
		}

		page, err := suite.repo.List(ctx, flags.ListOptions{Status: flags.StatusOpen, Limit: 2})
		suite.Require().NoError(err)
		suite.Assert().Len(page, 2)

		second, err := suite.repo.List(ctx, flags.ListOptions{Status: flags.StatusOpen, Limit: 2, Offset: 2})
		suite.Require().NoError(err)
		suite.Assert().Len(second, 1)

		count, err := suite.repo.Count(ctx, flags.ListOptions{Status: flags.StatusOpen})
		suite.Require().NoError(err)
		suite.Assert().Equal(3, count)
	})

	suite.Run("scopes to flags on machines assigned to a given user", func() {
		assignee := suite.insertUser(ctx, "assignee")
		other := suite.insertUser(ctx, "other")
		suite.insertMachineAssignedTo(ctx, "SN-MINE", assignee)
		suite.insertMachineAssignedTo(ctx, "SN-THEIRS", other)
		suite.insertMachine(ctx, "SN-UNASSIGNED")

		mine := &flags.Flag{MachineSerialNumber: "SN-MINE", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, mine)
		theirs := &flags.Flag{MachineSerialNumber: "SN-THEIRS", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, theirs)
		unassigned := &flags.Flag{MachineSerialNumber: "SN-UNASSIGNED", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, unassigned)

		scoped, err := suite.repo.List(ctx, flags.ListOptions{Status: flags.StatusOpen, AssignedUserID: &assignee, Limit: 10})
		suite.Require().NoError(err)
		suite.Require().Len(scoped, 1)
		suite.Assert().Equal(mine.ID, scoped[0].ID)

		scopedCount, err := suite.repo.Count(ctx, flags.ListOptions{Status: flags.StatusOpen, AssignedUserID: &assignee})
		suite.Require().NoError(err)
		suite.Assert().Equal(1, scopedCount)

		// Without the scope every machine's flags are visible, as for an admin.
		unscopedCount, err := suite.repo.Count(ctx, flags.ListOptions{Status: flags.StatusOpen})
		suite.Require().NoError(err)
		suite.Assert().Equal(3, unscopedCount)
	})

	suite.Run("keeps open flags but only the caller's own resolutions", func() {
		assignee := suite.insertUser(ctx, "own_res_assignee")
		admin := suite.insertUser(ctx, "own_res_admin")
		suite.insertMachineAssignedTo(ctx, "SN-OWN", assignee)

		// Closed by somebody else, most likely before the assignment.
		theirs := &flags.Flag{MachineSerialNumber: "SN-OWN", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, theirs)
		suite.Require().NoError(suite.repo.Resolve(ctx, theirs.ID, admin, ""))

		ours := &flags.Flag{MachineSerialNumber: "SN-OWN", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, ours)
		suite.Require().NoError(suite.repo.Resolve(ctx, ours.ID, assignee, ""))

		stillOpen := &flags.Flag{MachineSerialNumber: "SN-OWN", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, stillOpen)

		opts := flags.ListOptions{
			AssignedUserID:       &assignee,
			OwnResolutionsUserID: &assignee,
			Limit:                10,
		}
		listed, err := suite.repo.List(ctx, opts)
		suite.Require().NoError(err)
		ids := make([]string, 0, len(listed))
		for _, f := range listed {
			ids = append(ids, f.ID)
		}
		suite.Assert().ElementsMatch([]string{stillOpen.ID, ours.ID}, ids)

		count, err := suite.repo.Count(ctx, opts)
		suite.Require().NoError(err)
		suite.Assert().Equal(2, count, "the count matches the listing")

		// Asking for resolved flags narrows to the caller's own.
		opts.Status = flags.StatusResolved
		resolved, err := suite.repo.List(ctx, opts)
		suite.Require().NoError(err)
		suite.Require().Len(resolved, 1)
		suite.Assert().Equal(ours.ID, resolved[0].ID)

		// An admin, who sets neither filter, still sees the whole history.
		everything, err := suite.repo.List(ctx, flags.ListOptions{Status: flags.StatusResolved, Limit: 10})
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(everything)
		found := false
		for _, f := range everything {
			if f.ID == theirs.ID {
				found = true
			}
		}
		suite.Assert().True(found, "the other user's resolution is not hidden from admins")
	})
}

func (suite *FlagsRepositoryTestSuite) TestResolve() {
	ctx := context.Background()

	suite.Run("resolving an unknown id returns ErrNotFound", func() {
		suite.insertMachine(ctx, "SN-X")
		adminID := suite.insertUser(ctx, "admin")

		err := suite.repo.Resolve(ctx, uuid.NewString(), adminID, "")
		suite.Require().ErrorIs(err, flags.ErrNotFound)
	})

	suite.Run("stores a resolution note and reads it back", func() {
		suite.insertMachine(ctx, "SN-RN")
		adminID := suite.insertUser(ctx, "admin")

		f := &flags.Flag{MachineSerialNumber: "SN-RN", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, f)
		suite.Require().Empty(f.ResolutionNote, "an open flag has no resolution note")

		suite.Require().NoError(suite.repo.Resolve(ctx, f.ID, adminID, "filled in the PPM date"))

		resolved, err := suite.repo.GetByID(ctx, f.ID)
		suite.Require().NoError(err)
		suite.Assert().Equal(flags.StatusResolved, resolved.Status)
		suite.Assert().Equal("filled in the PPM date", resolved.ResolutionNote)
		suite.Require().NotNil(resolved.ResolvedByUsername)
	})

	suite.Run("an empty note leaves the column null, read back as empty", func() {
		suite.insertMachine(ctx, "SN-RN2")
		adminID := suite.insertUser(ctx, "admin")

		f := &flags.Flag{MachineSerialNumber: "SN-RN2", Reason: flags.ReasonMissingValues}
		suite.mustCreate(ctx, f)
		suite.Require().NoError(suite.repo.Resolve(ctx, f.ID, adminID, ""))

		var isNull bool
		suite.Require().NoError(suite.db.PostgresClient.QueryRow(ctx,
			`SELECT resolution_note IS NULL FROM machine_flags WHERE id = $1`, f.ID,
		).Scan(&isNull))
		suite.Assert().True(isNull)

		resolved, err := suite.repo.GetByID(ctx, f.ID)
		suite.Require().NoError(err)
		suite.Assert().Empty(resolved.ResolutionNote)
	})

	suite.Run("resolving an already-resolved flag returns ErrNotFound", func() {
		suite.insertMachine(ctx, "SN-Y")
		adminID := suite.insertUser(ctx, "admin")

		f := &flags.Flag{MachineSerialNumber: "SN-Y", Reason: flags.ReasonOther, Note: "note"}
		suite.mustCreate(ctx, f)
		suite.Require().NoError(suite.repo.Resolve(ctx, f.ID, adminID, ""))

		err := suite.repo.Resolve(ctx, f.ID, adminID, "")
		suite.Require().ErrorIs(err, flags.ErrNotFound)
	})
}

func TestFlagsRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(FlagsRepositoryTestSuite))
}
