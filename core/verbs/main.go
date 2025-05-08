package verbs

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/core/world"
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
	Execute(*world.World)
	ExecuteOnItem(*world.World, interfaces.ItemInterface)
	//ExecuteOnNpc(*world.World, *world.Location)
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

var VerbTalk = Verb{
	Name:        "talk",
	Description: "Talk to someone",
}

var VerbInventory = Verb{
	Name:        "inventory",
	Description: "Check your inventory",
}

var VerbDrop = Verb{
	Name:        "drop",
	Description: "Drop an item on the ground",
}

var VerbTeleport = Verb{
	Name:        "teleport",
	Description: "Teleport to a different location",
}

var VerbQuit = Verb{
	Name:        "quit",
	Description: "Quit the game",
}

var VerbCommands = Verb{
	Name:        "commands",
	Description: "List all available commands",
}

var AllVerbs = []Verb{
	VerbCommands,
	VerbLook,
	VerbExamine,
	VerbGo,
	VerbTeleport,
	VerbTake,
	VerbDrop,
	VerbUse,
	VerbTalk,
	VerbInventory,
	VerbQuit,
}
