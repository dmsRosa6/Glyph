BIN := bin/glyph
MAIN := ./main

build:
	go build -o $(BIN) $(MAIN)

run:
	go run $(MAIN)

run-bin: build
	./$(BIN)
