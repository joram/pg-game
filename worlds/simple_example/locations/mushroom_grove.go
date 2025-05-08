package locations

import (
	"fmt"
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example/items"
)

type MushroomGrove struct {
	mushroomTaken bool
	interfaces.BaseLocation
}

var LocationNameMushroomGrove = "Mushroom Grove"

func NewMushroomGrove(world interfaces.WorldInterface) *MushroomGrove {
	return &MushroomGrove{
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

func (m *MushroomGrove) Examine(name string) string {
	if name == "glowing mushroom" && !m.mushroomTaken {
		return "A glowing mushroom. It pulses gently like it’s alive."
	}
	if name == "mushroom grove" {
		return "Mushrooms of all shapes and sizes glow gently. The air is peaceful here."
	}
	return fmt.Sprintf("You don't notice anything special about the %s.", name)
}

func (m *MushroomGrove) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (m *MushroomGrove) Name() string { return LocationNameMushroomGrove }

func (m *MushroomGrove) Describe() string {
	return "Mushrooms of all shapes and sizes glow gently. The air is peaceful here. To the north, you can see the goblin camp. To the south, a dark cave system beckons."
}

func (m *MushroomGrove) TakeItemByName(name string) (interfaces.ItemInterface, string) {
	if name == "glowing mushroom" && !m.mushroomTaken {
		m.mushroomTaken = true
		return items.GlowingMushroom{}, "You pick a glowing mushroom."
	}
	return nil, "You can't take that."
}

func (m *MushroomGrove) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "You can't use that here.", true
}

func (m *MushroomGrove) Go(dir string) (string, *interfaces.LocationInterface) {
	if dir == "north" {
		return "You return to the goblin camp.", m.BaseLocation.World.GetLocationByName(LocationNameGoblinCamp)
	}
	if dir == "south" {
		return "You head deeper into a cave system, arriving at a junction.", m.BaseLocation.World.GetLocationByName(LocationNameCaveJunction)
	}
	return "You can't go that way.", nil
}
