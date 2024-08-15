package gapi

import (
	"fmt"

	db "github.com/taylordurden/go-simple-bank/db/sqlc"
	pb "github.com/taylordurden/go-simple-bank/pb"
	"github.com/taylordurden/go-simple-bank/token"
	"github.com/taylordurden/go-simple-bank/util"
)

// serve http requests for our banking service
type Server struct {
	pb.UnimplementedSimpleBankServer
	config    util.Config
	store     db.Store
	tokenAuth token.Authenticator
}

// creates a new gRPC server
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

	return server, nil
}
