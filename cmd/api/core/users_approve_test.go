package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/lib/pq"
)

func TestApproveUserSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	createdAt := time.Now().UTC()
	updatedAt := createdAt

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE users`).
		WithArgs("admin", "active", "u-approve-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "phone", "password_hash", "role", "status", "email_verified", "created_at", "updated_at"}).
			AddRow("u-approve-1", "Pending", "User", "pending@example.com", "+10000000011", "hash", "admin", "active", false, createdAt, updatedAt))
	mock.ExpectExec(`INSERT INTO user_tenants`).
		WithArgs("u-approve-1", "tenant-1", "admin", "active").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	app := newTestApp()
	app.model = data.NewModels(db)

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	payload, _ := json.Marshal(map[string]string{
		"tenant_id": "tenant-1",
		"role":      "admin",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users/u-approve-1/approve", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestApproveUserReturnsServiceUnavailableWhenTenantAccessDependencyMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	createdAt := time.Now().UTC()
	updatedAt := createdAt

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE users`).
		WithArgs("admin", "active", "u-approve-2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "phone", "password_hash", "role", "status", "email_verified", "created_at", "updated_at"}).
			AddRow("u-approve-2", "Pending", "User", "pending2@example.com", "+10000000012", "hash", "admin", "active", false, createdAt, updatedAt))
	mock.ExpectExec(`INSERT INTO user_tenants`).
		WithArgs("u-approve-2", "tenant-1", "admin", "active").
		WillReturnError(&pq.Error{Code: "42P01"})
	mock.ExpectRollback()

	app := newTestApp()
	app.model = data.NewModels(db)

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	payload, _ := json.Marshal(map[string]string{
		"tenant_id": "tenant-1",
		"role":      "admin",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users/u-approve-2/approve", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d body=%s", rr.Code, rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestApproveUserReturnsBadRequestForInvalidUserIdentifier(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE users`).
		WithArgs("admin", "active", "not-a-uuid").
		WillReturnError(&pq.Error{Code: "22P02"})
	mock.ExpectRollback()

	app := newTestApp()
	app.model = data.NewModels(db)

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	payload, _ := json.Marshal(map[string]string{
		"tenant_id": "tenant-1",
		"role":      "admin",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users/not-a-uuid/approve", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
