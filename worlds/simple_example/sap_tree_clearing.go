package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

type StickySap struct{}

func (s StickySap) Examine() string {
	return "A thick, sticky sap that could glue things together."
}

func (s StickySap) Name() string { return "sticky sap" }
func (s StickySap) Description() string {
	return "A thick, sticky sap that could glue things together."
}

var LocationNameSapTreeClearing = "Sap Tree Clearing"

type SapTreeClearing struct {
	sapTaken bool
	world    interfaces.WorldInterface
}

func NewSapTreeClearing(world interfaces.WorldInterface) *SapTreeClearing {
	return &SapTreeClearing{
		sapTaken: false,
		world:    world,
	}
}

func (l *SapTreeClearing) Examine(name string) string {
	if name == "sticky sap" && !l.sapTaken {
		return "A thick, sticky sap that could glue things together."
	}
	if name == "sap tree" {
		if l.sapTaken {
			return "A tall tree oozes dried sap. The clearing smells sweet."
		}
		return "A tall tree oozes thick, amber-colored sap. It drips slowly onto a flat rock."
	}
	return "You can’t see that here."
}

func (l *SapTreeClearing) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (l *SapTreeClearing) Name() string { return LocationNameSapTreeClearing }

func (l *SapTreeClearing) Describe() string {
	if l.sapTaken {
		return "A tall tree oozes dried sap. The clearing smells sweet."
	}
	return "A tall tree oozes thick, amber-colored sap. It drips slowly onto a flat rock."
}

func (l *SapTreeClearing) ListKnownItems() []interfaces.ItemInterface {
	if !l.sapTaken {
		return []interfaces.ItemInterface{StickySap{}}
	}
	return nil
}

func (l *SapTreeClearing) TakeItemByName(world interfaces.WorldInterface, name string) (interfaces.ItemInterface, string) {
	if name == "sticky sap" && !l.sapTaken {
		l.sapTaken = true
		return StickySap{}, "You collect some sticky sap in a small container."
	}
	return nil, "You can't take that."
}

func (l *SapTreeClearing) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "That doesn’t seem to do anything here.", true
}

func (l *SapTreeClearing) Go(world interfaces.WorldInterface, dir string) (bool, string, interfaces.LocationInterface) {
	if dir == "west" {
		return true, "You head back toward the mushroom grove.", world.GetLocationByName("Mushroom Grove")
	}
	return false, "You can't go that way.", nil
}
