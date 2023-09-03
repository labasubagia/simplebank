package restful_api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	mock_db "github.com/labasubagia/simplebank/db/mock"
	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/util"
	"github.com/labasubagia/simplebank/util/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRefreshToken(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_db.NewMockStore(ctrl)

	server := newTestServer(t, store)

	user, _ := randomUser(t)
	refreshToken, refreshTokenPayload, err := server.tokenMaker.CreateToken(user.Username, time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)
	require.NotEmpty(t, refreshTokenPayload)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(t *testing.T, store *mock_db.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"refresh_token": refreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				session := db.Session{
					Username:     user.Username,
					RefreshToken: refreshToken,
					ExpiredAt:    refreshTokenPayload.ExpiredAt,
					CreatedAt:    time.Now(),
				}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InvalidInput",
			body: gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FailedVerify",
			body: gin.H{
				"refresh_token": util.RandomString(32),
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "SessionNotFound",
			body: gin.H{
				"refresh_token": refreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(db.Session{}, db.ErrRecordNotFound)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "FailedGetSession",
			body: gin.H{
				"refresh_token": refreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(db.Session{}, pgx.ErrTxClosed)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "MismatchUser",
			body: gin.H{
				"refresh_token": refreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "other_user", time.Minute)
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				session := db.Session{
					Username:     "other_user",
					RefreshToken: refreshToken,
					ExpiredAt:    refreshTokenPayload.ExpiredAt,
					CreatedAt:    time.Now(),
				}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Blocked",
			body: gin.H{
				"refresh_token": refreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				session := db.Session{
					Username:     user.Username,
					RefreshToken: refreshToken,
					IsBlocked:    true,
					ExpiredAt:    refreshTokenPayload.ExpiredAt,
					CreatedAt:    time.Now(),
				}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "MismatchToken",
			body: gin.H{
				"refresh_token": refreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				otherToken, otherTokenPayload, err := server.tokenMaker.CreateToken(user.Username, time.Second)
				require.NoError(t, err)
				session := db.Session{
					Username:     user.Username,
					RefreshToken: otherToken,
					ExpiredAt:    otherTokenPayload.ExpiredAt,
					CreatedAt:    time.Now(),
				}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredSession",
			body: gin.H{
				"refresh_token": refreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				session := db.Session{
					Username:     user.Username,
					RefreshToken: refreshToken,
					ExpiredAt:    time.Now().Add(-time.Minute),
					CreatedAt:    time.Now(),
				}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(t, store)

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/v1/token/renew_access"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
