package middleware

// TxnKey used for storing and pulling a database transaction from the fiber context
var TxnKey = "txn"

// Transaction is an interface for wrapping one or more database actions together
type Transaction interface {
	Commit() error
	Rollback() error
}

// AddTxnToCtx attaches a datastore transaction to a fiber request context
