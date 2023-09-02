package db

import (
	"context"
	"testing"

	"github.com/labasubagia/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestCreateUserTx(t *testing.T) {
	hashedPassword, err := util.HashPassword(util.RandomString(8))
	require.NoError(t, err)
	arg := CreateUserTxParams{
		CreateUserParams: CreateUserParams{
			Username:       util.RandomOwner(),
			HashedPassword: hashedPassword,
			FullName:       util.RandomString(10),
			Email:          util.RandomEmail(),
		},
		AfterCreate: func(user User) error {
			return nil
		},
	}
	result, err := testStore.CreateUserTx(context.Background(), arg)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, arg.CreateUserParams.Username, result.User.Username)
	require.Equal(t, arg.CreateUserParams.Email, result.User.Email)
	require.Equal(t, arg.CreateUserParams.FullName, result.User.FullName)
	require.Equal(t, arg.CreateUserParams.HashedPassword, result.User.HashedPassword)
	require.Equal(t, false, result.User.IsEmailVerified)
}
