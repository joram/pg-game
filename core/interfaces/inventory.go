package interfaces

type Inventory struct {
	Items []ItemInterface
}

func (i *Inventory) AddItem(item ItemInterface) {
	i.Items = append(i.Items, item)
}

func (i *Inventory) RemoveItem(name string) ItemInterface {
	for j, invItem := range i.Items {
		if invItem.Name() == name {
			i.Items = append(i.Items[:j], i.Items[j+1:]...)
			return invItem
		}
	}
	return nil
}

func (i *Inventory) ListItems() []ItemInterface {
	return i.Items
}

func (i *Inventory) HaveItem(name string) bool {
	for _, item := range i.Items {
		if item.Name() == name {
			return true
		}
	}
	return false
}
