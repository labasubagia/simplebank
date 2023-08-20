package gapi

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/pb"
	"github.com/labasubagia/simplebank/token"
	"github.com/labasubagia/simplebank/util"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	store      db.Store
	router     *gin.Engine
	config     util.Config
	tokenMaker token.Maker
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		store:      store,
		config:     config,
		tokenMaker: tokenMaker,
	}

	return server, nil
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
