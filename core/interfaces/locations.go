package interfaces

type LocationInterface interface {
	Name() string
	Describe() string
	ListKnownItems() []ItemInterface
	TakeItemByName(string) ItemInterface
	Go(name string) (bool, string, LocationInterface)
	UseItem(item ItemInterface, targetName string) string
}
