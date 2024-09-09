FORCE:

DOCKER_COMPOSE=docker-compose.yml
DOCKER_COMPOSE_TESTS=docker-compose.tests.yml

tests-build:
	docker-compose -f ${DOCKER_COMPOSE_TESTS} build --force-rm

tests-up:
	docker-compose -f ${DOCKER_COMPOSE_TESTS} up --remove-orphans
	docker-compose -f ${DOCKER_COMPOSE_TESTS} down

tests-down:
	docker-compose -f ${DOCKER_COMPOSE_TESTS} down

build:
	docker-compose -f ${DOCKER_COMPOSE} build --force-rm

up:
	docker-compose -f ${DOCKER_COMPOSE} up  -d

down:
	docker-compose -f ${DOCKER_COMPOSE} down 

logs:
	docker-compose -f ${DOCKER_COMPOSE} logs 

coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

gen-rsa:
	openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:4096
	openssl rsa -pubout -in private_key.pem -out public_key.pem

gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/proto/metric.proto 