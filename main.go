package main

import (
	"database/sql"
	"log"

	"github.com/caleberi/simple-bank/api"
	db "github.com/caleberi/simple-bank/db/sqlc"
	"github.com/caleberi/simple-bank/pkg/utils"
	_ "github.com/lib/pq"
)

func main() {
	cfg := utils.LoadConfig(".", "dev", "env")

	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal("[ERROR] cannot connect to database :", err)
	}
	store := db.NewStore(conn)
	server := api.NewServer(store)

	if err := server.Start(cfg.VBankAddr); err != nil {
		log.Fatal("[ERROR] cannot start server :", err)
	}
}
