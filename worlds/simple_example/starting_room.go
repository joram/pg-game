package simple_example

import (
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
func (f FirstDoorKey) Examine() string {
	return "This is a key to a door. it looks fragile, like it could break easily."
}

var LocationNameStarting = "Starting Location"

type StartingRoom struct {
	doorUnlocked bool
	keyObtained  bool
	world        interfaces.WorldInterface
}

func NewStartingRoom(world interfaces.WorldInterface) *StartingRoom {
	return &StartingRoom{
		doorUnlocked: false,
		world:        world,
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

func (l *StartingRoom) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (l *StartingRoom) UseItem(item interfaces.ItemInterface, targetName string) (string, bool) {
	if item.Name() == "key" && targetName == "door" {
		l.doorUnlocked = true
		return "You unlock the door with the key. The key breaks in the lock, but you manage to open the door before it does.", false
	}
	return "You can't use that item on that target.", true
}

func (l *StartingRoom) Go(world interfaces.WorldInterface, name string) (bool, string, interfaces.LocationInterface) {
	if name == "north" {
		if l.doorUnlocked {
			return true, "You go through the door to the north.", world.GetLocationByName("Front Steps")
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

func (l *StartingRoom) TakeItemByName(world interfaces.WorldInterface, name string) (interfaces.ItemInterface, string) {
	if name == "key" && !l.keyObtained {
		l.keyObtained = true
		return FirstDoorKey{}, "You take the key."
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
