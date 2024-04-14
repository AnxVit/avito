PROJECTNAME=avito_banners

BINARY_API=./bin/api
BINARY_MIGRATE=./bin/migrate

export PG_USER=postgres
export PG_PASSWORD=1234
export PG_DBNAME=banner

## : 
## run: Launch application. Runs `go run` internally.
run:
	go run src/migrate/main.go up && go run src/avito/main.go
## : Create PG_PASSWORD ( make run PG_PASSWORD={} )

## : 
## build: Build application. Runs `docker build` internally.
build:
	go build -o ${BINARY_API} ./src/avito && go build -o ${BINARY_MIGRATE} ./src/migrate

## : 
## install: Launch binary files. Runs `sh -c` internally.
install:
	sh -c ./bin/migrate up && ./bin/api

## : 
## d.up: Up container. Runs `docker-compose up` internally.
d.up:
	docker-compose up

## : 
## d.down: Down container. Runs `docker-compose down` internally.
d.down:
	docker-compose down -v
	docker rmi avito_server

## : 
## d.up.build: Build container. Runs `docker-compose --build up` internally.
d.up.build:
	docker-compose --build up

## : 
## lint: Lauch golangci-lint. Runs `golangci-lint run` internally.
lint:
	golangci-lint run ./... --config=./.golangci.yml

## : 
## test: Launch test. Runs `go test` internally.
test:
	go test tests/banner_test.go

## : 
## dep: Download dependencies. Runs `go mod download` internally.
dep:
	go mod download

## : 
## clean: Clean build files. Runs `go clean` internally.
clean:
	go clean
	rm ${BINARY_API} 
	rm ${BINARY_MIGRATE}

help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo