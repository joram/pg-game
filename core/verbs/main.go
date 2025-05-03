package verbs

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

type VerbName string

const (
	VerbNameLook    VerbName = "look"
	VerbNameExamine VerbName = "examine"
	VerbNameGo      VerbName = "go"
	VerbNameTake    VerbName = "take"
	VerbNameUse     VerbName = "use"
)

type VerbInterface interface {
	Execute(*interfaces.WorldInterface)
	ExecuteOnItem(*interfaces.WorldInterface, interfaces.ItemInterface)
	//ExecuteOnNpc(*world.WorldInterface, *world.Location)
}

type Verb struct {
	Name        VerbName
	Description string
	VerbInterface
}

var VerbLook = Verb{
	Name:        VerbNameLook,
	Description: "Look at your surroundings",
}

var VerbExamine = Verb{
	Name:        VerbNameExamine,
	Description: "Examine an item",
}

var VerbGo = Verb{
	Name:        VerbNameGo,
	Description: "Go to a different location",
}

var VerbTake = Verb{
	Name:        VerbNameTake,
	Description: "Take an item",
}

var VerbUse = Verb{
	Name:        VerbNameUse,
	Description: "Use an item",
}

var AllVerbs = []Verb{
	VerbLook,
	//VerbExamine,
	VerbGo,
	VerbTake,
	VerbUse,
}
