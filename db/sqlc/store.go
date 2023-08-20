package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// Store provides all necessary information to execute db queries and transactions
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore returns a new Store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// ExecuteTrxn executes function with a database transaction
func (store *Store) executeTrxn(ctx context.Context, fn func(*Queries) error) error {
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
func (store *Store) PerformTransactionTrxn(ctx context.Context, arg TransferTxnParams) (TransferTrxResult, error) {

	var result TransferTrxResult

	transactFnx := func(q *Queries) error {
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
			result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:      arg.FromAccountID,
				Balance: -arg.Amount,
			})

			if err != nil {
				return err
			}

			result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:      arg.ToAccountID,
				Balance: +arg.Amount,
			})

			if err != nil {
				return err
			}
		} else {
			result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:      arg.ToAccountID,
				Balance: +arg.Amount,
			})

			if err != nil {
				return err
			}

			result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:      arg.FromAccountID,
				Balance: -arg.Amount,
			})

			if err != nil {
				return err
			}
		}

		return nil
	}

	err := store.executeTrxn(ctx, transactFnx)
	return result, err
}
