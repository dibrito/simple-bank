package db

import (
	"context"
	"database/sql"
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

func TestUpdateUserOnlyFullName(t *testing.T) {
	oldUser := createRandomUser(t)
	newFullName := util.RandomOwner()
	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		FullName: sql.NullString{
			Valid:  true,
			String: newFullName,
		},
	})
	require.NoError(t, err)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.NotEqual(t, oldUser.FullName, newFullName)
	// make sure nothing else changed
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateUserOnlyEmail(t *testing.T) {
	oldUser := createRandomUser(t)
	newEmail := util.RandomEmail()
	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Email: sql.NullString{
			Valid:  true,
			String: newEmail,
		},
	})
	require.NoError(t, err)
	require.NotEqual(t, oldUser.Email, newEmail)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateUserOnlyPassword(t *testing.T) {
	oldUser := createRandomUser(t)
	newPassword := util.RandomString(6)
	hashed, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		HashedPassword: sql.NullString{
			Valid:  true,
			String: hashed,
		},
	})
	require.NoError(t, err)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, hashed, updatedUser.HashedPassword)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
}

func TestUpdateUserAllFields(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullName := util.RandomOwner()
	newEmail := util.RandomEmail()
	newPassword := util.RandomString(6)

	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		HashedPassword: sql.NullString{
			Valid:  true,
			String: newHashedPassword,
		},
		FullName: sql.NullString{
			Valid:  true,
			String: newFullName,
		},
		Email: sql.NullString{
			Valid:  true,
			String: newEmail,
		},
	})
	require.NoError(t, err)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)

	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, newFullName, updatedUser.FullName)
}
