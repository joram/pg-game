package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example/items"
)

var LocationNameFallenTree = "Fallen Tree"

type FallenTree struct {
	plankTaken bool
	interfaces.BaseLocation
}

func NewFallenTree(world interfaces.WorldInterface) *FallenTree {
	return &FallenTree{
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

func (l *FallenTree) Examine(name string) string {
	if name == "wooden plank" && !l.plankTaken {
		return "A long, flat plank — perfect for patching something, is part of a log."
	}
	if name == "fallen tree" {
		if l.plankTaken {
			return "A mossy fallen tree lies split and broken."
		}
		return "A mossy fallen tree blocks part of the path. One of its planks is loose and looks usable."
	}
	return "You can’t see that here."
}

func (l *FallenTree) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (l *FallenTree) Name() string { return LocationNameFallenTree }

func (l *FallenTree) Describe() string {
	if l.plankTaken {
		return "A mossy fallen tree lies split and broken."
	}
	return "A mossy fallen tree blocks part of the path. One of its planks is loose and looks usable."
}

func (l *FallenTree) TakeItemByName(name string) (interfaces.ItemInterface, string) {
	if name == "wooden plank" && !l.plankTaken {
		l.plankTaken = true
		return items.WoodenPlank{}, "You pry a plank loose from the tree trunk."
	}
	return nil, "You can't take that."
}

func (l *FallenTree) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "Nothing happens.", true
}

func (l *FallenTree) Go(dir string) (string, *interfaces.LocationInterface) {
	if dir == "east" {
		return "You walk back toward the mushroom grove.", l.BaseLocation.World.GetLocationByName("Mushroom Grove")
	}
	return "There’s too much brush that way.", nil
}
