package main

import (
	"database/sql"
	"log"
	"net"
	"sync"

	"github.com/labasubagia/simplebank/api"
	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/gapi"
	"github.com/labasubagia/simplebank/pb"
	"github.com/labasubagia/simplebank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := util.LoadConfig(".env")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)

	// Serve both server for now

	var wg sync.WaitGroup

	wg.Add(1)
	go runGrpcServer(config, store)

	wg.Add(2)
	go runGinServer(config, store)

	wg.Wait()
}

func runGrpcServer(config util.Config, store db.Store) {
	grpcServer := grpc.NewServer()
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener")
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start grpc server")
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
