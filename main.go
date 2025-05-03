package main

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/veilstream/psql-text-based-adventure/core/actions"
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type Engine struct {
	Worlds       []interfaces.WorldInterface
	currentWorld *interfaces.WorldInterface
	psqlBackend  *pgproto3.Backend
}

func main() {
	go func() {
		http.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("OK"))
			if err != nil {
				fmt.Printf("error responding to healthcheck: %v", err)
				return
			}
			log.Printf("Health check OK\n")
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
	world := simple_example.NewSimpleWorld(&interfaces.Inventory{})
	engine := Engine{
		currentWorld: &world,
		Worlds:       []interfaces.WorldInterface{world},
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
			query = strings.Replace(query, "\n", "", -1)
			query = strings.Replace(query, ";", "", -1)
			engine.handleQuery(query)

			if strings.HasPrefix(query, "SELECT ") {
				engine.psqlBackend.Send(&pgproto3.RowDescription{
					Fields: []pgproto3.FieldDescription{},
				})
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

func (engine *Engine) handleQuery(query string) {
	fmt.Printf("Handling query: %s\n", query)

	if strings.HasPrefix(query, "look") {
		engine.Look()
		return
	}

	if strings.HasPrefix(query, "commands") {
		engine.Commands()
		return
	}

	if strings.HasPrefix(query, "take") {
		engine.Take(query)
		return
	}

	if strings.HasPrefix(query, "inventory") {
		engine.Inventory()
		return
	}

	if strings.HasPrefix(query, "go") {
		engine.Go(query)
		return
	}

	if strings.HasPrefix(query, "use") {
		engine.Use(query)
		return
	}

	// Unknown command
	engine.psqlBackend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  fmt.Sprintf("I don't know how to '%s'", query),
	})
}

func (engine *Engine) Say(msg string) {
	engine.psqlBackend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  msg,
	})
}

func (engine *Engine) Sayf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	engine.psqlBackend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  s,
	})
}

func (engine *Engine) Look() {
	action := actions.LookAction{}
	action.Execute(engine.psqlBackend, engine.currentWorld)
}

func (engine *Engine) Commands() {
	action := actions.ListCommandsAction{}
	action.Execute(engine.psqlBackend, engine.currentWorld)
}

func (engine *Engine) Take(query string) {
	itemName := strings.TrimSpace(strings.TrimPrefix(query, "take"))
	item, msg := engine.currentWorld.CurrentLocation.TakeItemByName(*engine.currentWorld, itemName)
	engine.Say(msg)
	if item == nil {
		return
	}
	engine.currentWorld.Inventory.AddItem(item)
	engine.Sayf("You now have a %s", item.Name())
}

func (engine *Engine) Inventory() {
	items := engine.currentWorld.Inventory.ListItems()
	if len(items) == 0 {
		engine.Say("Your inventory is empty.")
		return
	}
	inventoryList := "You have the following items in your inventory:"
	for _, item := range items {
		inventoryList = fmt.Sprintf("%s, %s", inventoryList, item.Name())
	}
	engine.Say(inventoryList)

}

func (engine *Engine) Go(query string) {
	locationName := strings.TrimSpace(strings.TrimPrefix(query, "go"))
	fmt.Printf("attempting to go to '%s'\n", locationName)
	success, msg, New := engine.currentWorld.CurrentLocation.Go(*engine.currentWorld, locationName)
	if !success {
		engine.Say(msg)
		return
	}
	engine.Sayf("You go %s", locationName)
	engine.currentWorld.CurrentLocation = New
}

func (engine *Engine) Use(query string) {
	removedUseStr := strings.TrimSpace(strings.TrimPrefix(query, "use"))
	parts := strings.Split(removedUseStr, " on ")
	if len(parts) != 2 {
		engine.Say("Usage: use <item> on <target>")
		return
	}
	itemName := strings.TrimSpace(parts[0])
	targetName := strings.TrimSpace(parts[1])
	item := engine.currentWorld.Inventory.RemoveItem(itemName)
	if item == nil {
		engine.Sayf("You do not have a  '%s' in your inventory", itemName)
		return
	}

	msg, keep := engine.currentWorld.CurrentLocation.UseItem(item, targetName)
	engine.Say(msg)
	if keep {
		engine.currentWorld.Inventory.AddItem(item)
	}
}

func (engine *Engine) Examine(query string) {
	itemName := strings.TrimSpace(strings.TrimPrefix(query, "examine"))
	// First try inventory
	for _, item := range engine.currentWorld.Inventory.ListItems() {
		if item.Name() == itemName {
			engine.Say(item.Examine())
			return
		}
	}
	// Then try the location
	engine.Say(engine.currentWorld.CurrentLocation.Examine(itemName))
}

func (engine *Engine) TalkTo(query string) {
	name := strings.TrimSpace(strings.TrimPrefix(query, "talk to"))
	response := engine.currentWorld.CurrentLocation.TalkTo(name)
	if response == "" {
		engine.Sayf("There's no one named '%s' here to talk to.", name)
		return
	}
	engine.Say(response)
}
