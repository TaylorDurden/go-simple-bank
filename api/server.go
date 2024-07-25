package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/taylordurden/go-simple-bank/db/sqlc"
)

// serve http requests for our banking service
type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	router.GET("/users/:username", server.getUserHandler)
	router.POST("/users", server.createUserHandler)

	router.POST("/accounts", server.createAccountHandler)
	router.GET("/accounts/:id", server.getAccountHandler)
	router.GET("/accounts", server.listPagedAccountHandler)
	router.DELETE("/accounts/:id", server.deleteAccountHandler)
	router.PATCH("/accounts/:id", server.updateAccountHandler)

	router.POST("/transfers", server.createTransferHandler)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
