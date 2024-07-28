package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/taylordurden/go-simple-bank/db/sqlc"
	"github.com/taylordurden/go-simple-bank/token"
	"github.com/taylordurden/go-simple-bank/util"
)

// serve http requests for our banking service
type Server struct {
	config    util.Config
	store     db.Store
	tokenAuth token.Authenticator
	router    *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenAuth, err := token.NewPasetoAuthenticator(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		store:     store,
		tokenAuth: tokenAuth,
		config:    config,
	}

	// custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupServerRouter()
	return server, nil
}

func (server *Server) setupServerRouter() {
	router := gin.Default()

	router.GET("/users/:username", server.getUserHandler)
	router.POST("/users", server.createUserHandler)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenAuth))

	authRoutes.POST("/accounts", server.createAccountHandler)
	authRoutes.GET("/accounts/:id", server.getAccountHandler)
	authRoutes.GET("/accounts", server.listPagedAccountHandler)
	authRoutes.DELETE("/accounts/:id", server.deleteAccountHandler)
	authRoutes.PATCH("/accounts/:id", server.updateAccountHandler)

	authRoutes.POST("/transfers", server.createTransferHandler)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
