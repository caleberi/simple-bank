package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:4600/simple_bank?sslmode=disable"
)

var db *sql.DB
var testQueries *Queries
var err error

func init() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("[ERROR] error opening database connection: ", err)
	}
	db = conn
}

func TestMain(m *testing.M) {
	if err != nil {
		log.Fatal("[ERROR] cannot connect tot db: ", err)
	}
	testQueries = New(db)
	os.Exit(m.Run())
}
