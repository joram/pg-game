package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example/items"
)

var LocationNameSapTreeClearing = "Sap Tree Clearing"

type SapTreeClearing struct {
	*interfaces.BaseLocation

	warningCount int
	sapTaken     bool
}

func NewSapTreeClearing(world interfaces.WorldInterface) *SapTreeClearing {
	return &SapTreeClearing{
		sapTaken: false,
		BaseLocation: &interfaces.BaseLocation{
			World: world,
		},
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

func (l *SapTreeClearing) Name() string { return LocationNameSapTreeClearing }

func (l *SapTreeClearing) Describe() string {
	directions := " To the west, you see the cave entrance. To the east, a path leads deeper into the forest."
	if l.sapTaken {
		return "A tall tree oozes dried sap. The clearing smells sweet." + directions
	}
	return "A tall tree oozes thick, amber-colored sap. It drips slowly onto a flat rock." + directions
}

func (l *SapTreeClearing) TakeItemByName(name string) (interfaces.ItemInterface, string) {
	if name == "sticky sap" && !l.sapTaken {
		l.sapTaken = true
		return items.StickySap{}, "You collect some sticky sap in a small container."
	}
	return nil, "You can't take that."
}

func (l *SapTreeClearing) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "That doesn’t seem to do anything here.", true
}

func (l *SapTreeClearing) Go(dir string) (string, *interfaces.LocationInterface) {
	if dir == "west" {
		return "You head back to the cave entrance.", l.BaseLocation.World.GetLocationByName("Cave Entrance")
	}
	if dir == "east" {
		l.warningCount++
		if l.warningCount == 1 {
			return "You hear a rustling in the bushes. You should be careful going that way.", nil
		}
		if l.warningCount == 2 {
			return "You hear a growl from the bushes. You should be careful going that way.", nil
		}
		if l.warningCount == 3 {
			l.SetDead(true)
			return "You venture deeper into the forest, you are eaten by a grue.", nil
		}
	}
	return "You can't go that way.", nil
}
