package items

type FirstDoorKey struct {
}

func (f FirstDoorKey) Name() string {
	return "key"
}

func (f FirstDoorKey) Description() string {
	return "This is a key to a door."
}
func (f FirstDoorKey) Examine() string {
	return "This is a key to a door. it looks fragile, like it could break easily."
}
