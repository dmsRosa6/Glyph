BIN := bin/glyph
MAIN := ./main

build:
	go build -o $(BIN) $(MAIN)

run:
	go run $(MAIN)

run-bin: build
	./$(BIN)

debug: build
	dlv debug $(MAIN) --tty=$$(tty)

debug-attach:
	dlv attach $$(pgrep glyph)
