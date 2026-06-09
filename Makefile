APP=pilot
BIN=bin/$(APP)

.PHONY: test build run clean

test:
	go test ./...

build:
	mkdir -p bin
	go build -o $(BIN) ./cmd/pilot

run:
	go run ./cmd/pilot

clean:
	rm -rf bin
