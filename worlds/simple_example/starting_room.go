package simple_example

import (
	"fmt"
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

type FirstDoorKey struct {
}

func (f FirstDoorKey) Name() string {
	return "key"
}

func (f FirstDoorKey) Description() string {
	return "This is a key to a door."
}

var LocationNameStarting = "Starting Location"

type LocationStartingRoom struct {
	items        []interfaces.ItemInterface
	doorUnlocked bool
}

func (l *LocationStartingRoom) UseItem(item interfaces.ItemInterface, targetName string) string {
	if item == firstDoorKey && targetName == "door" {
		l.doorUnlocked = true
		return "You unlock the door with the key."
	}
	return "You can't use that item on that target."
}

func (l *LocationStartingRoom) Go(name string) (bool, string, interfaces.LocationInterface) {
	if name == "north" {
		if l.doorUnlocked {
			return true, "You go through the door to the north.", frontSteps
		}
		return false, "The door is locked.", nil
	}
	if name == "south" {
		return true, "You are in a room, there is no way to go south.", nil
	}
	if name == "east" {
		return true, "You are in a room, there is no way to go east.", nil
	}
	if name == "west" {
		return true, "You are in a room, there is no way to go west.", nil
	}
	return false, "You can't go that way.", nil
}

func (l *LocationStartingRoom) TakeItemByName(name string) interfaces.ItemInterface {
	for i, item := range l.items {
		if item.Name() == name {
			takenItem := l.items[i]
			l.items = append(l.items[:i], l.items[i+1:]...)
			return takenItem
		}
	}
	return nil
}

func (l *LocationStartingRoom) ListKnownItems() []interfaces.ItemInterface {
	return l.items
}

func (l *LocationStartingRoom) Name() string {
	return LocationNameStarting
}
func (l *LocationStartingRoom) Describe() string {
	s := "You are in a small room with a door to the north."

	if len(l.ListKnownItems()) > 0 {
		s += "\nYou see the following items:"
		for _, item := range l.ListKnownItems() {
			s = fmt.Sprintf("%s %s", s, item.Name())
		}
	}
	return s
}

var firstDoorKey = FirstDoorKey{}
var start = LocationStartingRoom{
	items: []interfaces.ItemInterface{
		firstDoorKey,
	},
}
