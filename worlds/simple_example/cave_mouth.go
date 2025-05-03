package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

var LocationNameCaveMouth = "Cave Mouth"

type CaveMouth struct {
	world interfaces.WorldInterface
}

func NewCaveMouth(world interfaces.WorldInterface) *CaveMouth {
	return &CaveMouth{
		world: world,
	}
}

func (c CaveMouth) Examine(name string) string {
	if name == "cave" {
		return "A yawning cave entrance leads into impenetrable darkness. A chill wind whistles from the depths."
	}
	if name == "cave mouth" {
		return "A yawning cave entrance leads into impenetrable darkness. A chill wind whistles from the depths."
	}
	return "You can’t see that here."
}

func (c CaveMouth) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (c CaveMouth) Name() string { return LocationNameCaveMouth }
func (c CaveMouth) Describe() string {
	return "A yawning cave entrance leads into impenetrable darkness. A chill wind whistles from the depths. up to the garden, east to the go deeper into the darkness."
}

func (c CaveMouth) ListKnownItems() []interfaces.ItemInterface { return nil }
func (c CaveMouth) TakeItemByName(interfaces.WorldInterface, string) (interfaces.ItemInterface, string) {
	return nil, "Nothing here to take."
}

func (c CaveMouth) Go(world interfaces.WorldInterface, dir string) (bool, string, interfaces.LocationInterface) {
	switch dir {
	case "up":
		return true, "You climb back up to the garden.", world.GetLocationByName("Overgrown Garden")
	case "east":
		return true, "You peer into the depths, and venture forward.", world.GetLocationByName("Goblin Camp")
	default:
		return false, "You can't go that way.", nil
	}
}

func (c CaveMouth) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "Nothing happens.", true
}
