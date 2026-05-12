BIN_DIR ?= bin
PRIMARY_BIN := foldersguard
ALIAS_BIN := fg

.PHONY: build webui macos-release frontend-build test clean

build:
	go build -o $(BIN_DIR)/$(PRIMARY_BIN) ./cmd/foldersguard
	ln -sf $(PRIMARY_BIN) $(BIN_DIR)/$(ALIAS_BIN)

webui:
	wails build

macos-release:
	./scripts/build-macos-release.sh

frontend-build:
	cd frontend && npm run build

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
	rm -rf frontend/dist/assets frontend/dist/index.html
