package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyEmailTx(t *testing.T) {
	verifyEmail := createRandomVerifyEmail(t)
	arg := VerifyEmailTxParams{
		EmailID:    verifyEmail.ID,
		SecretCode: verifyEmail.SecretCode,
	}
	result, err := testStore.VerifyEmailTx(context.Background(), arg)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, verifyEmail.ID, result.VerifyEmail.ID)
	require.Equal(t, verifyEmail.Username, result.User.Username)
	require.True(t, result.VerifyEmail.IsUsed)
	require.True(t, result.User.IsEmailVerified)
}
