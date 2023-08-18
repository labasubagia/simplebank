env_up:
	docker compose up -d

env_down:
	docker compose down


migrate_up:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose up
 
migrate_down:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose down

migrate_drop:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose drop

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

k8s_run:
	skaffold dev

deploy_systemd:
	go build

	# create workdir
	mkdir -p /opt/simplebank
	cp -u simplebank /opt/simplebank/simplebank
	touch /opt/simplebank/.env # just so the binary not error
	
	# add systemd config
	cp -u infra/systemd/* /lib/systemd/system/
	
	# run service
	systemctl start simplebank
	systemctl status simplebank

 