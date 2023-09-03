package api

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	mock_db "github.com/labasubagia/simplebank/db/mock"
	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/grpc/pb"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRenewAccessToken(t *testing.T) {

	storeCtrl := gomock.NewController(t)
	defer storeCtrl.Finish()
	store := mock_db.NewMockStore(storeCtrl)
	server := newTestServer(t, store, nil)

	user, _ := randomUser(t)
	refreshToken, refreshTokenPayload, err := server.tokenMaker.CreateToken(user.Username, server.config.RefreshTokenDuration)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		req           *pb.RenewAccessTokenRequest
		buildStubs    func(t *testing.T, store *mock_db.MockStore)
		checkResponse func(t *testing.T, res *pb.RenewAccessTokenResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.RenewAccessTokenRequest{
				RefreshToken: refreshToken,
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				session := db.Session{Username: user.Username, RefreshToken: refreshToken, ExpiredAt: refreshTokenPayload.ExpiredAt}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(session, nil)
			},
			checkResponse: func(t *testing.T, res *pb.RenewAccessTokenResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
			},
		},
		{
			name: "Blocked",
			req: &pb.RenewAccessTokenRequest{
				RefreshToken: refreshToken,
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				session := db.Session{Username: user.Username, RefreshToken: refreshToken, ExpiredAt: refreshTokenPayload.ExpiredAt, IsBlocked: true}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(session, nil)
			},
			checkResponse: func(t *testing.T, res *pb.RenewAccessTokenResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "IncorrectUser",
			req: &pb.RenewAccessTokenRequest{
				RefreshToken: refreshToken,
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				session := db.Session{Username: "incorrect_user", RefreshToken: refreshToken, ExpiredAt: refreshTokenPayload.ExpiredAt}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(session, nil)
			},
			checkResponse: func(t *testing.T, res *pb.RenewAccessTokenResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "MismatchToken",
			req: &pb.RenewAccessTokenRequest{
				RefreshToken: refreshToken,
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				differentToken, payload, err := server.tokenMaker.CreateToken("other_user", server.config.RefreshTokenDuration)
				require.NoError(t, err)
				session := db.Session{Username: payload.Username, RefreshToken: differentToken, ExpiredAt: refreshTokenPayload.ExpiredAt}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(session, nil)
			},
			checkResponse: func(t *testing.T, res *pb.RenewAccessTokenResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "Expired",
			req: &pb.RenewAccessTokenRequest{
				RefreshToken: refreshToken,
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				session := db.Session{Username: user.Username, RefreshToken: refreshToken}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(session, nil)
			},
			checkResponse: func(t *testing.T, res *pb.RenewAccessTokenResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "Invalid",
			req: &pb.RenewAccessTokenRequest{
				RefreshToken: "invalid_token",
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				session := db.Session{Username: user.Username, RefreshToken: refreshToken}
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(0).Return(session, nil)
			},
			checkResponse: func(t *testing.T, res *pb.RenewAccessTokenResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "NoSession",
			req: &pb.RenewAccessTokenRequest{
				RefreshToken: refreshToken,
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(db.Session{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, res *pb.RenewAccessTokenResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "ErrGetSession",
			req: &pb.RenewAccessTokenRequest{
				RefreshToken: refreshToken,
			},
			buildStubs: func(t *testing.T, store *mock_db.MockStore) {
				store.EXPECT().GetSession(gomock.Any(), gomock.Any()).Times(1).Return(db.Session{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, res *pb.RenewAccessTokenResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(t, store)
			res, err := server.RenewAccessToken(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
