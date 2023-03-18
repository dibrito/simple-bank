package api

import (
	"fmt"

	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/dibrito/simple-bank/token"
	"github.com/dibrito/simple-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
	config     util.Config
}

// NewServer creates a new HTTP server and setup routing.
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

	// bind custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()
	return server, nil
}

func (s *Server) setupRouter() {
	router := gin.Default()
	router.POST("/users", s.createUser)
	router.POST("/users/login", s.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(s.tokenMaker))

	// note POST receives multipe funcs and and last is the handler
	// others are middlewares
	authRoutes.POST("/accounts", s.createAccount)
	authRoutes.GET("/accounts/:id", s.getAccount)
	authRoutes.GET("/accounts/", s.listAccount)
	authRoutes.POST("/transfers", s.createTransfer)
	s.router = router
}

// Start runs the HTTP server on a specif andresss to start handling api requests.
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}
