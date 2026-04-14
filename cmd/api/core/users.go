package core

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ebitezion/Nduracore/cmd/api/core/security"
	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) issueToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.logAuditEvent(r, "auth.token.issue", "failed", map[string]interface{}{"reason": "invalid_payload"})
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.Email != "", "email", "must be provided")
	v.Check(validator.Matches(input.Email, validator.EmailRX), "email", "must be a valid email address")
	v.Check(input.Password != "", "password", "must be provided")
	if !v.Valid() {
		app.logAuditEvent(r, "auth.token.issue", "failed", map[string]interface{}{"reason": "validation_error", "email": strings.ToLower(input.Email)})
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.model.Users.GetByEmail(strings.ToLower(input.Email))
	if err != nil || !security.VerifyPasswordHash(input.Password, user.PasswordHash) {
		app.logAuditEvent(r, "auth.token.issue", "failed", map[string]interface{}{"reason": "invalid_credentials", "email": strings.ToLower(input.Email)})
		app.unauthorizedResponse(w, r)
		return
	}

	if user.Status != "active" {
		app.logAuditEvent(r, "auth.token.issue", "failed", map[string]interface{}{"reason": "user_inactive", "user_id": user.ID, "role": user.Role})
		app.forbiddenResponse(w, r)
		return
	}

	token, err := app.security.GenerateToken(user.ID, user.Role, app.config.security.tokenTTL)
	if err != nil {
		app.logAuditEvent(r, "auth.token.issue", "failed", map[string]interface{}{"reason": "token_generation_error", "user_id": user.ID, "role": user.Role})
		app.serverErrorResponse(w, r)
		return
	}

	app.logAuditEvent(r, "auth.token.issue", "success", map[string]interface{}{"user_id": user.ID, "role": user.Role})
	_ = app.writeJSON(w, http.StatusCreated, envelope{"auth": envelope{
		"token":      token,
		"expires_in": int(app.config.security.tokenTTL.Seconds()),
		"role":       user.Role,
	}}, nil)
}

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		Email         string `json:"email"`
		Phone         string `json:"phone"`
		Password      string `json:"password"`
		Role          string `json:"role"`
		Status        string `json:"status"`
		EmailVerified bool   `json:"email_verified"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.logAuditEvent(r, "user.create", "failed", map[string]interface{}{"reason": "invalid_payload"})
		app.badRequestResponse(w, r, err)
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if strings.TrimSpace(input.Role) == "" {
		input.Role = "user"
	}
	if strings.TrimSpace(input.Status) == "" {
		input.Status = "active"
	}

	v := validator.New()
	v.Check(input.FirstName != "", "first_name", "must be provided")
	v.Check(len(input.FirstName) <= 50, "first_name", "must not be more than 50 bytes long")
	v.Check(input.LastName != "", "last_name", "must be provided")
	v.Check(len(input.LastName) <= 50, "last_name", "must not be more than 50 bytes long")
	v.Check(input.Email != "", "email", "must be provided")
	v.Check(len(input.Email) <= 255, "email", "must not be more than 255 bytes long")
	v.Check(validator.Matches(input.Email, validator.EmailRX), "email", "must be a valid email address")
	if strings.TrimSpace(input.Phone) != "" {
		v.Check(len(strings.TrimSpace(input.Phone)) <= 30, "phone", "must not be more than 30 bytes long")
	}
	v.Check(input.Password != "", "password", "must be provided")
	v.Check(len(input.Password) >= 8, "password", "must be at least 8 characters")
	v.Check(validator.In(input.Role, "user", "admin", "manager"), "role", "must be a valid role")
	v.Check(validator.In(input.Status, "active", "disabled", "suspended"), "status", "must be a valid status")
	if !v.Valid() {
		app.logAuditEvent(r, "user.create", "failed", map[string]interface{}{"reason": "validation_error", "email": input.Email})
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		app.logAuditEvent(r, "user.create", "failed", map[string]interface{}{"reason": "hash_error", "email": input.Email})
		app.serverErrorResponse(w, r)
		return
	}

	user := data.User{
		FirstName:     strings.TrimSpace(input.FirstName),
		LastName:      strings.TrimSpace(input.LastName),
		Email:         input.Email,
		Phone:         strings.TrimSpace(input.Phone),
		PasswordHash:  string(passwordHash),
		Role:          strings.TrimSpace(input.Role),
		Status:        strings.TrimSpace(input.Status),
		EmailVerified: input.EmailVerified,
	}

	created, err := app.model.Users.InsertAndReturn(user)
	if err != nil {
		if errors.Is(err, data.ErrDuplicateRecord) {
			app.errorResponse(w, r, http.StatusConflict, "user with same email or phone already exists")
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	app.logAuditEvent(r, "user.create", "success", map[string]interface{}{
		"user_id": created.ID,
		"email":   created.Email,
		"role":    created.Role,
	})

	_ = app.writeJSON(w, http.StatusCreated, envelope{"user": created}, nil)
}

func (app *application) listUsers(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	v := validator.New()
	filters := data.Filters{
		Page:         app.readInt(qs, "page", 1, v),
		PageSize:     app.readInt(qs, "page_size", 20, v),
		Sort:         app.readString(qs, "sort", "created_at"),
		SortSafelist: []string{"created_at", "-created_at", "email", "-email", "first_name", "-first_name"},
	}

	data.ValidateFilters(v, filters)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	users, metadata, err := app.model.Users.List(filters)
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{
		"users":    users,
		"metadata": metadata,
	}, nil)
}

func (app *application) enqueueAuditJob(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Action string `json:"action"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.logAuditEvent(r, "jobs.audit.enqueue", "failed", map[string]interface{}{"reason": "invalid_payload"})
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Action == "" {
		input.Action = "unknown"
	}

	if err := app.queue.Publish(r.Context(), Job{Name: "audit.log", Payload: map[string]string{"action": input.Action}}); err != nil {
		app.logAuditEvent(r, "jobs.audit.enqueue", "failed", map[string]interface{}{"reason": "queue_publish_error", "action": input.Action})
		app.serverErrorResponse(w, r)
		return
	}
	app.logAuditEvent(r, "jobs.audit.enqueue", "accepted", map[string]interface{}{"action": input.Action})
	_ = app.writeJSON(w, http.StatusAccepted, envelope{"message": "job enqueued"}, nil)
}
