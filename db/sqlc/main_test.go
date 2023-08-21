package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/caleberi/simple-bank/pkg/utils"
	_ "github.com/lib/pq"
)

var db *sql.DB
var testQueries *Queries
var err error

func init() {
	cfg := utils.LoadConfig("../../", "dev", "env")
	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
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
