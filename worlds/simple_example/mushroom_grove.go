package simple_example

import (
	"fmt"
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

type GlowingMushroom struct{}

func (g GlowingMushroom) Name() string        { return "glowing mushroom" }
func (g GlowingMushroom) Description() string { return "It glows with a soft, magical light." }
func (m GlowingMushroom) Examine() string {
	return "You look closely at the glowing mushroom. It pulses gently like it’s alive."
}

type MushroomGrove struct {
	mushroomTaken bool
	world         interfaces.WorldInterface
}

func NewMushroomGrove(world interfaces.WorldInterface) *MushroomGrove {
	return &MushroomGrove{
		world: world,
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

var LocationNameMushroomGrove = "Mushroom Grove"

func (m *MushroomGrove) Name() string { return LocationNameMushroomGrove }

func (m *MushroomGrove) Describe() string {
	return "Mushrooms of all shapes and sizes glow gently. The air is peaceful here."
}

func (m *MushroomGrove) ListKnownItems() []interfaces.ItemInterface {
	var items []interfaces.ItemInterface
	if !m.mushroomTaken {
		items = append(items, GlowingMushroom{})
	}
	return items
}

func (m *MushroomGrove) TakeItemByName(world interfaces.WorldInterface, name string) (interfaces.ItemInterface, string) {
	if name == "glowing mushroom" && !m.mushroomTaken {
		m.mushroomTaken = true
		return GlowingMushroom{}, "You pick a glowing mushroom."
	}
	return nil, "You can't take that."
}

func (m *MushroomGrove) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "You can't use that here.", true
}

func (m *MushroomGrove) Go(world interfaces.WorldInterface, dir string) (bool, string, interfaces.LocationInterface) {
	if dir == "north" {
		return true, "You return to the goblin camp.", world.GetLocationByName("Goblin Camp")
	}
	if dir == "west" {
		return true, "You head into a cavern glittering with crystals.", nil
	}
	return false, "You can't go that way.", nil
}

func (g *GoblinCamp) Examine(name string) string {
	if name == "goblin" {
		return "The goblin is small, green, and surprisingly tidy. He looks up hopefully at your satchel."
	}
	return fmt.Sprintf("You don't notice anything special about the %s.", name)
}

func (g *GoblinCamp) TalkTo(name string) string {
	if name == "goblin" {
		if g.helped {
			return "The goblin grins. 'Thanks again, friend!'"
		}
		return "The goblin grumbles. 'Glowy... need glowy mushroom...'"
	}
	return ""
}
