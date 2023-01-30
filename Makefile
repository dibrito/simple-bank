postgress:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

mysql:
	docker run --name mysql8 -p 3306:3306 -e MYSQL_ROOT_PASSWORD=secret -e MYSQL_DATABASE=simple_bank -d mysql:latest

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
#docker exec -it postgres12 dropdb simple_bank
	docker stop postgres12
	docker rm postgres12
	docker stop mysql8
	docker rm mysql8

migrateup:
	migrate -path=db/migration -database="postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path=db/migration -database="postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test: sqlc
	go test -v -cover ./...

.PHONY: dropdb createdb createdb migrateup migratedown
