BIN := bin/lrcx
CMD := ./cmd/lrcx

.PHONY: build test lint install clean

build:
	go build -o $(BIN) $(CMD)

test:
	go test ./...

lint:
	go vet ./...
	golangci-lint run

install:
	go install $(CMD)

clean:
	rm -f $(BIN)
