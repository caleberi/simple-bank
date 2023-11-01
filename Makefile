postgres:
	docker run --name postgres12 -p 4600:5432  -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine
createdb:
	docker exec -it postgres12 createdb --username=root --owner=root  simple_bank
dropdb:
	docker exec -it postgres12 dropdb simple_bank
migrateup:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:4600/simple_bank?sslmode=disable" -verbose up
migratedown:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:4600/simple_bank?sslmode=disable" -verbose down
migrateup_1:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:4600/simple_bank?sslmode=disable" -verbose up 1
migratedown_1:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:4600/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate
test:
	go test -v -race -cover ./...
start_server:
	go run main.go
mock_db:
	mockgen --build_flags=--mod=mod -package mockdb -destination db/mock/store.go github.com/caleberi/simple-bank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown migrateup_1 migratedown_1 sqlc test start_server mock_db