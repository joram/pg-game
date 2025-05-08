package locations

import (
	"fmt"
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example/items"
)

type GoblinCamp struct {
	tookShard    bool
	helped       bool
	goblinRanOff bool
	interfaces.BaseLocation
}

func NewGoblinCamp(world interfaces.WorldInterface) *GoblinCamp {
	return &GoblinCamp{
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

var LocationNameGoblinCamp = "Goblin Camp"

func (g *GoblinCamp) Name() string { return LocationNameGoblinCamp }

func (g *GoblinCamp) Describe() string {
	directionsDescription := "\nYou can see a path leading south deeper into the cave, and a path leading west back to the cave mouth."
	if g.goblinRanOff {
		return "The goblin has run off, leaving his stew behind. The camp is now empty." + directionsDescription
	}
	if !g.helped {
		return "A small goblin stirs a pot and looks up at you. He seems hungry and mutters something about glowing mushrooms." + directionsDescription
	}
	if g.helped && !g.tookShard {
		return "The goblin smiles happily, stirring his now glowing stew. He left a crystal shard nearby." + directionsDescription
	}
	return "The goblin is busy stirring his stew. He seems happy and content." + directionsDescription
}
func (g *GoblinCamp) TakeItemByName(name string) (interfaces.ItemInterface, string) {
	if name == "crystal shard" && g.helped && !g.tookShard {
		g.tookShard = true
		return items.CrystalShard{}, "You take the crystal shard the goblin left for you."
	}
	return nil, "You can't take that."
}

func (g *GoblinCamp) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	if item.Name() == "glowing mushroom" && !g.helped && target == "goblin" {
		g.helped = true
		return "The goblin gobbles the mushroom and grins. He tosses you a shiny shard from his satchel!", false
	}
	if item.Name() == "old sword" && target == "goblin" {
		return "The goblin looks at you in horror and runs away, leaving his stew behind.", true
	}
	if item.Name() == "old sword" && target == "stew" {
		return "You dip your sword in the stew, nothing happens.", true
	}

	bowl, ok := item.(items.ItemBowl)
	if ok && target == "stew" {
		if bowl.Full {
			return "You cannot do that with a full bowl.", true
		}
		if !g.goblinRanOff {
			return "The goblin swats your hand away. He looks annoyed.", true
		}
		bowl.Full = true
		bowl.Contents = "stew"
		return "You fill the bowl with stew.", true
	}

	return "That doesn't do anything here.", true
}

func (g *GoblinCamp) Go(dir string) (string, *interfaces.LocationInterface) {
	if dir == "south" {
		return "You head south to the mushroom grove.", g.BaseLocation.World.GetLocationByName("Mushroom Grove")
	}
	if dir == "west" {
		return "You go back to the cave mouth.", g.BaseLocation.World.GetLocationByName("Cave Mouth")
	}
	return "You can't go that way.", nil
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
