package main

import (
	"database/sql"
	"log"

	"github.com/DEVunderdog/gobank/api"
	db "github.com/DEVunderdog/gobank/database/sqlc"
	"github.com/DEVunderdog/gobank/utils"
	_ "github.com/lib/pq"
)

func main() {

	config, err := utils.LoadConfig(".")

	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config,store)
	if err != nil {
		log.Fatal("Server error, %w", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
