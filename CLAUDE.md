# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture

This is a WebAssembly (WASM) NATS server implementation that runs on wasmtime. The project creates a NATS server proxy that:

- Runs a NATS server in-process using `github.com/nats-io/nats-server/v2`
- Accepts connections via a WASM file descriptor (default: fd 3)
- Proxies TCP connections between clients and the embedded NATS server
- Supports optional JetStream persistence with a configurable store directory

The main components:
- `main()`: Sets up server configuration, starts NATS server, and binds to wasmtime socket
- `acceptLoop()`: Handles incoming connections concurrently
- `proxyConnection()`: Bidirectional proxy between client and NATS server
- `wasmListen()`: Creates listener from WASM file descriptor

## Development Commands

**Build and run:**
```bash
make
```

**Build only:**
```bash
make build
# Compiles to nats.wasm targeting GOOS=wasip1 GOARCH=wasm
```

**Run only:**
```bash
make run
# Runs with wasmtime, listening on 127.0.0.1:4222
```

## Requirements

- wasmtime
- Go 1.24+

## Key Configuration Options

- `-store <dir>`: Enable JetStream with persistence directory
- `-client-advertise <addr>`: Client advertise address (default: 127.0.0.1:4222)
- `-wasm-fd <fd>`: WASM file descriptor (default: 3)
- `-ready-timeout <duration>`: Server ready timeout (default: 15s)
