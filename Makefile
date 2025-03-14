postgres:
	docker run --name postgres17 --network bank_network -p 127.0.0.1:5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:17-alpine

simple_bank:
	docker run --name simple_bank -e GIN_MODE=release -e DB_SOURCE="postgresql://root:secret@postgres17/simple_bank?sslmode=disable" -e SERVER_ADDRESS="0.0.0.0:6000" --network bank_network -p 127.0.0.1:6000:6000 simple_bank:latest

createdb:
	docker exec -it postgres17 createdb -U root -O root simple_bank

dropdb:
	docker exec -it postgres17 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@127.0.0.1:5432/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@127.0.0.1:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@127.0.0.1:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@127.0.0.1:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/techschool/simplebank/db/sqlc Store

.PHONY: postgres simple_bank createdb dropdb  migrateup migratedown migrateup1 migratedown1 sqlc test server mock