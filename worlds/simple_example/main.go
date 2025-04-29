package simple_example

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

var SimpleExampleWorld = interfaces.World{
	Name:             "Small Example World",
	StartingLocation: &start,
	CurrentLocation:  &start,
	Locations: map[string]interfaces.LocationInterface{
		LocationNameStarting: &start,
	},
}
