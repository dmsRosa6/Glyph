BIN  := bin/glyph
MAIN := ./main

.PHONY: build run run-bin debug test clean

build:
	go build -o $(BIN) $(MAIN)

run:
	go run $(MAIN)

run-bin: build
	./$(BIN)

debug:
	dlv debug $(MAIN)

test:
	go test ./...

clean:
	rm -f $(BIN)
