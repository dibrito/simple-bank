package main

import (
	"database/sql"
	"log"

	"github.com/dibrito/simple-bank/api"
	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/dibrito/simple-bank/util"
	_ "github.com/lib/pq"
)

// const (
// 	dbDriver = "postgres"
// 	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
// 	address  = "0.0.0.0:8080"
// )

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config:%v", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("unable to open db connection:%v", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatalf("cannot create server:%v", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatalf("cannot start server:%v", err)
	}
}
