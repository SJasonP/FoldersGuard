BIN_DIR ?= bin
PRIMARY_BIN := foldersguard
ALIAS_BIN := fg

.PHONY: build test clean

build:
	go build -o $(BIN_DIR)/$(PRIMARY_BIN) ./cmd/foldersguard
	ln -sf $(PRIMARY_BIN) $(BIN_DIR)/$(ALIAS_BIN)

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
