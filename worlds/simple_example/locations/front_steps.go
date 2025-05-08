package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example/items"
)

type FrontSteps struct {
	interfaces.BaseLocation
	lantern items.ItemLantern
}

func NewFrontSteps(world interfaces.WorldInterface) interfaces.LocationInterface {
	return &FrontSteps{
		lantern: items.ItemLantern{},
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

func (f FrontSteps) Examine(name string) string {
	if name == "lantern" {
		return f.lantern.Examine()
	}
	if name == "door" {
		return "The door is locked."
	}
	if name == "path" {
		return "A small path leads away from the house."
	}
	if name == "woods" {
		return "A small path leads into the woods."
	}
	return "I am not sure what you mean."
}

func (f FrontSteps) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (f FrontSteps) Name() string {
	return "Front Steps"
}

func (f FrontSteps) Describe() string {
	return "You are standing on the front steps of a small house. It's dark out, the moon is obscured by clouds, " +
		"and the wind is howling. You can hear the sound of wind in the trees.\n" +
		"There's a small radius of light around you, it is cast by a lantern hanging on the wall beside the door. \n" +
		"You can see a small path leading away from the house.\n" +
		"To the north, you can see a small path leading into the woods.\n" +
		"To the south, behind you, there is the door you just exited.\n" +
		"To the east, you can see a small path leading to the back of the house.\n" +
		"To the west, darkness."
}

func (f FrontSteps) TakeItemByName(s string) (interfaces.ItemInterface, string) {
	if s == "lantern" && f.lantern.Attached {
		return nil, "you can't remove it, it's attached to the wall."
	}
	if s == "lantern" {
		f.lantern.Attached = false
		return f.lantern, "You take the lantern."
	}
	return nil, "You can't take that item."
}

func (f FrontSteps) Go(name string) (string, *interfaces.LocationInterface) {
	if name == "north" {
		if f.BaseLocation.World.HasInInventory("lantern") {
			return "You walk into the woods, the path is dark but you can see a little bit ahead of you.", f.BaseLocation.World.GetLocationByName("Woods Entrance")
		}
		return "You take a few steps, and realise you will get lost without a light. you go back to the house.", nil
	}
	if name == "south" {
		return "You walk to the front door of the house.", f.BaseLocation.World.GetLocationByName("Starting Location")
	}
	if name == "east" {
		return "You walk to the back of the house.", f.BaseLocation.World.GetLocationByName("Beside House")
	}
	if name == "west" {
		return "You get a bad feeling about this place, and decide to turn back.", nil
	}
	return "You can't go that way.", nil
}

func (f FrontSteps) UseItem(item interfaces.ItemInterface, targetName string) (string, bool) {
	if item.Name() == "screwdriver" && targetName == "lantern" {
		if f.lantern.Attached {
			f.lantern.Attached = false
			return "You use the screwdriver to remove the lantern from the sconce. The lantern is now loosely hanging from the wall.", true
		}
	}
	return "You can't use that item on that target.", true
}
