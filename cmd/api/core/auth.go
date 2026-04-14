package core

import (
	"context"
	"net/http"
	"strings"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			app.unauthorizedResponse(w, r)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			app.unauthorizedResponse(w, r)
			return
		}

		claims, err := app.security.ParseToken(parts[1])
		if err != nil {
			app.unauthorizedResponse(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userRoleContextKey, claims.Role)
		ctx = context.WithValue(ctx, userIDContextKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) requireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(userRoleContextKey).(string)
			if _, ok := allowed[role]; !ok {
				app.forbiddenResponse(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
