package main

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/veilstream/psql-text-based-adventure/core/actions"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
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
		log.Printf("HTTP server listening on %s\n", addr)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			fmt.Printf("error listening to healthcheck: %v", err)
			return
		}
	}()

	listenAddr := "0.0.0.0:5432"
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", listenAddr, err)
	}
	log.Printf("listening on %s", listenAddr)

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

	world := simple_example.NewSimpleWorld(backend)
	engine := Engine{
		currentWorld: world,
		psqlBackend:  backend,
	}

	backend.Send(&pgproto3.AuthenticationOk{})
	backend.Send(&pgproto3.ParameterStatus{Name: "server_version", Value: "16.8"})
	backend.Send(&pgproto3.ParameterStatus{Name: "client_encoding", Value: "UTF8"})
	backend.Send(&pgproto3.BackendKeyData{ProcessID: 1234, SecretKey: 5678})
	backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
	helpAction := actions.ListCommandsAction{}
	helpAction.Execute(backend, engine.currentWorld)
	backend.Send(&pgproto3.ReadyForQuery{})
	err = backend.Flush()
	if err != nil {
		fmt.Printf("Error flushing psql backend: %v\n", err)
		return
	}

	for {
		msg, err := backend.Receive()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr())
				return
			}
			fmt.Printf("Error receiving message: %v\n", err)
			return
		}

		switch m := msg.(type) {
		case *pgproto3.Query:
			query := m.String
			fmt.Printf("Received query: %s\n", query)
			if strings.HasPrefix(query, "SELECT version()") {
				engine.psqlBackend.Send(&pgproto3.RowDescription{
					Fields: []pgproto3.FieldDescription{
						{Name: []byte("version")},
					},
				})
				engine.psqlBackend.Send(&pgproto3.DataRow{
					Values: [][]byte{[]byte("foo")},
				})
				engine.psqlBackend.Send(&pgproto3.CommandComplete{
					CommandTag: []byte("SELECT 1"),
				})
				engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
				err := engine.psqlBackend.Flush()
				if err != nil {
					fmt.Printf("Error flushing psql backend: %v\n", err)
					return
				}
				continue
			}
			if strings.HasPrefix(query, "SELECT ") && strings.HasSuffix(query, "as type;") {
				engine.psqlBackend.Send(&pgproto3.RowDescription{
					Fields: []pgproto3.FieldDescription{
						{Name: []byte("type")},
					},
				})
				engine.psqlBackend.Send(&pgproto3.DataRow{
					Values: [][]byte{
						[]byte("log"),
					},
				})
				engine.psqlBackend.Send(&pgproto3.CommandComplete{
					CommandTag: []byte("SELECT 1"),
				})
				engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
				err = engine.psqlBackend.Flush()
				if err != nil {
					fmt.Printf("Error flushing psql backend: %v\n", err)
					return
				}
				continue

			}
			if strings.HasPrefix(query, "SELECT ") {
				engine.psqlBackend.Send(&pgproto3.RowDescription{
					Fields: []pgproto3.FieldDescription{},
				})
				engine.psqlBackend.Send(&pgproto3.DataRow{
					Values: [][]byte{},
				})
				engine.psqlBackend.Send(&pgproto3.CommandComplete{
					CommandTag: []byte("SELECT 0"),
				})
				engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
				err = engine.psqlBackend.Flush()
				if err != nil {
					fmt.Printf("Error flushing psql backend: %v\n", err)
					return
				}
				continue
			}
			if strings.HasPrefix(query, "SET ") {
				engine.psqlBackend.Send(&pgproto3.CommandComplete{
					CommandTag: []byte("SET"),
				})
				engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
				err = engine.psqlBackend.Flush()
				if err != nil {
					fmt.Printf("Error flushing psql backend: %v\n", err)
					return
				}
				continue
			}
			query = strings.Replace(query, "\n", "", -1)
			query = strings.Replace(query, ";", "", -1)
			engine.handleQuery(query)

			if engine.gameOver {
				for _, row := range tombStone {
					engine.Say(row)
				}
				engine.psqlBackend.Send(&pgproto3.ErrorResponse{
					Severity: "FATAL",
					Message:  "You have died. Please restart the game.",
					Code:     "ADMIN_SHUTDOWN",
				})
				err := engine.psqlBackend.Flush()
				if err != nil {
					fmt.Printf("Error flushing psql backend: %v\n", err)
					return
				}
				return
			}
			engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
			err := engine.psqlBackend.Flush()
			if err != nil {
				fmt.Printf("Error flushing psql backend: %v\n", err)
				return
			}

		case *pgproto3.Terminate:
			fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr())
		case *pgproto3.Sync:
			fmt.Printf("Received Sync message\n")
			engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
			err := engine.psqlBackend.Flush()
			if err != nil {
				fmt.Printf("Error flushing psql backend: %v\n", err)
			}
		default:
			fmt.Printf("Unhandled message type: %T\n", m)
		}
	}
}

var tombStone = []string{
	"         _.-'\"\"-._    ",
	"       .-'        '-.    ",
	"     .'   Rest In     '.    ",
	"    /      Peace       \\    ",
	"   /--------------------\\    ",
	"  |  .----------------.  |    ",
	"  |  |   _________    |  |    ",
	"  |  |  /         \\   |  |    ",
	"  |  | |  o     o |   |  |    ",
	"  |  | |     ^    |   |  |    ",
	"  |  | |    ___   |   |  |    ",
	"  |  |  \\_________/   |  |    ",
	"  |  |                |  |    ",
	"  |  |  ???? - 2025   |  |    ",
	"  |  '----------------'  |    ",
	"  \\____________________/    ",
	"  /\\/\\/\\/\\/\\/\\/\\/\\/\\    ",
	"  / ___  ___  ___  __\\    ",
	"  /_/   \\/   \\/   \\/  /    ",
	" \\__________________/    ",
}
