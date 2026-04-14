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

func TestCreateUserSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	createdAt := time.Now().UTC()
	updatedAt := createdAt

	mock.ExpectQuery(`INSERT INTO users\(first_name, last_name, email, phone, password_hash, role, status, email_verified\)`).
		WithArgs(
			"Jane",
			"Doe",
			"jane@example.com",
			"+10000000010",
			sqlmock.AnyArg(),
			"manager",
			"active",
			true,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "phone", "role", "status", "email_verified", "created_at", "updated_at"}).
			AddRow("u-create-1", "Jane", "Doe", "jane@example.com", "+10000000010", "manager", "active", true, createdAt, updatedAt))

	app := newTestApp()
	app.model = data.NewModels(db)

	token, err := app.security.GenerateToken("admin-user-id", "admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	payload := map[string]interface{}{
		"first_name":     "Jane",
		"last_name":      "Doe",
		"email":          "jane@example.com",
		"phone":          "+10000000010",
		"password":       "StrongPass#2026",
		"role":           "manager",
		"status":         "active",
		"email_verified": true,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCreateUserConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`INSERT INTO users\(first_name, last_name, email, phone, password_hash, role, status, email_verified\)`).
		WithArgs(
			"Jane",
			"Doe",
			"jane@example.com",
			"+10000000010",
			sqlmock.AnyArg(),
			"user",
			"active",
			false,
		).
		WillReturnError(&pq.Error{Code: "23505"})

	app := newTestApp()
	app.model = data.NewModels(db)

	token, err := app.security.GenerateToken("admin-user-id", "admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	payload := map[string]interface{}{
		"first_name": "Jane",
		"last_name":  "Doe",
		"email":      "jane@example.com",
		"phone":      "+10000000010",
		"password":   "StrongPass#2026",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateUserRequiresAdminRole(t *testing.T) {
	app := newTestApp()
	token, err := app.security.GenerateToken("manager-user-id", "manager", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	payload := map[string]interface{}{
		"first_name": "Jane",
		"last_name":  "Doe",
		"email":      "jane@example.com",
		"password":   "StrongPass#2026",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rr.Code, rr.Body.String())
	}
}
