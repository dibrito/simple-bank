package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Different types of error returned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	IssuedAt time.Time `json:"issuedAt"`
	ExpireAt time.Time `json:"expiredAt"`
}

// NewPayload creates a new token payload with the given username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:       tokenID,
		Username: username,
		IssuedAt: time.Now(),
		ExpireAt: time.Now().Add(duration),
	}

	return payload, nil
}

// Valid checks if the token payload is expired
func (p Payload) Valid() error {
	if time.Now().After(p.ExpireAt) {
		return ErrExpiredToken
	}
	return nil
}
