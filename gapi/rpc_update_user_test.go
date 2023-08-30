package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	mock_db "github.com/labasubagia/simplebank/db/mock"
	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/pb"
	"github.com/labasubagia/simplebank/token"
	"github.com/labasubagia/simplebank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateUser(t *testing.T) {
	user, _ := randomUser(t)

	newName := util.RandomOwner()
	newEmail := util.RandomEmail()
	invalidEmail := "invalid_email"

	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStubs    func(store *mock_db.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mock_db.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					Email: pgtype.Text{
						String: newEmail,
						Valid:  true,
					},
					FullName: pgtype.Text{
						String: newName,
						Valid:  true,
					},
				}
				updatedUser := db.User{
					Username:          arg.Username,
					HashedPassword:    user.HashedPassword,
					FullName:          newName,
					Email:             newEmail,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   user.IsEmailVerified,
				}
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updatedUser, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				updatedUser := res.GetUser()
				require.Equal(t, user.Username, updatedUser.GetUsername())
				require.Equal(t, newEmail, updatedUser.GetEmail())
				require.Equal(t, newName, updatedUser.GetFullName())
			},
		},
		{
			name: "UserNotFound",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, db.ErrRecordNotFound)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "InternalError",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "InvalidEmail",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &invalidEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "PermissionDenied",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, invalidEmail, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "NoAuthorization",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mock_db.NewMockStore(storeCtrl)

			tc.buildStubs(store)

			server := newTestServer(t, store, nil)
			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.UpdateUser(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
