package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labasubagia/simplebank/util"
	"github.com/stretchr/testify/require"
)

func createRandomSession(t *testing.T, expires time.Duration) Session {
	user := createRandomUser(t)
	arg := CreateSessionParams{
		ID:           uuid.New(),
		Username:     user.Username,
		RefreshToken: util.RandomString(32),
		ExpiredAt:    time.Now().Add(expires),
		UserAgent:    util.RandomString(4),
		ClientIp:     util.RandomString(4),
		IsBlocked:    false,
	}
	session, err := testStore.CreateSession(context.Background(), arg)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, arg.ID, session.ID)
	require.Equal(t, arg.Username, session.Username)
	require.Equal(t, arg.RefreshToken, session.RefreshToken)
	require.Equal(t, arg.UserAgent, session.UserAgent)
	require.Equal(t, arg.ClientIp, session.ClientIp)
	return session
}

func TestCreateSession(t *testing.T) {
	createRandomSession(t, time.Millisecond)
}

func TestGetSession(t *testing.T) {
	session := createRandomSession(t, time.Millisecond)
	result, err := testStore.GetSession(context.Background(), session.ID)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, session.ID, result.ID)
	require.Equal(t, session.Username, result.Username)
	require.Equal(t, session.RefreshToken, result.RefreshToken)
	require.Equal(t, session.UserAgent, result.UserAgent)
	require.Equal(t, session.ClientIp, result.ClientIp)
}
