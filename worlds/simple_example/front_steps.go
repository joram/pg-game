package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

type ItemLantern struct {
	attached bool
}

func (i ItemLantern) Examine() string {
	return "The lantern is small, and it is casting a small radius of light around you."
}

func (i ItemLantern) Name() string {
	return "lantern"
}

func (i ItemLantern) Description() string {
	if i.attached {
		return "A small lantern, it is hanging on the wall beside the door. It is casting a small radius of light around you."
	}
	return "a small lantern"
}

func (i ItemLantern) Use(target interfaces.ItemInterface) string {
	return "You can't use that item on that target."
}

type FrontSteps struct {
	world interfaces.WorldInterface
}

func NewFrontSteps(world interfaces.WorldInterface) *FrontSteps {
	return &FrontSteps{
		world: world,
	}
}

func (f FrontSteps) Examine(name string) string {
	if name == "lantern" {
		return lantern.Examine()
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

var lantern = ItemLantern{attached: true}

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

func (f FrontSteps) ListKnownItems() []interfaces.ItemInterface {
	return []interfaces.ItemInterface{
		lantern,
	}
}

func (f FrontSteps) TakeItemByName(world interfaces.WorldInterface, s string) (interfaces.ItemInterface, string) {
	if s == "lantern" && lantern.attached {
		return nil, "you can't remove it, it's attached to the wall."
	}
	if s == "lantern" {
		lantern.attached = false
		return lantern, "You take the lantern."
	}
	return nil, "You can't take that item."
}

func (f FrontSteps) Go(world interfaces.WorldInterface, name string) (bool, string, interfaces.LocationInterface) {
	if name == "north" {
		if world.Inventory.HaveItem("lantern") {
			return true, "You walk into the woods, the path is dark but you can see a little bit ahead of you.", world.GetLocationByName("Woods Entrance")
		}
		return false, "You take a few steps, and realise you will get lost without a light. you go back to the house.", nil
	}
	if name == "south" {
		return true, "You walk to the front door of the house.", world.GetLocationByName("Starting Location")
	}
	if name == "east" {
		return true, "You walk to the back of the house.", world.Locations["Beside House"]
	}
	if name == "west" {
		return false, "You get a bad feeling about this place, and decide to turn back.", nil
	}
	return false, "You can't go that way.", nil
}

func (f FrontSteps) UseItem(item interfaces.ItemInterface, targetName string) (string, bool) {
	if item.Name() == "screwdriver" && targetName == "lantern" {
		if lantern.attached {
			lantern.attached = false
			return "You use the screwdriver to remove the lantern from the sconce. The lantern is now loosely hanging from the wall.", true
		}
	}
	return "You can't use that item on that target.", true
}
