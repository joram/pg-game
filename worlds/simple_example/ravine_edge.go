package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

var LocationNameRavineEdge = "Ravine Edge"

type RavineEdge struct {
	ladderFixed bool
	world       interfaces.WorldInterface
}

func NewRavineEdge(world interfaces.WorldInterface) *RavineEdge {
	return &RavineEdge{
		ladderFixed: false,
		world:       world,
	}
}

func (r *RavineEdge) Examine(name string) string {
	if name == "ladder" {
		if r.ladderFixed {
			return "The ladder is now fixed and safe to use."
		}
		return "The ladder is broken and unsafe to use."
	}
	if name == "ravine" {
		return "A deep ravine stretches before you, with steep sides."
	}
	return "You can’t see that here."
}

func (r *RavineEdge) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (r *RavineEdge) Name() string { return LocationNameRavineEdge }

func (r *RavineEdge) Describe() string {
	if r.ladderFixed {
		return "A repaired ladder now spans the ravine, leading down safely."
	}
	return "You reach a ravine. A broken ladder lies splintered nearby. It's too dangerous to climb without fixing it."
}

func (r *RavineEdge) ListKnownItems() []interfaces.ItemInterface {
	return nil
}

func (r *RavineEdge) TakeItemByName(world interfaces.WorldInterface, name string) (interfaces.ItemInterface, string) {
	return nil, "There’s nothing to take."
}

func (r *RavineEdge) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	if item.Name() == "wooden plank" && target == "ladder" && !r.ladderFixed {
		if r.world.Inventory.HaveItem("sticky sap") {
			r.ladderFixed = true
			return "You patch the broken ladder with the plank and seal it with sticky sap.", true
		}
		return "You need something to hold the plank in place.", true
	}
	return "That doesn't work here.", true
}

func (r *RavineEdge) Go(world interfaces.WorldInterface, dir string) (bool, string, interfaces.LocationInterface) {
	if dir == "north" {
		return true, "You head back to the mushroom grove.", world.GetLocationByName("Mushroom Grove")
	}
	if dir == "down" {
		if r.ladderFixed {
			return true, "You carefully descend the ladder into the ravine.", world.GetLocationByName("Hidden Forest Shrine")
		}
		return false, "The ladder is broken — it's too dangerous to climb down.", nil
	}
	return false, "You can't go that way.", nil
}
