package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRolesAllowsSuperAdminOverride(t *testing.T) {
	app := newTestApp()

	h := app.requireRoles("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), userRoleContextKey, "super_admin"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for super_admin override, got %d", rr.Code)
	}
}

func TestRequireTenantAccessRequiresTenantHeaderForScopedUsers(t *testing.T) {
	app := newTestApp()

	h := app.requireTenantAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), userRoleContextKey, "manager"))
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, "user-1"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when tenant header missing, got %d", rr.Code)
	}
}

func TestRequireTenantAccessBypassesForSuperAdmin(t *testing.T) {
	app := newTestApp()

	h := app.requireTenantAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), userRoleContextKey, "super_admin"))
	req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, "sa-1"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for super_admin bypass, got %d", rr.Code)
	}
}
