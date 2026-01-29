.PHONY: run test build clean

run:
	go run cmd/api/main.go

test:
	go test ./... -v -cover

build:
	go build -o bin/calculator-api cmd/api/main.go

clean:
	rm -rf bin/
