.PHONY: run build lint lint-fix test swag migrate-up migrate-down migrate-status migrate-create docker-build docker-up docker-down docker-rebuild install-lint-deps

## run: запуск приложения локально
run:
	go run ./cmd/app/...

## build: сборка бинаря
build:
	go build -o bin/app ./cmd/app/...

## test: запуск тестов
test:
	go test -race -count=1 ./...

## swag: генерация swagger-документации
swag:
	swag init -g cmd/app/main.go -o docs/swagger

## migrate-up: применить все миграции
migrate-up:
	go run ./cmd/cli migrate:up

## migrate-down: откатить последнюю миграцию
migrate-down:
	go run ./cmd/cli migrate:down

## migrate-status: показать статус миграций
migrate-status:
	go run ./cmd/cli migrate:status

## migrate-create: создать новый файл миграции (запросит имя)
migrate-create:
	@read -p "Migration name: " name; \
	go run ./cmd/cli migrate:create $$name

## docker-build: сборка docker-образа
docker-build:
	docker build -t golang-boilerplate:latest .

## docker-up: запуск через docker-compose
docker-up:
	docker-compose up -d

## docker-down: остановка docker-compose
docker-down:
	docker-compose down

## docker-rebuild: пересборка и перезапуск контейнеров
docker-rebuild:
	docker-compose down && docker-compose up -d --build

## install-lint-deps: установить golangci-lint (если не установлен)
install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

## lint: запустить линтер (установит golangci-lint при необходимости)
lint: install-lint-deps
	golangci-lint run ./...

## lint-fix: запустить линтер с автоисправлением
lint-fix: install-lint-deps
	golangci-lint run ./... --fix

## help: список доступных команд
help:
	@grep -E '^## ' Makefile | sed 's/## //'