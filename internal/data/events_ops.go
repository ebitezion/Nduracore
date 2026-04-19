package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type EventModel struct {
	DB *sql.DB
}

type ProviderEvent struct {
	ID             string                 `json:"id"`
	Provider       string                 `json:"provider"`
	TenantID       string                 `json:"tenant_id"`
	DedupeKey      string                 `json:"dedupe_key"`
	SignatureValid bool                   `json:"signature_valid"`
	Payload        map[string]interface{} `json:"payload"`
	Headers        map[string]string      `json:"headers"`
	Status         string                 `json:"status"`
	ReceivedAt     time.Time              `json:"received_at"`
}

type NormalizedEvent struct {
	ID              string                 `json:"id"`
	ProviderEventID string                 `json:"provider_event_id"`
	Provider        string                 `json:"provider"`
	TenantID        string                 `json:"tenant_id"`
	EventType       string                 `json:"event_type"`
	WalletID        string                 `json:"wallet_id,omitempty"`
	WithdrawalID    string                 `json:"withdrawal_id,omitempty"`
	TxHash          string                 `json:"tx_hash"`
	AmountMinor     int64                  `json:"amount_minor"`
	Status          string                 `json:"status"`
	Payload         map[string]interface{} `json:"payload"`
	CreatedAt       time.Time              `json:"created_at"`
}

type EventDLQ struct {
	ID              string                 `json:"id"`
	ProviderEventID string                 `json:"provider_event_id"`
	TenantID        string                 `json:"tenant_id"`
	Provider        string                 `json:"provider"`
	Reason          string                 `json:"reason"`
	Payload         map[string]interface{} `json:"payload"`
	RetryCount      int                    `json:"retry_count"`
	Status          string                 `json:"status"`
	LastError       string                 `json:"last_error"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type ReconcileModel struct {
	DB *sql.DB
}

type ReconciliationRun struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	Provider  string                 `json:"provider"`
	Status    string                 `json:"status"`
	Summary   map[string]interface{} `json:"summary"`
	StartedAt time.Time              `json:"started_at"`
	EndedAt   *time.Time             `json:"ended_at,omitempty"`
}

type ReconciliationItem struct {
	ID             string    `json:"id"`
	RunID          string    `json:"run_id"`
	TenantID       string    `json:"tenant_id"`
	ObjectType     string    `json:"object_type"`
	ObjectID       string    `json:"object_id"`
	ExpectedState  string    `json:"expected_state"`
	ProviderState  string    `json:"provider_state"`
	MismatchReason string    `json:"mismatch_reason"`
	CreatedAt      time.Time `json:"created_at"`
}

func (m EventModel) CreateProviderEvent(ctx context.Context, event *ProviderEvent) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	payloadRaw, _ := json.Marshal(event.Payload)
	headersRaw, _ := json.Marshal(event.Headers)

	stmt := `INSERT INTO provider_events (provider, tenant_id, dedupe_key, signature_valid, payload, headers, status)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7)
		ON CONFLICT (provider, dedupe_key)
		DO UPDATE SET received_at = NOW()
		RETURNING id, received_at`

	err := m.DB.QueryRowContext(ctx, stmt,
		strings.ToLower(strings.TrimSpace(event.Provider)),
		strings.TrimSpace(event.TenantID),
		strings.TrimSpace(event.DedupeKey),
		event.SignatureValid,
		string(payloadRaw),
		string(headersRaw),
		strings.TrimSpace(event.Status),
	).Scan(&event.ID, &event.ReceivedAt)
	return err
}

func (m EventModel) CreateNormalizedEvent(ctx context.Context, ev *NormalizedEvent) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	payloadRaw, _ := json.Marshal(ev.Payload)
	stmt := `INSERT INTO normalized_events (provider_event_id, provider, tenant_id, event_type, wallet_id, withdrawal_id, tx_hash, amount_minor, status, payload)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::uuid, NULLIF($6, '')::uuid, $7, $8, $9, $10::jsonb)
		RETURNING id, created_at`

	return m.DB.QueryRowContext(ctx, stmt,
		strings.TrimSpace(ev.ProviderEventID),
		strings.ToLower(strings.TrimSpace(ev.Provider)),
		strings.TrimSpace(ev.TenantID),
		strings.TrimSpace(ev.EventType),
		strings.TrimSpace(ev.WalletID),
		strings.TrimSpace(ev.WithdrawalID),
		strings.TrimSpace(ev.TxHash),
		ev.AmountMinor,
		strings.TrimSpace(ev.Status),
		string(payloadRaw),
	).Scan(&ev.ID, &ev.CreatedAt)
}

func (m EventModel) AddEventFailure(ctx context.Context, providerEventID, tenantID, message string, retryCount int, nextRetryAt *time.Time) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, `INSERT INTO event_failures (provider_event_id, tenant_id, error_message, retry_count, next_retry_at)
		VALUES ($1, $2, $3, $4, $5)`,
		strings.TrimSpace(providerEventID),
		strings.TrimSpace(tenantID),
		strings.TrimSpace(message),
		retryCount,
		nextRetryAt,
	)
	return err
}

func (m EventModel) EnqueueDLQ(ctx context.Context, item *EventDLQ) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	payloadRaw, _ := json.Marshal(item.Payload)
	stmt := `INSERT INTO event_dlq (provider_event_id, tenant_id, provider, reason, payload, retry_count, status, last_error)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return m.DB.QueryRowContext(ctx, stmt,
		strings.TrimSpace(item.ProviderEventID),
		strings.TrimSpace(item.TenantID),
		strings.ToLower(strings.TrimSpace(item.Provider)),
		strings.TrimSpace(item.Reason),
		string(payloadRaw),
		item.RetryCount,
		strings.TrimSpace(item.Status),
		strings.TrimSpace(item.LastError),
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (m EventModel) GetDLQByID(ctx context.Context, tenantID, id string) (*EventDLQ, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	var item EventDLQ
	var payloadRaw []byte
	stmt := `SELECT id, provider_event_id::text, tenant_id, provider, reason, payload, retry_count, status, last_error, created_at, updated_at
		FROM event_dlq
		WHERE tenant_id = $1 AND id = $2::uuid`
	err := m.DB.QueryRowContext(ctx, stmt, strings.TrimSpace(tenantID), strings.TrimSpace(id)).Scan(
		&item.ID, &item.ProviderEventID, &item.TenantID, &item.Provider, &item.Reason, &payloadRaw, &item.RetryCount, &item.Status, &item.LastError, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(payloadRaw, &item.Payload)
	if item.Payload == nil {
		item.Payload = map[string]interface{}{}
	}
	return &item, nil
}

func (m EventModel) MarkDLQReplayed(ctx context.Context, tenantID, id string) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, `UPDATE event_dlq SET status = 'replayed', updated_at = NOW() WHERE tenant_id = $1 AND id = $2::uuid`,
		strings.TrimSpace(tenantID), strings.TrimSpace(id))
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (m ReconcileModel) CreateRun(ctx context.Context, run *ReconciliationRun) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	summaryRaw, _ := json.Marshal(run.Summary)
	stmt := `INSERT INTO reconciliation_runs (tenant_id, provider, status, summary)
		VALUES ($1, $2, $3, $4::jsonb)
		RETURNING id, started_at`

	return m.DB.QueryRowContext(ctx, stmt,
		strings.TrimSpace(run.TenantID),
		strings.ToLower(strings.TrimSpace(run.Provider)),
		strings.TrimSpace(run.Status),
		string(summaryRaw),
	).Scan(&run.ID, &run.StartedAt)
}

func (m ReconcileModel) AddItem(ctx context.Context, item *ReconciliationItem) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO reconciliation_items (run_id, tenant_id, object_type, object_id, expected_state, provider_state, mismatch_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	return m.DB.QueryRowContext(ctx, stmt,
		strings.TrimSpace(item.RunID),
		strings.TrimSpace(item.TenantID),
		strings.TrimSpace(item.ObjectType),
		strings.TrimSpace(item.ObjectID),
		strings.TrimSpace(item.ExpectedState),
		strings.TrimSpace(item.ProviderState),
		strings.TrimSpace(item.MismatchReason),
	).Scan(&item.ID, &item.CreatedAt)
}

func (m ReconcileModel) CompleteRun(ctx context.Context, runID, status string, summary map[string]interface{}) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	summaryRaw, _ := json.Marshal(summary)
	_, err := m.DB.ExecContext(ctx, `UPDATE reconciliation_runs
		SET status = $2, summary = $3::jsonb, ended_at = NOW()
		WHERE id = $1::uuid`, strings.TrimSpace(runID), strings.TrimSpace(status), string(summaryRaw))
	return err
}
