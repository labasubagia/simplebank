package db

import (
	"context"
	"testing"

	"github.com/labasubagia/simplebank/util"
	"github.com/stretchr/testify/require"
)

func createRandomVerifyEmail(t *testing.T) VerifyEmail {
	user := createRandomUser(t)
	arg := CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      util.RandomEmail(),
		SecretCode: util.RandomString(32),
	}
	verifyEmail, err := testStore.CreateVerifyEmail(context.Background(), arg)
	require.NoError(t, err)
	require.NotNil(t, verifyEmail)
	require.Equal(t, arg.Username, verifyEmail.Username)
	require.Equal(t, arg.Email, verifyEmail.Email)
	require.Equal(t, arg.SecretCode, verifyEmail.SecretCode)
	require.False(t, verifyEmail.IsUsed)
	return verifyEmail
}

func TestCreateVerifyEmail(t *testing.T) {
	createRandomVerifyEmail(t)
}

func TestUpdateVerifyEmail(t *testing.T) {
	old := createRandomVerifyEmail(t)
	arg := UpdateVerifyEmailParams{
		ID:         old.ID,
		SecretCode: old.SecretCode,
	}
	current, err := testStore.UpdateVerifyEmail(context.Background(), arg)
	require.NoError(t, err)
	require.NotNil(t, current)
	require.Equal(t, old.ID, current.ID)
	require.False(t, old.IsUsed)
	require.True(t, current.IsUsed)
}
