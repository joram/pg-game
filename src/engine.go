package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

type Engine struct {
	psqlBackend  *pgproto3.Backend
	db *pgxpool.Pool
	llm anthropic.Client
	model string
}

// GameResponse represents the structured JSON response from the LLM
type GameResponse struct {
	DungeonMasterResponse string        `json:"dungeon_master_response"`
	ItemsToAdd           []ItemUpdate   `json:"items_to_add,omitempty"`
	ItemsToUpdate        []ItemUpdate   `json:"items_to_update,omitempty"`
	ItemsToRemove        []int          `json:"items_to_remove,omitempty"`
	ItemsToAddToInventory []int         `json:"items_to_add_to_inventory,omitempty"` // Item IDs to add to player's inventory
	ItemsToRemoveFromInventory []int     `json:"items_to_remove_from_inventory,omitempty"` // Item IDs to remove from player's inventory
	NpcsToAdd            []NPCUpdate    `json:"npcs_to_add,omitempty"`
	NpcsToUpdate         []NPCUpdate    `json:"npcs_to_update,omitempty"`
	NpcsToRemove         []int          `json:"npcs_to_remove,omitempty"`
	NpcInteractions      []NPCInteraction `json:"npc_interactions,omitempty"` // New interactions to record
	LocationsToAdd       []LocationUpdate `json:"locations_to_add,omitempty"`
	LocationsToUpdate    []LocationUpdate `json:"locations_to_update,omitempty"`
	PlayerStateUpdates   map[string]interface{} `json:"player_state_updates,omitempty"`
}

type ItemUpdate struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LocationID  int    `json:"location_id,omitempty"`
}

type NPCUpdate struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LocationID  int    `json:"location_id,omitempty"`
}

type LocationUpdate struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type NPCInteraction struct {
	NpcID      int    `json:"npc_id"`      // Required: ID of the NPC
	PlayerID   int    `json:"player_id,omitempty"` // Optional: defaults to 1
	Interaction string `json:"interaction"` // Required: description of what happened
	Sentiment  string `json:"sentiment,omitempty"` // Optional: "positive", "negative", "neutral"
}

func NewEngine(psqlBackend *pgproto3.Backend) *Engine {
	psqlBackend.Send(&pgproto3.AuthenticationOk{})
	psqlBackend.Send(&pgproto3.ParameterStatus{Name: "server_version", Value: "16.8"})
	psqlBackend.Send(&pgproto3.ParameterStatus{Name: "client_encoding", Value: "UTF8"})
	psqlBackend.Send(&pgproto3.BackendKeyData{ProcessID: 1234, SecretKey: 5678})
	psqlBackend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
	err := psqlBackend.Flush()
	if err != nil {
		fmt.Printf("Error flushing psql backend: %v\n", err)
		return nil
	}

	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return nil
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Printf("Warning: ANTHROPIC_API_KEY not set\n")
	}

	llmClient := anthropic.NewClient(option.WithAPIKey(apiKey))

	return &Engine{
		psqlBackend: psqlBackend,
		db: db,
		llm: llmClient,
		model: "claude-opus-4-5-20251101",
	}
}


func (engine *Engine) Run() error {

	// INIT the database
	engine.initDatabase()

	// Run the game loop
	for {
		msg, err := engine.psqlBackend.Receive()
		if err != nil {
		if err == io.EOF {
			fmt.Printf("Client disconnected\n")
			return nil
		}
		fmt.Printf("Error receiving message: %v\n", err)
		return err
		}

		switch m := msg.(type) {
		case *pgproto3.Query:
			query := m.String
			fmt.Printf("Received: %s\n", query)
			if strings.HasPrefix(query, "SELECT version()") {
				engine.psqlBackend.Send(&pgproto3.RowDescription{
					Fields: []pgproto3.FieldDescription{
						{Name: []byte("version")},
					},
				})
				engine.psqlBackend.Send(&pgproto3.DataRow{
					Values: [][]byte{[]byte("foo")},
				})
				engine.psqlBackend.Send(&pgproto3.CommandComplete{
					CommandTag: []byte("SELECT 1"),
				})
				engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
				err := engine.psqlBackend.Flush()
				if err != nil {
					fmt.Printf("Error flushing psql backend: %v\n", err)
					return err
				}
				continue
			}
			if strings.HasPrefix(query, "SELECT ") && strings.HasSuffix(query, "as type;") {
				engine.psqlBackend.Send(&pgproto3.RowDescription{
					Fields: []pgproto3.FieldDescription{
						{Name: []byte("type")},
					},
				})
				engine.psqlBackend.Send(&pgproto3.DataRow{
					Values: [][]byte{
						[]byte("log"),
					},
				})
				engine.psqlBackend.Send(&pgproto3.CommandComplete{
					CommandTag: []byte("SELECT 1"),
				})
				engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
				err = engine.psqlBackend.Flush()
				if err != nil {
					fmt.Printf("Error flushing psql backend: %v\n", err)
					return err
				}
				continue
			}
			if strings.HasPrefix(query, "SELECT ") {
				engine.psqlBackend.Send(&pgproto3.RowDescription{
					Fields: []pgproto3.FieldDescription{},
				})
				engine.psqlBackend.Send(&pgproto3.DataRow{
					Values: [][]byte{},
				})
				engine.psqlBackend.Send(&pgproto3.CommandComplete{
					CommandTag: []byte("SELECT 0"),
				})
				engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
				err = engine.psqlBackend.Flush()
				if err != nil {
					fmt.Printf("Error flushing psql backend: %v\n", err)
					return err
				}
				continue
			}
			if strings.HasPrefix(query, "SET ") {
				engine.psqlBackend.Send(&pgproto3.CommandComplete{
					CommandTag: []byte("SET"),
				})
				engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
				err = engine.psqlBackend.Flush()
				if err != nil {
					fmt.Printf("Error flushing psql backend: %v\n", err)
					return err
				}
				continue
			}
		query = strings.Replace(query, "\n", "", -1)
		query = strings.Replace(query, ";", "", -1)
		engine.handleQuery(query)

		engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
		err := engine.psqlBackend.Flush()
		if err != nil {
			fmt.Printf("Error flushing psql backend: %v\n", err)
			return err
		}

		case *pgproto3.Terminate:
			fmt.Printf("Client disconnected\n")
		case *pgproto3.Sync:
			fmt.Printf("Received Sync message\n")
			engine.psqlBackend.Send(&pgproto3.ReadyForQuery{})
			err := engine.psqlBackend.Flush()
			if err != nil {
				fmt.Printf("Error flushing psql backend: %v\n", err)
			}
		default:
			fmt.Printf("Unhandled message type: %T\n", m)
		}
	}
}


func (engine *Engine) Sayf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
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


func (engine *Engine) handleQuery(query string) {
	world := engine.getWorld()
	items := engine.getItems()
	worldItems := engine.getWorldItems()
	
	// Get current player location
	currentLocationID := engine.getCurrentPlayerLocation()
	npcs := engine.getNpcsForLocation(currentLocationID)

	jsonSchema := `{
		"type": "object",
		"properties": {
			"dungeon_master_response": {
			"type": "string",
			"description": "The narrative response to show the player describing what happens"
			},
			"items_to_add": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
				"name": {"type": "string"},
				"description": {"type": "string"},
				"location_id": {"type": "integer"}
				},
				"required": ["name", "description"]
			}
			},
			"items_to_update": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
				"id": {"type": "integer"},
				"name": {"type": "string"},
				"description": {"type": "string"},
				"location_id": {"type": "integer"}
				},
				"required": ["id"]
			}
			},
			"items_to_remove": {
			"type": "array",
			"items": {"type": "integer"}
			},
			"items_to_add_to_inventory": {
			"type": "array",
			"items": {"type": "integer"},
			"description": "Item IDs to add to the player's inventory (must exist in items table)"
			},
			"items_to_remove_from_inventory": {
			"type": "array",
			"items": {"type": "integer"},
			"description": "Item IDs to remove from the player's inventory"
			},
			"npcs_to_add": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
				"name": {"type": "string"},
				"description": {"type": "string"},
				"location_id": {"type": "integer"}
				},
				"required": ["name", "description"]
			}
			},
			"npcs_to_update": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
				"id": {"type": "integer"},
				"name": {"type": "string"},
				"description": {"type": "string"},
				"location_id": {"type": "integer"}
				},
				"required": ["id"]
			}
			},
			"npcs_to_remove": {
			"type": "array",
			"items": {"type": "integer"}
			},
			"npc_interactions": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
				"npc_id": {"type": "integer"},
				"player_id": {"type": "integer"},
				"interaction": {"type": "string"},
				"sentiment": {"type": "string", "enum": ["positive", "negative", "neutral"]}
				},
				"required": ["npc_id", "interaction"]
			},
			"description": "Record new interactions between NPCs and the player. Use this when the player talks to, helps, or interacts with an NPC."
			},
			"locations_to_add": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
				"name": {"type": "string"},
				"description": {"type": "string"}
				},
				"required": ["name", "description"]
			}
			},
			"locations_to_update": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
				"id": {"type": "integer"},
				"name": {"type": "string"},
				"description": {"type": "string"}
				},
				"required": ["id"]
			}
			}
		},
		"required": ["dungeon_master_response"]
	}`

	// Get current player location for context
	var locationContext string
	if currentLocationID > 0 {
		var locationName string
		err := engine.db.QueryRow(context.Background(), "SELECT name FROM locations WHERE id = $1", currentLocationID).Scan(&locationName)
		if err == nil {
			locationContext = fmt.Sprintf("\n## Current Player Location: %s (ID: %d)\nNote: Only NPCs in this location will show their interaction history with the player.", locationName, currentLocationID)
		}
	}
	
	systemPrompt := fmt.Sprintf(`You are a dungeon master for a text adventure game. You must respond ONLY with valid JSON in the exact format specified below.

# Current World State
## Locations:
%s%s

## Items in Inventory:
%s

## Items in the World:
%s

## NPCs (in current location with interaction history):
%s

IMPORTANT: The "Interaction History" shown for each NPC contains the actual recorded history of interactions between the player and that NPC. When the player asks about their history with an NPC, you MUST reference the specific interactions listed in the Interaction History. Do not make up or ignore the interaction history - it is the factual record of what has happened.

# Response Format
You MUST respond with a JSON object matching this schema:
%s

# Rules:
1. Respond ONLY with valid JSON - no markdown, no code blocks, no explanations
2. The dungeon_master_response field is REQUIRED and should be the narrative description of what happens
3. Only include fields that have changes (omit empty arrays)
4. When adding items/NPCs/locations, provide name and description
5. When updating, you must include the id field
6. When removing, provide the id in the appropriate _to_remove array
7. When a player takes/picks up an item, add the item ID to items_to_add_to_inventory array
8. When a player drops/loses an item, add the item ID to items_to_remove_from_inventory array
9. When the player interacts with an NPC (talks, helps, threatens, etc.), add an entry to npc_interactions with:
   - npc_id: The ID of the NPC
   - interaction: A brief description of what happened (e.g., "Player was friendly and helpful", "Player insulted the NPC", "Player gave the NPC a gift")
   - sentiment: "positive", "negative", or "neutral" based on how the NPC would perceive the interaction
10. When the player moves to a new location, update player_state_updates with {"current_location_id": <location_id>}
11. NPCs remember past interactions - ALWAYS use the interaction history shown above to inform their responses
12. When the player asks about their history with an NPC, reference the specific interactions from the Interaction History field
13. Only NPCs in the player's current location will show their interaction history - this helps focus on relevant NPCs
14. Be creative and respond to player actions appropriately`, world, locationContext, items, worldItems, npcs, jsonSchema)

	response, err := engine.llm.Messages.New(
		context.Background(),
		anthropic.MessageNewParams{
			Model: anthropic.Model(engine.model),
			MaxTokens: 2048,
			System: []anthropic.TextBlockParam{
				{Text: systemPrompt},
			},
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(
					anthropic.NewTextBlock(fmt.Sprintf("Player action: %s\n\nRespond with JSON only:", query)),
				),
			},
		},
	)
	
	if err != nil {
		errStr := err.Error()
		fmt.Printf("Error calling LLM: %v\n", err)
		
		if strings.Contains(errStr, "429") || strings.Contains(errStr, "quota") || strings.Contains(errStr, "billing") {
			engine.Sayf("I'm sorry, but I'm unable to process your request right now due to API quota limits. Please check your Anthropic account billing and quota settings at https://console.anthropic.com/")
		} else if strings.Contains(errStr, "401") || strings.Contains(errStr, "403") {
			engine.Sayf("Authentication error with Anthropic API. Please check your API key in the .env file.")
		} else {
			engine.Sayf("I encountered an error processing your request: %v. Please try again later.", err)
		}
		return
	}
	
	// Extract text content from Claude's response
	var responseText string
	for _, content := range response.Content {
		if textBlock := content.AsText(); textBlock.Text != "" {
			responseText += textBlock.Text
		}
	}
	
	// Parse JSON from response (may be wrapped in markdown code blocks)
	jsonStr := engine.extractJSON(responseText)
	
	var gameResponse GameResponse
	if err := json.Unmarshal([]byte(jsonStr), &gameResponse); err != nil {
		fmt.Printf("Error parsing JSON response: %v\nRaw response: %s\n", err, responseText)
		// Fallback: show raw response if JSON parsing fails
		engine.Sayf("Error parsing game response. Raw: %s", responseText)
		return
	}
	
	// Update database based on the response
	engine.applyGameUpdates(&gameResponse)
	
	// Show the dungeon master response to the user
	engine.Sayf(gameResponse.DungeonMasterResponse)
}

// extractJSON extracts JSON from a string, handling markdown code blocks
func (engine *Engine) extractJSON(text string) string {
	// Remove markdown code blocks if present
	jsonBlockRegex := regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*?\\})\\s*```")
	matches := jsonBlockRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	
	// Try to find JSON object directly
	jsonObjRegex := regexp.MustCompile("(\\{.*\\})")
	matches = jsonObjRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	
	// Return as-is if no pattern matches
	return strings.TrimSpace(text)
}

// applyGameUpdates applies the game state changes from the LLM response to the database
func (engine *Engine) applyGameUpdates(response *GameResponse) {
	ctx := context.Background()
	
	// Upsert items (combine add and update)
	allItems := append(response.ItemsToAdd, response.ItemsToUpdate...)
	for _, item := range allItems {
		if item.ID > 0 {
			// Update existing item
			_, err := engine.db.Exec(ctx,
				`INSERT INTO items (id, name, description, location_id) 
				 VALUES ($1, $2, $3, CASE WHEN $4 > 0 AND EXISTS(SELECT 1 FROM locations WHERE id = $4) THEN $4 ELSE NULL END)
				 ON CONFLICT (id) 
				 DO UPDATE SET 
				   name = COALESCE(EXCLUDED.name, items.name),
				   description = COALESCE(EXCLUDED.description, items.description),
				   location_id = CASE 
				     WHEN EXCLUDED.location_id > 0 AND EXISTS(SELECT 1 FROM locations WHERE id = EXCLUDED.location_id) 
				     THEN EXCLUDED.location_id 
				     ELSE items.location_id 
				   END`,
				item.ID, item.Name, item.Description, item.LocationID,
			)
			if err != nil {
				fmt.Printf("Error upserting item %d: %v\n", item.ID, err)
			} else {
				fmt.Printf("Upserted item ID %d: %s\n", item.ID, item.Name)
			}
		} else {
			// Insert new item - handle location_id: use NULL if 0 or invalid
			var locationID interface{}
			if item.LocationID > 0 {
				// Verify location exists before using it
				var exists bool
				err := engine.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1)", item.LocationID).Scan(&exists)
				if err == nil && exists {
					locationID = item.LocationID
				} else {
					locationID = nil
				}
			} else {
				locationID = nil
			}
			
			_, err := engine.db.Exec(ctx,
				"INSERT INTO items (name, description, location_id) VALUES ($1, $2, $3)",
				item.Name, item.Description, locationID,
			)
			if err != nil {
				fmt.Printf("Error adding item %s: %v\n", item.Name, err)
			} else {
				fmt.Printf("Added item: %s\n", item.Name)
			}
		}
	}
	
	// Remove items
	for _, itemID := range response.ItemsToRemove {
		_, err := engine.db.Exec(ctx, "DELETE FROM items WHERE id = $1", itemID)
		if err != nil {
			fmt.Printf("Error removing item %d: %v\n", itemID, err)
		} else {
			fmt.Printf("Removed item ID: %d\n", itemID)
		}
	}
	
	// Ensure default player exists before adding items to inventory
	engine.ensureDefaultPlayer(ctx)
	
	// Add items to player inventory (player_id = 1 for now)
	for _, itemID := range response.ItemsToAddToInventory {
		// Check if item exists and is not already in inventory
		var exists bool
		err := engine.db.QueryRow(ctx, 
			"SELECT EXISTS(SELECT 1 FROM items WHERE id = $1)", 
			itemID,
		).Scan(&exists)
		if err != nil {
			fmt.Printf("Error checking if item %d exists: %v\n", itemID, err)
			continue
		}
		if !exists {
			fmt.Printf("Warning: Item %d does not exist, skipping inventory add\n", itemID)
			continue
		}
		
		// Check if already in inventory
		var inInventory bool
		err = engine.db.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM player_items WHERE player_id = 1 AND item_id = $1)",
			itemID,
		).Scan(&inInventory)
		if err != nil {
			fmt.Printf("Error checking inventory for item %d: %v\n", itemID, err)
			continue
		}
		
		if !inInventory {
			_, err = engine.db.Exec(ctx,
				"INSERT INTO player_items (player_id, item_id) VALUES (1, $1) ON CONFLICT DO NOTHING",
				itemID,
			)
			if err != nil {
				fmt.Printf("Error adding item %d to inventory: %v\n", itemID, err)
			} else {
				fmt.Printf("Added item %d to player inventory\n", itemID)
			}
		} else {
			fmt.Printf("Item %d already in inventory, skipping\n", itemID)
		}
	}
	
	// Remove items from player inventory
	for _, itemID := range response.ItemsToRemoveFromInventory {
		_, err := engine.db.Exec(ctx,
			"DELETE FROM player_items WHERE player_id = 1 AND item_id = $1",
			itemID,
		)
		if err != nil {
			fmt.Printf("Error removing item %d from inventory: %v\n", itemID, err)
		} else {
			fmt.Printf("Removed item %d from player inventory\n", itemID)
		}
	}
	
	// Upsert NPCs (combine add and update)
	allNpcs := append(response.NpcsToAdd, response.NpcsToUpdate...)
	for _, npc := range allNpcs {
		if npc.ID > 0 {
			// Update existing NPC
			_, err := engine.db.Exec(ctx,
				`INSERT INTO npcs (id, name, description, location_id) 
				 VALUES ($1, $2, $3, CASE WHEN $4 > 0 AND EXISTS(SELECT 1 FROM locations WHERE id = $4) THEN $4 ELSE NULL END)
				 ON CONFLICT (id) 
				 DO UPDATE SET 
				   name = COALESCE(EXCLUDED.name, npcs.name),
				   description = COALESCE(EXCLUDED.description, npcs.description),
				   location_id = CASE 
				     WHEN EXCLUDED.location_id > 0 AND EXISTS(SELECT 1 FROM locations WHERE id = EXCLUDED.location_id) 
				     THEN EXCLUDED.location_id 
				     ELSE npcs.location_id 
				   END`,
				npc.ID, npc.Name, npc.Description, npc.LocationID,
			)
			if err != nil {
				fmt.Printf("Error upserting NPC %d: %v\n", npc.ID, err)
			} else {
				fmt.Printf("Upserted NPC ID %d: %s\n", npc.ID, npc.Name)
			}
		} else {
			// Insert new NPC - handle location_id: use NULL if 0 or invalid
			var locationID interface{}
			if npc.LocationID > 0 {
				// Verify location exists before using it
				var exists bool
				err := engine.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1)", npc.LocationID).Scan(&exists)
				if err == nil && exists {
					locationID = npc.LocationID
				} else {
					locationID = nil
				}
			} else {
				locationID = nil
			}
			
			_, err := engine.db.Exec(ctx,
				"INSERT INTO npcs (name, description, location_id) VALUES ($1, $2, $3)",
				npc.Name, npc.Description, locationID,
			)
			if err != nil {
				fmt.Printf("Error adding NPC %s: %v\n", npc.Name, err)
			} else {
				fmt.Printf("Added NPC: %s\n", npc.Name)
			}
		}
	}
	
	// Remove NPCs
	for _, npcID := range response.NpcsToRemove {
		_, err := engine.db.Exec(ctx, "DELETE FROM npcs WHERE id = $1", npcID)
		if err != nil {
			fmt.Printf("Error removing NPC %d: %v\n", npcID, err)
		} else {
			fmt.Printf("Removed NPC ID: %d\n", npcID)
		}
	}
	
	// Save NPC interactions
	for _, interaction := range response.NpcInteractions {
		playerID := interaction.PlayerID
		if playerID == 0 {
			playerID = 1 // Default to player 1
		}
		
		// Verify NPC exists
		var npcExists bool
		err := engine.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM npcs WHERE id = $1)", interaction.NpcID).Scan(&npcExists)
		if err != nil {
			fmt.Printf("Error checking if NPC %d exists: %v\n", interaction.NpcID, err)
			continue
		}
		if !npcExists {
			fmt.Printf("Warning: NPC %d does not exist, skipping interaction\n", interaction.NpcID)
			continue
		}
		
		// Verify player exists
		var playerExists bool
		err = engine.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM players WHERE id = $1)", playerID).Scan(&playerExists)
		if err != nil {
			fmt.Printf("Error checking if player %d exists: %v\n", playerID, err)
			continue
		}
		if !playerExists {
			fmt.Printf("Warning: Player %d does not exist, skipping interaction\n", playerID)
			continue
		}
		
		// Validate sentiment if provided
		sentiment := interaction.Sentiment
		if sentiment != "" && sentiment != "positive" && sentiment != "negative" && sentiment != "neutral" {
			sentiment = "neutral" // Default to neutral if invalid
		}
		
		_, err = engine.db.Exec(ctx,
			"INSERT INTO npc_player_interactions (npc_id, player_id, interaction, sentiment) VALUES ($1, $2, $3, $4)",
			interaction.NpcID, playerID, interaction.Interaction, sentiment,
		)
		if err != nil {
			fmt.Printf("Error saving interaction with NPC %d: %v\n", interaction.NpcID, err)
		} else {
			fmt.Printf("Recorded interaction: NPC %d - %s [%s]\n", interaction.NpcID, interaction.Interaction, sentiment)
		}
	}
	
	// Handle player state updates (including location changes)
	if response.PlayerStateUpdates != nil {
		if currentLocationID, ok := response.PlayerStateUpdates["current_location_id"].(float64); ok {
			locationIDInt := int(currentLocationID)
			// Verify location exists
			var exists bool
			err := engine.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1)", locationIDInt).Scan(&exists)
			if err == nil && exists {
				_, err = engine.db.Exec(ctx, "UPDATE players SET current_location_id = $1 WHERE id = 1", locationIDInt)
				if err != nil {
					fmt.Printf("Error updating player location: %v\n", err)
				} else {
					fmt.Printf("Updated player location to %d\n", locationIDInt)
				}
			} else if err == nil {
				fmt.Printf("Warning: Location %d does not exist, skipping location update\n", locationIDInt)
			}
		}
		// Handle other player state updates here if needed
	}
	
	// Upsert locations (combine add and update)
	allLocations := append(response.LocationsToAdd, response.LocationsToUpdate...)
	for _, location := range allLocations {
		if location.ID > 0 {
			// Update existing location
			_, err := engine.db.Exec(ctx,
				`INSERT INTO locations (id, name, description) 
				 VALUES ($1, $2, $3)
				 ON CONFLICT (id) 
				 DO UPDATE SET 
				   name = COALESCE(EXCLUDED.name, locations.name),
				   description = COALESCE(EXCLUDED.description, locations.description)`,
				location.ID, location.Name, location.Description,
			)
			if err != nil {
				fmt.Printf("Error upserting location %d: %v\n", location.ID, err)
			} else {
				fmt.Printf("Upserted location ID %d: %s\n", location.ID, location.Name)
			}
		} else {
			// Insert new location
			_, err := engine.db.Exec(ctx,
				"INSERT INTO locations (name, description) VALUES ($1, $2)",
				location.Name, location.Description,
			)
			if err != nil {
				fmt.Printf("Error adding location %s: %v\n", location.Name, err)
			} else {
				fmt.Printf("Added location: %s\n", location.Name)
			}
		}
	}
}

func (engine *Engine) getWorld() string {
	ctx := context.Background()
	var locations []string
	
	rows, err := engine.db.Query(ctx, "SELECT name, description FROM locations ORDER BY id")
	if err != nil {
		fmt.Printf("Error querying locations: %v\n", err)
		return "Unable to load world information."
	}
	defer rows.Close()
	
	for rows.Next() {
		var name, description string
		if err := rows.Scan(&name, &description); err != nil {
			continue
		}
		locations = append(locations, fmt.Sprintf("%s: %s", name, description))
	}
	
	if len(locations) == 0 {
		return "No locations found in the world."
	}
	
	return strings.Join(locations, "\n\n")
}

func (engine *Engine) getItems() string {
	ctx := context.Background()
	var items []string
	
	// Get items in player's inventory (items linked via player_items table)
	// For now, using player_id = 1 as the default player
	// TODO: Track current player_id per connection
	rows, err := engine.db.Query(ctx, `
		SELECT i.id, i.name, i.description 
		FROM items i
		INNER JOIN player_items pi ON i.id = pi.item_id
		WHERE pi.player_id = 1
		ORDER BY i.id
	`)
	if err != nil {
		fmt.Printf("Error querying inventory items: %v\n", err)
		return "Unable to load inventory items."
	}
	defer rows.Close()
	
	for rows.Next() {
		var id int
		var name, description string
		if err := rows.Scan(&id, &name, &description); err != nil {
			continue
		}
		items = append(items, fmt.Sprintf("ID %d: %s: %s", id, name, description))
	}
	
	if len(items) == 0 {
		return "No items found in inventory."
	}
	
	return strings.Join(items, "\n\n")
}


func (engine *Engine) getWorldItems() string {
	ctx := context.Background()
	var items []string
	
	rows, err := engine.db.Query(ctx, `
		SELECT i.id, i.name, i.description, l.name as location_name 
		FROM items i 
		LEFT JOIN locations l ON i.location_id = l.id 
		ORDER BY i.id
	`)
	if err != nil {
		fmt.Printf("Error querying items: %v\n", err)
		return "Unable to load items."
	}
	defer rows.Close()
	
	for rows.Next() {
		var id int
		var name, description, locationName string
		if err := rows.Scan(&id, &name, &description, &locationName); err != nil {
			continue
		}
		if locationName != "" {
			items = append(items, fmt.Sprintf("ID %d: %s (at %s): %s", id, name, locationName, description))
		} else {
			items = append(items, fmt.Sprintf("ID %d: %s: %s", id, name, description))
		}
	}
	
	if len(items) == 0 {
		return "No items found in the world."
	}
	
	return strings.Join(items, "\n\n")
}

// getCurrentPlayerLocation returns the current location ID for player 1
func (engine *Engine) getCurrentPlayerLocation() int {
	ctx := context.Background()
	var locationID int
	err := engine.db.QueryRow(ctx, 
		"SELECT COALESCE(current_location_id, 0) FROM players WHERE id = 1",
	).Scan(&locationID)
	if err != nil {
		fmt.Printf("Error getting current player location: %v\n", err)
		return 0
	}
	return locationID
}

// getNpcsForLocation returns NPCs in the specified location with their interaction history
// Only returns NPCs that are in the specified location (locationID > 0)
func (engine *Engine) getNpcsForLocation(locationID int) string {
	ctx := context.Background()
	var npcs []string
	
	// Only return NPCs if we have a valid location
	if locationID <= 0 {
		return "No NPCs at current location (player location not set)."
	}
	
	query := `
		SELECT n.id, n.name, n.description, l.name as location_name, n.location_id
		FROM npcs n 
		LEFT JOIN locations l ON n.location_id = l.id 
		WHERE n.location_id = $1
		ORDER BY n.id
	`
	
	rows, err := engine.db.Query(ctx, query, locationID)
	if err != nil {
		fmt.Printf("Error querying NPCs: %v\n", err)
		return "Unable to load NPCs."
	}
	defer rows.Close()
	
	for rows.Next() {
		var id, npcLocationID int
		var name, description, locationName string
		if err := rows.Scan(&id, &name, &description, &locationName, &npcLocationID); err != nil {
			continue
		}
		
		// Get interaction history for NPCs in the current location
		interactions := engine.getNPCInteractions(ctx, id, 1)
		
		npcStr := fmt.Sprintf("ID %d: %s", id, name)
		if locationName != "" {
			npcStr += fmt.Sprintf(" (at %s)", locationName)
		}
		npcStr += fmt.Sprintf(": %s", description)
		
		if interactions != "" {
			npcStr += fmt.Sprintf("\n  Interaction History with Player: %s", interactions)
			fmt.Printf("NPC %d (%s) has interaction history: %s\n", id, name, interactions)
		} else {
			fmt.Printf("NPC %d (%s) has no interaction history\n", id, name)
		}
		
		npcs = append(npcs, npcStr)
	}
	
	if len(npcs) == 0 {
		return fmt.Sprintf("No NPCs found at current location (ID: %d).", locationID)
	}
	
	return strings.Join(npcs, "\n\n")
}

// getNpcs returns all NPCs (kept for backward compatibility if needed)
func (engine *Engine) getNpcs() string {
	return engine.getNpcsForLocation(0) // 0 means all locations
}

// getNPCInteractions returns a formatted string of interaction history between an NPC and player
func (engine *Engine) getNPCInteractions(ctx context.Context, npcID, playerID int) string {
	// First check if any interactions exist
	var count int
	err := engine.db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM npc_player_interactions 
		WHERE npc_id = $1 AND player_id = $2
	`, npcID, playerID).Scan(&count)
	if err != nil {
		fmt.Printf("Error counting interactions for NPC %d, player %d: %v\n", npcID, playerID, err)
		return ""
	}
	if count == 0 {
		fmt.Printf("No interactions found for NPC %d, player %d (count: 0)\n", npcID, playerID)
		return ""
	}
	
	fmt.Printf("Found %d interactions for NPC %d, player %d, retrieving...\n", count, npcID, playerID)
	
	rows, err := engine.db.Query(ctx, `
		SELECT interaction, COALESCE(sentiment, '') as sentiment, created_at
		FROM npc_player_interactions
		WHERE npc_id = $1 AND player_id = $2
		ORDER BY created_at DESC
		LIMIT 5
	`, npcID, playerID)
	if err != nil {
		fmt.Printf("Error querying interactions for NPC %d, player %d: %v\n", npcID, playerID, err)
		return ""
	}
	defer rows.Close()
	
	var interactions []string
	rowCount := 0
	for rows.Next() {
		rowCount++
		var interaction, sentiment string
		var createdAt interface{} // We'll format this if needed
		if err := rows.Scan(&interaction, &sentiment, &createdAt); err != nil {
			fmt.Printf("Error scanning interaction row %d for NPC %d: %v\n", rowCount, npcID, err)
			continue
		}
		sentimentStr := ""
		if sentiment != "" {
			sentimentStr = fmt.Sprintf(" [%s]", sentiment)
		}
		interactions = append(interactions, fmt.Sprintf("%s%s", interaction, sentimentStr))
		fmt.Printf("Scanned interaction %d for NPC %d: %s%s\n", rowCount, npcID, interaction, sentimentStr)
	}
	
	if err := rows.Err(); err != nil {
		fmt.Printf("Error iterating interaction rows for NPC %d: %v\n", npcID, err)
	}
	
	if len(interactions) == 0 {
		fmt.Printf("No interactions after scanning for NPC %d (scanned %d rows)\n", npcID, rowCount)
		return ""
	}
	
	result := strings.Join(interactions, "; ")
	fmt.Printf("Retrieved %d interactions for NPC %d: %s\n", len(interactions), npcID, result)
	return result
}

func (engine *Engine) initDatabase() {
	ctx := context.Background()
	
	// Create tables
	queries := []string{
		"CREATE TABLE IF NOT EXISTS players (id SERIAL PRIMARY KEY, name VARCHAR(255), current_location_id INT REFERENCES locations(id) ON DELETE SET NULL)",
		"CREATE TABLE IF NOT EXISTS locations (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT)",
		"CREATE TABLE IF NOT EXISTS items (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT, location_id INT REFERENCES locations(id) ON DELETE SET NULL)",
		"CREATE TABLE IF NOT EXISTS npcs (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT, location_id INT REFERENCES locations(id) ON DELETE SET NULL)",
		"CREATE TABLE IF NOT EXISTS player_items (id SERIAL PRIMARY KEY, player_id INT REFERENCES players(id), item_id INT REFERENCES items(id))",
		"CREATE TABLE IF NOT EXISTS player_notes (id SERIAL PRIMARY KEY, player_id INT REFERENCES players(id), note TEXT)",
		"CREATE TABLE IF NOT EXISTS npc_player_interactions (id SERIAL PRIMARY KEY, npc_id INT REFERENCES npcs(id) ON DELETE CASCADE, player_id INT REFERENCES players(id) ON DELETE CASCADE, interaction TEXT, sentiment VARCHAR(20), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)",
	}
	for _, query := range queries {
		_, err := engine.db.Exec(ctx, query)
		if err != nil {
			fmt.Printf("Error executing query: %v\n", err)
			return
		}
	}

	// Ensure location_id columns allow NULL (migration for existing tables)
	alterQueries := []string{
		"ALTER TABLE items ALTER COLUMN location_id DROP NOT NULL",
		"ALTER TABLE npcs ALTER COLUMN location_id DROP NOT NULL",
	}
	for _, query := range alterQueries {
		_, err := engine.db.Exec(ctx, query)
		// Ignore errors - column might already allow NULL or not exist
		if err != nil {
			// Only log if it's not a "column does not exist" or "already correct" error
			if !strings.Contains(err.Error(), "does not exist") && !strings.Contains(err.Error(), "is not null") {
				fmt.Printf("Note: Could not alter column (may already be correct): %v\n", err)
			}
		}
	}
	
	// Add current_location_id to players table if it doesn't exist (migration)
	// Check if column exists first
	var columnExists bool
	var err2 error
	err2 = engine.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_schema = 'public'
			AND table_name = 'players' 
			AND column_name = 'current_location_id'
		)
	`).Scan(&columnExists)
	
	if err2 != nil {
		fmt.Printf("Error checking for current_location_id column: %v\n", err2)
	} else if !columnExists {
		// Column doesn't exist, add it
		// First, add the column without the foreign key constraint
		_, err2 = engine.db.Exec(ctx, `ALTER TABLE players ADD COLUMN current_location_id INT`)
		if err2 != nil {
			fmt.Printf("Error adding current_location_id column: %v\n", err2)
		} else {
			fmt.Printf("Added current_location_id column to players table\n")
			
			// Then add the foreign key constraint (if locations table exists)
			_, err2 = engine.db.Exec(ctx, `
				DO $$
				BEGIN
					IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'locations') THEN
						ALTER TABLE players 
						ADD CONSTRAINT players_current_location_id_fkey 
						FOREIGN KEY (current_location_id) 
						REFERENCES locations(id) 
						ON DELETE SET NULL;
					END IF;
				END $$;
			`)
			if err2 != nil {
				// Constraint might already exist
				fmt.Printf("Note: Could not add foreign key constraint (may already exist): %v\n", err2)
			}
		}
	}

	// Ensure default player exists (player_id = 1)
	engine.ensureDefaultPlayer(ctx)

	// Check if database is empty and seed default data
	var locationCount int
	err := engine.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM locations").Scan(&locationCount)
	if err != nil {
		fmt.Printf("Error checking location count: %v\n", err)
		return
	}

	if locationCount == 0 {
		fmt.Printf("Database is empty, seeding default data...\n")
		// Only seed if database is truly empty - don't drop existing tables
		engine.seedDefaultData()
	} else {
		fmt.Printf("Database already has %d location(s), skipping seed.\n", locationCount)
	}
}

// ensureDefaultPlayer creates a default player with ID 1 if it doesn't exist
func (engine *Engine) ensureDefaultPlayer(ctx context.Context) {
	var exists bool
	err := engine.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM players WHERE id = 1)").Scan(&exists)
	if err != nil {
		fmt.Printf("Error checking for default player: %v\n", err)
		return
	}
	
	if !exists {
		// Get the first location as default (or NULL if no locations exist)
		var defaultLocationID interface{}
		err := engine.db.QueryRow(ctx, "SELECT id FROM locations ORDER BY id LIMIT 1").Scan(&defaultLocationID)
		if err != nil {
			defaultLocationID = nil // No locations yet, will be set when location is created
		}
		
		// Insert player with ID 1
		_, err = engine.db.Exec(ctx, 
			"INSERT INTO players (id, name, current_location_id) VALUES (1, 'Player', $1) ON CONFLICT (id) DO NOTHING",
			defaultLocationID,
		)
		if err != nil {
			fmt.Printf("Error creating default player: %v\n", err)
			return
		}
		
		// Ensure the sequence is set correctly for future inserts
		_, err = engine.db.Exec(ctx,
			"SELECT setval('players_id_seq', GREATEST((SELECT MAX(id) FROM players), 1), true)",
		)
		if err != nil {
			// Ignore sequence errors - not critical
			fmt.Printf("Note: Could not update players sequence (may not exist): %v\n", err)
		}
		
		fmt.Printf("Created default player (ID: 1)\n")
	} else {
		// Ensure player has a location if one exists but player doesn't have one set
		var hasLocation bool
		err := engine.db.QueryRow(ctx, "SELECT current_location_id IS NOT NULL FROM players WHERE id = 1").Scan(&hasLocation)
		if err == nil && !hasLocation {
			var firstLocationID interface{}
			err := engine.db.QueryRow(ctx, "SELECT id FROM locations ORDER BY id LIMIT 1").Scan(&firstLocationID)
			if err == nil {
				_, err = engine.db.Exec(ctx, "UPDATE players SET current_location_id = $1 WHERE id = 1", firstLocationID)
				if err == nil {
					fmt.Printf("Set player's location to first available location\n")
				}
			}
		}
	}
}

func (engine *Engine) recreateTables() {
	ctx := context.Background()
	
	// Drop tables in reverse order of dependencies
	dropQueries := []string{
		"DROP TABLE IF EXISTS npc_player_interactions",
		"DROP TABLE IF EXISTS player_notes",
		"DROP TABLE IF EXISTS player_items",
		"DROP TABLE IF EXISTS npcs",
		"DROP TABLE IF EXISTS items",
		"DROP TABLE IF EXISTS locations",
		"DROP TABLE IF EXISTS players",
	}
	
	for _, query := range dropQueries {
		_, err := engine.db.Exec(ctx, query)
		if err != nil {
			fmt.Printf("Warning: Error dropping table: %v\n", err)
		}
	}
	
	// Recreate tables with correct schema
	createQueries := []string{
		"CREATE TABLE players (id SERIAL PRIMARY KEY, name VARCHAR(255), current_location_id INT REFERENCES locations(id) ON DELETE SET NULL)",
		"CREATE TABLE locations (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT)",
		"CREATE TABLE items (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT, location_id INT REFERENCES locations(id))",
		"CREATE TABLE npcs (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT, location_id INT REFERENCES locations(id))",
		"CREATE TABLE player_items (id SERIAL PRIMARY KEY, player_id INT REFERENCES players(id), item_id INT REFERENCES items(id))",
		"CREATE TABLE player_notes (id SERIAL PRIMARY KEY, player_id INT REFERENCES players(id), note TEXT)",
		"CREATE TABLE npc_player_interactions (id SERIAL PRIMARY KEY, npc_id INT REFERENCES npcs(id) ON DELETE CASCADE, player_id INT REFERENCES players(id) ON DELETE CASCADE, interaction TEXT, sentiment VARCHAR(20), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)",
	}
	
	for _, query := range createQueries {
		_, err := engine.db.Exec(ctx, query)
		if err != nil {
			fmt.Printf("Error recreating table: %v\n", err)
			return
		}
	}
	fmt.Printf("Tables recreated with correct schema.\n")
}

func (engine *Engine) seedDefaultData() {
	ctx := context.Background()

	// Insert default location
	var locationID int
	err := engine.db.QueryRow(
		ctx,
		"INSERT INTO locations (name, description) VALUES ($1, $2) RETURNING id",
		"The Old Tavern",
		"You find yourself standing in a dimly lit tavern. The air is thick with the smell of ale and wood smoke. A crackling fireplace casts dancing shadows on the weathered wooden walls. A few patrons sit at tables, but the place feels mostly empty. Behind the bar, you can see shelves lined with bottles and mugs. An old wooden door leads outside, and a narrow staircase in the corner leads to the upper floor.",
	).Scan(&locationID)
	if err != nil {
		fmt.Printf("Error inserting default location: %v\n", err)
		return
	}
	fmt.Printf("Created default location: The Old Tavern (ID: %d)\n", locationID)
	
	// Set player's initial location to the default location
	_, err = engine.db.Exec(ctx,
		"UPDATE players SET current_location_id = $1 WHERE id = 1",
		locationID,
	)
	if err != nil {
		fmt.Printf("Warning: Could not set player's initial location: %v\n", err)
	} else {
		fmt.Printf("Set player's initial location to The Old Tavern\n")
	}

	// Insert default items at this location
	items := []struct {
		name        string
		description string
	}{
		{
			name:        "Rusty Key",
			description: "An old, tarnished key with intricate patterns. It looks like it might unlock something important.",
		},
		{
			name:        "Leather Journal",
			description: "A worn leather-bound journal. The pages are filled with handwritten notes, though many are faded and hard to read.",
		},
		{
			name:        "Candle",
			description: "A simple tallow candle. It's been used before but still has plenty of wax left. It could provide light in dark places.",
		},
		{
			name:        "Copper Coin",
			description: "A single copper coin, worn smooth from years of use. It's not worth much, but every coin counts.",
		},
	}

	for _, item := range items {
		var itemID int
		err := engine.db.QueryRow(
			ctx,
			"INSERT INTO items (name, description, location_id) VALUES ($1, $2, $3) RETURNING id",
			item.name,
			item.description,
			locationID,
		).Scan(&itemID)
		if err != nil {
			fmt.Printf("Error inserting item %s: %v\n", item.name, err)
			continue
		}
		fmt.Printf("Created item: %s (ID: %d)\n", item.name, itemID)
	}

	// Insert default NPCs at this location
	npcs := []struct {
		name        string
		description string
	}{
		{
			name:        "Bartender",
			description: "A grizzled old man with a bushy beard and kind eyes. He's wiping down the bar with a rag and seems to know everyone in town. He might have useful information if you're willing to chat.",
		},
		{
			name:        "Mysterious Stranger",
			description: "A cloaked figure sitting alone in the corner, nursing a drink. They seem to be watching everything carefully but haven't spoken to anyone. Something about them feels important.",
		},
	}

	for _, npc := range npcs {
		var npcID int
		err := engine.db.QueryRow(
			ctx,
			"INSERT INTO npcs (name, description, location_id) VALUES ($1, $2, $3) RETURNING id",
			npc.name,
			npc.description,
			locationID,
		).Scan(&npcID)
		if err != nil {
			fmt.Printf("Error inserting NPC %s: %v\n", npc.name, err)
			continue
		}
		fmt.Printf("Created NPC: %s (ID: %d)\n", npc.name, npcID)
	}

	fmt.Printf("Default data seeding completed!\n")
}

func (engine *Engine) Close() {
	if engine.db != nil {
		engine.db.Close()
	}
}