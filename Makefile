BIN := bin/glyph
BASE := ./examples/
EXAMPLE ?= clock-window
MAIN = $(BASE)$(EXAMPLE)/main.go
LOG_FOLDER := logs

.PHONY: build run debug test clean

build:
	go build -o $(BIN) $(MAIN)

run:
	go run $(MAIN)

clear-run:
	rm -rf $(LOG_FOLDER)
	go run $(MAIN)

debug:
	dlv debug $(MAIN)

test:
	go test ./...

clean-build:
	rm -f $(BIN)