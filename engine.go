package main

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/veilstream/psql-text-based-adventure/core/actions"
	"github.com/veilstream/psql-text-based-adventure/core/items"
	"github.com/veilstream/psql-text-based-adventure/core/world"
	"strings"
)

type Engine struct {
	Worlds       []*world.World
	currentWorld *world.World
	psqlBackend  *pgproto3.Backend
	gameOver     bool
}

func (engine *Engine) handleQuery(query string) {
	fmt.Printf("Handling query: %s\n", query)

	if strings.HasPrefix(query, "drop") {
		engine.Drop(query)
		return
	}

	if strings.HasPrefix(query, "teleport") {
		engine.Teleport(query)
		return
	}

	if strings.HasPrefix(query, "look") {
		engine.Look()
		return
	}

	if strings.HasPrefix(query, "commands") {
		engine.Commands()
		return
	}

	if strings.HasPrefix(query, "take") {
		engine.Take(query)
		return
	}

	if strings.HasPrefix(query, "inventory") {
		engine.Inventory()
		return
	}

	if strings.HasPrefix(query, "go") {
		engine.Go(query)
		return
	}

	if strings.HasPrefix(query, "use") {
		engine.Use(query)
		return
	}

	if strings.HasPrefix(query, "examine") {
		engine.Examine(query)
		return
	}

	if strings.HasPrefix(query, "talk to") {
		engine.TalkTo(query)
		return
	}

	// Unknown command
	engine.psqlBackend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  fmt.Sprintf("I don't know how to '%s'", query),
	})
}

func (engine *Engine) Say(msg string) {
	engine.psqlBackend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  msg,
	})
	err := engine.psqlBackend.Flush()
	if err != nil {
		fmt.Printf("Error flushing psql backend: %v\n", err)
		return
	}
}

func (engine *Engine) Sayf(format string, a ...any) {
	engine.Say(fmt.Sprintf(format, a...))
}

func (engine *Engine) Look() {
	action := actions.LookAction{}
	action.Execute(engine.psqlBackend, engine.currentWorld)

	loc := *engine.currentWorld.CurrentLocation
	if len(loc.ItemsOnGround()) > 0 {
		engine.Say("You see the following items on the ground:")
		for _, item := range loc.ItemsOnGround() {
			engine.Sayf(" - %s", item.Name())
		}
	}
}

func (engine *Engine) Commands() {
	action := actions.ListCommandsAction{}
	action.Execute(engine.psqlBackend, engine.currentWorld)
}

func (engine *Engine) Drop(query string) {
	itemName := strings.TrimSpace(strings.TrimPrefix(query, "drop"))
	loc := *engine.currentWorld.CurrentLocation

	// Drop a teleportation stone
	if itemName == "teleportation stone" {
		item := engine.currentWorld.Inventory.PeekItem("Bag of Rune Stones of Teleportation")
		if item == nil {
			engine.Say("You do not have a Bag of Rune Stones of Teleportation")
			return
		}
		bagOfTele, ok := item.(*items.BagOfTeleportationRuneStones)
		if !ok {
			engine.Say("You do not have a Bag of Rune Stones of Teleportation")
			return
		}
		engine.Say("You drop the teleportation stone. Looking around you try and memorize the location.")
		engine.Say("To teleport to this location, use the command 'teleport to <location name>' what would you like to name this location?")

		locationName, err := engine.ReceiveText()
		if err != nil {
			engine.Say("Error receiving location name")
			return
		}
		engine.Sayf("You have named this location '%s'", locationName)

		teleportationStone := items.NewPlacedTeleportationStone(locationName, loc)
		loc.DropItemByName(teleportationStone)
		bagOfTele.AddTeleportationStone(*teleportationStone)

		return
	}

	// Drop a normal Item
	has := engine.currentWorld.Inventory.HaveItem(itemName)
	if !has {
		engine.Sayf("You do not have a '%s' in your inventory", itemName)
		return
	}

	item := engine.currentWorld.Inventory.RemoveItem(itemName)
	loc.DropItemByName(item)
}

func (engine *Engine) ReceiveText() (string, error) {
	engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
	err := engine.psqlBackend.Flush()
	if err != nil {
		fmt.Printf("Error flushing psql backend: %v\n", err)
		return "", err
	}

	msg, err := engine.psqlBackend.Receive()
	fmt.Printf("received message: %v\n", msg)
	if err != nil {
		fmt.Printf("Error receiving message: %v\n", err)
		return "", err
	}
	if query, ok := msg.(*pgproto3.Query); ok {
		return strings.TrimSpace(strings.TrimSuffix(query.String, ";")), nil
	}
	return "", fmt.Errorf("Message is not a Query")
}

func (engine *Engine) Take(query string) {
	itemName := strings.TrimSpace(strings.TrimPrefix(query, "take"))
	loc := *engine.currentWorld.CurrentLocation

	item, msg := loc.TakeItemByName(itemName)
	if item != nil {
		engine.Say(msg)
		engine.currentWorld.Inventory.AddItem(item)
		return
	}

	item = loc.PickUpItemFromGround(itemName)
	if item != nil {
		engine.Sayf("You pick up the %s", item.Name())
		engine.currentWorld.Inventory.AddItem(item)
		return
	}

	engine.Sayf("You can't take that item.")
}

func (engine *Engine) Inventory() {
	items := engine.currentWorld.Inventory.ListItems()
	if len(items) == 0 {
		engine.Say("Your inventory is empty.")
		return
	}
	inventoryList := "You have the following items in your inventory:"
	for _, item := range items {
		inventoryList = fmt.Sprintf("%s, %s", inventoryList, item.Name())
	}
	engine.Say(inventoryList)

}

func (engine *Engine) Go(query string) {
	locationName := strings.TrimSpace(strings.TrimPrefix(query, "go"))
	loc := *engine.currentWorld.CurrentLocation

	fmt.Printf("attempting to go to '%s'\n", locationName)
	msg, newLocation := loc.Go(locationName)
	engine.Say(msg)
	if newLocation != nil {
		engine.currentWorld.CurrentLocation = newLocation
		engine.Look()
		return
	}
	if loc.GetDead() {
		engine.Sayf("You are dead. You cannot go %s", locationName)
		engine.Sayf("You go %s", locationName)
		engine.Say(msg)
		engine.gameOver = true
		return
	}
}

func (engine *Engine) Use(query string) {
	removedUseStr := strings.TrimSpace(strings.TrimPrefix(query, "use"))
	parts := strings.Split(removedUseStr, " on ")
	if len(parts) != 2 {
		engine.Say("Usage: use <item> on <target>")
		return
	}
	itemName := strings.TrimSpace(parts[0])
	targetName := strings.TrimSpace(parts[1])
	item := engine.currentWorld.Inventory.RemoveItem(itemName)
	if item == nil {
		engine.Sayf("You do not have a  '%s' in your inventory", itemName)
		return
	}
	loc := *engine.currentWorld.CurrentLocation
	msg, keep := loc.UseItem(item, targetName)
	engine.Say(msg)
	if keep {
		engine.currentWorld.Inventory.AddItem(item)
	}
}

func (engine *Engine) Examine(query string) {
	itemName := strings.TrimSpace(strings.TrimPrefix(query, "examine"))
	// First try inventory
	for _, item := range engine.currentWorld.Inventory.ListItems() {
		if item.Name() == itemName {
			engine.Say(item.Examine())
			return
		}
	}
	// Then try the location
	loc := *engine.currentWorld.CurrentLocation
	engine.Say(loc.Examine(itemName))
}

func (engine *Engine) TalkTo(query string) {
	name := strings.TrimSpace(strings.TrimPrefix(query, "talk to"))
	loc := *engine.currentWorld.CurrentLocation
	response := loc.TalkTo(name)
	if response == "" {
		engine.Sayf("There's no one named '%s' here to talk to.", name)
		return
	}
	engine.Say(response)
}

func (engine *Engine) Teleport(query string) {
	name := strings.TrimSpace(strings.TrimPrefix(query, "teleport to"))
	bag := engine.currentWorld.Inventory.PeekItem("Bag of Rune Stones of Teleportation")
	if bag == nil {
		engine.Say("You do not have a Bag of Rune Stones of Teleportation")
		return
	}
	bagOfTele, ok := bag.(*items.BagOfTeleportationRuneStones)
	if !ok {
		engine.Say("You do not have a Bag of Rune Stones of Teleportation")
		return
	}
	newLoc := bagOfTele.GetTeleportationStone(name)
	if newLoc == nil {
		engine.Sayf("You do not have a teleportation stone named '%s'", name)
		engine.Sayf("To teleport to a location, use the command 'teleport to <location name>'")
		for _, stone := range bagOfTele.PlacedStones {
			engine.Sayf(" - %s", stone.LocationName)
		}
		return
	}
	engine.Sayf("You close your eyes and think of the location you wanted to go, you remember it's name as '%s'. You feel the wind rush around you and pull you in a direction you did not know existed. You feel weightless for a moment, then it all stops, and your feet feel like they touched solid ground. You open your eyes.", newLoc.LocationName)
	engine.currentWorld.CurrentLocation = &newLoc.Location
	engine.Look()
}
