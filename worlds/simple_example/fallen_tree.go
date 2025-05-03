package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

type WoodenPlank struct{}

func (p WoodenPlank) Examine() string {
	return "A long, flat plank — perfect for patching something."
}

func (p WoodenPlank) Name() string { return "wooden plank" }
func (p WoodenPlank) Description() string {
	return "A long, flat plank — perfect for patching something."
}

var LocationNameFallenTree = "Fallen Tree"

type FallenTree struct {
	plankTaken bool
	world      interfaces.WorldInterface
}

func NewFallenTree(world interfaces.WorldInterface) *FallenTree {
	return &FallenTree{
		world: world,
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

func (l *FallenTree) ListKnownItems() []interfaces.ItemInterface {
	if !l.plankTaken {
		return []interfaces.ItemInterface{WoodenPlank{}}
	}
	return nil
}

func (l *FallenTree) TakeItemByName(world interfaces.WorldInterface, name string) (interfaces.ItemInterface, string) {
	if name == "wooden plank" && !l.plankTaken {
		l.plankTaken = true
		return WoodenPlank{}, "You pry a plank loose from the tree trunk."
	}
	return nil, "You can't take that."
}

func (l *FallenTree) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "Nothing happens.", true
}

func (l *FallenTree) Go(world interfaces.WorldInterface, dir string) (bool, string, interfaces.LocationInterface) {
	if dir == "east" {
		return true, "You walk back toward the mushroom grove.", world.GetLocationByName("Mushroom Grove")
	}
	return false, "There’s too much brush that way.", nil
}
