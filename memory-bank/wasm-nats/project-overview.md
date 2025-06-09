# WASM-NATS Project Overview

## Description
WebAssembly (WASM) NATS server implementation that runs on wasmtime. Creates a NATS server proxy that runs in-process and accepts connections via WASM file descriptors.

## Key Components
- Embedded NATS server using `github.com/nats-io/nats-server/v2`
- TCP proxy between clients and NATS server
- Optional JetStream persistence support
- WASM file descriptor binding (default: fd 3)

## Build Commands
- `make` - Build and run
- `make build` - Compile to nats.wasm (GOOS=wasip1 GOARCH=wasm)
- `make run` - Run with wasmtime on 127.0.0.1:4222

## Configuration Options
- `-store <dir>`: Enable JetStream with persistence
- `-client-advertise <addr>`: Client advertise address
- `-wasm-fd <fd>`: WASM file descriptor
- `-ready-timeout <duration>`: Server ready timeout

## Requirements
- wasmtime runtime
- Go 1.24+