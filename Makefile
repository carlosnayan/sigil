.PHONY: run build test

BIN_DIR := bin
BIN := $(BIN_DIR)/sigil

# Build the binary (run when sources change).
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) .

# Run the existing binary without rebuilding. Pass CLI args via ARGS, e.g.:
#   make run ARGS=init
#   make run ARGS='config --help'
# (Do not use "make run init" — Make treats init as a separate target.)
run:
	@test -f $(BIN) || (echo "$(BIN) missing — run: make build" >&2 && exit 1)
	./$(BIN) $(ARGS)

test:
	./scripts/test-ci.sh

