package locations

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/core/items"
)

type CaveExit struct {
	travellerTalkCount int
	interfaces.BaseLocation
}

func (c *CaveExit) UseItem(item interfaces.ItemInterface, targetName string) (string, bool) {
	return "Nothing happens.", true
}

var LocationNameCaveExit = "Cave Exit"

func NewCaveExit(world interfaces.WorldInterface) *CaveExit {
	return &CaveExit{
		travellerTalkCount: 0,
		BaseLocation: interfaces.BaseLocation{
			World: world,
		},
	}
}

func (c *CaveExit) Examine(name string) string {
	return "You can’t see that here."
}

func (c *CaveExit) TalkTo(name string) string {
	c.travellerTalkCount += 1
	if name == "traveler" {

		if c.travellerTalkCount == 1 {
			return "The traveler looks up at you, their face obscured by a hood. They mutter something about the forest being cursed."
		}
		if c.travellerTalkCount == 2 {
			return "The traveler looks up at you, their face obscured by a hood. They mutter something about the forest being cursed. 'You should leave this place,' they say, 'before it consumes you.'"
		}
		if c.travellerTalkCount == 3 {
			c.BaseLocation.CreateItemOnGround(items.NewBagOfTeleportationRuneStones())
			return "The traveler pulls himself up off the rock, using his staff for support. " +
				"His presence is imposing, and you feel a chill run down your spine, not from fear, but from a deep sense of wisdom and power eminating from his presence. " +
				"He looks at you with piercing eyes, and you feel as if he can see right through you. " +
				"After spending a moment examining you, he turns his gaze to the forest. and starts to talk, his voice gravelly and low. " +
				"\"You seem to have no intention of turning back, and I would be letting you walk to your own death if I did not intervene.\" " +
				"He pauses for a moment, and you feel a sense of dread wash over you. " +
				"He reaches into his pocket and pulls out a small ornate bag, and places it on the stone in between you. " +
				"\"Take this bag, it contains teleportation stones. It will allow you to escape the forest if you find yourself in danger.\" " +
				"He looks at you with a serious expression, and you can tell that he is not joking. " +
				"He then with a slight hint of a grin, closes his eyes, and you feel a wave of energy wash over you. " +
				"The wind seems to pick up around him. You watch carefully and after a moment he opens his eyes, and you see a small glimmer of light in the air, then he vanishes. " +
				"He teleports away, leaving you alone with the bag. "
		}
		return "The traveler is gone."
	}
	return "You can’t see that here."
}

func (c *CaveExit) Name() string { return LocationNameCaveExit }
func (c *CaveExit) Describe() string {
	return "To the south, you see a cave entrance. Outside the cave a dark and foreboding forest looms, " +
		"its trees twisted and gnarled seem to be leaning in towards you. " +
		"It takes you a moment, but you notice a weary traveler sitting on a rock, they're disheveled and hunched " +
		"in a tattered robe, you can not see their face, they don't seem to notice you. " +
		"A small trail leads to the north, deep into the dark forest. " +
		"To the east, you can see a small clearing with a large tree in the center. "
}

func (c *CaveExit) TakeItemByName(string) (interfaces.ItemInterface, string) {
	return nil, "Nothing here to take."
}

func (c *CaveExit) Go(dir string) (string, *interfaces.LocationInterface) {
	switch dir {
	case "south":
		return "You head back into the cave.", c.BaseLocation.World.GetLocationByName("Cave Junction")
	case "north":
		return "You can't go that way. John has not implemented that yet.", nil
		//return "You venture deeper into the dark forest, slashing at the small branches obscuring the trail, clearing a way.", world.GetLocationByName("Dark Woods Trail")
	case "east":
		return "You head towards the clearing, where a large tree stands.", c.BaseLocation.World.GetLocationByName("Sap Tree Clearing")
	default:
		return "You can't go that way.", nil
	}
}
