BIN := stress_test

.PHONY: build run clean

build:
	go build -o $(BIN) stress.go

run: build
	./$(BIN)

clean:
	rm -f $(BIN) stress_test.log
