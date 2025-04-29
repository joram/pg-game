package main

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

type Inventory struct {
	Items []interfaces.ItemInterface
}

func (i *Inventory) AddItem(item interfaces.ItemInterface) {
	i.Items = append(i.Items, item)
}

func (i *Inventory) RemoveItem(name string) interfaces.ItemInterface {
	for j, invItem := range i.Items {
		if invItem.Name() == name {
			i.Items = append(i.Items[:j], i.Items[j+1:]...)
			return invItem
		}
	}
	return nil
}

func (i *Inventory) ListItems() []interfaces.ItemInterface {
	return i.Items
}
