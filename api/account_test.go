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

	db_mock "github.com/dibrito/simple-bank/db/mocks"
	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/dibrito/simple-bank/db/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccountApi(t *testing.T) {
	account := randomAccount()
	tcs := []struct {
		name          string
		accountID     int64
		setStubs      func(store *db_mock.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			accountID: account.ID,
			name:      "when valid account should get account - OK",
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
			name:      "when no account found should return - 404",
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
			server := NewServer(store)
			// respose recorder
			recorder := httptest.NewRecorder()

			// build request
			url := fmt.Sprintf("/accounts/%v", tc.accountID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
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
