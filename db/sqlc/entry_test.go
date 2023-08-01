package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/caleberi/simple-bank/pkg/utils"
	assert "github.com/stretchr/testify/require"
)

var (
	testCreate = true
	testGet    = true
)

func createAndGetNewEntryWithAccount(t *testing.T) func(t *testing.T) (Account, Entry) {
	account := createRandomAccount(t)
	entryArgs := CreateEntryParams{
		AccountID: account.ID,
		Amount:    utils.RandomMoney(),
	}
	return func(t *testing.T) (Account, Entry) {

		entry, err := testQueries.CreateEntry(context.Background(), entryArgs)
		assert.NoError(t, err)

		assert.Equal(t, account.ID, entry.AccountID)
		assert.Equal(t, entryArgs.Amount, entry.Amount)
		assert.WithinDuration(t, account.CreatedAt, entry.CreatedAt, 60*time.Millisecond)
		return account, entry
	}
}

func TestCreateEntry(t *testing.T) {
	if !testCreate {
		return
	}
	createAndGetNewEntryWithAccount(t)(t)
}

func TestGetEntry(t *testing.T) {
	if !testGet {
		return
	}

	account, entry := createAndGetNewEntryWithAccount(t)(t)
	q := `
	SELECT * FROM entries 
	WHERE account_id = $1 AND id = $2
	LIMIT 1;`
	rows, err := testQueries.db.QueryContext(context.Background(), q, account.ID, entry.ID)
	assert.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&entry)
	}

	assert.NoError(t, err)
	assert.Equal(t, entry.AccountID, account.ID)
}

func TestListEntries(t *testing.T) {
	entryIds := []interface{}{}
	k := 5
	for i := 0; i < k; i++ {
		_, entry := createAndGetNewEntryWithAccount(t)(t)
		entryIds = append(entryIds, fmt.Sprintf("%d", entry.ID))
	}

	q := fmt.Sprintf(`SELECT * FROM entries 
	WHERE id IN ( '%v', '%v', '%v', '%v','%v') 
	ORDER BY created_at ASC`, entryIds...)

	rows, err := testQueries.db.QueryContext(context.Background(), q)
	assert.NoError(t, err)
	defer rows.Close()
	entries := make([]Entry, 0)
	// copy all values from database rows
	for rows.Next() {
		var entry Entry
		err = rows.Scan(&entry.ID, &entry.AccountID, &entry.Amount, &entry.CreatedAt)
		assert.NoError(t, err)
		entries = append(entries, entry)
	}
	assert.Equal(t, len(entries), len(entryIds))
	for i := 0; i < len(entries); i++ {
		assert.Equal(t, fmt.Sprintf("%v", entries[i].ID), entryIds[i])
	}
}
