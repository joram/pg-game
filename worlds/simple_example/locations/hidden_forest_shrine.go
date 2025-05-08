package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
)

var LocationNameHiddenForestShrine = "Hidden Forest Shrine"

type HiddenForestShrine struct {
	interfaces.BaseLocation
}

func NewHiddenForestShrine(world interfaces.WorldInterface) *HiddenForestShrine {
	return &HiddenForestShrine{
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

func (l *HiddenForestShrine) Examine(name string) string {
	if name == "shrine" {
		return "The shrine is made of ancient stone, covered in glowing runes. It seems to hum with energy."
	}
	if name == "crystals" {
		return "The crystals float in the air, casting soft rainbow glows on the surrounding trees."
	}
	if name == "trees" {
		return "The trees are tall and ancient, their leaves whispering secrets to one another."
	}
	return "You can’t see that here."
}

func (l *HiddenForestShrine) TalkTo(name string) string {
	return "there is nobody to talk to here"
}

func (l *HiddenForestShrine) Name() string { return LocationNameHiddenForestShrine }

func (l *HiddenForestShrine) Describe() string {
	return `You step into a glade filled with shimmering light. 
A stone shrine rises from the forest floor, vines curling gently around its base.
Crystals float silently in the air, casting soft rainbow glows on the surrounding trees.
A hush falls over everything — the wind, the birds, even your own thoughts.
	
On the pedestal, faint glowing runes pulse gently. They seem to respond to your presence.

A voice, warm and ancient, echoes in your mind:
"You have taken your first step, young traveler. The forest remembers those who are kind and clever."

You feel a path opening before you — not one you can see, but one you will choose.`
}

func (l *HiddenForestShrine) TakeItemByName(name string) (interfaces.ItemInterface, string) {
	return nil, "There is nothing to take here — only serenity."
}

func (l *HiddenForestShrine) UseItem(item interfaces.ItemInterface, target string) (string, bool) {
	return "You kneel at the shrine. It hums softly in response.", true
}

func (l *HiddenForestShrine) Go(dir string) (string, *interfaces.LocationInterface) {
	if dir == "up" {
		return "You climb back up the repaired ladder.", l.BaseLocation.World.GetLocationByName("Ravine Edge")
	}
	return "Only trees surround you.", nil
}
