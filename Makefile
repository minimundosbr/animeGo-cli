.PHONY: build test race vet fmt

build:
	go build -o bin/animego ./cmd/animego

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .