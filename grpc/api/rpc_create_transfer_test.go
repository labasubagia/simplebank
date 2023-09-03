package api

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	mock_db "github.com/labasubagia/simplebank/db/mock"
	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/grpc/pb"
	"github.com/labasubagia/simplebank/token"
	"github.com/labasubagia/simplebank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateTransfer(t *testing.T) {
	currency := util.USD

	user1, _ := randomUser(t)
	account1 := randomAccount(user1.Username)
	account1.Currency = currency

	user2, _ := randomUser(t)
	account2 := randomAccount(user2.Username)
	account2.Currency = currency

	amount := int64(10)

	testCases := []struct {
		name          string
		req           *pb.CreateTransferRequest
		buildStubs    func(store *mock_db.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.CreateTransferResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Currency:      currency,
				Amount:        amount,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(account2, nil)

				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}
				res := db.TransferTxResult{
					Transfer:    db.Transfer{FromAccountID: account1.ID, ToAccountID: account2.ID, Amount: arg.Amount},
					FromAccount: account1,
					ToAccount:   account2,
					FromEntry:   db.Entry{AccountID: account1.ID, Amount: -arg.Amount},
					ToEntry:     db.Entry{AccountID: account2.ID, Amount: arg.Amount},
				}
				store.EXPECT().TransferTx(gomock.Any(), arg).Times(1).Return(res, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)

				require.Equal(t, account1.ID, res.Transfer.FromAccountId)
				require.Equal(t, account1.ID, res.FromAccount.Id)
				require.Equal(t, account1.ID, res.FromEntry.AccountId)

				require.Equal(t, account2.ID, res.Transfer.ToAccountId)
				require.Equal(t, account2.ID, res.ToAccount.Id)
				require.Equal(t, account2.ID, res.ToEntry.AccountId)

				require.Equal(t, amount, res.Transfer.Amount)
				require.Equal(t, -amount, res.FromEntry.Amount)
				require.Equal(t, amount, res.ToEntry.Amount)
			},
		},
		{
			name: "ErrUnauthenticated",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Currency:      currency,
				Amount:        10,
			},
			buildStubs: func(store *mock_db.MockStore) {
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "ErrInvalidInput",
			req: &pb.CreateTransferRequest{
				FromAccountId: 0,
				ToAccountId:   0,
				Currency:      "invalid_currency",
				Amount:        -12,
			},
			buildStubs: func(store *mock_db.MockStore) {
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "ErrFromAccountNotFound",
			req: &pb.CreateTransferRequest{
				FromAccountId: 12,
				ToAccountId:   account2.ID,
				Currency:      currency,
				Amount:        amount,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(db.Account{}, db.ErrRecordNotFound)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "ErrFromAccountMismatchCurrency",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Currency:      util.EUR,
				Amount:        amount,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "ErrForbidden",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Currency:      currency,
				Amount:        amount,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(account1, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user2.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "ErrToAccountNotFound",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   12,
				Currency:      currency,
				Amount:        amount,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(db.Account{}, db.ErrRecordNotFound)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "ErrToAccountMismatchCurrency",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Currency:      util.EUR,
				Amount:        amount,
			},
			buildStubs: func(store *mock_db.MockStore) {
				acc1 := account1
				acc1.Currency = util.EUR
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(acc1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(account2, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "ErrTransfer",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Currency:      currency,
				Amount:        amount,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(account2, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(db.TransferTxResult{}, pgx.ErrTxClosed)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
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

			res, err := server.CreateTransfer(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
