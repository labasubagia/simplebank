package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestCreateAccount_OK(t *testing.T) {
	account := randomAccount()
	body := createAccountRequest{
		Owner:    account.Owner,
		Currency: account.Currency,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)
	store.EXPECT().
		CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{Owner: body.Owner, Currency: body.Currency})).
		Times(1).
		Return(account, nil)

	server := NewServer(store)
	recorder := httptest.NewRecorder()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	url := "/accounts"
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestCreateAccount_InternalError(t *testing.T) {
	account := randomAccount()
	body := createAccountRequest{
		Owner:    account.Owner,
		Currency: account.Currency,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)
	store.EXPECT().
		CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{Owner: body.Owner, Currency: body.Currency})).
		Times(1).
		Return(db.Account{}, errors.New("unknown error"))

	server := NewServer(store)
	recorder := httptest.NewRecorder()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	url := "/accounts"
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestCreateAccount_BodyInvalid(t *testing.T) {
	body := struct{}{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)
	store.EXPECT().
		CreateAccount(gomock.Any(), gomock.Any()).
		Times(0)

	server := NewServer(store)
	recorder := httptest.NewRecorder()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	url := "/accounts"
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
