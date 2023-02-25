package db

import (
	"context"
	"testing"
	"time"

	"github.com/dibrito/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	want := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	got, err := testQueries.CreateUser(context.Background(), want)
	// require stops the test if fails
	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Equal(t, want.Username, got.Username)
	require.Equal(t, want.HashedPassword, got.HashedPassword)
	require.Equal(t, want.FullName, got.FullName)
	require.Equal(t, want.Email, got.Email)

	require.NotZero(t, got.CreatedAt)
	require.NotZero(t, got.PasswordChangedAt)
	return got
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	want := createRandomUser(t)
	got, err := testQueries.GetUser(context.Background(), want.Username)
	require.NoError(t, err)
	require.NotEmpty(t, got)

	require.Equal(t, want.Username, got.Username)
	require.Equal(t, want.HashedPassword, got.HashedPassword)
	require.Equal(t, want.FullName, got.FullName)
	require.Equal(t, want.Email, got.Email)

	require.WithinDuration(t, want.CreatedAt, got.CreatedAt, time.Second)
	require.WithinDuration(t, want.PasswordChangedAt, got.PasswordChangedAt, time.Second)
}
