package api

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	mock_db "github.com/labasubagia/simplebank/db/mock"
	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/grpc/pb"
	"github.com/labasubagia/simplebank/util"
	"github.com/labasubagia/simplebank/util/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateAccount(t *testing.T) {
	user, _ := randomUser(t)
	currency := util.USD

	testCases := []struct {
		name          string
		req           *pb.CreateAccountRequest
		buildStubs    func(store *mock_db.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.CreateAccountResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateAccountRequest{
				Currency: currency,
			},
			buildStubs: func(store *mock_db.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    user.Username,
					Balance:  0,
					Currency: currency,
				}
				account := db.Account{
					Owner:     user.Username,
					Balance:   0,
					Currency:  arg.Currency,
					CreatedAt: time.Now(),
				}
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, res.Account.Currency, currency)
				require.Equal(t, res.Account.Owner, user.Username)
			},
		},
		{
			name: "Unauthenticated",
			req: &pb.CreateAccountRequest{
				Currency: currency,
			},
			buildStubs: func(store *mock_db.MockStore) {
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InvalidCurrency",
			req: &pb.CreateAccountRequest{
				Currency: "invalid_currency",
			},
			buildStubs: func(store *mock_db.MockStore) {
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "Duplicate",
			req: &pb.CreateAccountRequest{
				Currency: currency,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, db.ErrUniqueViolation)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.AlreadyExists, st.Code())
			},
		},
		{
			name: "Failed",
			req: &pb.CreateAccountRequest{
				Currency: currency,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, pgx.ErrTxClosed)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
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

			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mock_db.NewMockStore(storeCtrl)

			tc.buildStubs(store)

			server := newTestServer(t, store, nil)
			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.CreateAccount(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}
