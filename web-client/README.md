# PostgreSQL Text Adventure - Web Client

React-based web interface for the PostgreSQL Text Adventure game.

## Features

- **Auto-connect**: Automatically connects to the game server on page load
- **Real-time**: Uses WebSocket for instant communication
- **Query History**: Arrow up to recall previous commands
- **Results Display**: Shows query results in a formatted table
- **Connection Status**: Visual indicator of connection state

## Development

To run the web client in development mode:

```bash
cd web-client
npm install
npm run dev
```

The client will run on `http://localhost:3000` and connect to the game server's WebSocket at `ws://localhost:8080/ws`.

## Production

The web client is containerized and served via nginx. It's automatically built and started when you run `docker compose up`.

## Usage

1. Open `http://localhost:3000` in your browser
2. Wait for the connection indicator to show "Connected" (green dot)
3. Type game commands in the query box:
   - `look` - Look around your current location
   - `inventory` - Check your inventory
   - `take key` - Pick up an item
   - `go north` - Move to a new location
   - `talk to bartender` - Interact with NPCs
4. Press Enter or click Send to execute the command
5. Results appear in the results panel below

## Keyboard Shortcuts

- **Enter**: Send query
- **Arrow Up**: Recall previous query from history
