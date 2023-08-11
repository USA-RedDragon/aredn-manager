THIS_FILE := $(lastword $(MAKEFILE_LIST))

APP_NAME := aredn-manger
APP_PATH := github.com/USA-RedDragon/aredn-manger

GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)

fmt:
	gofmt -w $(GOFMT_FILES)

install-deps:
	@echo "--> Installing Golang dependencies"
	go get
	@echo "--> Done"

build-frontend:
	@echo "--> Installing JavaScript assets"
	@cd frontend && npm ci
	@echo "--> Building Vue application"
	@cd frontend && npm run build
	@echo "--> Done"

build: install-deps build-frontend
	@echo "--> Building"
	@go generate ./...
	@env CGO_ENABLED=0 go build -o bin/$(APP_NAME)
	@echo "--> Done"

# CI handles the frontend on its own so that
# we don't have to rebuild the frontend on each
# architecture
build-ci: install-deps
	@echo "--> Building"
	@go generate ./...
	@env CGO_ENABLED=0 go build -o bin/$(APP_NAME)
	@echo "--> Done"

run:
	@echo "--> Running"
	@go run .
	@echo "--> Done"

coverage:
	@echo "--> Running tests"
	@env CGO_ENABLED=0 go test -v -coverprofile=coverage.txt -coverpkg=./... -covermode=atomic ./...
	@echo "--> Done"

view-coverage:
	@echo "--> Viewing coverage"
	@go tool cover -html=coverage.txt
	@echo "--> Done"

lint:
	@echo "--> Linting"
	@golangci-lint run
	@echo "--> Done"

test:
	@echo "--> Running tests"
	@env CGO_ENABLED=0 go test -p 2 -v ./...
	@echo "--> Done"

benchmark:
	@echo "--> Running benchmarks"
	@env CGO_ENABLED=0 go test -run ^$ -benchmem -bench=. ./...
	@echo "--> Done"

frontend-unit-test:
	@cd frontend && npm run test:unit

frontend-e2e-test-electron:
	@cd frontend && npm ci
	@echo "--> Building Vue application"
	@cd frontend && env NODE_ENV=test npm run build
	@echo "--> Running end-to-end tests"
	@cd frontend && env NODE_ENV=test npm run test:e2e

frontend-e2e-test-chrome:
	@cd frontend && npm ci
	@echo "--> Building Vue application"
	@cd frontend && env NODE_ENV=test npm run build
	@echo "--> Running end-to-end tests"
	@cd frontend && env NODE_ENV=test npm run test:e2e:chrome

frontend-e2e-test-firefox:
	@cd frontend && npm ci
	@echo "--> Building Vue application"
	@cd frontend && env NODE_ENV=test npm run build
	@echo "--> Running end-to-end tests"
	@cd frontend && env NODE_ENV=test BROWSER=firefox npm run test:e2e:firefox
