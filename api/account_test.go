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
	"github.com/caleberi/simple-bank/token"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type result struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
}

func generateRandomAccount(owner string) db.Account {
	return db.Account{
		ID:           utils.RandomInt(1, 100),
		Owner:        owner,
		Balance:      utils.RandomMoney(),
		CurrencyCode: utils.RandomCurrencyCode(),
	}
}

func Test_CreateAccount(t *testing.T) {
	user, _ := randomUser(t)
	account := generateRandomAccount(user.Username)

	createAccountRequest := createAccountRequest{
		CurrencyCode: utils.RandomCurrencyCode(),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().CreateAccount(gomock.Any(), db.CreateAccountParams{
		CurrencyCode: createAccountRequest.CurrencyCode,
		Balance:      0,
		Owner:        user.Username,
	}).Times(1).Return(account, nil)

	url := "/accounts"

	//  new server with store
	server := newTestServer(t, store)

	buf := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(createAccountRequest)
	require.NoError(t, err)

	// build  response struct
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodPost, url, buf)
	request.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)

	addAuthorization(t, request, server.tokenGenerator, authorizationBearerType, user.Username, time.Minute)
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
	user, _ := randomUser(t)
	account := generateRandomAccount(user.Username)

	testcases := []struct {
		name          string
		accountId     int64
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationBearerType, user.Username, time.Minute)
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
			}, setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationBearerType, user.Username, time.Minute)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationBearerType, user.Username, time.Minute)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationBearerType, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountId: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationBearerType, "unauthorized_user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		}, {
			name:      "NoAuthorization",
			accountId: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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
			server := newTestServer(t, store)
			// build  response struct
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountId)

			// make request againt the url
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenGenerator)

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
