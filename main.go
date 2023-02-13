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
	c, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config:%v", err)
	}
	conn, err := sql.Open(c.DBDriver, c.DBSource)
	if err != nil {
		log.Fatalf("unable to open db connection:%v", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(c.ServerAddress)
	if err != nil {
		log.Fatalf("cannot start server:%v", err)
	}
}
