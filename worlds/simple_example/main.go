package simple_example

import (
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/core/world"
	locations2 "github.com/veilstream/psql-text-based-adventure/worlds/simple_example/locations"
)

func NewSimpleWorld(backend *pgproto3.Backend) *world.World {
	var worldd = world.NewWorld("Tutorial Region")

	var start = locations2.NewStartingRoom(worldd)
	var besideHouse = locations2.NewBesideHouse(worldd)
	var tombEntrance = locations2.NewTombEntrance(worldd)
	var forkedPath = locations2.NewForkedPath(worldd)
	var frontSteps = locations2.NewFrontSteps(worldd)
	var overgrownGarden = locations2.NewOvergrownGarden(worldd)
	var shedExterior = locations2.NewShedExterior(worldd)
	var shedInterior = locations2.NewShedInterior(worldd)
	var woodsEntrance = locations2.NewWoodsEntrance(worldd)
	var goblinCamp = locations2.NewGoblinCamp(worldd)
	var mushroomGrove = locations2.NewMushroomGrove(worldd)
	var fallenTree = locations2.NewFallenTree(worldd)
	var hiddenForestShrine = locations2.NewHiddenForestShrine(worldd)
	var ravineEdge = locations2.NewRavineEdge(worldd)
	var sapTreeClearing = locations2.NewSapTreeClearing(worldd)
	var caveJunction = locations2.NewCaveJunction(worldd)
	var caveExit = locations2.NewCaveExit(worldd)

	var locations = map[string]*interfaces.LocationInterface{}
	for _, loc := range []interfaces.LocationInterface{
		start,
		besideHouse,
		tombEntrance,
		forkedPath,
		frontSteps,
		overgrownGarden,
		shedExterior,
		shedInterior,
		woodsEntrance,
		goblinCamp,
		mushroomGrove,
		fallenTree,
		hiddenForestShrine,
		ravineEdge,
		sapTreeClearing,
		caveJunction,
		caveExit,
	} {
		locations[loc.Name()] = &loc
	}

	worldd.StartingLocation = locations[start.Name()]
	worldd.CurrentLocation = locations[start.Name()]
	worldd.Locations = locations
	return worldd
}
