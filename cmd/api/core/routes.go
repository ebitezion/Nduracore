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
	routes.HandlerFunc(http.MethodPost, "/v1/auth/register", app.registerUser)
	routes.Handler(http.MethodPost, "/v1/users", app.authenticate(app.requireRoles("admin")(http.HandlerFunc(app.createUser))))
	routes.Handler(http.MethodGet, "/v1/users", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.listUsers))))
	routes.Handler(http.MethodGet, "/v1/admin/users/pending", app.authenticate(app.requireRoles("super_admin")(http.HandlerFunc(app.listPendingUsers))))
	routes.Handler(http.MethodPost, "/v1/admin/users/:id/approve", app.authenticate(app.requireRoles("super_admin")(http.HandlerFunc(app.approveUser))))
	routes.Handler(http.MethodPost, "/v1/admin/users/:id/reject", app.authenticate(app.requireRoles("super_admin")(http.HandlerFunc(app.rejectUser))))
	routes.Handler(http.MethodPost, "/v1/treasuries", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.createTreasury)))))
	routes.Handler(http.MethodGet, "/v1/treasuries", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.listTreasuries)))))
	routes.Handler(http.MethodGet, "/v1/treasuries/:id", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getTreasury)))))
	routes.Handler(http.MethodPost, "/v1/treasuries/:id/vaults", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.createVault)))))
	routes.Handler(http.MethodGet, "/v1/treasuries/:id/vaults", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.listVaults)))))
	routes.Handler(http.MethodGet, "/v1/vaults/:id", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getVault)))))
	routes.Handler(http.MethodPost, "/v1/vaults/:id/assets", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.createVaultAsset)))))
	routes.Handler(http.MethodGet, "/v1/vaults/:id/assets", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.listVaultAssets)))))
	routes.Handler(http.MethodPost, "/v1/vaults/:id/aliases", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.createVaultAlias)))))
	routes.Handler(http.MethodGet, "/v1/vaults/:id/aliases", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.listVaultAliases)))))
	routes.Handler(http.MethodGet, "/v1/vault-aliases/resolve", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.resolveVaultAlias)))))
	routes.Handler(http.MethodPut, "/v1/fees/policies/:network/:asset", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.upsertFeePolicy)))))
	routes.Handler(http.MethodGet, "/v1/fees/policies/:network/:asset", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getFeePolicy)))))
	routes.Handler(http.MethodPost, "/v1/jobs/audit", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.enqueueAuditJob))))
	routes.Handler(http.MethodPost, "/v1/wallets", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.createWallet)))))
	routes.Handler(http.MethodGet, "/v1/wallets", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.listWallets)))))
	routes.Handler(http.MethodGet, "/v1/wallets/:id", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getWallet)))))
	routes.Handler(http.MethodGet, "/v1/wallets/:id/balance", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getWalletBalance)))))
	routes.Handler(http.MethodGet, "/v1/wallets/:id/transactions", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getWalletTransactions)))))
	routes.Handler(http.MethodGet, "/v1/wallets/:id/gas-estimate", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getWalletGasEstimate)))))
	routes.Handler(http.MethodPost, "/v1/wallets/:id/withdrawals/simulate", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.simulateWalletWithdrawal)))))
	routes.Handler(http.MethodGet, "/v1/wallets/:id/token-allowances", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getWalletTokenAllowances)))))
	routes.Handler(http.MethodGet, "/v1/wallets/:id/risk-score", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getWalletRiskScore)))))
	routes.Handler(http.MethodGet, "/v1/wallets/:id/nonces", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getWalletNonces)))))
	routes.Handler(http.MethodGet, "/v1/wallets/:id/deposits", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.listWalletDeposits)))))
	routes.HandlerFunc(http.MethodPost, "/v1/wallets/:id/deposits/webhook", app.ingestAlchemyDepositWebhook)
	routes.HandlerFunc(http.MethodPost, "/v1/webhooks/:provider", app.ingestProviderWebhook)
	routes.Handler(http.MethodPost, "/v1/withdrawals", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.createWithdrawal)))))
	routes.Handler(http.MethodGet, "/v1/withdrawals", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.listWithdrawals)))))
	routes.Handler(http.MethodGet, "/v1/withdrawals/:id", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getWithdrawal)))))
	routes.Handler(http.MethodGet, "/v1/withdrawals/:id/trace", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.getWithdrawalTrace)))))
	routes.Handler(http.MethodPost, "/v1/withdrawals/:id/approve", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.approveWithdrawal)))))
	routes.Handler(http.MethodPost, "/v1/withdrawals/:id/reject", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.rejectWithdrawal)))))
	routes.Handler(http.MethodGet, "/v1/networks/:network/status", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.getNetworkStatus))))
	routes.Handler(http.MethodGet, "/v1/assets/:network/:asset/metadata", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.getAssetMetadata))))
	routes.Handler(http.MethodPost, "/v1/addresses/validate", app.authenticate(app.requireRoles("admin", "manager")(http.HandlerFunc(app.validateAddress))))
	routes.Handler(http.MethodPost, "/v1/ops/reconcile/withdrawals/run", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.runWithdrawalReconciliation)))))
	routes.Handler(http.MethodPost, "/v1/ops/events/dlq/:id/replay", app.authenticate(app.requireRoles("admin", "manager")(app.requireTenantAccess(http.HandlerFunc(app.replayDLQEvent)))))

	return app.chain(routes)
}
