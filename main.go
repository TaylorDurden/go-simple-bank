package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/taylordurden/go-simple-bank/api"
	db "github.com/taylordurden/go-simple-bank/db/sqlc"
	"github.com/taylordurden/go-simple-bank/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("load config err:", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("connection err to db:", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("start server err:", err)
	}
}
