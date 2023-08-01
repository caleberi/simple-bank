package db

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransferTrxn(t *testing.T) {
	store := NewStore(db)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	log.Printf(">>before trxn :%v %v\n", account1.Balance, account2.Balance)

	// run n concurrent transfer transfer
	numberOfTransferToMake := 2
	amountToTransfer := int64(10)
	errs := make(chan error)
	results := make(chan TransferTrxResult)
	done := make(chan bool)

	for i := 0; i < numberOfTransferToMake; i++ {
		i := i
		fromAccountID := account2.ID
		toAccountID := account1.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}
		trxnHash := fmt.Sprintf("tx %d", i+1)
		go func(i int) {
			ctx := context.WithValue(context.Background(), txKey, trxnHash)
			result, err := store.PerformTransactionTrxn(ctx, TransferTxnParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amountToTransfer,
			})

			results <- result
			errs <- err

			if i == numberOfTransferToMake-1 {
				done <- true
			}
		}(i)
	}

	existed := make(map[int]bool)
	terminate := false
	for {
		select {
		case err = <-errs:
			assert.NoError(t, err)
		case result := <-results:
			assert.NotEmpty(t, result)
			transfer := result.Transfer
			assert.Equal(t, account1.ID, transfer.FromAccountID)
			assert.Equal(t, account2.ID, transfer.ToAccountID)
			assert.Equal(t, amountToTransfer, transfer.Amount)
			assert.NotZero(t, transfer.ID)
			assert.NotZero(t, transfer.CreatedAt)

			_, err := store.GetTransfer(context.Background(), transfer.ID)
			assert.NoError(t, err)

			fromEntry := result.FromEntry
			assert.NotEmpty(t, fromEntry)
			assert.Equal(t, account1.ID, fromEntry.AccountID)
			assert.Equal(t, -amountToTransfer, fromEntry.Amount)
			assert.NotZero(t, fromEntry.ID)
			assert.NotZero(t, fromEntry.CreatedAt)

			_, err = store.GetEntry(context.Background(), fromEntry.ID)
			assert.NoError(t, err)

			toEntry := result.ToEntry
			assert.NotEmpty(t, toEntry)
			assert.Equal(t, account2.ID, toEntry.AccountID)
			assert.Equal(t, amountToTransfer, toEntry.Amount)
			assert.NotZero(t, toEntry.ID)
			assert.NotZero(t, toEntry.CreatedAt)

			_, err = store.GetEntry(context.Background(), toEntry.ID)
			assert.NoError(t, err)

			// TODO: check the account
			fromAccount := result.FromAccount
			assert.NotEmpty(t, fromAccount)
			assert.Equal(t, account1.ID, fromAccount.ID)

			toAccount := result.ToAccount
			assert.NotEmpty(t, toAccount)
			assert.Equal(t, account2.ID, toAccount.ID)

			// check the account balance
			log.Printf(">>trxn :%v %v\n", account1.Balance, account2.Balance)
			diff1 := account1.Balance - fromAccount.Balance
			diff2 := toAccount.Balance - account2.Balance

			assert.Equal(t, diff1, diff2)
			assert.True(t, diff1 > 0)
			assert.True(t, diff1%amountToTransfer == 0)

			k := int(diff1 / amountToTransfer)
			assert.True(t, k >= 1 && k <= numberOfTransferToMake)
			assert.NotContains(t, existed, k)
			existed[k] = true
		case <-done:
			terminate = true
		}
		if terminate {
			break
		}
	}

	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	assert.NoError(t, err)
	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	assert.NoError(t, err)
	log.Printf(">>trxn :%v %v\n", account1.Balance, account2.Balance)
	assert.Equal(t, account1.Balance-int64(numberOfTransferToMake)*amountToTransfer, updatedAccount1.Balance)
	assert.Equal(t, account2.Balance+int64(numberOfTransferToMake)*amountToTransfer, updatedAccount2.Balance)
}
