package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example/items"
)

var LocationNameShedInterior = "Shed Interior"

type ShedInterior struct {
	sword      items.ItemOldSword
	bowl       items.ItemBowl
	swordTaken bool
	bowlTaken  bool
	interfaces.BaseLocation
}

func NewShedInterior(world interfaces.WorldInterface) *ShedInterior {
	return &ShedInterior{
		sword: items.ItemOldSword{},
		bowl: items.ItemBowl{
			Full: false,
		},
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
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

func (l *ShedInterior) TakeItemByName(name string) (interfaces.ItemInterface, string) {
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

func (l *ShedInterior) Go(dir string) (string, *interfaces.LocationInterface) {
	switch dir {
	case "out":
		return "You step back outside.", l.BaseLocation.World.GetLocationByName("Shed Exterior")
	default:
		return "You bump into a wall.", nil
	}
}

func (l *ShedInterior) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "No useful interaction here.", true
}
