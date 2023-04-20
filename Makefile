DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable

postgress:
	docker run --name postgres12 --network simplebank -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker stop postgres12
	docker rm postgres12

# if "error: pq: role "root" does not exist" u probably have psql locally running
migrateup:
	migrate -path=db/migration -database="${DB_URL}" -verbose up
# migrate -path=db/migration -database="postgres://root:aYYa6Ij9aXXlQrBBuId6SRQdgU8ccIWe@dpg-cgi1ssak728s1brfp1cg-a.frankfurt-postgres.render.com/simple_bank_kmbe" -verbose up

migrateup1:
	migrate -path=db/migration -database="${DB_URL}" -verbose up 1

migratedown:
	migrate -path=db/migration -database="${DB_URL}" -verbose down

migratedown1:
	migrate -path=db/migration -database="${DB_URL}" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package db_mock --destination=./db/mocks/store.go github.com/dibrito/simple-bank/db/sqlc Store

dbdocs:
	dbdocs build docs/db/db.dbml

dbschema:
	dbml2sql --postgres -o ./docs/db/schema.sql ./docs/db/db.dbml

.PHONY: postgress createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server mock dbdocs dbschema
