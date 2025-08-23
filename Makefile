APP_NAME := demoserv
CMD_DIR := ./cmd
BIN_DIR := ./bin

.PHONY: build run clean test

## Собрать бинарник
build:
	@echo ">>> Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)

## Запустить сервер
run: build
	@echo ">>> Running $(APP_NAME)..."
	@$(BIN_DIR)/$(APP_NAME)

## Запустить тесты
test:
	@go test ./... -v

## Очистить бинарники
clean:
	@rm -rf $(BIN_DIR)
