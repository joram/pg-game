package interfaces

type World struct {
	Name             string
	StartingLocation LocationInterface
	CurrentLocation  LocationInterface
	Locations        map[string]LocationInterface
}
