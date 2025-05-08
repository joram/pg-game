package world

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

type World struct {
	Name             string
	StartingLocation *interfaces.LocationInterface
	CurrentLocation  *interfaces.LocationInterface
	Locations        map[string]*interfaces.LocationInterface
	Inventory        *interfaces.Inventory
	Over             bool
	PsqlBackend      *pgproto3.Backend
}

func (i World) HasInInventory(name string) bool {
	for _, item := range i.Inventory.Items {
		if item.Name() == name {
			return true
		}
	}
	return false
}

func (i World) Sayf(format string, a ...any) {
	i.PsqlBackend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  fmt.Sprintf(format, a...),
	})
}

func (i World) ReceiveMessage() string {
	response, err := i.PsqlBackend.Receive()
	if err != nil {
		i.Sayf("Error receiving message: %v", err)
		return ""
	}
	switch msg := response.(type) {
	case *pgproto3.Query:
		return msg.String
	default:
		i.Sayf("Received unexpected message type: %T", msg)
	}
	return ""
}

func NewWorld(name string) *World {
	inventory := interfaces.Inventory{
		Items: []interfaces.ItemInterface{},
	}
	return &World{
		Name:      name,
		Inventory: &inventory,
		Locations: make(map[string]*interfaces.LocationInterface),
		Over:      false,
	}
}

func (i World) GetLocationByName(s string) *interfaces.LocationInterface {
	for k, loc := range i.Locations {
		if k == s {
			return loc
		}
	}
	fmt.Printf("Location %s not found in world %s\n", s, i.Name)
	return nil
}

func (i World) Say(s string) {

}

func (i World) SetGameOver() {
	i.Over = true
}
