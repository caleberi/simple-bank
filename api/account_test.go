package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/caleberi/simple-bank/db/mock"
	db "github.com/caleberi/simple-bank/db/sqlc"
	"github.com/caleberi/simple-bank/pkg/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type result struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
}

func generateRandomAccount() db.Account {
	return db.Account{
		ID:           utils.RandomInt(1, 100),
		Owner:        utils.RandomOwner(),
		Balance:      utils.RandomMoney(),
		CurrencyCode: utils.RandomCurrencyCode(),
	}
}

func Test_CreateAccount(t *testing.T) {

	// create a new request body
	accountId := utils.RandomInt(1, 100)
	createAccountRequest := createAccountRequest{
		Owner:        utils.RandomOwner(),
		CurrencyCode: utils.RandomCurrencyCode(),
	}

	//  expect account creation results
	account := db.Account{
		ID:           accountId,
		Balance:      0,
		Owner:        createAccountRequest.Owner,
		CurrencyCode: createAccountRequest.CurrencyCode,
		CreatedAt:    time.Now(),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().CreateAccount(gomock.Any(), db.CreateAccountParams{
		Owner:        createAccountRequest.Owner,
		CurrencyCode: createAccountRequest.CurrencyCode,
		Balance:      0,
	}).Times(1).Return(account, nil)

	url := "/accounts"

	//  new server with store
	server := NewServer(store)

	buf := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(createAccountRequest)
	require.NoError(t, err)

	// build  response struct
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodPost, url, buf)
	request.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	data, err := io.ReadAll(recorder.Body)

	require.NoError(t, err)

	var aux result

	err = json.Unmarshal(data, &aux)
	require.NoError(t, err)

	var resultAccount db.Account

	mp, err := json.Marshal(aux.Data)
	require.NoError(t, err)

	json.Unmarshal(mp, &resultAccount)

	require.Equal(t, resultAccount.ID, account.ID)
	require.Equal(t, resultAccount.Balance, account.Balance)
	require.Equal(t, resultAccount.CurrencyCode, account.CurrencyCode)
	require.Equal(t, resultAccount.Owner, account.Owner)
	require.GreaterOrEqual(t, resultAccount.CreatedAt, account.CreatedAt)

}

func Test_GetAccount(t *testing.T) {
	account := generateRandomAccount()

	testcases := []struct {
		name          string
		accountId     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountId: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status
				require.Equal(t, http.StatusOK, recorder.Code)
				// check response body
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "Not Found",
			accountId: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "Internal Server Error",
			accountId: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountId: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			//  new server with store
			server := NewServer(store)
			// build  response struct
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountId)
			// make request againt the url
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {

	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var aux result
	err = json.Unmarshal(data, &aux)
	require.NoError(t, err)

	var resultAccount db.Account

	mp, err := json.Marshal(aux.Data)
	require.NoError(t, err)

	json.Unmarshal(mp, &resultAccount)

	require.Equal(t, account, resultAccount)

}
