package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	mock_db "github.com/labasubagia/simplebank/db/mock"
	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

func TestGetAccount_OK(t *testing.T) {
	account := randomAccount()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)
	store.EXPECT().
		GetAccount(gomock.Any(), gomock.Eq(account.ID)).
		Times(1).
		Return(account, nil)

	server := NewServer(store)
	recorder := httptest.NewRecorder()

	url := fmt.Sprintf("/accounts/%d", account.ID)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	requireBodyMatchAccount(t, recorder.Body, account)
}

func TestGetAccount_InvalidID(t *testing.T) {
	account := randomAccount()
	account.ID = 0

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)
	store.EXPECT().
		GetAccount(gomock.Any(), gomock.Any()).
		Times(0)

	server := NewServer(store)
	recorder := httptest.NewRecorder()

	url := fmt.Sprintf("/accounts/%d", account.ID)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetAccount_NotFound(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)
	store.EXPECT().
		GetAccount(gomock.Any(), gomock.Any()).
		Times(1).
		Return(db.Account{}, sql.ErrNoRows)

	server := NewServer(store)
	recorder := httptest.NewRecorder()

	url := fmt.Sprintf("/accounts/%d", 200)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetAccount_InternalError(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)
	store.EXPECT().
		GetAccount(gomock.Any(), gomock.Any()).
		Times(1).
		Return(db.Account{}, errors.New("unknown error"))

	server := NewServer(store)
	recorder := httptest.NewRecorder()

	url := fmt.Sprintf("/accounts/%d", 200)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestUpdateAccount_OK(t *testing.T) {
	account := randomAccount()
	body := updateAccountRequest{ID: account.ID, Balance: account.Balance}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)
	store.EXPECT().
		UpdateAccount(gomock.Any(), gomock.Eq(db.UpdateAccountParams{ID: body.ID, Balance: body.Balance})).
		Times(1).
		Return(account, nil)

	server := NewServer(store)
	recorder := httptest.NewRecorder()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	url := fmt.Sprintf("/accounts/%d", account.ID)
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestUpdateAccount_NotFound(t *testing.T) {
	body := updateAccountRequest{ID: 200, Balance: 200}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)
	store.EXPECT().
		UpdateAccount(gomock.Any(), gomock.Eq(db.UpdateAccountParams{ID: body.ID, Balance: body.Balance})).
		Times(1).
		Return(db.Account{}, sql.ErrNoRows)

	server := NewServer(store)
	recorder := httptest.NewRecorder()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	url := fmt.Sprintf("/accounts/%d", body.ID)
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestUpdateAccount_BodyInvalid(t *testing.T) {
	body := struct{}{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)
	store.EXPECT().
		UpdateAccount(gomock.Any(), gomock.Eq(body)).
		Times(0)

	server := NewServer(store)
	recorder := httptest.NewRecorder()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	url := fmt.Sprintf("/accounts/%d", 200)
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestCreateAccount(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			buildStubs: func(store *mock_db.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			buildStubs: func(store *mock_db.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Account{}, errors.New("internal error"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid",
			body: gin.H{},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_db.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/accounts"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListAccounts(t *testing.T) {
	n := 20
	accounts := make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccount()
	}

	type Query struct {
		PageID   int
		PageSize int
	}

	testCases := []struct {
		name          string
		query         *Query
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: &Query{PageID: 2, PageSize: 5},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{Limit: 5, Offset: 5})).
					Times(1).
					Return(accounts[5:10], nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts[5:10])
			},
		},
		{
			name:  "InvalidQuery",
			query: &Query{PageID: 0, PageSize: 10_000},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NoQuery",
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_db.NewMockStore(ctrl)
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/accounts"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			if tc.query != nil {
				query := request.URL.Query()
				query.Add("page_id", strconv.Itoa(tc.query.PageID))
				query.Add("page_size", strconv.Itoa(tc.query.PageSize))
				request.URL.RawQuery = query.Encode()
			}

			tc.buildStubs(store)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}

}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}
