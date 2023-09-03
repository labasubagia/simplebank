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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestVerifyEmail(t *testing.T) {
	user, _ := randomUser(t)
	verifyEmail := randomVerifyEmail(user.Username, time.Minute)

	testCases := []struct {
		name          string
		req           *pb.VerifyEmailRequest
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, res *pb.VerifyEmailResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.VerifyEmailRequest{
				EmailId:    verifyEmail.ID,
				SecretCode: verifyEmail.SecretCode,
			},
			buildStubs: func(store *mock_db.MockStore) {
				arg := db.VerifyEmailTxParams{
					EmailID:    verifyEmail.ID,
					SecretCode: verifyEmail.SecretCode,
				}
				u := user
				u.IsEmailVerified = true
				res := db.VerifyEmailTxResult{
					User:        u,
					VerifyEmail: verifyEmail,
				}
				store.EXPECT().VerifyEmailTx(gomock.Any(), arg).Times(1).Return(res, nil)
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, true, res.GetIsVerified())
			},
		},
		{
			name: "ErrInvalidInput",
			req: &pb.VerifyEmailRequest{
				EmailId:    -1,
				SecretCode: "",
			},
			buildStubs: func(store *mock_db.MockStore) {
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "ErrVerify",
			req: &pb.VerifyEmailRequest{
				EmailId:    verifyEmail.ID,
				SecretCode: verifyEmail.SecretCode,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().VerifyEmailTx(gomock.Any(), gomock.Any()).Times(1).Return(db.VerifyEmailTxResult{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
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

			res, err := server.VerifyEmail(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}

func randomVerifyEmail(username string, expires time.Duration) db.VerifyEmail {
	return db.VerifyEmail{
		ID:         util.RandomInt(1, 200),
		Username:   username,
		Email:      util.RandomEmail(),
		SecretCode: util.RandomString(32),
		CreatedAt:  time.Now(),
		ExpiredAt:  time.Now().Add(expires),
	}
}
