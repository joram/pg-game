package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

type FrontSteps struct {
}

func (f FrontSteps) Name() string {
	return "Front Steps"
}

func (f FrontSteps) Describe() string {
	return "You are standing on the front steps of a small house. It's dark out, the moon is obscured by clouds, and the wind is howling. You can hear the sound of wind in the trees.\nThere's a small radius of light around you, and you can see a small path leading away from the house.\nTo the north, you can see a small path leading into the woods. To the south, you can see a small path leading to the front door of the house.\nTo the east, you can see a small path leading to the back of the house.\nTo the west, darkness."
}

func (f FrontSteps) ListKnownItems() []interfaces.ItemInterface {
	return []interfaces.ItemInterface{}
}

func (f FrontSteps) TakeItemByName(s string) interfaces.ItemInterface {
	return nil
}

func (f FrontSteps) Go(name string) (bool, string, interfaces.LocationInterface) {
	if name == "north" {
		return true, "You walk into the woods.", nil
	}
	if name == "south" {
		return true, "You walk to the front door of the house.", nil
	}
	if name == "east" {
		return true, "You walk to the back of the house.", nil
	}
	if name == "west" {
		return true, "You walk to the street.", nil
	}
	return false, "You can't go that way.", nil
}

func (f FrontSteps) UseItem(item interfaces.ItemInterface, targetName string) string {
	return "You can't use that item on that target."
}

var frontSteps = FrontSteps{}
