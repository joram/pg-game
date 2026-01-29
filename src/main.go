package main

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgproto3"
	"log"
	"net"
	"net/http"
)

func main() {
	go func() {
		http.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("OK"))
			if err != nil {
				fmt.Printf("error responding to healthcheck: %v", err)
				return
			}
		})

		addr := "0.0.0.0:80"
		log.Printf("HTTP %s\n", addr)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			fmt.Printf("error listening to healthcheck: %v", err)
			return
		}
	}()

	// WebSocket server (same container, connects to localhost:5432)
	go StartWebSocketServer("0.0.0.0:8080")

	listenAddr := "0.0.0.0:5432"
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", listenAddr, err)
	}
	log.Printf("PSQL %s", listenAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("New connection from %s\n", conn.RemoteAddr())
	backend := pgproto3.NewBackend(conn, conn)
	backend, err := upgradeBackendToTls(*backend, conn)
	if err != nil {
		fmt.Printf("Error upgrading to TLS: %v\n", err)
		return
	}
	_, err = backend.ReceiveStartupMessage()
	if err != nil {
		fmt.Printf("Error receiving startup message: %v\n", err)
		return
	}

	engine := NewEngine(backend)
	defer engine.Close()
	err = engine.Run()
	if err != nil {
		fmt.Printf("Error running engine: %v\n", err)
		return
	}
}

