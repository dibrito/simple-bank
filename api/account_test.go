package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db_mock "github.com/dibrito/simple-bank/db/mocks"
	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/dibrito/simple-bank/token"
	"github.com/dibrito/simple-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccountApi(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	tcs := []struct {
		name          string
		accountID     int64
		setStubs      func(store *db_mock.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
		setAuth       func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			accountID: account.ID,
			name:      "when valid account should get account - OK",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			accountID: account.ID,
			name:      "Unauthorized user",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			accountID: account.ID,
			name:      "noAuthorization",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			accountID: account.ID,
			name:      "when no account found should return - 404",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			accountID: 0,
			name:      "when invalid account request should return - 400",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0).Return(db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			accountID: account.ID,
			name:      "when store error should return - 500",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// mock store
			store := db_mock.NewMockStore(ctrl)
			tc.setStubs(store)

			// build server
			server := newTestServer(t, store)
			// respose recorder
			recorder := httptest.NewRecorder()

			// build request
			url := fmt.Sprintf("/accounts/%v", tc.accountID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// set auth
			tc.setAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccountApi(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	tcs := []struct {
		name          string
		body          gin.H
		accountID     int64
		setStubs      func(store *db_mock.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
		setAuth       func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			accountID: account.ID,
			body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			name: "when valid request should create account - OK",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), db.CreateAccountParams{
					Owner:    user.Username,
					Currency: account.Currency,
				}).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			accountID: account.ID,
			body: gin.H{
				"owner":    account.Owner,
				"currency": "XXX",
			},
			name: "when ivalid request should return - 400",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Times(0).Return(db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			accountID: account.ID,
			body: gin.H{
				"owner":    user.Username,
				"currency": account.Currency,
			},
			name: "when error in store should return - 500",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), db.CreateAccountParams{
					Owner:    user.Username,
					Currency: account.Currency,
				}).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// mock store
			store := db_mock.NewMockStore(ctrl)
			tc.setStubs(store)

			// build server
			server := newTestServer(t, store)
			// respose recorder
			recorder := httptest.NewRecorder()

			// build request
			url := "/accounts"

			body, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			require.NoError(t, err)

			tc.setAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountApi(t *testing.T) {
	user, _ := randomUser(t)
	accounts := make([]db.Account, 5)
	for i := 0; i < 5; i++ {
		accounts[i] = randomAccount(user.Username)
	}

	type Query struct {
		PageID   int32
		PageSize int32
	}

	tcs := []struct {
		name          string
		query         Query
		setStubs      func(store *db_mock.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
		setAuth       func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			query: Query{
				PageID:   1,
				PageSize: 5,
			},
			name: "when valid request should list account - OK",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().ListAccounts(gomock.Any(), db.ListAccountsParams{
					Owner: user.Username,
					// limit is page size
					Limit: 5,
					// offset is n records skipped
					Offset: 0,
				}).Times(1).Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			query: Query{
				PageID:   0,
				PageSize: 5,
			},
			name: "when invalid list request should return - 400",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Any()).Times(0).Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			query: Query{
				PageID:   1,
				PageSize: 5,
			},
			name: "when eror in store should return - 500",
			setAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			setStubs: func(store *db_mock.MockStore) {
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Any()).Times(1).Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := db_mock.NewMockStore(ctrl)
			tc.setStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// build request
			url := "/accounts/"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			// Add query parameters to request URL
			q := req.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.PageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.PageSize))
			req.URL.RawQuery = q.Encode()

			tc.setAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandonMoney(),
		Currency: util.RandonCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	b, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	var got db.Account
	json.Unmarshal(b, &got)
	require.Equal(t, account, got)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, account []db.Account) {
	b, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	var got []db.Account
	json.Unmarshal(b, &got)
	require.Equal(t, account, got)
}
