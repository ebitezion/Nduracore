package data

import (
	"database/sql"
	"errors"
)

// Define a custom ErrRecordNotFound error.
var (
	ErrRecordNotFound        = errors.New("record not found")
	ErrDuplicateRecord       = errors.New("duplicate record")
	ErrInvalidInput          = errors.New("invalid input")
	ErrDependencyUnavailable = errors.New("dependency unavailable")
)

// Create a Models struct which wraps the UserModel and others
type Models struct {
	Users     UserModel
	Wallets   WalletModel
	Treasury  TreasuryModel
	Assets    AssetRegistryModel
	Events    EventModel
	Reconcile ReconcileModel
	Tx        TxManager
}

// New() method returns a Models struct containing.
func NewModels(db *sql.DB) Models {
	return Models{
		Users:     UserModel{DB: db},
		Wallets:   WalletModel{DB: db},
		Treasury:  TreasuryModel{DB: db},
		Assets:    AssetRegistryModel{DB: db},
		Events:    EventModel{DB: db},
		Reconcile: ReconcileModel{DB: db},
		Tx:        TxManager{DB: db},
	}
}
