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
	npcs := engine.getNpcs()

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

	systemPrompt := fmt.Sprintf(`You are a dungeon master for a text adventure game. You must respond ONLY with valid JSON in the exact format specified below.

# Current World State
## Locations:
%s

## Items in Inventory:
%s

## Items in the World:
%s

## NPCs:
%s

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
9. Be creative and respond to player actions appropriately`, world, items, worldItems, npcs, jsonSchema)

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

func (engine *Engine) getNpcs() string {
	ctx := context.Background()
	var npcs []string
	
	rows, err := engine.db.Query(ctx, `
		SELECT n.name, n.description, l.name as location_name 
		FROM npcs n 
		LEFT JOIN locations l ON n.location_id = l.id 
		ORDER BY n.id
	`)
	if err != nil {
		fmt.Printf("Error querying NPCs: %v\n", err)
		return "Unable to load NPCs."
	}
	defer rows.Close()
	
	for rows.Next() {
		var name, description, locationName string
		if err := rows.Scan(&name, &description, &locationName); err != nil {
			continue
		}
		if locationName != "" {
			npcs = append(npcs, fmt.Sprintf("%s (at %s): %s", name, locationName, description))
		} else {
			npcs = append(npcs, fmt.Sprintf("%s: %s", name, description))
		}
	}
	
	if len(npcs) == 0 {
		return "No NPCs found in the world."
	}
	
	return strings.Join(npcs, "\n\n")
}

func (engine *Engine) initDatabase() {
	ctx := context.Background()
	
	// Create tables
	queries := []string{
		"CREATE TABLE IF NOT EXISTS players (id SERIAL PRIMARY KEY, name VARCHAR(255))",
		"CREATE TABLE IF NOT EXISTS locations (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT)",
		"CREATE TABLE IF NOT EXISTS items (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT, location_id INT REFERENCES locations(id) ON DELETE SET NULL)",
		"CREATE TABLE IF NOT EXISTS npcs (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT, location_id INT REFERENCES locations(id) ON DELETE SET NULL)",
		"CREATE TABLE IF NOT EXISTS player_items (id SERIAL PRIMARY KEY, player_id INT REFERENCES players(id), item_id INT REFERENCES items(id))",
		"CREATE TABLE IF NOT EXISTS player_notes (id SERIAL PRIMARY KEY, player_id INT REFERENCES players(id), note TEXT)",
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
		// Insert player with ID 1
		_, err := engine.db.Exec(ctx, 
			"INSERT INTO players (id, name) VALUES (1, 'Player') ON CONFLICT (id) DO NOTHING",
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
	}
}

func (engine *Engine) recreateTables() {
	ctx := context.Background()
	
	// Drop tables in reverse order of dependencies
	dropQueries := []string{
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
		"CREATE TABLE players (id SERIAL PRIMARY KEY, name VARCHAR(255))",
		"CREATE TABLE locations (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT)",
		"CREATE TABLE items (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT, location_id INT REFERENCES locations(id))",
		"CREATE TABLE npcs (id SERIAL PRIMARY KEY, name VARCHAR(255), description TEXT, location_id INT REFERENCES locations(id))",
		"CREATE TABLE player_items (id SERIAL PRIMARY KEY, player_id INT REFERENCES players(id), item_id INT REFERENCES items(id))",
		"CREATE TABLE player_notes (id SERIAL PRIMARY KEY, player_id INT REFERENCES players(id), note TEXT)",
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