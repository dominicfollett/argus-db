ARCH := arm64
OUT_DIR := bin
BINARY := argusdb

.PHONY: all test build clean dev

all: test build

test:
	go clean -testcache && go test -race ./...

build:
	GOARCH=$(ARCH) go build -o $(OUT_DIR)/$(BINARY) ./main.go

clean:
	rm -rf $(OUT_DIR)

dev:
	docker-compose up --build
