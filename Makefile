postgress:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker stop postgres12
	docker rm postgres12

migrateup:
	migrate -path=db/migration -database="postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path=db/migration -database="postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path=db/migration -database="postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path=db/migration -database="postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package db_mock --destination=./db/mocks/store.go github.com/dibrito/simple-bank/db/sqlc Store

.PHONY: dropdb createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server
