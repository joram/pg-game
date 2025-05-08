package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

var LocationNameWoodsEntrance = "Woods Entrance"

type WoodsEntrance struct {
	interfaces.BaseLocation
}

func NewWoodsEntrance(world interfaces.WorldInterface) *WoodsEntrance {
	return &WoodsEntrance{
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

func (w WoodsEntrance) Examine(name string) string {
	if name == "woods" {
		return "Tall firs loom ahead, their branches choking out the moonlight. A narrow path disappears into the darkness to the north."
	}
	if name == "path" {
		return "A narrow path disappears into the darkness to the north."
	}
	if name == "fir trees" {
		return "Tall firs loom ahead, their branches choking out the moonlight."
	}
	return "You can’t see that here."
}

func (w WoodsEntrance) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (w WoodsEntrance) Name() string { return LocationNameWoodsEntrance }

func (w WoodsEntrance) Describe() string {
	return "Tall firs loom ahead, their branches choking out the moonlight. A narrow path disappears into the darkness to the north. To the south, the house is visible."
}

func (w WoodsEntrance) TakeItemByName(string) (interfaces.ItemInterface, string) {
	return nil, "There is nothing to take."
}

func (w WoodsEntrance) Go(dir string) (string, *interfaces.LocationInterface) {
	switch dir {
	case "south":
		return "You head back to the front steps.", w.BaseLocation.World.GetLocationByName("Front Steps")
	case "north":
		// The engine should enforce the player has a lantern; we allow for now.
		return "You follow the dark path deeper into the woods.", w.BaseLocation.World.GetLocationByName("Forked Path")
	default:
		return "You can't go that way.", nil
	}
}

func (w WoodsEntrance) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "There's nothing here that responds to that item.", true
}
