package simple_example

import "github.com/veilstream/psql-text-based-adventure/core/interfaces"

var LocationNameShedExterior = "Shed Exterior"

type ShedExterior struct {
	hingeRemoved bool
}

func NewShedExterior(world interfaces.WorldInterface) *ShedExterior {
	return &ShedExterior{
		hingeRemoved: false,
	}
}

func (s *ShedExterior) Examine(name string) string {
	if name == "shed" {
		if s.hingeRemoved {
			return "The shed door hangs loose on one side; you could easily slip inside."
		}
		return "A dilapidated wooden shed leans at odd angles. The door is secured by a single rusty hinge plate."
	}
	if name == "door" {
		if s.hingeRemoved {
			return "The door hangs loose on one side; you could easily slip inside."
		}
		return "The door is secured by a single rusty hinge plate."
	}
	if name == "hinge" {
		if s.hingeRemoved {
			return "The hinge is already removed."
		}
		return "A rusty hinge plate secures the door."
	}
	return "You can’t see that here."
}

func (s *ShedExterior) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (s *ShedExterior) Name() string { return LocationNameShedExterior }

func (s *ShedExterior) Describe() string {
	desc := "A dilapidated wooden shed leans at odd angles. The door is secured by a single rusty hinge plate."
	if s.hingeRemoved {
		desc = "The shed door hangs loose on one side; you could easily slip inside."
	}
	return desc
}

func (s *ShedExterior) ListKnownItems() []interfaces.ItemInterface { return nil }
func (s *ShedExterior) TakeItemByName(interfaces.WorldInterface, string) (interfaces.ItemInterface, string) {
	return nil, "There's nothing here worth taking."
}

func (s *ShedExterior) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	if item.Name() == "screwdriver" && target == "hinge" && !s.hingeRemoved {
		s.hingeRemoved = true
		return "You pry off the rusty hinge plate—the door creaks open a crack.", true
	}
	return "That doesn't seem to work.", true
}

func (s *ShedExterior) Go(world interfaces.WorldInterface, dir string) (bool, string, interfaces.LocationInterface) {
	switch dir {
	case "west":
		return true, "You head back to the fork.", world.GetLocationByName("Forked Path")
	case "in":
		if s.hingeRemoved {
			return true, "You slip into the dark shed interior.", world.GetLocationByName("Shed Interior")
		}
		return false, "The door won't budge; the hinge is still holding.", nil
	default:
		return false, "You can't go that way.", nil
	}
}
