

all: build

.PHONY: build

build:
	go build -o build/go-semantic-release-temp ./cmd/go-semantic-release/

lint:
	golangci-lint run --print-issued-lines=false --fix ./...

test:
	go test --coverprofile coverage.out -v -parallel 20 ./...
