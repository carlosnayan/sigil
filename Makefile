.PHONY: run build

BIN_DIR := bin
BIN := $(BIN_DIR)/sigil

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) .

run: build
	./$(BIN) $(ARGS)

test:
	./scripts/test-ci.sh