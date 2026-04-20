package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ebitezion/Nduracore/internal/validator"
	"github.com/lib/pq"
)

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type UserModel struct {
	DB *sql.DB
}
type User struct {
	ID            string    `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	TenantID      string    `json:"tenant_id,omitempty"`
	PasswordHash  string    `json:"-"`
	Role          string    `json:"role"`
	Status        string    `json:"status"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UserTenant struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const dbTimeout = 3 * time.Second

func ValidateUsers(v *validator.Validator, user *User) {

	// First Name
	v.Check(user.FirstName != "", "first_name", "must be provided")
	v.Check(len(user.FirstName) <= 50, "first_name", "must not be more than 50 bytes long")

	// Last Name
	v.Check(user.LastName != "", "last_name", "must be provided")
	v.Check(len(user.LastName) <= 50, "last_name", "must not be more than 50 bytes long")

	// Email
	v.Check(user.Email != "", "email", "must be provided")
	v.Check(len(user.Email) <= 255, "email", "must not be more than 255 bytes long")
	v.Check(validator.Matches(user.Email, validator.EmailRX), "email", "must be a valid email address")

	// Phone (optional but must be reasonable length if provided)
	if user.Phone != "" {
		v.Check(len(user.Phone) <= 30, "phone", "must not be more than 30 bytes long")
	}

	// Password Hash
	v.Check(user.PasswordHash != "", "password", "must be provided")

	// Role validation
	v.Check(user.Role != "", "role", "must be provided")
	v.Check(validator.In(user.Role, "user", "admin", "manager", "super_admin"), "role", "must be a valid role")

	// Status validation
	v.Check(user.Status != "", "status", "must be provided")
	v.Check(validator.In(user.Status, "pending", "active", "disabled", "suspended", "rejected"), "status", "must be a valid status")
}

func (u UserModel) Insert(user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `INSERT INTO users(first_name, last_name, email, phone, password_hash, role, status, email_verified)
			 VALUES($1,$2,$3,$4,$5,$6,$7,$8)
			 RETURNING id, created_at
	`
	args := []interface{}{
		user.FirstName,
		user.LastName,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.EmailVerified,
	}

	return u.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.CreatedAt)
}

func (u UserModel) InsertAndReturn(user User) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `INSERT INTO users(first_name, last_name, email, phone, password_hash, role, status, email_verified)
			 VALUES($1,$2,$3,$4,$5,$6,$7,$8)
			 RETURNING id, first_name, last_name, email, phone, role, status, email_verified, created_at, updated_at`
	args := []interface{}{
		user.FirstName,
		user.LastName,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.EmailVerified,
	}

	var created User
	err := u.DB.QueryRowContext(ctx, stmt, args...).Scan(
		&created.ID,
		&created.FirstName,
		&created.LastName,
		&created.Email,
		&created.Phone,
		&created.Role,
		&created.Status,
		&created.EmailVerified,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrDuplicateRecord
		}
		return nil, err
	}

	return &created, nil
}

func (u UserModel) Get(id int64) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if id < 1 {
		return nil, ErrRecordNotFound
	}
	stmt := `SELECT id, first_name, last_name, email, phone, role, status, email_verified, created_at 
			 FROM users
			 WHERE id = $1
	`

	var user User

	err := u.DB.QueryRowContext(ctx, stmt, id).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &user.Role, &user.Status, &user.EmailVerified, &user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	return &user, nil
}

func (u UserModel) Update(user User, id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `UPDATE users
			SET first_name = $1, last_name = $2, email = $3, phone = $4, password_hash = $5, role = $6, status = $7, email_verified = $8 
			WHERE id = $9
			RETURNING id
	     `
	args := []interface{}{
		user.FirstName,
		user.LastName,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.EmailVerified,
		id,
	}
	return u.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID)
}

func (u UserModel) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, ErrRecordNotFound
	}

	stmt := `SELECT id, first_name, last_name, email, phone, password_hash, role, status, email_verified, created_at, updated_at
		FROM users
		WHERE lower(email) = $1`

	var user User
	err := u.DB.QueryRowContext(ctx, stmt, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u UserModel) GetByID(id string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrRecordNotFound
	}

	stmt := `SELECT id, first_name, last_name, email, phone, password_hash, role, status, email_verified, created_at, updated_at
		FROM users
		WHERE id = $1`

	var user User
	err := u.DB.QueryRowContext(ctx, stmt, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u UserModel) List(filters Filters) ([]User, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query, err := userListQueryForSort(filters.Sort, false)
	if err != nil {
		return nil, Metadata{}, err
	}

	args := []interface{}{filters.PageSize, (filters.Page - 1) * filters.PageSize}

	rows, err := u.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	users := []User{}

	for rows.Next() {
		var user User
		if err := rows.Scan(
			&totalRecords,
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Phone,
			&user.Role,
			&user.Status,
			&user.EmailVerified,
			&user.TenantID,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, Metadata{}, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return users, metadata, nil
}

func (u UserModel) ListByStatus(status string, filters Filters) ([]User, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query, err := userListQueryForSort(filters.Sort, true)
	if err != nil {
		return nil, Metadata{}, err
	}

	args := []interface{}{strings.TrimSpace(status), filters.PageSize, (filters.Page - 1) * filters.PageSize}

	rows, err := u.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&totalRecords,
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Phone,
			&user.Role,
			&user.Status,
			&user.EmailVerified,
			&user.TenantID,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, Metadata{}, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	return users, CalculateMetadata(totalRecords, filters.Page, filters.PageSize), nil
}

func (u UserModel) UpdateRoleAndStatus(userID, role, status string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	user, err := updateRoleAndStatusQuery(ctx, u.DB, strings.TrimSpace(userID), strings.TrimSpace(role), strings.TrimSpace(status))
	if err != nil {
		return nil, mapUserWriteError(err)
	}

	return user, nil
}

func (u UserModel) UpsertTenantAccess(userID, tenantID, role, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	err := upsertTenantAccessExec(ctx, u.DB, strings.TrimSpace(userID), strings.TrimSpace(tenantID), strings.TrimSpace(role), strings.TrimSpace(status))
	if err != nil {
		return mapUserWriteError(err)
	}
	return nil
}

func (u UserModel) ApproveWithTenantAccess(userID, tenantID, role string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	tx, err := u.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	updatedUser, err := updateRoleAndStatusQuery(ctx, tx, strings.TrimSpace(userID), strings.TrimSpace(role), "active")
	if err != nil {
		return nil, mapUserWriteError(err)
	}

	if err := upsertTenantAccessExec(ctx, tx, strings.TrimSpace(updatedUser.ID), strings.TrimSpace(tenantID), strings.TrimSpace(updatedUser.Role), "active"); err != nil {
		return nil, mapUserWriteError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (u UserModel) HasTenantAccess(parent context.Context, userID, tenantID string) (bool, error) {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, dbTimeout)
	defer cancel()

	stmt := `SELECT EXISTS(
		SELECT 1
		FROM user_tenants
		WHERE user_id = $1 AND tenant_id = $2 AND status = 'active'
	)`

	var allowed bool
	err := u.DB.QueryRowContext(ctx, stmt, strings.TrimSpace(userID), strings.TrimSpace(tenantID)).Scan(&allowed)
	return allowed, err
}

func userListQueryForSort(sort string, withStatus bool) (string, error) {
	limitPlaceholder := "$1"
	offsetPlaceholder := "$2"
	whereClause := ""
	if withStatus {
		whereClause = "WHERE LOWER(TRIM(users.status)) = LOWER(TRIM($1))"
		limitPlaceholder = "$2"
		offsetPlaceholder = "$3"
	}

	orderClause := ""
	switch sort {
	case "created_at":
		orderClause = "users.created_at ASC, users.id ASC"
	case "-created_at":
		orderClause = "users.created_at DESC, users.id ASC"
	case "email":
		orderClause = "users.email ASC, users.id ASC"
	case "-email":
		orderClause = "users.email DESC, users.id ASC"
	case "first_name":
		orderClause = "users.first_name ASC, users.id ASC"
	case "-first_name":
		orderClause = "users.first_name DESC, users.id ASC"
	default:
		return "", ErrRecordNotFound
	}

	query := fmt.Sprintf(`SELECT count(*) OVER(),
			users.id,
			users.first_name,
			users.last_name,
			users.email,
			users.phone,
			users.role,
			users.status,
			users.email_verified,
			COALESCE(uta.tenant_id, '') AS tenant_id,
			users.created_at,
			users.updated_at
		FROM users
		LEFT JOIN LATERAL (
			SELECT tenant_id
			FROM user_tenants
			WHERE user_id = users.id AND status = 'active'
			ORDER BY updated_at DESC, created_at DESC
			LIMIT 1
		) AS uta ON true
		%s
		ORDER BY %s
		LIMIT %s OFFSET %s`, whereClause, orderClause, limitPlaceholder, offsetPlaceholder)

	return query, nil
}

type userRoleWriter interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type userTenantWriter interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func updateRoleAndStatusQuery(ctx context.Context, writer userRoleWriter, userID, role, status string) (*User, error) {
	stmt := `UPDATE users
		SET role = $1, status = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, first_name, last_name, email, phone, password_hash, role, status, email_verified, created_at, updated_at`

	var user User
	err := writer.QueryRowContext(ctx, stmt, role, status, userID).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func upsertTenantAccessExec(ctx context.Context, writer userTenantWriter, userID, tenantID, role, status string) error {
	stmt := `INSERT INTO user_tenants (user_id, tenant_id, role, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, tenant_id)
		DO UPDATE SET role = EXCLUDED.role, status = EXCLUDED.status, updated_at = NOW()`

	_, err := writer.ExecContext(ctx, stmt, userID, tenantID, role, status)
	return err
}

func mapUserWriteError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrRecordNotFound
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "22P02":
			return ErrInvalidInput
		case "42P01":
			return ErrDependencyUnavailable
		case "23505":
			return ErrDuplicateRecord
		case "23503":
			return ErrRecordNotFound
		}
	}

	return err
}
