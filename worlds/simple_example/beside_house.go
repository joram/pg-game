package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

type LogPile struct{}

func (l LogPile) Examine() string {
	return "The log pile is neatly stacked and ready for use."
}

func (l LogPile) Name() string        { return "log pile" }
func (l LogPile) Description() string { return "A neatly stacked pile of firewood beside the house." }

type Screwdriver struct{}

func (s Screwdriver) Examine() string {
	return "The screwdriver is a bit rusty, it's a flathead, and it still seems to works."
}

func (s Screwdriver) Name() string        { return "screwdriver" }
func (s Screwdriver) Description() string { return "A flathead screwdriver, a bit rusty but usable." }

var logPile = LogPile{}

var LocationNameBesideHouse = "Beside House"

type BesideHouse struct {
	takenScrewdriver bool
	screwdriver      *Screwdriver
}

func NewBesideHouse(worldInterface interfaces.WorldInterface) *BesideHouse {
	return &BesideHouse{
		takenScrewdriver: false,
		screwdriver:      &Screwdriver{},
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

func (l *BesideHouse) ListKnownItems() []interfaces.ItemInterface {
	items := []interfaces.ItemInterface{logPile}
	if !l.takenScrewdriver {
		items = append(items, l.screwdriver)
	}
	return items
}

func (l *BesideHouse) TakeItemByName(w interfaces.WorldInterface, name string) (interfaces.ItemInterface, string) {
	if name == "log pile" {
		return nil, "The log pile is too heavy to take."
	}
	if name == "screwdriver" && !l.takenScrewdriver {
		l.takenScrewdriver = true
		return l.screwdriver, "You take the screwdriver."
	}
	return nil, "You can't take that item."
}

func (l *BesideHouse) Go(world interfaces.WorldInterface, direction string) (bool, string, interfaces.LocationInterface) {
	if direction == "west" {
		return true, "You go back to the front steps.", world.GetLocationByName("Front Steps")
	}
	return false, "You can't go that way.", nil
}

func (l *BesideHouse) UseItem(item interfaces.ItemInterface, targetName string) (string, bool) {
	return "You can't use that item here.", true
}
