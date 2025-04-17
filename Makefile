all: build run

.PHONY: build
build:
	GOOS=wasip1 GOARCH=wasm go build -o nats.wasm

.PHONY: run
run:
	wasmtime run \
		-S preview2=n \
		-S tcplisten=127.0.0.1:4222 \
	./nats.wasm
