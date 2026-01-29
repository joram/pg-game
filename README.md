# PostgreSQL Text-Based Adventure Game

A text-based adventure game server that uses PostgreSQL as the protocol interface and Anthropic Claude for dynamic game logic and narrative generation.

## Overview

This project implements a game server that:
- Accepts PostgreSQL protocol connections (port 5432)
- Uses Anthropic Claude LLM to generate dynamic game responses
- Persists game state in PostgreSQL (locations, items, NPCs, player inventory, interactions)
- Supports hot-reloading during development with Air
- Tracks player location, inventory, and NPC interaction history

## Architecture

- **Game Server**: Go application that implements PostgreSQL wire protocol
- **Database**: PostgreSQL 15 for game state persistence
- **LLM**: Anthropic Claude 3.5 Sonnet for game logic and narrative generation
- **Hot Reload**: Air (Cosmtrek/air) for automatic code reloading during development
- **Containerization**: Docker Compose for easy deployment

## Prerequisites

- Docker and Docker Compose
- Anthropic API key ([Get one here](https://console.anthropic.com/))

## Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd psql-text-based-adventure
   ```

2. **Create a `.env` file** in the project root:
   ```bash
   ANTHROPIC_API_KEY=your-api-key-here
   ```

3. **Start the services**
   ```bash
   docker compose up
   ```

   This will start:
   - `game-server`: The game server on port 5432 (PostgreSQL protocol)
   - `db`: PostgreSQL database on internal port 5432

4. **Connect to the game** using any PostgreSQL client:
   ```bash
   psql -h localhost -p 5432 -U postgres -d postgres
   ```

   Or use any PostgreSQL-compatible client/tool.

## Usage

Once connected, you can interact with the game using natural language commands:

```sql
-- Look around
describe the world;

-- Check your inventory
inventory;

-- Interact with the world
take key;
talk to bartender;
go north;
use key on door;

-- Ask about NPC history
what history do I have with the bartender?
```

The game will:
- Generate dynamic responses based on your actions
- Update your inventory and location
- Remember your interactions with NPCs
- Persist all game state in the database

## Features

- **Dynamic World**: Game world is generated and managed by the LLM
- **State Persistence**: All game state (locations, items, NPCs, inventory) persists across sessions
- **NPC Memory**: NPCs remember past interactions with players
- **Inventory Management**: Track items in your inventory and in the world
- **Location System**: Navigate between locations in the game world
- **Hot Reload**: Code changes automatically reload during development

## Development

The project uses Air for hot-reloading. When you edit files in the `src/` directory, the server will automatically rebuild and restart.

### Project Structure

```
.
├── src/              # Go source code
│   ├── main.go      # Entry point and PostgreSQL protocol handler
│   ├── engine.go    # Game engine, LLM integration, database logic
│   └── ssl.go       # TLS/SSL handling
├── docker-compose.yml
├── Dockerfile
├── .air.toml        # Air configuration
├── .env             # Environment variables (not in git)
└── README.md
```

### Database Schema

The game uses the following main tables:
- `locations`: Game locations/rooms
- `items`: Items in the world
- `npcs`: Non-player characters
- `players`: Player information and current location
- `player_items`: Player inventory (junction table)
- `npc_player_interactions`: History of player-NPC interactions

### Environment Variables

- `ANTHROPIC_API_KEY`: Required. Your Anthropic API key for Claude access
- `DATABASE_URL`: Optional. Defaults to `postgresql://postgres:postgres@db:5432/postgres`

## Troubleshooting

- **Connection refused**: Ensure Docker Compose services are running (`docker compose ps`)
- **LLM errors**: Check that `ANTHROPIC_API_KEY` is set correctly in `.env`
- **Database errors**: The database will auto-initialize on first run. Check logs with `docker compose logs db`
- **Code not reloading**: Check Air logs with `docker compose logs game-server`

## License

[Add your license here]
