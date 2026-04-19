package core

import (
	"net/http"
	"strings"

	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/txstate"
)

func (app *application) runWithdrawalReconciliation(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	run := data.ReconciliationRun{
		TenantID: tenantID,
		Provider: "alchemy",
		Status:   "running",
		Summary:  map[string]interface{}{},
	}
	if err := app.model.Reconcile.CreateRun(r.Context(), &run); err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	itemsChecked := 0
	mismatches := 0
	filters := data.Filters{Page: 1, PageSize: 200, Sort: "-created_at", SortSafelist: []string{"created_at", "-created_at"}}
	withdrawals, _, err := app.model.Wallets.ListWithdrawals(r.Context(), tenantID, "", filters)
	if err != nil {
		_ = app.model.Reconcile.CompleteRun(r.Context(), run.ID, "failed", map[string]interface{}{"error": err.Error()})
		app.serverErrorResponse(w, r)
		return
	}

	for _, wd := range withdrawals {
		if strings.TrimSpace(wd.ProviderTxHash) == "" || strings.ToLower(strings.TrimSpace(wd.Network)) == "" {
			continue
		}
		itemsChecked++

		providerState := "pending"
		expected := txstate.Normalize(wd.Status)
		if chainFamily(wd.Network) == "evm" && app.alchemy != nil {
			receipt, err := app.alchemy.GetTxReceiptState(r.Context(), wd.Network, wd.ProviderTxHash)
			if err == nil {
				if receipt.Found && receipt.Confirmed && receipt.Success {
					providerState = txstate.StateConfirmed
				} else if receipt.Found && !receipt.Success {
					providerState = txstate.StateFailed
				} else {
					providerState = txstate.StateBroadcasted
				}
			}
		}

		if expected != providerState && expected != txstate.StatePolicyPending {
			mismatches++
			_ = app.model.Reconcile.AddItem(r.Context(), &data.ReconciliationItem{
				RunID:          run.ID,
				TenantID:       tenantID,
				ObjectType:     "withdrawal",
				ObjectID:       wd.ID,
				ExpectedState:  expected,
				ProviderState:  providerState,
				MismatchReason: "status_drift",
			})
		}
	}

	summary := map[string]interface{}{
		"checked":    itemsChecked,
		"mismatches": mismatches,
	}
	_ = app.model.Reconcile.CompleteRun(r.Context(), run.ID, "completed", summary)

	_ = app.writeJSON(w, http.StatusAccepted, envelope{"run_id": run.ID, "summary": summary}, nil)
}

func (app *application) replayDLQEvent(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	dlqID := app.pathParam(r, "id")
	if !isValidUUID(dlqID) {
		app.errorResponse(w, r, http.StatusBadRequest, "dlq id must be a valid UUID")
		return
	}

	item, err := app.model.Events.GetDLQByID(r.Context(), tenantID, dlqID)
	if err != nil {
		if err == data.ErrRecordNotFound {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	event := data.NormalizedEvent{
		ProviderEventID: item.ProviderEventID,
		Provider:        item.Provider,
		TenantID:        item.TenantID,
		EventType:       "dlq.replayed",
		Status:          "replayed",
		Payload:         item.Payload,
	}
	if err := app.model.Events.CreateNormalizedEvent(r.Context(), &event); err != nil {
		app.serverErrorResponse(w, r)
		return
	}
	if err := app.model.Events.MarkDLQReplayed(r.Context(), tenantID, dlqID); err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusAccepted, envelope{"dlq_id": dlqID, "status": "replayed"}, nil)
}
