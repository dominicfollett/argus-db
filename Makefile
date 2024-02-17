ARCH := arm64
OUT_DIR := bin
BINARY := argusdb

.PHONY: all test build clean

all: test build

test:
	go test ./...

build:
	GOARCH=$(ARCH) go build -o $(OUT_DIR)/$(BINARY) ./main.go

clean:
	rm -rf $(OUT_DIR)
