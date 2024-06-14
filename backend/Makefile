include .env

run:
	go run cmd/app/main.go -cfg configs/local

build-models:
	docker run --rm --user $(shell id -u):$(shell id -g) -e GOPATH=$(go env GOPATH):/go -v ${HOME}:${HOME} -w $(shell pwd) quay.io/goswagger/swagger generate model --spec=../swagger.yaml
	go mod tidy

postgres.start:
	if [ ! "$(shell docker ps -q -f name=bdd-postgres)" ]; then \
        if [ "$(shell docker ps -aq -f status=exited -f name=bdd-postgres)" ]; then \
            docker rm bdd-postgres; \
        fi; \
		docker run --restart unless-stopped -d -p 1339:5432 -e POSTGRES_PASSWORD=1337 --name bdd-postgres postgres:13; \
        sleep 5; \
    fi;
	-docker exec bdd-postgres psql -U postgres -c "create user bdd password '1337'"
	-docker exec bdd-postgres psql -U postgres -c "create database bdd"
	-docker exec bdd-postgres psql -U postgres -c "grant all privileges on database bdd to bdd"
	-docker exec bdd-postgres psql -U postgres -c "create user public_bdd password '1'"
	-docker exec bdd-postgres psql -U postgres -c "GRANT CONNECT ON DATABASE bdd TO public_bdd;"
	-docker exec bdd-postgres psql -U postgres -c "GRANT USAGE ON SCHEMA public TO public_bdd;"
	-docker exec bdd-postgres psql -U postgres -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO public_bdd;"


postgres.stop:
	docker stop bdd-postgres
	docker rm bdd-postgres

swagger.start:
	docker run --restart unless-stopped -d --name bdd-swagger -p 9804:8080 -e SWAGGER_JSON=/specs/swagger.yaml -v $(shell pwd)/../:/specs swaggerapi/swagger-ui

swagger.stop:
	docker stop bdd-swagger
	docker rm bdd-swagger

migrate.up:
	goose -dir migrations postgres "host=localhost port=1339 user=bdd password=1337 dbname=bdd sslmode=disable" up

migrate.down:
	goose -dir migrations postgres "host=localhost port=1339 user=bdd password=1337 dbname=bdd sslmode=disable" down

migrate.status:
	goose -dir migrations postgres "host=localhost port=1339 user=bdd password=1337 dbname=bdd sslmode=disable" status

tests:
	go test -v ./integrational_tests/... 

# To use one check one test like `make tt test=TestSuiteRun/TestTagsList`
tt:
	go test -v ./integrational_tests/... -run $$test