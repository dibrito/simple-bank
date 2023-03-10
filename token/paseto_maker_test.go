package token

import (
	"fmt"
	"testing"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/dibrito/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	//create maker
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	user := util.RandomOwner()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := time.Now().Add(duration)

	// create token
	token, err := maker.CreateToken(user, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// verify token
	p, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, p)

	// assert payload
	require.NotZero(t, p.ID)
	require.Equal(t, user, p.Username)
	require.WithinDuration(t, issuedAt, p.IssuedAt, time.Minute)
	require.WithinDuration(t, expiredAt, p.ExpireAt, time.Minute)
}

func TestExpiredPasetoToken(t *testing.T) {
	//create maker
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	// create token
	token, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// verify token
	p, err := maker.VerifyToken(token)
	require.Error(t, err, ErrExpiredToken)
	require.Nil(t, p)
}

func TestInvalidPasetoToken(t *testing.T) {
	//create maker
	maker, err := NewPasetoMaker(util.RandomString(31))
	require.EqualError(t, err, fmt.Errorf("invalid key size: must be exactly %d", chacha20poly1305.KeySize).Error())
	require.Empty(t, maker)

	// // create token
	// token, err := maker.CreateToken(util.RandomOwner(), time.Minute)
	// require.NoError(t, err)
	// require.NotEmpty(t, token)

	// // verify token
	// p, err := maker.VerifyToken(token)
	// require.Error(t, err, ErrExpiredToken)
	// require.Nil(t, p)
}
