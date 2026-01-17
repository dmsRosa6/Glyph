BIN  := bin/glyph
MAIN := ./main

.PHONY: build run run-bin debug debug-attach clean

build:
	go build -o $(BIN) $(MAIN)

run:
	go run $(MAIN)

run-bin: build
	./$(BIN)
	
clean:
	rm -f $(BIN)
