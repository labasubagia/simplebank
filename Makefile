DB_URL=postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable
DB_MIGRATION_PATH=db/migration

env_up:
	docker compose up -d

env_down:
	docker compose down


migrate_up:
	migrate -path "$(DB_MIGRATION_PATH)" -database "$(DB_URL)" -verbose up

migrate_down:
	migrate -path "$(DB_MIGRATION_PATH)" -database "$(DB_URL)" -verbose down

migrate_drop:
	migrate -path "$(DB_MIGRATION_PATH)" -database "$(DB_URL)" -verbose drop

db_dump:
	docker compose exec -T db pg_dump -s -U postgres simple_bank > doc/db/schema.sql

db_dbml:
	sql2dbml --postgres doc/db/schema.sql -o doc/db/schema.dbml

db_doc:
	# run `make db_dump`
	# run `make db_dbml`
	# edit doc/db/schema.sql if error, usually line 183 `password_changed_at` remove default value, dbml error
	dbdocs build doc/db/schema.dbml

sqlc:
	sqlc generate

protoc:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		proto/*.proto

test:
	go test -v -cover ./...

server:
	go run main.go

k8s_run:
	skaffold dev

deploy_systemd:
	bash infra/systemd/install.sh

evans:
	evans --host localhost --port 6000 -r repl
