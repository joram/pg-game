package interfaces

import "fmt"

type ItemInterface interface {
	Name() string
	Description() string
	Examine() string
}

type LocationInterface interface {
	Name() string
	Describe() string
	TakeItemByName(WorldInterface, string) (ItemInterface, string)
	Go(world WorldInterface, name string) (bool, string, LocationInterface)
	UseItem(item ItemInterface, targetName string) (string, bool)
	Examine(name string) string
	TalkTo(name string) string
}

type WorldInterface struct {
	Name             string
	StartingLocation LocationInterface
	CurrentLocation  LocationInterface
	Locations        map[string]LocationInterface
	Inventory        *Inventory
}

func (i WorldInterface) GetLocationByName(s string) LocationInterface {
	for _, loc := range i.Locations {
		if loc.Name() == s {
			return loc
		}
	}
	fmt.Printf("Location %s not found in world %s\n", s, i.Name)
	return nil
}
