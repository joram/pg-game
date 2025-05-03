package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

func NewSimpleWorld(inventory *interfaces.Inventory) interfaces.WorldInterface {
	var world = interfaces.WorldInterface{}

	var start = NewStartingRoom(world)
	var besideHouse = NewBesideHouse(world)
	var caveMouth = NewCaveMouth(world)
	var forkedPath = NewForkedPath(world)
	var frontSteps = NewFrontSteps(world)
	var overgrownGarden = NewOvergrownGarden(world)
	var shedExterior = NewShedExterior(world)
	var shedInterior = NewShedInterior(world)
	var woodsEntrance = NewWoodsEntrance(world)
	var goblinCamp = NewGoblinCamp(world)
	var mushroomGrove = NewMushroomGrove(world)
	var fallenTree = NewFallenTree(world)
	var hiddenForestShrine = NewHiddenForestShrine(world)
	var ravineEdge = NewRavineEdge(world)
	var sapTreeClearing = NewSapTreeClearing(world)

	var locations = map[string]interfaces.LocationInterface{}
	for _, loc := range []interfaces.LocationInterface{
		start,
		besideHouse,
		caveMouth,
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
	} {
		locations[loc.Name()] = loc
	}

	world.StartingLocation = start
	world.CurrentLocation = start
	world.Locations = locations
	return world
}
