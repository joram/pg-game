package main

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/veilstream/psql-text-based-adventure/core/actions"
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds"
	"io"
	"log"
	"net"
	"strings"
)

type PsqlTextBasedAdventure struct {
	Worlds       []interfaces.World
	currentWorld *interfaces.World
	inventory    Inventory
	psqlBackend  *pgproto3.Backend
}

func main() {
	engine := PsqlTextBasedAdventure{
		Worlds:       worlds.Worlds,
		currentWorld: &worlds.Worlds[0],
	}
	engine.Start()
}

func (engine *PsqlTextBasedAdventure) Start() {
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
		go engine.handleConnection(conn)
	}
}

func (engine *PsqlTextBasedAdventure) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("New connection from %s\n", conn.RemoteAddr())
	backend := pgproto3.NewBackend(conn, conn)
	backend, err := upgradeBackendToTls(*backend, conn)
	if err != nil {
		panic(err)
	}
	_, err = backend.ReceiveStartupMessage()
	if err != nil {
		panic(err)
	}
	engine.psqlBackend = backend

	backend.Send(&pgproto3.AuthenticationOk{})
	backend.Send(&pgproto3.ParameterStatus{Name: "server_version", Value: "14.0"})
	backend.Send(&pgproto3.ParameterStatus{Name: "client_encoding", Value: "UTF8"})
	backend.Send(&pgproto3.BackendKeyData{ProcessID: 1234, SecretKey: 5678})
	helpAction := actions.ListCommandsAction{}
	helpAction.Execute(backend, engine.currentWorld)
	backend.Send(&pgproto3.ReadyForQuery{})
	err = backend.Flush()
	if err != nil {
		panic(err)
	}

	for {
		msg, err := backend.Receive()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr())
				return
			}
			panic(err)
		}

		switch m := msg.(type) {
		case *pgproto3.Query:
			query := m.String
			query = strings.Replace(query, "\n", "", -1)
			query = strings.Replace(query, ";", "", -1)
			engine.handleQuery(query)
			engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
			err := engine.psqlBackend.Flush()
			if err != nil {
				panic(err)
			}

		case *pgproto3.Terminate:
			fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr())

		default:
			fmt.Printf("Unhandled message type: %T\n", m)
		}
	}
}

func (engine *PsqlTextBasedAdventure) handleQuery(query string) {
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

func (engine *PsqlTextBasedAdventure) Look() {
	action := actions.LookAction{}
	action.Execute(engine.psqlBackend, engine.currentWorld)
}

func (engine *PsqlTextBasedAdventure) Commands() {
	action := actions.ListCommandsAction{}
	action.Execute(engine.psqlBackend, engine.currentWorld)
}

func (engine *PsqlTextBasedAdventure) Take(query string) {
	itemName := strings.TrimSpace(strings.TrimPrefix(query, "take"))
	item := engine.currentWorld.CurrentLocation.TakeItemByName(itemName)
	if item == nil {
		engine.psqlBackend.Send(&pgproto3.NoticeResponse{
			Severity: "",
			Message:  fmt.Sprintf("Item '%s' not found", itemName),
		})
		return
	}
	engine.inventory.AddItem(item)
	engine.psqlBackend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  fmt.Sprintf("You now have a %s", item.Name()),
	})
}

func (engine *PsqlTextBasedAdventure) Inventory() {
	items := engine.inventory.ListItems()
	if len(items) == 0 {
		engine.psqlBackend.Send(&pgproto3.NoticeResponse{
			Severity: "",
			Message:  "Your inventory is empty.",
		})
		return
	}
	inventoryList := "You have the following items in your inventory:"
	for _, item := range items {
		inventoryList = fmt.Sprintf("%s %s", inventoryList, item.Name())
	}
	engine.psqlBackend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  inventoryList,
	})
}

func (engine *PsqlTextBasedAdventure) Go(query string) {
	locationName := strings.TrimSpace(strings.TrimPrefix(query, "go"))
	fmt.Printf("attempting to go to '%s'\n", locationName)
	success, msg, newLocation := engine.currentWorld.CurrentLocation.Go(locationName)
	if !success {
		engine.psqlBackend.Send(&pgproto3.NoticeResponse{
			Severity: "",
			Message:  msg,
		})
		return
	}
	engine.psqlBackend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  fmt.Sprintf("You go to %s", locationName),
	})
	engine.currentWorld.CurrentLocation = newLocation
}

func (engine *PsqlTextBasedAdventure) Use(query string) {
	removedUseStr := strings.TrimSpace(strings.TrimPrefix(query, "use"))
	parts := strings.Split(removedUseStr, " on ")
	if len(parts) != 2 {
		engine.psqlBackend.Send(&pgproto3.NoticeResponse{
			Severity: "",
			Message:  "Usage: use <item> on <target>",
		})
		return
	}
	itemName := strings.TrimSpace(parts[0])
	targetName := strings.TrimSpace(parts[1])
	item := engine.inventory.RemoveItem(itemName)
	if item == nil {
		engine.psqlBackend.Send(&pgproto3.NoticeResponse{
			Severity: "",
			Message:  fmt.Sprintf("You do not have a  '%s' in your inventory", itemName),
		})
		return
	}

	msg := engine.currentWorld.CurrentLocation.UseItem(item, targetName)
	engine.psqlBackend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  msg,
	})

}
