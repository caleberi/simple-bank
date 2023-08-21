package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Store interface {
	Querier
	PerformTransactionTrxn(ctx context.Context, arg TransferTxnParams) (TransferTrxResult, error)
}

// Store provides all necessary information to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

// NewStore returns a new Store
func NewStore(db *sql.DB) *SQLStore {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// ExecuteTrxn executes function with a database transaction
func (store *SQLStore) executeTrxn(ctx context.Context, fn func(*Queries) error) error {
	trxn, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(trxn)
	err = fn(q)
	if err != nil {
		if rberr := trxn.Rollback(); err != nil {
			return fmt.Errorf("[ERROR] transaction rollback error: %v , rb err: %v ", err, rberr)
		}
		return err
	}

	return trxn.Commit()
}

// TransferTxnParams  contains the input parameters of the transfer transaction.
type TransferTxnParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxnResult is the result of the  transfer transaction.
type TransferTrxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// PerformTransactionTrxn performs a money from one account to the other .
// It creates a transfer record , add account entries and update accounts balance within a single database transaction
func (store *SQLStore) PerformTransactionTrxn(ctx context.Context, arg TransferTxnParams) (TransferTrxResult, error) {
	var result TransferTrxResult

	err := store.executeTrxn(ctx, func(q *Queries) error {
		var err error

		transfer := CreateTransferParams{}
		bt, _ := json.Marshal(arg)
		_ = json.Unmarshal(bt, &transfer)

		result.Transfer, err = q.CreateTransfer(ctx, transfer)
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		return err
	})

	return result, err
}

func addMoney(ctx context.Context, q *Queries, accountID1 int64, amount1 int64, accountID2 int64, amount2 int64) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return
}
