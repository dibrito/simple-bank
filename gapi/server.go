package gapi

import (
	"fmt"

	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/dibrito/simple-bank/pb"
	"github.com/dibrito/simple-bank/token"
	"github.com/dibrito/simple-bank/util"
)

// Server serves gRPC requests for our banking service.
type Server struct {
	// Its main purpose is to enable forward compatibility,
	// Which means that the server can already accept the calls
	// to the CreateUser and LoginUser RPCs before they are actually implemented.

	// it easier for a team
	// to work on multiple RPCs in parallel without blocking or conflicting with each other.
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

// NewServer creates a new gRPC server.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("can not create token maker: %v", err)
	}
	server := &Server{
		store:      store,
		config:     config,
		tokenMaker: tokenMaker,
	}
	return server, nil
}
