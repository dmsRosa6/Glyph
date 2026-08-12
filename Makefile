BIN  := bin/glyph
MAIN := ./main
LOG_FOLDER := logs

.PHONY: build run run-bin debug test clean

build:
	go build -o $(BIN) $(MAIN)

run:
	go run $(MAIN)

clear-run:
	rm -rf $(LOG_FOLDER)
	go run $(MAIN)

run-bin: build
	./$(BIN)

debug:
	dlv debug $(MAIN)

test:
	go test ./...

clean-build:
	rm -f $(BIN)
