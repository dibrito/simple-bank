package api

import (
	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store db.Store) *Server {
	s := &Server{
		store: store,
	}
	router := gin.Default()

	// bind custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	router.POST("/users", s.createUser)
	// note POST receives multipe funcs and and last is the handler
	// others are middlewares
	router.POST("/accounts", s.createAccount)
	router.GET("/accounts/:id", s.getAccount)
	router.GET("/accounts/", s.listAccount)

	router.POST("/transfers", s.createTransfer)
	s.router = router
	return s
}

// Start runs the HTTP server on a specif andresss to start handling api requests.
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}
