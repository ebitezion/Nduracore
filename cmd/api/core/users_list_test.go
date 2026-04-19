package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ebitezion/Nduracore/internal/data"
)

func TestListUsersIncludesTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	mock.ExpectQuery(`SELECT count\(\*\) OVER\(\)`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"count",
			"id",
			"first_name",
			"last_name",
			"email",
			"phone",
			"role",
			"status",
			"email_verified",
			"tenant_id",
			"created_at",
			"updated_at",
		}).AddRow(1, "u-list-1", "Jane", "Doe", "jane@example.com", "+10000000010", "admin", "active", true, "tenant-1", now, now))

	app := newTestApp()
	app.model = data.NewModels(db)

	token, err := app.security.GenerateToken("admin-user-id", "admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/users?page=1&page_size=20&sort=created_at", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var payload struct {
		Users []struct {
			TenantID string `json:"tenant_id"`
		} `json:"users"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(payload.Users) != 1 {
		t.Fatalf("expected one user, got %d", len(payload.Users))
	}
	if payload.Users[0].TenantID != "tenant-1" {
		t.Fatalf("expected tenant_id tenant-1, got %q", payload.Users[0].TenantID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
