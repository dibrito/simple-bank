package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/dibrito/simple-bank/db/util"
	_ "github.com/lib/pq"
)

// const (
// 	dbDriver = "postgres"
// 	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
// )

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	c, err := util.LoadConfig("./../..")
	if err != nil {
		log.Fatalf("cannot load config:%v", err)
	}
	// var err error
	testDB, err = sql.Open(c.DBDriver, c.DBSource)
	if err != nil {
		log.Fatalf("unable to open db connection:%v", err)
	}
	testQueries = New(testDB)
	os.Exit(m.Run())
}
