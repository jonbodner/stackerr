
.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: fmt
	go vet ./...

.PHONY:vet

test: vet
	go test -trimpath -v -cover ./...
.PHONY:test
