package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

var LocationNameCaveMouth = "Tomb Entrance"

type TombEntrance struct {
	interfaces.BaseLocation
}

func NewTombEntrance(world interfaces.WorldInterface) *TombEntrance {
	return &TombEntrance{
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

func (c TombEntrance) Examine(name string) string {
	if name == "cave" {
		return "A yawning cave entrance leads into impenetrable darkness. A chill wind whistles from the depths."
	}
	if name == "cave mouth" {
		return "A yawning cave entrance leads into impenetrable darkness. A chill wind whistles from the depths."
	}
	return "You can’t see that here."
}

func (c TombEntrance) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (c TombEntrance) Name() string { return LocationNameCaveMouth }
func (c TombEntrance) Describe() string {
	return "A yawning cave entrance leads into impenetrable darkness. A chill wind whistles from the depths. up to the garden, east to the go deeper into the darkness."
}

func (c TombEntrance) TakeItemByName(string) (interfaces.ItemInterface, string) {
	return nil, "Nothing here to take."
}

func (c TombEntrance) Go(dir string) (string, *interfaces.LocationInterface) {
	switch dir {
	case "up":
		return "You climb back up to the garden.", c.BaseLocation.World.GetLocationByName("Overgrown Garden")
	case "east":
		return "You peer into the depths, and venture forward.", c.BaseLocation.World.GetLocationByName("Goblin Camp")
	default:
		return "You can't go that way.", nil
	}
}

func (c TombEntrance) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "Nothing happens.", true
}
