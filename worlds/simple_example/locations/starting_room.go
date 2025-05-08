package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example/items"
)

var LocationNameStarting = "Starting Location"

type StartingRoom struct {
	interfaces.BaseLocation
	doorUnlocked bool
	keyObtained  bool
}

func NewStartingRoom(world interfaces.WorldInterface) *StartingRoom {
	return &StartingRoom{
		doorUnlocked: false,
		keyObtained:  false,
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

func (l *StartingRoom) Examine(name string) string {
	if name == "door" {
		if l.doorUnlocked {
			return "The door is unlocked."
		}
		return "The door is locked."
	}
	return "I am not sure what you mean."
}

func (l *StartingRoom) UseItem(item interfaces.ItemInterface, targetName string) (string, bool) {
	if item.Name() == "key" && targetName == "door" {
		l.doorUnlocked = true
		return "You unlock the door with the key. The key breaks in the lock, but you manage to open the door before it does.", false
	}
	return "You can't use that item on that target.", true
}

func (l *StartingRoom) Go(name string) (string, *interfaces.LocationInterface) {
	if name == "north" {
		if l.doorUnlocked {
			return "You go through the door to the north.", l.BaseLocation.World.GetLocationByName("Front Steps")
		}
		return "The door is locked.", nil
	}
	if name == "south" {
		return "You are in a room, there is no way to go south.", nil
	}
	if name == "east" {
		return "You are in a room, there is no way to go east.", nil
	}
	if name == "west" {
		return "You are in a room, there is no way to go west.", nil
	}
	return "You can't go that way.", nil
}

func (l *StartingRoom) TakeItemByName(name string) (interfaces.ItemInterface, string) {
	if name == "key" && !l.keyObtained {
		l.keyObtained = true
		return items.FirstDoorKey{}, "You take the key."
	}
	return nil, "You can't take that item."
}

func (l *StartingRoom) Name() string {
	return LocationNameStarting
}
func (l *StartingRoom) Describe() string {
	s := "You are in a small room with a door to the north."

	if !l.doorUnlocked {
		s += " The door is locked."
	} else {
		s += " The door is unlocked."
	}

	if !l.keyObtained {
		s += " There is a key on the floor."
	}
	return s
}
