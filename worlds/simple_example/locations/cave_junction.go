package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

type CaveJunction struct {
	interfaces.BaseLocation
}

func (c *CaveJunction) UseItem(item interfaces.ItemInterface, targetName string) (string, bool) {
	return "Nothing happens.", true
}

var LocationNameCaveJunction = "Cave Junction"

func NewCaveJunction(world interfaces.WorldInterface) interfaces.LocationInterface {
	return &CaveJunction{
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

func (c *CaveJunction) Examine(name string) string {
	if name == "cave" {
		return "A dark cave system stretches out in all directions. The air is damp and musty."
	}
	if name == "junction" {
		return "A dark cave system stretches out in all directions. The air is damp and musty."
	}
	return "You can’t see that here."
}

func (c *CaveJunction) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (c *CaveJunction) Name() string { return LocationNameCaveJunction }
func (c *CaveJunction) Describe() string {
	return "A dark cave system stretches out in all directions. The air is damp and musty."
}

func (c *CaveJunction) TakeItemByName(string) (interfaces.ItemInterface, string) {
	return nil, "Nothing here to take."
}

func (c *CaveJunction) Go(dir string) (string, *interfaces.LocationInterface) {
	switch dir {
	case "north":
		return "You head deeper into the cave system.", c.BaseLocation.World.GetLocationByName(LocationNameCaveExit)
	case "south":
		return "You head back to the goblin camp.", c.BaseLocation.World.GetLocationByName(LocationNameGoblinCamp)
	case "west":
		return "You venture further into the darkness. After wandering for hours, you arrive back where you started, at the same cave junction.", c.BaseLocation.World.GetLocationByName(LocationNameCaveJunction)
	case "east":
		return "You venture further into the darkness. After wandering for hours, you arrive back where you started, or could this be a different cave junction?.", c.BaseLocation.World.GetLocationByName(LocationNameCaveJunction)
	default:
		return "You can't go that way.", nil
	}
}
