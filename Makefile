clean-test:
	go clean -testcache

test:
	go test --tags=!integration ./...

integration_test:
	sudo docker compose -f docker/docker-compose-test.yml up -d --build
	go test --tags=integration ./...

# if you don't have mysql on your system run docker first
run-docker:
	sudo docker compose -f docker/docker-compose.yml up -d --build

start:
	go run cmd/web-api/main.go --config cmd/web-api/config/config.dev.yml --env cmd/web-api/config/.env
