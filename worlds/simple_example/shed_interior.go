package simple_example

import (
	"fmt"
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

var LocationNameShedInterior = "Shed Interior"

type ItemOldSword struct{}

func (i ItemOldSword) Name() string { return "old sword" }
func (i ItemOldSword) Description() string {
	return "A heavy, time‑worn longsword. Its edge is still sharp enough to hack through thick vines."
}
func (i ItemOldSword) Examine() string {
	return "A heavy, time‑worn longsword. Its edge is still sharp enough to hack through thick vines. or scare off small creatures"
}

type ItemBowl struct {
	Full     bool
	Contents string
}

func (b ItemBowl) Name() string {
	if b.Full {
		return "bowl full of " + b.Contents
	}
	return "an empty bowl"
}

func (b ItemBowl) Description() string {
	if b.Full {
		return fmt.Sprintf("Bowl full of %s", b.Contents)
	}

	return fmt.Sprintf("A bowl, it looks like it could hold something. Probably %s.", b.Contents)
}
func (b ItemBowl) Examine() string {
	if b.Full {
		return fmt.Sprintf("A bowl full of %s", b.Contents)
	}
	return "An empty bowl."
}

type ShedInterior struct {
	sword      ItemOldSword
	bowl       ItemBowl
	swordTaken bool
	bowlTaken  bool
	world      interfaces.WorldInterface
}

func NewShedInterior(world interfaces.WorldInterface) *ShedInterior {
	return &ShedInterior{
		sword: ItemOldSword{},
		bowl: ItemBowl{
			Full: false,
		},
		world: world,
	}
}

func (l *ShedInterior) Examine(name string) string {
	if name == "bowl" {
		return l.bowl.Examine()
	}
	if name == "cloth" && !l.swordTaken {
		return "A long, cloth‑wrapped object rests on a rack."
	}
	if name == "old sword" && !l.swordTaken {
		return l.sword.Examine()
	}
	if name == "sword" && !l.swordTaken {
		return l.sword.Examine()
	}
	return "You can’t see that here."
}

func (l *ShedInterior) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (l *ShedInterior) Name() string { return LocationNameShedInterior }

func (l *ShedInterior) Describe() string {
	desc := "Dust motes dance in the lantern light. Cobwebbed tools line the walls."
	if !l.swordTaken {
		desc += " A long, cloth‑wrapped object rests on a rack."
	}
	if l.bowlTaken {
		desc += " A bowl lies on the floor, empty."
	}
	return desc
}

func (l *ShedInterior) ListKnownItems() []interfaces.ItemInterface {

	var oldSword = ItemOldSword{}

	if l.swordTaken {
		return nil
	}
	return []interfaces.ItemInterface{oldSword}
}

func (l *ShedInterior) TakeItemByName(world interfaces.WorldInterface, name string) (interfaces.ItemInterface, string) {
	if name == "bowl" {
		return &l.bowl, "You take empty the bowl."
	}

	if name == "cloth" && !l.swordTaken {
		l.swordTaken = true
		return l.sword, "You unwrap the bundle, as you do it falls apart in your hands revealing an old sword. You take the sword."
	}
	if name == "old sword" && !l.swordTaken {
		l.swordTaken = true
		return l.sword, "You unwrap the bundle, revealing an old sword."
	}
	return nil, "You can't take that."
}

func (l *ShedInterior) Go(world interfaces.WorldInterface, dir string) (bool, string, interfaces.LocationInterface) {
	switch dir {
	case "out":
		return true, "You step back outside.", world.GetLocationByName("Shed Exterior")
	default:
		return false, "You bump into a wall.", nil
	}
}

func (l *ShedInterior) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "No useful interaction here.", true
}
