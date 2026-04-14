package main

import (
	"net/http"
	"time"
)

func (app *application) logAuditEvent(r *http.Request, action, outcome string, details map[string]interface{}) {
	payload := map[string]interface{}{
		"event":      "audit",
		"action":     action,
		"outcome":    outcome,
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": app.requestIDFromContext(r.Context()),
		"ip":         app.readClientIP(r),
		"at":         time.Now().UTC().Format(time.RFC3339),
	}

	for key, value := range details {
		payload[key] = value
	}

	app.logJSON(payload)
}
