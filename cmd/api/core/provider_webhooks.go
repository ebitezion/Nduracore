package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/wallet"
)

func (app *application) ingestProviderWebhook(w http.ResponseWriter, r *http.Request) {
	provider := strings.ToLower(strings.TrimSpace(app.pathParam(r, "provider")))
	if provider == "" {
		app.notFoundErrorResponse(w, r)
		return
	}
	walletID := strings.TrimSpace(app.readString(r.URL.Query(), "wallet_id", ""))
	app.processProviderWebhook(w, r, provider, walletID)
}

func (app *application) processProviderWebhook(w http.ResponseWriter, r *http.Request, provider, walletID string) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	headers := map[string]string{
		"x-alchemy-signature": strings.TrimSpace(r.Header.Get("X-Alchemy-Signature")),
		"content-type":        strings.TrimSpace(r.Header.Get("Content-Type")),
	}
	signatureValid := true
	if provider == "alchemy" {
		signatureValid = app.alchemy != nil && app.alchemy.VerifyWebhookSignature(body, headers["x-alchemy-signature"])
	}
	if !signatureValid {
		app.errorResponse(w, r, http.StatusUnauthorized, "invalid webhook signature")
		return
	}

	dedupe := sha256.Sum256([]byte(provider + ":" + tenantID + ":" + string(body)))
	providerEvent := data.ProviderEvent{
		Provider:       provider,
		TenantID:       tenantID,
		DedupeKey:      hex.EncodeToString(dedupe[:]),
		SignatureValid: signatureValid,
		Status:         "accepted",
		Payload:        map[string]interface{}{},
		Headers:        headers,
	}
	_ = json.Unmarshal(body, &providerEvent.Payload)

	if app.model.Events.DB != nil {
		if err := app.model.Events.CreateProviderEvent(r.Context(), &providerEvent); err != nil {
			app.serverErrorResponse(w, r)
			return
		}
	}

	normalizedEvents, err := app.normalizeProviderEvents(r.Context(), provider, tenantID, walletID, body, providerEvent.ID)
	if err != nil {
		if app.model.Events.DB != nil {
			nextRetry := time.Now().UTC().Add(5 * time.Minute)
			_ = app.model.Events.AddEventFailure(r.Context(), providerEvent.ID, tenantID, err.Error(), 1, &nextRetry)
			_ = app.model.Events.EnqueueDLQ(r.Context(), &data.EventDLQ{
				ProviderEventID: providerEvent.ID,
				TenantID:        tenantID,
				Provider:        provider,
				Reason:          "normalization_failed",
				Payload:         providerEvent.Payload,
				RetryCount:      1,
				Status:          "queued",
				LastError:       err.Error(),
			})
		}
		app.badRequestResponse(w, r, err)
		return
	}

	_ = app.writeJSON(w, http.StatusAccepted, envelope{"provider_event_id": providerEvent.ID, "events": normalizedEvents}, nil)
}

func (app *application) normalizeProviderEvents(ctx context.Context, provider, tenantID, walletID string, body []byte, providerEventID string) ([]data.NormalizedEvent, error) {
	if provider == "alchemy" {
		if app.alchemy == nil {
			return nil, errors.New("alchemy provider is not configured")
		}
		if !isValidUUID(walletID) {
			return nil, errors.New("wallet_id query parameter is required for alchemy webhooks")
		}
		parsed, err := app.alchemy.ParseWebhookDeposits(body)
		if err != nil {
			return nil, err
		}

		out := make([]data.NormalizedEvent, 0, len(parsed))
		for _, dep := range parsed {
			recorded, err := app.walletService.RecordDeposit(ctx, wallet.RecordDepositInput{
				TenantID:      tenantID,
				WalletID:      walletID,
				TxHash:        dep.TxHash,
				AmountMinor:   dep.AmountMinor,
				Confirmations: dep.Confirmations,
				Status:        dep.Status,
			})
			if err != nil {
				return nil, err
			}

			ev := data.NormalizedEvent{
				ProviderEventID: providerEventID,
				Provider:        provider,
				TenantID:        tenantID,
				EventType:       "deposit.detected",
				WalletID:        walletID,
				TxHash:          dep.TxHash,
				AmountMinor:     dep.AmountMinor,
				Status:          dep.Status,
				Payload: map[string]interface{}{
					"confirmations": dep.Confirmations,
					"recorded_id":   recorded.ID,
				},
			}
			if app.model.Events.DB != nil {
				if err := app.model.Events.CreateNormalizedEvent(ctx, &ev); err != nil {
					return nil, err
				}
			}
			out = append(out, ev)
		}
		return out, nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	ev := data.NormalizedEvent{
		ProviderEventID: providerEventID,
		Provider:        provider,
		TenantID:        tenantID,
		EventType:       strings.TrimSpace(asString(payload["event_type"])),
		WalletID:        strings.TrimSpace(asString(payload["wallet_id"])),
		WithdrawalID:    strings.TrimSpace(asString(payload["withdrawal_id"])),
		TxHash:          strings.TrimSpace(asString(payload["tx_hash"])),
		AmountMinor:     asInt64(payload["amount_minor"]),
		Status:          strings.TrimSpace(asString(payload["status"])),
		Payload:         payload,
	}
	if ev.EventType == "" {
		ev.EventType = "provider.event"
	}
	if app.model.Events.DB != nil {
		if err := app.model.Events.CreateNormalizedEvent(ctx, &ev); err != nil {
			return nil, err
		}
	}

	return []data.NormalizedEvent{ev}, nil
}

func asString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(t, 10)
	case int:
		return strconv.Itoa(t)
	default:
		if v == nil {
			return ""
		}
		b, _ := json.Marshal(v)
		return strings.Trim(string(b), "\"")
	}
}

func asInt64(v interface{}) int64 {
	s := strings.TrimSpace(asString(v))
	if s == "" {
		return 0
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}
