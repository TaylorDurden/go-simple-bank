package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/taylordurden/go-simple-bank/api"
	db "github.com/taylordurden/go-simple-bank/db/sqlc"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:postgres@127.0.0.1:5432/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("connection err to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("start server err:", err)
	}
}
