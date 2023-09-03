package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/grpc/pb"
	"github.com/labasubagia/simplebank/util"
	"github.com/labasubagia/simplebank/util/token"
	"github.com/labasubagia/simplebank/worker"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	store           db.Store
	router          *gin.Engine
	config          util.Config
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		store:           store,
		config:          config,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	return server, nil
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
