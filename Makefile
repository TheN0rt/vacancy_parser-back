.PHONY: build
build:
	go build -v ./cmd/vacancy-parser

.PHONY: test
test:
	go test -v -race -timeout 30s ./ ...

.DEFAULT_GOAL := build
