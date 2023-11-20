build:
	go build

test:
	go test -v ./...
lint:
	golangci-lint run

run:
	go run main.go

destroy:
	docker stop $$(docker ps -aq) || true
	docker container stop $$(docker container ls -aq) || true
	docker container prune --force || true

run-dependencies:
	docker compose --file docker-compose.yaml up --detach --force-recreate --renew-anon-volumes --build
