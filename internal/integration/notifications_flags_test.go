//go:build integration

package integration

import (
	"context"
	"net/http"
	"strings"
	"time"

	"ralts-cms/internal/flags"
	"ralts-cms/internal/handlers"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/notifications"
	"ralts-cms/internal/users"

	"github.com/google/uuid"
)

// flagListResponse mirrors the body of GET /machines/{serial}/flags, which the
// handler writes as an untyped map.
type flagListResponse struct {
	Flags []*flags.Flag `json:"flags"`
}

// directoryResponse mirrors the body of GET /users/directory.
type directoryResponse struct {
	Users []*users.DirectoryEntry `json:"users"`
}

// notificationSettleWindow is how long we let the asynchronous notification
// writer run before asserting that a notification was *not* delivered. Notify
// spawns a goroutine that writes immediately, so this only needs to cover
// scheduling plus one round trip to Postgres.
const notificationSettleWindow = time.Second

func (s *IntegrationSuite) TestTier2_MachineAssigneeRegisteredUserOrFreeText() {
	_, _, adminTok := s.newUser("adm", users.RoleAdmin)
	tech, techID, _ := s.newUser("tech", users.RoleNonAdmin)

	s.Run("registered assignee derives person_in_charge from the username", func() {
		serial := newSerial()
		body := s.machineJSON(serial, "AssigneeCo")
		body["assigned_user_id"] = techID
		body["person_in_charge"] = "client supplied value that must be ignored"

		created := s.createMachine(adminTok, body)
		s.Require().NotNil(created.AssignedUserID)
		s.Equal(techID, *created.AssignedUserID)
		s.Equal(tech, created.PersonInCharge)
		s.Require().NotNil(created.AssignedUser)
		s.Equal(tech, created.AssignedUser.Username)

		var fetched machines.Machine
		s.expectJSON(s.get("/api/v1/machines/"+serial, adminTok), http.StatusOK, &fetched)
		s.Require().NotNil(fetched.AssignedUserID)
		s.Equal(techID, *fetched.AssignedUserID)
		s.Equal(tech, fetched.PersonInCharge)

		stored := s.assignedUserIDInDB(serial)
		s.Require().NotNil(stored)
		s.Equal(techID, *stored)
	})

	s.Run("free-text assignee is trimmed and stored without a user link", func() {
		serial := newSerial()
		body := s.machineJSON(serial, "GhostCo")
		body["person_in_charge"] = "  Jane Contractor  "

		created := s.createMachine(adminTok, body)
		s.Nil(created.AssignedUserID)
		s.Nil(created.AssignedUser)
		s.Equal("Jane Contractor", created.PersonInCharge)

		var fetched machines.Machine
		s.expectJSON(s.get("/api/v1/machines/"+serial, adminTok), http.StatusOK, &fetched)
		s.Nil(fetched.AssignedUserID)
		s.Nil(fetched.AssignedUser)
		s.Equal("Jane Contractor", fetched.PersonInCharge)
		s.Nil(s.assignedUserIDInDB(serial))
	})

	s.Run("free text matching a username is not linked to that account", func() {
		// Matching typed names against the directory is the frontend's job; the
		// API links an assignee only when it is given an explicit user id.
		serial := newSerial()
		body := s.machineJSON(serial, "GhostCo")
		body["person_in_charge"] = tech

		created := s.createMachine(adminTok, body)
		s.Equal(tech, created.PersonInCharge)
		s.Nil(created.AssignedUserID)
		s.Nil(s.assignedUserIDInDB(serial))
	})

	s.Run("unknown assigned_user_id is rejected", func() {
		body := s.machineJSON(newSerial(), "AssigneeCo")
		body["assigned_user_id"] = techID + 1_000_000
		s.expectJSON(s.postJSON("/api/v1/machines", body, adminTok), http.StatusBadRequest, nil)
	})

	s.Run("over-long free-text assignee is rejected", func() {
		body := s.machineJSON(newSerial(), "AssigneeCo")
		body["person_in_charge"] = strings.Repeat("x", maxFreeTextAssigneeLen+1)
		s.expectJSON(s.postJSON("/api/v1/machines", body, adminTok), http.StatusBadRequest, nil)
	})

	s.Run("update swaps between a registered user and free text", func() {
		serial := s.createMachineAssignedTo(adminTok, techID)

		toGhost := s.machineJSON(serial, "AssigneeCo")
		toGhost["person_in_charge"] = "Ghost Contractor"
		var updated machines.Machine
		s.expectJSON(s.putJSON("/api/v1/machines/"+serial, toGhost, adminTok), http.StatusOK, &updated)
		s.Nil(updated.AssignedUserID)
		s.Equal("Ghost Contractor", updated.PersonInCharge)
		s.Nil(s.assignedUserIDInDB(serial))

		toUser := s.machineJSON(serial, "AssigneeCo")
		toUser["assigned_user_id"] = techID
		s.expectJSON(s.putJSON("/api/v1/machines/"+serial, toUser, adminTok), http.StatusOK, &updated)
		s.Require().NotNil(updated.AssignedUserID)
		s.Equal(techID, *updated.AssignedUserID)
		s.Equal(tech, updated.PersonInCharge)
		stored := s.assignedUserIDInDB(serial)
		s.Require().NotNil(stored)
		s.Equal(techID, *stored)
	})

	s.Run("an update that names no assignee keeps the current one", func() {
		serial := s.createMachineAssignedTo(adminTok, techID)

		// Editing an unrelated field must not unlink the assignee, so this body
		// mentions neither assignee field.
		body := s.machineJSON(serial, "RenamedCo")
		delete(body, "person_in_charge")

		var updated machines.Machine
		s.expectJSON(s.putJSON("/api/v1/machines/"+serial, body, adminTok), http.StatusOK, &updated)
		s.Equal("RenamedCo", updated.Customer)
		s.Require().NotNil(updated.AssignedUserID)
		s.Equal(techID, *updated.AssignedUserID)
		s.Equal(tech, updated.PersonInCharge)
		stored := s.assignedUserIDInDB(serial)
		s.Require().NotNil(stored)
		s.Equal(techID, *stored)
	})

	s.Run("an account that cannot log in cannot be given a machine", func() {
		// The assignee dropdown only offers approved, active users; anything else
		// would be sent notifications it can never come back and act on.
		suspended := "susp_" + uuid.NewString()[:8]
		s.expectJSON(s.postJSON("/api/v1/users", map[string]any{
			"username": suspended,
			"email":    suspended + "@t.example",
			"password": integrationPassword,
			"role":     users.RoleNonAdmin,
			"approved": true,
			"status":   users.StatusSuspended,
		}, ""), http.StatusCreated, nil)

		body := s.machineJSON(newSerial(), "AssigneeCo")
		body["assigned_user_id"] = s.userIDByUsername(suspended)
		s.expectJSON(s.postJSON("/api/v1/machines", body, adminTok), http.StatusBadRequest, nil)
	})
}

func (s *IntegrationSuite) TestTier2_FlagCreationIsAdminOnly() {
	admin, _, adminTok := s.newUser("adm", users.RoleAdmin)
	_, techID, techTok := s.newUser("tech", users.RoleNonAdmin)
	serial := s.createMachineAssignedTo(adminTok, techID)

	s.Run("non-admin cannot raise a flag", func() {
		body := handlers.CreateFlagRequest{Reason: flags.ReasonMissingValues}
		s.expectJSON(s.postJSON("/api/v1/machines/"+serial+"/flags", body, techTok), http.StatusForbidden, nil)
	})

	var created flags.Flag
	s.Run("admin raises a missing_values flag", func() {
		created = s.raiseFlag(adminTok, serial, flags.ReasonMissingValues, "")
		s.NotEmpty(created.ID)
		s.Equal(serial, created.MachineSerialNumber)
		s.Equal(flags.ReasonMissingValues, created.Reason)
		s.Equal(flags.StatusOpen, created.Status)
	})

	s.Run("reason=other requires a note", func() {
		body := handlers.CreateFlagRequest{Reason: flags.ReasonOther}
		s.expectJSON(s.postJSON("/api/v1/machines/"+serial+"/flags", body, adminTok), http.StatusBadRequest, nil)

		// The machine is already flagged, so this rewrites its open flag and the
		// API answers 200 rather than 201.
		withNote := handlers.CreateFlagRequest{Reason: flags.ReasonOther, Note: "Customer disputes the reading"}
		var flag flags.Flag
		s.expectJSON(s.postJSON("/api/v1/machines/"+serial+"/flags", withNote, adminTok), http.StatusOK, &flag)
		s.Equal("Customer disputes the reading", flag.Note)
		s.Equal(created.ID, flag.ID)
	})

	s.Run("the machine carries a single open flag however often it is flagged", func() {
		open := s.listFlags(adminTok, serial, false)
		s.Require().Len(open, 1)
		s.Equal(created.ID, open[0].ID)
	})

	s.Run("requires_attention is reserved for the system", func() {
		body := handlers.CreateFlagRequest{Reason: flags.ReasonRequiresAttention}
		s.expectJSON(s.postJSON("/api/v1/machines/"+serial+"/flags", body, adminTok), http.StatusBadRequest, nil)
	})

	s.Run("unknown reason is rejected", func() {
		body := handlers.CreateFlagRequest{Reason: flags.Reason("not_a_reason")}
		s.expectJSON(s.postJSON("/api/v1/machines/"+serial+"/flags", body, adminTok), http.StatusBadRequest, nil)
	})

	s.Run("flagging a machine that does not exist is a 404", func() {
		body := handlers.CreateFlagRequest{Reason: flags.ReasonMissingValues}
		s.expectJSON(s.postJSON("/api/v1/machines/"+newSerial()+"/flags", body, adminTok), http.StatusNotFound, nil)
	})

	s.Run("any authenticated user can read a machine's flags", func() {
		list := s.listFlags(techTok, serial, false)
		s.Require().NotEmpty(list)
		found := false
		for _, f := range list {
			if f.ID == created.ID {
				found = true
				s.Require().NotNil(f.CreatedByUsername)
				s.Equal(admin, *f.CreatedByUsername)
			}
		}
		s.True(found, "expected the flag raised by the admin to be listed")

		s.expectJSON(s.get("/api/v1/machines/"+serial+"/flags", ""), http.StatusUnauthorized, nil)
	})
}

func (s *IntegrationSuite) TestTier2_FlagResolutionAllowsAdminOrAssignee() {
	_, _, adminTok := s.newUser("adm", users.RoleAdmin)
	tech, techID, techTok := s.newUser("tech", users.RoleNonAdmin)
	_, _, strangerTok := s.newUser("other", users.RoleNonAdmin)
	serial := s.createMachineAssignedTo(adminTok, techID)

	flag := s.raiseFlag(adminTok, serial, flags.ReasonMissingValues, "")
	resolvePath := "/api/v1/machines/" + serial + "/flags/" + flag.ID + "/resolve"

	s.Run("a non-admin who is not the assignee is refused", func() {
		s.expectJSON(s.postJSON(resolvePath, nil, strangerTok), http.StatusForbidden, nil)
		s.Equal(flags.StatusOpen, s.flagByID(techTok, serial, flag.ID).Status)
	})

	s.Run("the assignee can resolve a flag on their own machine", func() {
		s.expectJSON(s.postJSON(resolvePath, nil, techTok), http.StatusNoContent, nil)

		resolved := s.flagByID(techTok, serial, flag.ID)
		s.Equal(flags.StatusResolved, resolved.Status)
		s.Require().NotNil(resolved.ResolvedByUsername)
		s.Equal(tech, *resolved.ResolvedByUsername)

		// The two events are dated separately: closing the flag stamps its own
		// time and leaves the raise time as it was.
		s.Require().NotNil(resolved.ResolvedAt)
		s.False(resolved.CreatedAt.IsZero(), "the API reports when it was flagged")
		s.WithinDuration(flag.CreatedAt, resolved.CreatedAt, time.Second)
		s.False(resolved.ResolvedAt.Before(resolved.CreatedAt), "resolved no earlier than raised")
	})

	s.Run("a resolved flag drops out of the open lists", func() {
		for _, f := range s.listFlags(techTok, serial, false) {
			s.NotEqual(flag.ID, f.ID)
		}
		s.NotContains(s.openFlagsBySerial(adminTok, serial), serial)
	})

	s.Run("resolving twice reports not found", func() {
		s.expectJSON(s.postJSON(resolvePath, nil, techTok), http.StatusNotFound, nil)
	})

	s.Run("an unknown flag id is not found", func() {
		path := "/api/v1/machines/" + serial + "/flags/" + uuid.NewString() + "/resolve"
		s.expectJSON(s.postJSON(path, nil, techTok), http.StatusNotFound, nil)
	})

	s.Run("an admin can resolve a flag on someone else's machine", func() {
		other := s.raiseFlag(adminTok, serial, flags.ReasonOther, "Needs a site visit")
		path := "/api/v1/machines/" + serial + "/flags/" + other.ID + "/resolve"
		s.expectJSON(s.postJSON(path, nil, adminTok), http.StatusNoContent, nil)
		s.Equal(flags.StatusResolved, s.flagByID(adminTok, serial, other.ID).Status)
	})

	s.Run("an optional resolution note is stored on the flag", func() {
		noted := s.raiseFlag(adminTok, serial, flags.ReasonMissingValues, "")
		path := "/api/v1/machines/" + serial + "/flags/" + noted.ID + "/resolve"
		body := handlers.ResolveFlagRequest{Note: "  Filled in the district  "}
		s.expectJSON(s.postJSON(path, body, techTok), http.StatusNoContent, nil)

		resolved := s.flagByID(techTok, serial, noted.ID)
		s.Equal("Filled in the district", resolved.ResolutionNote, "stored trimmed")
	})

	s.Run("a resolution note over the limit is rejected and the flag stays open", func() {
		tooLong := s.raiseFlag(adminTok, serial, flags.ReasonMissingValues, "")
		path := "/api/v1/machines/" + serial + "/flags/" + tooLong.ID + "/resolve"
		body := handlers.ResolveFlagRequest{Note: strings.Repeat("a", flags.MaxNoteLength+1)}
		s.expectJSON(s.postJSON(path, body, techTok), http.StatusBadRequest, nil)

		s.Equal(flags.StatusOpen, s.flagByID(techTok, serial, tooLong.ID).Status)
	})

	s.Run("resolving without a note leaves the note empty", func() {
		// Its own machine: the flag rejected above is still open on serial, and a
		// machine holds only one open flag.
		own := s.createMachineAssignedTo(adminTok, techID)
		plain := s.raiseFlag(adminTok, own, flags.ReasonMissingValues, "")
		path := "/api/v1/machines/" + own + "/flags/" + plain.ID + "/resolve"
		s.expectJSON(s.postJSON(path, nil, techTok), http.StatusNoContent, nil)

		s.Empty(s.flagByID(techTok, own, plain.ID).ResolutionNote)
	})

	s.Run("flags on a free-text assignee's machine are admin-only", func() {
		ghostSerial := newSerial()
		body := s.machineJSON(ghostSerial, "GhostCo")
		body["person_in_charge"] = tech // same name, but no account link
		s.createMachine(adminTok, body)

		ghostFlag := s.raiseFlag(adminTok, ghostSerial, flags.ReasonMissingValues, "")
		path := "/api/v1/machines/" + ghostSerial + "/flags/" + ghostFlag.ID + "/resolve"
		s.expectJSON(s.postJSON(path, nil, techTok), http.StatusForbidden, nil)
		s.expectJSON(s.postJSON(path, nil, adminTok), http.StatusNoContent, nil)
	})
}

func (s *IntegrationSuite) TestTier2_OpenFlagBatchLookup() {
	_, _, adminTok := s.newUser("adm", users.RoleAdmin)
	_, _, techTok := s.newUser("tech", users.RoleNonAdmin)

	flagged := newSerial()
	alsoFlagged := newSerial()
	clean := newSerial()
	for _, serial := range []string{flagged, alsoFlagged, clean} {
		s.createMachine(adminTok, s.machineJSON(serial, "BatchCo"))
	}
	first := s.raiseFlag(adminTok, flagged, flags.ReasonMissingValues, "")
	s.raiseFlag(adminTok, alsoFlagged, flags.ReasonOther, "Awaiting parts")

	s.Run("only flagged serials appear in the batch response", func() {
		byMachine := s.openFlagsBySerial(adminTok, flagged, alsoFlagged, clean)
		s.Len(byMachine[flagged], 1)
		s.Equal(first.ID, byMachine[flagged][0].ID)
		s.Len(byMachine[alsoFlagged], 1)
		s.NotContains(byMachine, clean)
	})

	s.Run("the endpoint is open to any authenticated user", func() {
		// Badges are shown to every signed-in user, on any machine, so this has
		// to answer a non-admin token with the same data.
		s.Len(s.openFlagsBySerial(techTok, flagged)[flagged], 1)
	})

	s.Run("no serials yields an empty map", func() {
		var out handlers.BatchOpenByMachineResponse
		s.expectJSON(s.get("/api/v1/machines/flags/open-by-machine", adminTok), http.StatusOK, &out)
		s.Empty(out.Flags)
	})

	s.Run("too many serials is rejected", func() {
		serials := make([]string, 201)
		for i := range serials {
			serials[i] = newSerial()
		}
		resp := s.get("/api/v1/machines/flags/open-by-machine?serials="+strings.Join(serials, ","), adminTok)
		s.expectJSON(resp, http.StatusBadRequest, nil)
	})
}

func (s *IntegrationSuite) TestTier2_FlaggedRecordsListing() {
	_, _, adminTok := s.newUser("adm", users.RoleAdmin)
	_, techID, techTok := s.newUser("tech", users.RoleNonAdmin)
	_, _, strangerTok := s.newUser("other", users.RoleNonAdmin)

	unassignedSerial := newSerial()
	cleanSerial := newSerial()
	s.createMachine(adminTok, s.machineJSON(unassignedSerial, "ListCo"))
	assigned := s.createMachineAssignedTo(adminTok, techID)

	unassignedFlag := s.raiseFlag(adminTok, unassignedSerial, flags.ReasonMissingValues, "")
	resolvedFlag := s.raiseFlag(adminTok, assigned, flags.ReasonOther, "Awaiting parts")
	s.expectJSON(
		s.postJSON("/api/v1/machines/"+assigned+"/flags/"+resolvedFlag.ID+"/resolve", nil, adminTok),
		http.StatusNoContent, nil,
	)
	assignedFlag := s.raiseFlag(adminTok, assigned, flags.ReasonMissingValues, "")
	s.createMachine(adminTok, s.machineJSON(cleanSerial, "ListCo")) // never flagged

	// The suite shares one database across tests, so the admin (unscoped) view
	// also carries flags other tests left behind. Assertions on it therefore ask
	// where this test's own flags appear rather than how many rows came back; the
	// scoped view belongs to a user created here, so it can be checked exactly.
	listFlags := func(token, query string) handlers.ListFlagsResponse {
		var out handlers.ListFlagsResponse
		s.expectJSON(s.get("/api/v1/flags"+query+"&limit=200", token), http.StatusOK, &out)
		return out
	}
	indexOf := func(list []*flags.Flag, id string) int {
		for i, f := range list {
			if f.ID == id {
				return i
			}
		}
		return -1
	}

	s.Run("admins see open flags across every machine", func() {
		out := listFlags(adminTok, "?status=open")
		s.Equal(handlers.FlagScopeAll, out.Scope)

		newest, older := indexOf(out.Flags, assignedFlag.ID), indexOf(out.Flags, unassignedFlag.ID)
		s.Require().NotEqual(-1, newest, "the flag on the assigned machine is listed")
		s.Require().NotEqual(-1, older, "the flag on the unassigned machine is listed too")
		s.Less(newest, older, "newest first")
		s.Require().NotNil(out.Flags[newest].CreatedByUsername)
		s.Equal(-1, indexOf(out.Flags, resolvedFlag.ID), "resolved flags are not open")
	})

	s.Run("non-admins only see flags on machines assigned to them", func() {
		out := listFlags(techTok, "?status=open")
		s.Equal(handlers.FlagScopeAssigned, out.Scope)
		s.Equal(1, out.Count)
		s.Require().Len(out.Flags, 1)
		s.Equal(assignedFlag.ID, out.Flags[0].ID)
		s.Equal(assigned, out.Flags[0].MachineSerialNumber)

		// Only the open flag: the resolved one on the same machine was closed by
		// the admin, so it is not part of this user's history.
		all := listFlags(techTok, "?status=all")
		s.Equal(1, all.Count)
		s.Equal(-1, indexOf(all.Flags, resolvedFlag.ID))

		// Someone with no machines has an empty page rather than an error.
		s.Equal(0, listFlags(strangerTok, "?status=all").Count)
	})

	s.Run("resolved history is limited to what the caller closed themselves", func() {
		// Being assigned a machine does not hand over what others closed on it.
		resolved := listFlags(techTok, "?status=resolved")
		s.Zero(resolved.Count)
		s.Equal(-1, indexOf(resolved.Flags, resolvedFlag.ID))

		// A second machine of theirs, since the first one's flag is still open and
		// a machine holds only one.
		secondSerial := s.createMachineAssignedTo(adminTok, techID)
		own := s.raiseFlag(adminTok, secondSerial, flags.ReasonMissingValues, "")
		s.expectJSON(
			s.postJSON("/api/v1/machines/"+secondSerial+"/flags/"+own.ID+"/resolve", nil, techTok),
			http.StatusNoContent, nil,
		)

		mine := listFlags(techTok, "?status=resolved")
		s.Require().Len(mine.Flags, 1)
		s.Equal(own.ID, mine.Flags[0].ID)
		s.Require().NotNil(mine.Flags[0].ResolvedByUsername)

		all := listFlags(adminTok, "?status=all")
		s.NotEqual(-1, indexOf(all.Flags, resolvedFlag.ID), "an admin sees every resolution")
		s.NotEqual(-1, indexOf(all.Flags, assignedFlag.ID))
	})

	s.Run("limit and offset page through the results", func() {
		var page handlers.ListFlagsResponse
		s.expectJSON(s.get("/api/v1/flags?status=all&limit=1", techTok), http.StatusOK, &page)
		s.Require().Len(page.Flags, 1)
		s.Equal(2, page.Count, "count reports the whole result set, not the page")

		var next handlers.ListFlagsResponse
		s.expectJSON(s.get("/api/v1/flags?status=all&limit=1&offset=1", techTok), http.StatusOK, &next)
		s.Require().Len(next.Flags, 1)
		s.NotEqual(page.Flags[0].ID, next.Flags[0].ID)
	})

	s.Run("a non-admin resolves a flag from their own listing", func() {
		path := "/api/v1/machines/" + assigned + "/flags/" + assignedFlag.ID + "/resolve"
		s.expectJSON(s.postJSON(path, nil, techTok), http.StatusNoContent, nil)

		s.Empty(listFlags(techTok, "?status=open").Flags, "the resolved flag leaves their open list")

		adminView := listFlags(adminTok, "?status=open")
		s.Equal(-1, indexOf(adminView.Flags, assignedFlag.ID), "and the admin's view as well")
		s.NotEqual(-1, indexOf(adminView.Flags, unassignedFlag.ID), "the other flag is untouched")
	})

	s.Run("authentication is required", func() {
		s.expectJSON(s.get("/api/v1/flags", ""), http.StatusUnauthorized, nil)
	})

	s.Run("invalid parameters are rejected", func() {
		for _, query := range []string{"?status=bogus", "?limit=0", "?limit=201", "?offset=-1"} {
			s.expectJSON(s.get("/api/v1/flags"+query, adminTok), http.StatusBadRequest, nil)
		}
	})
}

func (s *IntegrationSuite) TestTier3_AssignmentNotifiesTheNewAssignee() {
	admin, adminID, adminTok := s.newUser("adm", users.RoleAdmin)
	_, techID, techTok := s.newUser("tech", users.RoleNonAdmin)
	_, standInID, standInTok := s.newUser("standin", users.RoleNonAdmin)

	serial := s.createMachineAssignedTo(adminTok, techID)

	n := s.awaitNotification(techTok, func(n *notifications.Notification) bool {
		return n.Type == notifications.TypeAssigned && n.MachineSerialNumber != nil && *n.MachineSerialNumber == serial
	})
	s.Equal("You were assigned to machine "+serial, n.Title)
	s.Equal("Customer: FlagCo", n.Body)
	s.Require().NotNil(n.ActorUserID)
	s.Equal(adminID, *n.ActorUserID)
	s.Require().NotNil(n.ActorUsername)
	s.Equal(admin, *n.ActorUsername)
	s.Nil(n.ReadAt)
	s.Nil(n.FlagID)

	s.Run("the unread count reflects the new notification", func() {
		var out map[string]int
		s.expectJSON(s.get("/api/v1/notifications/unread-count", techTok), http.StatusOK, &out)
		s.GreaterOrEqual(out["unread_count"], 1)
	})

	s.Run("saving the machine again does not re-notify", func() {
		body := s.machineJSON(serial, "FlagCo")
		body["assigned_user_id"] = techID
		s.expectJSON(s.putJSON("/api/v1/machines/"+serial, body, adminTok), http.StatusOK, nil)

		time.Sleep(notificationSettleWindow)
		s.Equal(1, s.countNotifications(techTok, func(n *notifications.Notification) bool {
			return n.Type == notifications.TypeAssigned && n.MachineSerialNumber != nil && *n.MachineSerialNumber == serial
		}))
	})

	s.Run("reassignment notifies the new assignee only", func() {
		body := s.machineJSON(serial, "FlagCo")
		body["assigned_user_id"] = standInID
		s.expectJSON(s.putJSON("/api/v1/machines/"+serial, body, adminTok), http.StatusOK, nil)

		s.awaitNotification(standInTok, func(n *notifications.Notification) bool {
			return n.Type == notifications.TypeAssigned && n.MachineSerialNumber != nil && *n.MachineSerialNumber == serial
		})
		s.Equal(1, s.countNotifications(techTok, func(n *notifications.Notification) bool {
			return n.Type == notifications.TypeAssigned && n.MachineSerialNumber != nil && *n.MachineSerialNumber == serial
		}))
	})

	s.Run("assigning a machine to yourself notifies nobody", func() {
		selfSerial := s.createMachineAssignedTo(adminTok, adminID)
		time.Sleep(notificationSettleWindow)
		s.Zero(s.countNotifications(adminTok, func(n *notifications.Notification) bool {
			return n.MachineSerialNumber != nil && *n.MachineSerialNumber == selfSerial
		}))
	})
}

// TestTier3_ReflaggingReplacesTheOpenFlag covers the one-open-flag-per-machine
// rule: flagging an already-flagged machine rewrites its open flag rather than
// stacking a second one, and the assignee hears about the new note.
func (s *IntegrationSuite) TestTier3_ReflaggingReplacesTheOpenFlag() {
	_, _, adminTok := s.newUser("adm", users.RoleAdmin)
	_, techID, techTok := s.newUser("tech", users.RoleNonAdmin)
	serial := s.createMachineAssignedTo(adminTok, techID)

	first := s.raiseFlag(adminTok, serial, flags.ReasonMissingValues, "District is blank")
	s.awaitNotification(techTok, func(n *notifications.Notification) bool {
		return n.FlagID != nil && *n.FlagID == first.ID
	})

	again := s.reflag(adminTok, serial, flags.ReasonOther, "Still incomplete after the site visit")
	s.Equal(first.ID, again.ID, "the open flag keeps its identity")
	s.Equal(flags.ReasonOther, again.Reason)
	s.Equal("Still incomplete after the site visit", again.Note)
	s.False(again.CreatedAt.Before(first.CreatedAt), "the raise time moves forward")

	s.Run("the machine holds one flag, carrying the latest reason and note", func() {
		all := s.listFlags(adminTok, serial, true)
		s.Require().Len(all, 1)
		s.Equal(first.ID, all[0].ID)
		s.Equal(flags.ReasonOther, all[0].Reason)
		s.Equal("Still incomplete after the site visit", all[0].Note)
	})

	s.Run("the assignee is notified about the new note too", func() {
		s.awaitNotification(techTok, func(n *notifications.Notification) bool {
			return n.FlagID != nil && *n.FlagID == first.ID &&
				n.Body == "Still incomplete after the site visit"
		})
		s.Equal(2, s.countNotifications(techTok, func(n *notifications.Notification) bool {
			return n.FlagID != nil && *n.FlagID == first.ID
		}), "one notification per flagging, both pointing at the same flag")
	})

	s.Run("once resolved, flagging again starts a new record", func() {
		path := "/api/v1/machines/" + serial + "/flags/" + first.ID + "/resolve"
		s.expectJSON(s.postJSON(path, nil, techTok), http.StatusNoContent, nil)

		fresh := s.raiseFlag(adminTok, serial, flags.ReasonMissingValues, "")
		s.NotEqual(first.ID, fresh.ID)
		s.Len(s.listFlags(adminTok, serial, true), 2, "the resolved flag is kept as history")
	})
}

// TestTier3_FlaggingYourOwnMachineWorksButDoesNotNotifyYou pins the admin→admin
// case: an admin can flag a machine assigned to themselves and work it off from
// the flags page, but is not sent a notification about their own action.
func (s *IntegrationSuite) TestTier3_FlaggingYourOwnMachineWorksButDoesNotNotifyYou() {
	_, adminID, adminTok := s.newUser("adm", users.RoleAdmin)
	serial := s.createMachineAssignedTo(adminTok, adminID)

	flag := s.raiseFlag(adminTok, serial, flags.ReasonOther, "Needs a firmware update")
	s.Equal(flags.StatusOpen, flag.Status)

	s.Run("the flag is stored and listed on the machine", func() {
		s.Equal(flag.ID, s.flagByID(adminTok, serial, flag.ID).ID)
		s.Contains(s.openFlagsBySerial(adminTok, serial), serial)
	})

	s.Run("no notification is delivered for your own action", func() {
		time.Sleep(notificationSettleWindow)
		s.Zero(s.countNotifications(adminTok, func(n *notifications.Notification) bool {
			return n.FlagID != nil && *n.FlagID == flag.ID
		}))
	})

	s.Run("it is still resolvable, with a note", func() {
		path := "/api/v1/machines/" + serial + "/flags/" + flag.ID + "/resolve"
		body := handlers.ResolveFlagRequest{Note: "Updated the firmware"}
		s.expectJSON(s.postJSON(path, body, adminTok), http.StatusNoContent, nil)

		resolved := s.flagByID(adminTok, serial, flag.ID)
		s.Equal(flags.StatusResolved, resolved.Status)
		s.Equal("Updated the firmware", resolved.ResolutionNote)
	})
}

func (s *IntegrationSuite) TestTier3_FlaggingNotifiesTheAssigneeWhoResolvesFromTheirFlagsPage() {
	admin, _, adminTok := s.newUser("adm", users.RoleAdmin)
	_, techID, techTok := s.newUser("tech", users.RoleNonAdmin)
	serial := s.createMachineAssignedTo(adminTok, techID)

	flag := s.raiseFlag(adminTok, serial, flags.ReasonOther, "Meter reading looks wrong")
	n := s.awaitNotification(techTok, func(n *notifications.Notification) bool {
		return n.FlagID != nil && *n.FlagID == flag.ID
	})

	s.Equal(notifications.TypeFlagged, n.Type)
	s.Equal("Machine "+serial+" flagged", n.Title)
	s.Equal("Meter reading looks wrong", n.Body)
	s.Require().NotNil(n.ActorUsername)
	s.Equal(admin, *n.ActorUsername)
	s.Require().NotNil(n.FlagStatus)
	s.Equal(string(flags.StatusOpen), *n.FlagStatus, "the inbox shows the flag as still needing work")

	s.Run("the flag appears on the assignee's own flagged records page", func() {
		var listing handlers.ListFlagsResponse
		s.expectJSON(s.get("/api/v1/flags", techTok), http.StatusOK, &listing)
		s.Equal(handlers.FlagScopeAssigned, listing.Scope)
		s.Require().Len(listing.Flags, 1)
		s.Equal(flag.ID, listing.Flags[0].ID)
	})

	s.Run("resolving it there flips the status the inbox reports", func() {
		path := "/api/v1/machines/" + serial + "/flags/" + flag.ID + "/resolve"
		s.expectJSON(s.postJSON(path, nil, techTok), http.StatusNoContent, nil)

		after := s.awaitNotification(techTok, func(c *notifications.Notification) bool {
			return c.ID == n.ID && c.FlagStatus != nil && *c.FlagStatus == string(flags.StatusResolved)
		})
		s.Equal(flag.ID, *after.FlagID)
	})

	s.Run("marking the notification read clears it from the unread list", func() {
		s.expectJSON(s.postJSON("/api/v1/notifications/"+n.ID+"/read", nil, techTok), http.StatusNoContent, nil)

		var unread handlers.ListNotificationsResponse
		s.expectJSON(s.get("/api/v1/notifications?unread_only=true", techTok), http.StatusOK, &unread)
		for _, item := range unread.Notifications {
			s.NotEqual(n.ID, item.ID)
		}

		read := s.awaitNotification(techTok, func(c *notifications.Notification) bool { return c.ID == n.ID })
		s.NotNil(read.ReadAt)
	})
}

// TestTier3_ResolutionByAnAssigneeNotifiesTheRaisingAdmin covers the reply half
// of the flag conversation: the admin who asked for the work hears when an
// assignee says it is done, while admins clearing flags themselves stay quiet.
func (s *IntegrationSuite) TestTier3_ResolutionByAnAssigneeNotifiesTheRaisingAdmin() {
	_, _, adminTok := s.newUser("adm", users.RoleAdmin)
	tech, techID, techTok := s.newUser("tech", users.RoleNonAdmin)
	serial := s.createMachineAssignedTo(adminTok, techID)

	flag := s.raiseFlag(adminTok, serial, flags.ReasonMissingValues, "District is blank")
	path := "/api/v1/machines/" + serial + "/flags/" + flag.ID + "/resolve"

	s.Run("the raiser is told what was done", func() {
		body := handlers.ResolveFlagRequest{Note: "Filled in the district"}
		s.expectJSON(s.postJSON(path, body, techTok), http.StatusNoContent, nil)

		n := s.awaitNotification(adminTok, func(n *notifications.Notification) bool {
			return n.FlagID != nil && *n.FlagID == flag.ID
		})
		s.Equal(notifications.TypeFlagResolved, n.Type)
		s.Equal("Machine "+serial+": flag resolved", n.Title)
		s.Equal("Filled in the district", n.Body)
		s.Require().NotNil(n.ActorUsername)
		s.Equal(tech, *n.ActorUsername)
		s.Require().NotNil(n.MachineSerialNumber)
		s.Equal(serial, *n.MachineSerialNumber)
		s.Nil(n.ReadAt, "it arrives unread")
	})

	s.Run("the assignee is not told about their own resolution", func() {
		time.Sleep(notificationSettleWindow)
		s.Zero(s.countNotifications(techTok, func(n *notifications.Notification) bool {
			return n.Type == notifications.TypeFlagResolved
		}))
	})

	s.Run("an admin resolving somebody else's flag notifies nobody", func() {
		_, _, otherAdminTok := s.newUser("adm2", users.RoleAdmin)
		second := s.raiseFlag(otherAdminTok, serial, flags.ReasonOther, "Check the serial plate")

		secondPath := "/api/v1/machines/" + serial + "/flags/" + second.ID + "/resolve"
		s.expectJSON(s.postJSON(secondPath, nil, adminTok), http.StatusNoContent, nil)

		time.Sleep(notificationSettleWindow)
		s.Zero(s.countNotifications(otherAdminTok, func(n *notifications.Notification) bool {
			return n.FlagID != nil && *n.FlagID == second.ID
		}))
	})
}

func (s *IntegrationSuite) TestTier3_FreeTextAssigneeIsNeverNotified() {
	_, _, adminTok := s.newUser("adm", users.RoleAdmin)
	tech, _, techTok := s.newUser("tech", users.RoleNonAdmin)

	// The machine names the user as free text but carries no assigned_user_id,
	// so there is no recipient for either notification type.
	serial := newSerial()
	body := s.machineJSON(serial, "GhostCo")
	body["person_in_charge"] = tech
	s.createMachine(adminTok, body)
	s.raiseFlag(adminTok, serial, flags.ReasonMissingValues, "")

	time.Sleep(notificationSettleWindow)

	var out handlers.ListNotificationsResponse
	s.expectJSON(s.get("/api/v1/notifications", techTok), http.StatusOK, &out)
	s.Empty(out.Notifications)
	s.Zero(out.UnreadCount)
}

func (s *IntegrationSuite) TestTier2_NotificationsAreScopedToTheCaller() {
	_, _, adminTok := s.newUser("adm", users.RoleAdmin)
	_, firstID, firstTok := s.newUser("first", users.RoleNonAdmin)
	_, secondID, secondTok := s.newUser("second", users.RoleNonAdmin)

	firstSerial := s.createMachineAssignedTo(adminTok, firstID)
	secondSerial := s.createMachineAssignedTo(adminTok, secondID)

	firstNote := s.awaitNotification(firstTok, func(n *notifications.Notification) bool {
		return n.MachineSerialNumber != nil && *n.MachineSerialNumber == firstSerial
	})
	s.awaitNotification(secondTok, func(n *notifications.Notification) bool {
		return n.MachineSerialNumber != nil && *n.MachineSerialNumber == secondSerial
	})

	s.Run("a user only sees their own inbox", func() {
		var out handlers.ListNotificationsResponse
		s.expectJSON(s.get("/api/v1/notifications", firstTok), http.StatusOK, &out)
		s.Require().NotEmpty(out.Notifications)
		for _, n := range out.Notifications {
			s.Equal(firstID, n.UserID)
			if n.MachineSerialNumber != nil {
				s.NotEqual(secondSerial, *n.MachineSerialNumber)
			}
		}
	})

	s.Run("marking another user's notification read is a no-op", func() {
		s.expectJSON(s.postJSON("/api/v1/notifications/"+firstNote.ID+"/read", nil, secondTok), http.StatusNoContent, nil)

		still := s.awaitNotification(firstTok, func(n *notifications.Notification) bool { return n.ID == firstNote.ID })
		s.Nil(still.ReadAt, "another user must not be able to read the recipient's notification")
	})

	s.Run("read-all only touches the caller's rows", func() {
		var updated map[string]int
		s.expectJSON(s.postJSON("/api/v1/notifications/read-all", nil, firstTok), http.StatusOK, &updated)
		s.GreaterOrEqual(updated["updated"], 1)

		var firstCount, secondCount map[string]int
		s.expectJSON(s.get("/api/v1/notifications/unread-count", firstTok), http.StatusOK, &firstCount)
		s.Zero(firstCount["unread_count"])
		s.expectJSON(s.get("/api/v1/notifications/unread-count", secondTok), http.StatusOK, &secondCount)
		s.GreaterOrEqual(secondCount["unread_count"], 1)
	})

	s.Run("the inbox requires authentication", func() {
		s.expectJSON(s.get("/api/v1/notifications", ""), http.StatusUnauthorized, nil)
		s.expectJSON(s.get("/api/v1/notifications/unread-count", ""), http.StatusUnauthorized, nil)
	})
}

func (s *IntegrationSuite) TestTier2_AssigneeDirectoryListsApprovedUsersOnly() {
	approved, _, techTok := s.newUser("tech", users.RoleNonAdmin)

	pending := "pend_" + uuid.NewString()[:8]
	s.expectJSON(s.postJSON("/api/v1/users", map[string]any{
		"username": pending,
		"email":    pending + "@t.example",
		"password": integrationPassword,
		"role":     users.RoleNonAdmin,
	}, ""), http.StatusCreated, nil)

	var out directoryResponse
	s.expectJSON(s.get("/api/v1/users/directory", techTok), http.StatusOK, &out)

	usernames := make([]string, 0, len(out.Users))
	for _, e := range out.Users {
		s.NotZero(e.ID)
		usernames = append(usernames, e.Username)
	}
	s.Contains(usernames, approved, "the picker must offer approved users to any authenticated caller")
	s.NotContains(usernames, pending, "users awaiting approval cannot be assigned")

	s.expectJSON(s.get("/api/v1/users/directory", ""), http.StatusUnauthorized, nil)
}

// maxFreeTextAssigneeLen mirrors the handler's person_in_charge limit, which
// matches the width of the machines."personInCharge" column.
const maxFreeTextAssigneeLen = 200

func newSerial() string {
	return "SN_" + uuid.NewString()[:8]
}

// createMachine posts a machine record and returns the created representation.
func (s *IntegrationSuite) createMachine(token string, body map[string]any) machines.Machine {
	s.T().Helper()
	var created machines.Machine
	s.expectJSON(s.postJSON("/api/v1/machines", body, token), http.StatusCreated, &created)
	return created
}

// createMachineAssignedTo creates a machine owned by a registered user and
// returns its serial number.
func (s *IntegrationSuite) createMachineAssignedTo(token string, assigneeID int64) string {
	s.T().Helper()
	serial := newSerial()
	body := s.machineJSON(serial, "FlagCo")
	body["assigned_user_id"] = assigneeID
	s.createMachine(token, body)
	return serial
}

// raiseFlag flags a machine as the given (admin) caller and returns the flag,
// expecting the machine to have had no open flag (201).
func (s *IntegrationSuite) raiseFlag(token, serial string, reason flags.Reason, note string) flags.Flag {
	s.T().Helper()
	var created flags.Flag
	body := handlers.CreateFlagRequest{Reason: reason, Note: note}
	s.expectJSON(s.postJSON("/api/v1/machines/"+serial+"/flags", body, token), http.StatusCreated, &created)
	return created
}

// reflag flags a machine that already carries an open flag, which the API
// rewrites in place and reports with 200 rather than 201.
func (s *IntegrationSuite) reflag(token, serial string, reason flags.Reason, note string) flags.Flag {
	s.T().Helper()
	var updated flags.Flag
	body := handlers.CreateFlagRequest{Reason: reason, Note: note}
	s.expectJSON(s.postJSON("/api/v1/machines/"+serial+"/flags", body, token), http.StatusOK, &updated)
	return updated
}

// listFlags reads a machine's flags, optionally including resolved ones.
func (s *IntegrationSuite) listFlags(token, serial string, includeResolved bool) []*flags.Flag {
	s.T().Helper()
	path := "/api/v1/machines/" + serial + "/flags"
	if includeResolved {
		path += "?include_resolved=true"
	}
	var out flagListResponse
	s.expectJSON(s.get(path, token), http.StatusOK, &out)
	return out.Flags
}

// flagByID finds a flag (open or resolved) on a machine, failing if it is gone.
func (s *IntegrationSuite) flagByID(token, serial, id string) *flags.Flag {
	s.T().Helper()
	for _, f := range s.listFlags(token, serial, true) {
		if f.ID == id {
			return f
		}
	}
	s.FailNowf("flag missing", "flag %s not found on machine %s", id, serial)
	return nil
}

// openFlagsBySerial calls the batch badge endpoint for the given serials.
func (s *IntegrationSuite) openFlagsBySerial(token string, serials ...string) map[string][]*flags.Flag {
	s.T().Helper()
	var out handlers.BatchOpenByMachineResponse
	path := "/api/v1/machines/flags/open-by-machine?serials=" + strings.Join(serials, ",")
	s.expectJSON(s.get(path, token), http.StatusOK, &out)
	return out.Flags
}

// awaitNotification polls the caller's inbox until a notification satisfying
// match appears. Notifications are written asynchronously, so a single read
// right after the triggering request would be racy.
func (s *IntegrationSuite) awaitNotification(token string, match func(*notifications.Notification) bool) *notifications.Notification {
	s.T().Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var out handlers.ListNotificationsResponse
		s.expectJSON(s.get("/api/v1/notifications?limit=200", token), http.StatusOK, &out)
		for _, n := range out.Notifications {
			if match(n) {
				return n
			}
		}
		if time.Now().After(deadline) {
			s.FailNowf("notification not delivered", "no match among %d notifications", len(out.Notifications))
			return nil
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// countNotifications counts the caller's notifications satisfying match. Use it
// with a settle window when asserting that nothing extra was delivered.
func (s *IntegrationSuite) countNotifications(token string, match func(*notifications.Notification) bool) int {
	s.T().Helper()
	var out handlers.ListNotificationsResponse
	s.expectJSON(s.get("/api/v1/notifications?limit=200", token), http.StatusOK, &out)
	count := 0
	for _, n := range out.Notifications {
		if match(n) {
			count++
		}
	}
	return count
}

// assignedUserIDInDB reads the assignee FK straight from Postgres to confirm the
// handler persisted (or cleared) the link rather than only echoing it back.
func (s *IntegrationSuite) assignedUserIDInDB(serial string) *int64 {
	s.T().Helper()
	var id *int64
	err := s.db.PostgresClient.QueryRow(context.Background(),
		`SELECT "assignedUserId" FROM machines WHERE "serialNumber" = $1`, serial).Scan(&id)
	s.Require().NoError(err)
	return id
}
