package core

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	routes := httprouter.New()

	routes.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFoundErrorResponse(w, r)
	})
	routes.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.methodNotAllowedErrorResponse(w, r)
	})

	routes.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheck)
	routes.HandlerFunc(http.MethodGet, "/liveness", app.liveness)
	routes.HandlerFunc(http.MethodGet, "/readiness", app.readiness)
	routes.HandlerFunc(http.MethodGet, "/metrics", app.metricsHandler)

	routes.HandlerFunc(http.MethodPost, "/v1/auth/token", app.issueToken)
	routes.Handler(http.MethodPost, "/v1/users", app.authenticate(app.requireRoles("admin")(http.HandlerFunc(app.createUser))))
	routes.Handler(http.MethodGet, "/v1/users", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.listUsers))))
	routes.Handler(http.MethodPost, "/v1/jobs/audit", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.enqueueAuditJob))))
	routes.Handler(http.MethodPost, "/v1/wallets", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.createWallet))))
	routes.Handler(http.MethodGet, "/v1/wallets", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.listWallets))))
	routes.Handler(http.MethodGet, "/v1/wallets/:id", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.getWallet))))
	routes.Handler(http.MethodGet, "/v1/wallets/:id/deposits", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.listWalletDeposits))))
	routes.HandlerFunc(http.MethodPost, "/v1/wallets/:id/deposits/webhook", app.ingestAlchemyDepositWebhook)
	routes.Handler(http.MethodPost, "/v1/withdrawals", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.createWithdrawal))))
	routes.Handler(http.MethodGet, "/v1/withdrawals/:id", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.getWithdrawal))))
	routes.Handler(http.MethodPost, "/v1/withdrawals/:id/approve", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.approveWithdrawal))))
	routes.Handler(http.MethodPost, "/v1/withdrawals/:id/reject", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.rejectWithdrawal))))

	return app.chain(routes)
}
