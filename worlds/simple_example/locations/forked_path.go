package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

var LocationNameForkedPath = "Forked Path"

type ForkedPath struct {
	interfaces.BaseLocation
}

func NewForkedPath(world interfaces.WorldInterface) *ForkedPath {
	return &ForkedPath{
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
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

func (l ForkedPath) TakeItemByName(string) (interfaces.ItemInterface, string) {
	return nil, "There is nothing to take here."
}

func (l ForkedPath) Go(dir string) (string, *interfaces.LocationInterface) {
	switch dir {
	case "south":
		return "You retrace your steps toward the house.", l.BaseLocation.World.GetLocationByName("Woods Entrance")
	case "east":
		return "You step toward the listing shed.", l.BaseLocation.World.GetLocationByName("Shed Exterior")
	case "west":
		return "You push through the undergrowth toward the garden.", l.BaseLocation.World.GetLocationByName("Overgrown Garden")
	default:
		return "That direction leads nowhere discernible.", nil
	}
}

func (l ForkedPath) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "That achieves nothing here.", true
}
