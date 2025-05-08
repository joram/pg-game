package items

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

type PlacedTeleportationStone struct {
	LocationName string
	Location     interfaces.LocationInterface
}

func (p PlacedTeleportationStone) Name() string {
	return "Teleportation Stone"
}

func (p PlacedTeleportationStone) Description() string {
	return "A stone that allows you to teleport to a different location."
}

func (p PlacedTeleportationStone) Examine() string {
	return "The stone is engraved with runes that glow faintly."
}

func NewPlacedTeleportationStone(locationName string, location interfaces.LocationInterface) *PlacedTeleportationStone {
	return &PlacedTeleportationStone{
		LocationName: locationName,
		Location:     location,
	}
}

type BagOfTeleportationRuneStones struct {
	PlacedStones []PlacedTeleportationStone
}

func (b *BagOfTeleportationRuneStones) Name() string {
	return "Bag of Rune Stones of Teleportation"
}

func (b *BagOfTeleportationRuneStones) Description() string {
	return "A bag of rune stones that can be used to teleport to different locations."
}

func (b *BagOfTeleportationRuneStones) Examine() string {
	return "The bag is filled with rune stones that glow with a faint light."
}

func (b *BagOfTeleportationRuneStones) GetTeleportationStone(name string) *PlacedTeleportationStone {
	for _, stone := range b.PlacedStones {
		if stone.LocationName == name {
			return &stone
		}
	}
	return nil
}

func (b *BagOfTeleportationRuneStones) AddTeleportationStone(stone PlacedTeleportationStone) {
	b.PlacedStones = append(b.PlacedStones, stone)
}

func NewBagOfTeleportationRuneStones() *BagOfTeleportationRuneStones {
	return &BagOfTeleportationRuneStones{
		PlacedStones: []PlacedTeleportationStone{},
	}
}
