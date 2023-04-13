package token

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dibrito/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	//create maker
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	user := util.RandomOwner()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := time.Now().Add(duration)

	// create token
	token, payload, err := maker.CreateToken(user, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

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

func TestExpiredJWTToken(t *testing.T) {
	//create maker
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	// create token
	token, payload, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	// verify token
	p, err := maker.VerifyToken(token)
	require.Error(t, err, ErrExpiredToken)
	require.Nil(t, p)
}

func TesInvalidJWTTokenAlgNone(t *testing.T) {
	// create payload
	payload, err := NewPayload(util.RandomOwner(), time.Minute)
	require.NoError(t, err)

	// create unsafe token
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}
