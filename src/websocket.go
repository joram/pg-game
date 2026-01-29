package main

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgproto3"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSMessage is the JSON message format for WebSocket client ↔ game server.
type WSMessage struct {
	Type     string          `json:"type"`
	Query    string          `json:"query,omitempty"`
	Content  string          `json:"content,omitempty"`
	Message  string          `json:"message,omitempty"`
	Rows     [][]interface{} `json:"rows,omitempty"`
	Columns  []string        `json:"columns,omitempty"`
	RowCount int             `json:"rowCount,omitempty"`
}

type wsQueryState struct {
	columns []string
	rows    [][]interface{}
	query   string
}

// connectToGameServer performs SSL handshake and startup against the local game server.
func connectToGameServer(addr string) (*pgproto3.Frontend, net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	sslReq := make([]byte, 8)
	binary.BigEndian.PutUint32(sslReq[0:4], 8)
	binary.BigEndian.PutUint32(sslReq[4:8], 80877103)
	if _, err := conn.Write(sslReq); err != nil {
		conn.Close()
		return nil, nil, err
	}

	buf := make([]byte, 1)
	if _, err := conn.Read(buf); err != nil || buf[0] != 'S' {
		conn.Close()
		return nil, nil, fmt.Errorf("server did not accept SSL: %v", err)
	}

	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	tlsConn := tls.Client(conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("TLS handshake: %w", err)
	}

	frontend := pgproto3.NewFrontend(tlsConn, tlsConn)
	frontend.Send(&pgproto3.StartupMessage{
		ProtocolVersion: 196608,
		Parameters:      map[string]string{"user": "postgres", "database": "postgres"},
	})
	if err := frontend.Flush(); err != nil {
		tlsConn.Close()
		return nil, nil, err
	}

	for {
		msg, err := frontend.Receive()
		if err != nil {
			tlsConn.Close()
			return nil, nil, err
		}
		switch msg.(type) {
		case *pgproto3.ReadyForQuery:
			return frontend, tlsConn, nil
		case *pgproto3.ErrorResponse:
			tlsConn.Close()
			return nil, nil, fmt.Errorf("startup error from server")
		default:
			continue
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("New WebSocket connection from %s", r.RemoteAddr)

	// Connect to our own game server on localhost
	frontend, pgConn, err := connectToGameServer("127.0.0.1:5432")
	if err != nil {
		log.Printf("Failed to connect to game server: %v", err)
		writeWSError(conn, fmt.Sprintf("Failed to connect to game server: %v", err))
		return
	}
	defer pgConn.Close()

	var currentQuery *wsQueryState

	go func() {
		for {
			var msg WSMessage
			if err := conn.ReadJSON(&msg); err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}
			if msg.Type == "query" && msg.Query != "" {
				currentQuery = &wsQueryState{
					query:   msg.Query,
					columns: []string{},
					rows:    [][]interface{}{},
				}
				frontend.SendQuery(&pgproto3.Query{String: msg.Query})
				if err := frontend.Flush(); err != nil {
					log.Printf("Error sending query: %v", err)
					writeWSError(conn, fmt.Sprintf("Error sending query: %v", err))
					return
				}
			}
		}
	}()

	for {
		msg, err := frontend.Receive()
		if err != nil {
			log.Printf("Error receiving from game server: %v", err)
			return
		}
		switch m := msg.(type) {
		case *pgproto3.DataRow:
			if currentQuery != nil {
				row := make([]interface{}, len(m.Values))
				for i, val := range m.Values {
					if val == nil {
						row[i] = nil
					} else {
						row[i] = string(val)
					}
				}
				currentQuery.rows = append(currentQuery.rows, row)
			}
		case *pgproto3.RowDescription:
			if currentQuery != nil {
				columns := make([]string, len(m.Fields))
				for i, field := range m.Fields {
					columns[i] = string(field.Name)
				}
				currentQuery.columns = columns
			}
		case *pgproto3.CommandComplete:
			if currentQuery != nil {
				response := WSMessage{
					Type:     "result",
					Query:    currentQuery.query,
					Columns:  currentQuery.columns,
					Rows:     currentQuery.rows,
					RowCount: len(currentQuery.rows),
				}
				if err := conn.WriteJSON(response); err != nil {
					log.Printf("Error sending response: %v", err)
					return
				}
				currentQuery = nil
			}
		case *pgproto3.ErrorResponse:
			errorMsg := m.Message
			if errorMsg == "" {
				errorMsg = m.Severity
			}
			writeWSError(conn, errorMsg)
			currentQuery = nil
		case *pgproto3.ReadyForQuery:
			// ready for next query
		case *pgproto3.NoticeResponse:
			noticeMsg := m.Message
			if noticeMsg == "" {
				noticeMsg = m.Severity
			}
			if noticeMsg != "" {
				conn.WriteJSON(WSMessage{Type: "text", Content: noticeMsg})
			}
		case *pgproto3.CopyInResponse, *pgproto3.CopyOutResponse:
			// no-op
		default:
			// ignore other backend messages
		}
	}
}

func writeWSError(conn *websocket.Conn, message string) {
	conn.WriteJSON(WSMessage{Type: "error", Message: message})
}

// handleTTS handles POST requests to /tts, generating speech audio from text using espeak-ng.
func handleTTS(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	// Piper reads stdin line-by-line and only speaks the first line by default.
	// Normalize to a single line so the full message is spoken (replace newlines with space).
	textForPiper := strings.ReplaceAll(req.Text, "\n", " ")
	textForPiper = strings.TrimSpace(strings.Join(strings.Fields(textForPiper), " "))

	// Run piper with the neural voice model.
	// length_scale > 1 = slower speech (more gravitas); 1.35 gives a deliberate, weighty delivery.
	// noise_scale slightly below default = less variation, more consistent gravitas (if supported).
	cmd := exec.Command("piper",
		"--model", "/opt/piper-voices/en_US-lessac-medium.onnx",
		"--output_file", "-",
		"--length_scale", "0.9",
		"--noise_scale", "0.667")
	cmd.Stdin = strings.NewReader(textForPiper)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("TTS error: %v", err)
		http.Error(w, "TTS generation failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/wav")
	w.Write(output)
}

// StartWebSocketServer starts the HTTP server that serves /ws and /tts on the given addr.
func StartWebSocketServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleWebSocket)
	mux.HandleFunc("/tts", handleTTS)
	log.Printf("WebSocket server %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("WebSocket server error: %v", err)
	}
}
