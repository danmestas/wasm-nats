//go:build wasip1

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
	_ "unsafe"

	"github.com/nats-io/nats-server/v2/server"
)

func main() {
	var (
		storeDir        string
		readyTimeout    time.Duration
		wasmFd          int
		clientAdvertise string
	)

	flag.StringVar(&storeDir, "store", "", "store directory")
	flag.StringVar(&clientAdvertise, "client-advertise", "127.0.0.1:4222", "client advertise")
	flag.IntVar(&wasmFd, "wasm-fd", 3, "wasm file descriptor")
	flag.DurationVar(&readyTimeout, "ready-timeout", 15*time.Second, "ready timeout")

	flag.Parse()

	opts := &server.Options{
		ServerName:      "wash",
		DontListen:      true,
		ClientAdvertise: clientAdvertise,
		JetStream:       (storeDir != ""),
		StoreDir:        storeDir,
		NoSigs:          true,
		JetStreamDomain: "default",
	}

	s, err := server.NewServer(opts)
	if err != nil {
		panic(err)
	}

	s.ConfigureLogger()
	if !opts.JetStream {
		s.Logger().Warnf("Running without JetStream")
	}

	s.Start()

	if !s.ReadyForConnections(readyTimeout) {
		s.Shutdown()
		s.Logger().Fatalf("not ready after %s", readyTimeout)
		os.Exit(1)
	}

	s.Logger().Noticef("nats server ready, binding to wasmtime")
	socket, err := wasmListen(wasmFd)
	if err != nil {
		s.Logger().Fatalf("failed to create listener: %s", err)
		os.Exit(1)
	}
	defer socket.Close()

	// NOTE(lxf): we can't bind to signals here.
	// wasmtime will stop handling sockets correctly.
	ctx := context.Background()

	go func() {
		<-ctx.Done()
		s.Logger().Noticef("signal received, shutting down")
		s.Shutdown()
		socket.Close()
	}()

	s.Logger().Noticef("accepting connections")
	acceptLoop(socket, s)
	s.WaitForShutdown()
}

func acceptLoop(socket net.Listener, s *server.Server) {
	var wg sync.WaitGroup
	for {
		conn, err := socket.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			break
		}

		s.Logger().Noticef("accepted connection from %s", conn.RemoteAddr())

		wg.Add(1)
		go func() {
			defer wg.Done()
			proxyConnection(conn, s)
		}()
	}
	wg.Wait()
}

func proxyConnection(conn net.Conn, s *server.Server) {
	defer conn.Close()

	// Create a new connection to the NATS server
	natsConn, err := s.InProcessConn()
	if err != nil {
		fmt.Println("Error creating NATS connection:", err)
		return
	}
	defer natsConn.Close()

	// Start a goroutine to copy data from the NATS connection to the client
	go func() {
		io.Copy(conn, natsConn)
	}()

	// Copy data from the client to the NATS connection
	io.Copy(natsConn, conn)
}

func wasmListen(wasmFd int) (net.Listener, error) {
	if err := syscall.SetNonblock(int(wasmFd), true); err != nil {
		return nil, err
	}

	file := os.NewFile(uintptr(wasmFd), "wasm")
	if file == nil {
		return nil, errors.New("invalid wasm file descriptor")
	}
	defer file.Close()

	listener, err := net.FileListener(file)
	if err != nil {
		return nil, err
	}

	return listener, nil
}
