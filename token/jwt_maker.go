package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil
}

// CreateToken creates and assing a token for a user
func (j *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	p, err := NewPayload(username, duration)
	if err != nil {
		return "", p, err
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, p)
	token, err := jwtToken.SignedString([]byte(j.secretKey))
	return token, p, err
}

// VerifyToken check if a token is valid
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// receives the parsed but unverified token,
	// You should verify its header to make sure that
	// the signing algorithm matches with
	// what you normally use to sign the tokens.
	// Then if it matches, you return the key

	// when we create the token with siging method
	// SigningMethodHS256 from type SigningMethodHMAC
	// that's why we try to convert back to SigningMethodHMAC
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			// the algorithm of the token doesn’t match with our signing algorithm
			return nil, ErrInvalidToken
		}
		// If the conversion is successful,
		// then it means the algorithm matches.
		// We can just return the secret key
		// that we’re using to sign the token
		return []byte(maker.secretKey), nil

	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		// Either the token is invalid or it is expired.
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
