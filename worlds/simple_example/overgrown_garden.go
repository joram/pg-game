package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

var LocationNameOvergrownGarden = "Overgrown Garden"

type OvergrownGarden struct {
	vinesCut bool
	world    interfaces.WorldInterface
}

func NewOvergrownGarden(world interfaces.WorldInterface) *OvergrownGarden {
	return &OvergrownGarden{
		vinesCut: false,
		world:    world,
	}
}

func (g *OvergrownGarden) Examine(name string) string {
	if name == "vines" && !g.vinesCut {
		return "Thick vines coil around a waist high stone plinth at the center."
	}
	if name == "plinth" {
		if g.vinesCut {
			return "An old stone plinth stands cleared of vines. Its lid lies ajar, revealing a stairway descending into darkness."
		}
		return "A waist high stone plinth is choked with thick vines."
	}
	return "You can’t see that here."
}

func (g *OvergrownGarden) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (g *OvergrownGarden) Name() string { return LocationNameOvergrownGarden }

func (g *OvergrownGarden) Describe() string {
	if g.vinesCut {
		return "An old stone plinth stands cleared of vines. Its lid lies ajar, revealing a stairway descending into darkness."
	}
	return "Broken fountains and knee high ivy choke the clearing. Thick vines coil around a waist high stone plinth at the center."
}

func (g *OvergrownGarden) ListKnownItems() []interfaces.ItemInterface { return nil }
func (g *OvergrownGarden) TakeItemByName(interfaces.WorldInterface, string) (interfaces.ItemInterface, string) {
	return nil, "There's nothing here to pick up."
}

func (g *OvergrownGarden) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	if item.Name() == "old sword" && target == "vines" && !g.vinesCut {
		g.vinesCut = true
		return "You hack through the vines, clearing the plinth and revealing a hidden stairway leading down.", true
	}
	return "That doesn't work.", true
}

func (g *OvergrownGarden) Go(world interfaces.WorldInterface, dir string) (bool, string, interfaces.LocationInterface) {
	switch dir {
	case "east":
		return true, "You return to the fork in the path.", world.GetLocationByName("Forked Path")
	case "down":
		if g.vinesCut {
			return true, "You descend the narrow stone steps.", world.GetLocationByName("Cave Mouth")
		}
		return false, "Thick vines block any passage downward.", nil
	default:
		return false, "Shrubs and trees block your path.", nil
	}
}
