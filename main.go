package main

import (
	"database/sql"
	"log"

	"github.com/labasubagia/simplebank/api"
	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/util"
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
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
