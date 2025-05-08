package interfaces

type WorldInterface interface {
	HasInInventory(name string) bool
	GetLocationByName(name string) *LocationInterface
	Sayf(format string, a ...any)
	ReceiveMessage() string
}

type ItemInterface interface {
	Name() string
	Description() string
	Examine() string
}

type LocationInterface interface {
	Name() string
	Describe() string
	TakeItemByName(string) (ItemInterface, string)
	PickUpItemFromGround(name string) ItemInterface
	DropItemByName(ItemInterface)
	ItemsOnGround() []ItemInterface
	Go(name string) (string, *LocationInterface)
	UseItem(item ItemInterface, targetName string) (string, bool)
	Examine(name string) string
	TalkTo(name string) string
	SetDead(bool)
	GetDead() bool
}

type BaseLocation struct {
	Dead          bool
	itemsOnGround []ItemInterface
	World         WorldInterface
}

func (l *BaseLocation) SetDead(d bool) {
	l.Dead = d
}

func (l *BaseLocation) GetDead() bool {
	return l.Dead
}

func (l *BaseLocation) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (l *BaseLocation) ItemsOnGround() []ItemInterface {
	return l.itemsOnGround
}

func (l *BaseLocation) PickUpItemFromGround(name string) ItemInterface {
	for i, item := range l.itemsOnGround {
		if item.Name() == name {
			l.itemsOnGround = append(l.itemsOnGround[:i], l.itemsOnGround[i+1:]...)
			return item
		}
	}
	return nil
}

func (l *BaseLocation) sayf(format string, a ...any) {
	l.World.Sayf(format, a...)
}

func (l *BaseLocation) DropItemByName(item ItemInterface) {
	if item.Name() == "teleportation rune stone" {
		l.itemsOnGround = append(l.itemsOnGround[:1], l.itemsOnGround[1:]...)
		l.sayf("You drop the teleportation rune stone.")
		l.sayf("what would you like to name this location?")
	}
	l.itemsOnGround = append(l.itemsOnGround, item)
}

func (l *BaseLocation) GetLocationByName(name string) *LocationInterface {
	return nil
}

func (l *BaseLocation) CreateItemOnGround(item ItemInterface) {
	l.itemsOnGround = append(l.itemsOnGround, item)
}
