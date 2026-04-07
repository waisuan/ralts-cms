//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"ralts-cms/internal/handlers"
	"ralts-cms/internal/users"

	"github.com/google/uuid"
)

func (s *IntegrationSuite) TestTier1_Health() {
	resp, err := http.Get(s.apiURL("/health"))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *IntegrationSuite) TestTier1_ProtectedRouteRequiresAuth() {
	resp := s.get("/api/v1/machines", "")
	defer resp.Body.Close()
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *IntegrationSuite) TestTier1_PendingUserCannotLogin() {
	u := "u_" + uuid.NewString()[:8]
	body := map[string]any{
		"username": u, "email": u + "@t.example", "password": "pw12345678", "role": users.RoleNonAdmin,
	}
	resp := s.postJSON("/api/v1/users", body, "")
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode, string(raw))

	resp2 := s.postJSON("/api/v1/users/login", map[string]string{"username": u, "password": "pw12345678"}, "")
	defer resp2.Body.Close()
	s.Equal(http.StatusForbidden, resp2.StatusCode)
}

func (s *IntegrationSuite) TestTier1_AdminUserLoginAndMachineCRUD() {
	admin := "adm_" + uuid.NewString()[:8]
	s.createApprovedUser(admin, "pw12345678", admin+"@t.example", users.RoleAdmin)
	token := s.login(admin, "pw12345678")

	sn := "SN_" + uuid.NewString()[:8]
	resp := s.postJSON("/api/v1/machines", s.machineJSON(sn, "Acme"), token)
	defer resp.Body.Close()
	rb, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode, string(rb))

	respG := s.get("/api/v1/machines/"+sn, token)
	defer respG.Body.Close()
	rg, err := io.ReadAll(respG.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, respG.StatusCode, string(rg))

	upd := s.machineJSON(sn, "AcmeRenamed")
	respU := s.putJSON("/api/v1/machines/"+sn, upd, token)
	defer respU.Body.Close()
	ru, err := io.ReadAll(respU.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, respU.StatusCode, string(ru))

	respD := s.delete("/api/v1/machines/"+sn, token)
	defer respD.Body.Close()
	rd, err := io.ReadAll(respD.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, respD.StatusCode, string(rd))
}

func (s *IntegrationSuite) TestTier2_MachineListPaginationAndSearch() {
	admin := "adm_" + uuid.NewString()[:8]
	s.createApprovedUser(admin, "pw12345678", admin+"@t.example", users.RoleAdmin)
	token := s.login(admin, "pw12345678")

	tokenTag := "Tok_" + uuid.NewString()[:8]
	for i := range 3 {
		sn := "SN_" + uuid.NewString()[:8]
		resp := s.postJSON("/api/v1/machines", s.machineJSON(sn, tokenTag+"_"+string(rune('A'+i))), token)
		_, err := io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, resp.StatusCode)
	}

	resp := s.get("/api/v1/machines?limit=2&offset=0", token)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)
	var list handlers.ListMachinesResponse
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&list))
	s.Len(list.Machines, 2)
	s.GreaterOrEqual(list.Count, int32(3))

	resp2 := s.get("/api/v1/machines?limit=2&offset=2", token)
	defer resp2.Body.Close()
	s.Equal(http.StatusOK, resp2.StatusCode)
	s.Require().NoError(json.NewDecoder(resp2.Body).Decode(&list))
	s.GreaterOrEqual(len(list.Machines), 1)

	resp3 := s.get("/api/v1/machines?q="+tokenTag, token)
	defer resp3.Body.Close()
	s.Equal(http.StatusOK, resp3.StatusCode)
	s.Require().NoError(json.NewDecoder(resp3.Body).Decode(&list))
	s.GreaterOrEqual(list.Count, int32(3))
}

func (s *IntegrationSuite) TestTier2_InvalidMachineLimitRejected() {
	admin := "adm_" + uuid.NewString()[:8]
	s.createApprovedUser(admin, "pw12345678", admin+"@t.example", users.RoleAdmin)
	token := s.login(admin, "pw12345678")

	resp := s.get("/api/v1/machines?limit=9999", token)
	defer resp.Body.Close()
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *IntegrationSuite) TestTier2_MaintenanceCRUD() {
	admin := "adm_" + uuid.NewString()[:8]
	s.createApprovedUser(admin, "pw12345678", admin+"@t.example", users.RoleAdmin)
	token := s.login(admin, "pw12345678")

	sn := "SN_" + uuid.NewString()[:8]
	resp := s.postJSON("/api/v1/machines", s.machineJSON(sn, "MaintCo"), token)
	_, err := io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	wo := "WO_" + uuid.NewString()[:8]
	body := map[string]any{
		"work_order_number": wo,
		"work_order_date":   time.Now().UTC().Format(time.RFC3339),
		"action_taken":      "Replaced filter",
		"reported_by":       "tech",
		"work_order_type":   "Preventive",
	}
	respC := s.postJSON("/api/v1/machines/"+sn+"/maintenance", body, token)
	defer respC.Body.Close()
	rc, err := io.ReadAll(respC.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, respC.StatusCode, string(rc))

	respL := s.get("/api/v1/machines/"+sn+"/maintenance?limit=10&offset=0", token)
	defer respL.Body.Close()
	s.Equal(http.StatusOK, respL.StatusCode)

	respG := s.get("/api/v1/machines/"+sn+"/maintenance/"+wo, token)
	defer respG.Body.Close()
	s.Equal(http.StatusOK, respG.StatusCode)
}

func (s *IntegrationSuite) TestTier2_AdminUsersAndNonAdminDenied() {
	admin := "adm_" + uuid.NewString()[:8]
	s.createApprovedUser(admin, "pw12345678", admin+"@t.example", users.RoleAdmin)
	adminTok := s.login(admin, "pw12345678")

	reg := "reg_" + uuid.NewString()[:8]
	s.createApprovedUser(reg, "pw12345678", reg+"@t.example", users.RoleNonAdmin)
	regID := s.userIDByUsername(reg)
	s.Require().Greater(regID, int64(0))

	userTok := s.login(reg, "pw12345678")
	respDeny := s.get("/api/v1/admin/users?limit=10&offset=0", userTok)
	defer respDeny.Body.Close()
	s.Equal(http.StatusForbidden, respDeny.StatusCode)

	respList := s.get("/api/v1/admin/users?limit=10&offset=0", adminTok)
	defer respList.Body.Close()
	s.Equal(http.StatusOK, respList.StatusCode)
	var lu handlers.ListUsersResponse
	s.Require().NoError(json.NewDecoder(respList.Body).Decode(&lu))
	s.GreaterOrEqual(lu.TotalCount, 2)

	respUp := s.putJSON("/api/v1/admin/users/"+strconv.FormatInt(regID, 10)+"/status", map[string]string{
		"status": users.StatusSuspended,
	}, adminTok)
	defer respUp.Body.Close()
	ru, err := io.ReadAll(respUp.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, respUp.StatusCode, string(ru))
}

func (s *IntegrationSuite) TestTier2_CSVExportMachinesAndMaintenance() {
	admin := "adm_" + uuid.NewString()[:8]
	s.createApprovedUser(admin, "pw12345678", admin+"@t.example", users.RoleAdmin)
	token := s.login(admin, "pw12345678")

	sn := "SN_" + uuid.NewString()[:8]
	resp := s.postJSON("/api/v1/machines", s.machineJSON(sn, "CSVCustomer"), token)
	_, err := io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	respCSV := s.get("/api/v1/machines/export/csv", token)
	defer respCSV.Body.Close()
	s.Equal(http.StatusOK, respCSV.StatusCode)
	s.Contains(respCSV.Header.Get("Content-Type"), "text/csv")
	body, err := io.ReadAll(respCSV.Body)
	s.Require().NoError(err)
	s.Contains(string(body), sn)

	wo := "WO_" + uuid.NewString()[:8]
	maintBody := map[string]any{
		"work_order_number": wo,
		"work_order_date":   time.Now().UTC().Format(time.RFC3339),
		"action_taken":      "Oil change",
		"reported_by":       "tech",
		"work_order_type":   "Corrective",
	}
	respM := s.postJSON("/api/v1/machines/"+sn+"/maintenance", maintBody, token)
	_, cerr := io.Copy(io.Discard, respM.Body)
	_ = respM.Body.Close()
	s.Require().NoError(cerr)
	s.Equal(http.StatusCreated, respM.StatusCode)

	respMC := s.get("/api/v1/machines/"+sn+"/maintenance/export/csv", token)
	defer respMC.Body.Close()
	s.Equal(http.StatusOK, respMC.StatusCode)
	mb, err := io.ReadAll(respMC.Body)
	s.Require().NoError(err)
	s.Contains(string(mb), wo)
}

func (s *IntegrationSuite) TestTier2_AdminBulkStatus() {
	admin := "adm_" + uuid.NewString()[:8]
	s.createApprovedUser(admin, "pw12345678", admin+"@t.example", users.RoleAdmin)
	tok := s.login(admin, "pw12345678")

	u1 := "b1_" + uuid.NewString()[:8]
	u2 := "b2_" + uuid.NewString()[:8]
	s.createApprovedUser(u1, "pw12345678", u1+"@t.example", users.RoleNonAdmin)
	s.createApprovedUser(u2, "pw12345678", u2+"@t.example", users.RoleNonAdmin)
	id1 := s.userIDByUsername(u1)
	id2 := s.userIDByUsername(u2)

	resp := s.putJSON("/api/v1/admin/users/bulk-status", map[string]any{
		"user_ids": []int64{id1, id2},
		"status":   users.StatusInactive,
	}, tok)
	defer resp.Body.Close()
	rb, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode, string(rb))
}

func (s *IntegrationSuite) TestTier3_AuditLogsWrittenAsync() {
	ctx := context.Background()
	var before int
	s.Require().NoError(s.db.PostgresClient.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs`).Scan(&before))

	admin := "adm_" + uuid.NewString()[:8]
	s.createApprovedUser(admin, "pw12345678", admin+"@t.example", users.RoleAdmin)
	s.login(admin, "pw12345678")

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var after int
		s.Require().NoError(s.db.PostgresClient.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs`).Scan(&after))
		if after > before {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	s.FailNow("expected new audit row after login")
}

func (s *IntegrationSuite) TestTier3_MachineAttachmentUploadAndDownload() {
	admin := "adm_" + uuid.NewString()[:8]
	s.createApprovedUser(admin, "pw12345678", admin+"@t.example", users.RoleAdmin)
	token := s.login(admin, "pw12345678")

	sn := "SN_" + uuid.NewString()[:8]
	resp := s.postJSON("/api/v1/machines", s.machineJSON(sn, "AttCo"), token)
	_, err := io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	pdf := []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
	filename := "doc_" + uuid.NewString()[:8] + ".pdf"
	respUp := s.postMultipart("/api/v1/machines/"+sn+"/attachments", "file", filename, pdf, token)
	defer respUp.Body.Close()
	upb, err := io.ReadAll(respUp.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, respUp.StatusCode, string(upb))

	respGet := s.get("/api/v1/machines/"+sn+"/attachments/"+filename, token)
	defer respGet.Body.Close()
	got, err := io.ReadAll(respGet.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, respGet.StatusCode, string(got))
	s.Equal(pdf, got)
}

func (s *IntegrationSuite) TestTier3_MaintenanceAttachmentRoundTrip() {
	admin := "adm_" + uuid.NewString()[:8]
	s.createApprovedUser(admin, "pw12345678", admin+"@t.example", users.RoleAdmin)
	token := s.login(admin, "pw12345678")

	sn := "SN_" + uuid.NewString()[:8]
	resp := s.postJSON("/api/v1/machines", s.machineJSON(sn, "AttCo2"), token)
	_, err := io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	wo := "WO_" + uuid.NewString()[:8]
	maintBody := map[string]any{
		"work_order_number": wo,
		"work_order_date":   time.Now().UTC().Format(time.RFC3339),
		"action_taken":      "Check",
		"reported_by":       "tech",
		"work_order_type":   "Inspection",
	}
	respM := s.postJSON("/api/v1/machines/"+sn+"/maintenance", maintBody, token)
	_, cerr := io.Copy(io.Discard, respM.Body)
	_ = respM.Body.Close()
	s.Require().NoError(cerr)
	s.Equal(http.StatusCreated, respM.StatusCode)

	pdf := []byte("%PDF-1.4\n2 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
	filename := "m_" + uuid.NewString()[:8] + ".pdf"
	path := "/api/v1/machines/" + sn + "/maintenance/" + wo + "/attachments"
	respUp := s.postMultipart(path, "file", filename, pdf, token)
	defer respUp.Body.Close()
	upb, err := io.ReadAll(respUp.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, respUp.StatusCode, string(upb))

	respGet := s.get(path+"/"+filename, token)
	defer respGet.Body.Close()
	got, err := io.ReadAll(respGet.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, respGet.StatusCode, string(got))
	s.Equal(pdf, got)
}

func (s *IntegrationSuite) userIDByUsername(username string) int64 {
	ctx := context.Background()
	var id int64
	err := s.db.PostgresClient.QueryRow(ctx,
		`SELECT id FROM users WHERE username = $1`, username).Scan(&id)
	s.Require().NoError(err)
	return id
}
