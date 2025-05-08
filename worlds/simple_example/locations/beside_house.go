package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example/items"
)

var logPile = items.LogPile{}

var LocationNameBesideHouse = "Beside House"

type BesideHouse struct {
	interfaces.BaseLocation
	takenScrewdriver bool
	screwdriver      *items.Screwdriver
}

func NewBesideHouse(world interfaces.WorldInterface) interfaces.LocationInterface {
	return &BesideHouse{
		takenScrewdriver: false,
		screwdriver:      &items.Screwdriver{},
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

func (l *BesideHouse) Examine(name string) string {
	if name == "log pile" {
		return logPile.Examine()
	}
	if name == "screwdriver" && !l.takenScrewdriver {
		return l.screwdriver.Examine()
	}
	if name == "house" {
		return "The house is old and weathered, with a small porch and a few windows."
	}
	return "You can’t see that here."
}

func (l *BesideHouse) TalkTo(name string) string {
	return "there is nobody here to talk to."
}

func (l *BesideHouse) Name() string {
	return LocationNameBesideHouse
}

func (l *BesideHouse) Describe() string {
	desc := "You are standing beside the house. There is a log pile stacked against the wall."
	if !l.takenScrewdriver {
		desc += " A screwdriver lies on top of the pile."
	}
	return desc
}

func (l *BesideHouse) TakeItemByName(name string) (interfaces.ItemInterface, string) {
	if name == "log pile" {
		return nil, "The log pile is too heavy to take."
	}
	if name == "screwdriver" && !l.takenScrewdriver {
		l.takenScrewdriver = true
		return l.screwdriver, "You take the screwdriver."
	}
	return nil, "You can't take that item."
}

func (l *BesideHouse) Go(direction string) (string, *interfaces.LocationInterface) {
	if direction == "west" {
		return "You go back to the front steps.", l.BaseLocation.World.GetLocationByName("Front Steps")
	}
	return "You can't go that way.", nil
}

func (l *BesideHouse) UseItem(item interfaces.ItemInterface, targetName string) (string, bool) {
	return "You can't use that item here.", true
}
