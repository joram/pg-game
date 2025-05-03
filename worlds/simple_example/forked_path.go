package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

var LocationNameForkedPath = "Forked Path"

type ForkedPath struct {
	world interfaces.WorldInterface
}

func NewForkedPath(world interfaces.WorldInterface) *ForkedPath {
	return &ForkedPath{
		world: world,
	}
}

func (l ForkedPath) Examine(name string) string {
	if name == "forked path" {
		return "The path splits here: to the east stands a decrepit wooden shed; to the west a tangle of ivy hides a small garden clearing."
	}
	if name == "shed" {
		return "The shed is old and weathered, with peeling paint and a sagging roof."
	}
	if name == "garden" {
		return "The garden is overgrown with weeds, but you can see hints of flowers peeking through."
	}
	return "You can’t see that here."
}

func (l ForkedPath) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (l ForkedPath) Name() string { return LocationNameForkedPath }

func (l ForkedPath) Describe() string {
	return "The path splits here: to the east stands a decrepit wooden shed; to the west a tangle of ivy hides a small garden clearing."
}

func (l ForkedPath) ListKnownItems() []interfaces.ItemInterface { return nil }
func (l ForkedPath) TakeItemByName(interfaces.WorldInterface, string) (interfaces.ItemInterface, string) {
	return nil, "There is nothing to take here."
}

func (l ForkedPath) Go(world interfaces.WorldInterface, dir string) (bool, string, interfaces.LocationInterface) {
	switch dir {
	case "south":
		return true, "You retrace your steps toward the house.", world.GetLocationByName("Woods Entrance")
	case "east":
		return true, "You step toward the listing shed.", world.GetLocationByName("Shed Exterior")
	case "west":
		return true, "You push through the undergrowth toward the garden.", world.GetLocationByName("Overgrown Garden")
	default:
		return false, "That direction leads nowhere discernible.", nil
	}
}

func (l ForkedPath) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "That achieves nothing here.", true
}
