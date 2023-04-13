package token

import "time"

// Maker will manager tokens
type Maker interface {
	// CreateToken creates and assing a token for a user
	CreateToken(username string, duration time.Duration) (string, *Payload, error)
	// VerifyToken check if a token is valid
	VerifyToken(token string) (*Payload, error)
}
