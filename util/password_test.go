package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	password := RandomString(6)

	hashed, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashed)

	err = CheckPassword(password, hashed)
	require.NoError(t, err)

	wrongPassword := RandomString(6)
	err = CheckPassword(wrongPassword, hashed)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
}

func TestPasswordCheckDiffHashesSamePass(t *testing.T) {
	password := RandomString(6)

	hashed, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashed)

	hashed2, err2 := HashPassword(password)
	require.NoError(t, err2)
	require.NotEmpty(t, hashed2)

	require.NotEqual(t, hashed, hashed2)
}
