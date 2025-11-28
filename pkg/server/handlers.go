package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go-ai-agent-v2/go-cli/pkg/types"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by default.
		return true
	},
}

// handleWebSocket is the HTTP handler for WebSocket connections.
func (s *Server) handleWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}
		defer s.Broadcaster.RemoveClient(conn)

		s.Broadcaster.AddClient(conn)

		// Loop to handle incoming messages
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Client disconnected: %v", err)
				break
			}

			prompt := string(message)
			log.Printf("Received prompt: %s", prompt)

			// Generate a new session ID for each new WebSocket message
			sessionID := s.SessionService.GenerateSessionID()

			eventChan, err := s.ChatService.SendMessage(r.Context(), sessionID, prompt)
			if err != nil {
				log.Printf("Error sending message to chat service: %v", err)
				// Optionally send an error message back to the client
				continue
			}

			go s.broadcastEvents(eventChan, sessionID)
		}
	}
}

// broadcastEvents reads events from the chat service channel and broadcasts them to all clients.
func (s *Server) broadcastEvents(eventChan <-chan any, sessionID string) {
	for event := range eventChan {
		var eventData struct {
			Type      string      `json:"type"`
			Payload   interface{} `json:"payload"`
			SessionID string      `json:"sessionId"` // Include session ID
		}
		eventData.SessionID = sessionID

		switch e := event.(type) {
		case types.Part:
			eventData.Type = "chunk"
			eventData.Payload = e
		case types.StreamingStartedEvent:
			eventData.Type = "streaming_started"
			eventData.Payload = e
		case types.ThinkingEvent:
			eventData.Type = "thinking"
			eventData.Payload = e
		case types.FinalResponseEvent:
			eventData.Type = "final_response"
			eventData.Payload = e
		case types.ToolCallStartEvent:
			eventData.Type = "tool_call_start"
			eventData.Payload = e
		case types.ToolCallEndEvent:
			eventData.Type = "tool_call_end"
			eventData.Payload = e
		case types.ErrorEvent:
			eventData.Type = "error"
			eventData.Payload = e
		case types.TokenCountEvent:
			eventData.Type = "token_count"
			eventData.Payload = e
		case types.ModelSwitchEvent:
			eventData.Type = "model_switch"
			eventData.Payload = e
		case types.ToolConfirmationRequestEvent:
			eventData.Type = "tool_confirmation_request"
			eventData.Payload = e
		case types.TodosSummaryUpdateEvent:
			eventData.Type = "todos_summary_update"
			eventData.Payload = e
		default:
			log.Printf("Unknown event type: %T", e)
			continue
		}

		jsonData, err := json.Marshal(eventData)
		if err != nil {
			log.Printf("Error marshaling event to JSON: %v", err)
			continue
		}
		s.Broadcaster.Broadcast(websocket.TextMessage, jsonData)
	}
}

// handleTaskWebhook is the HTTP handler for incoming task webhooks.
func (s *Server) handleTaskWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var taskReq types.WebhookTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&taskReq); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if taskReq.Prompt == "" {
			http.Error(w, "Prompt cannot be empty", http.StatusBadRequest)
			return
		}

		log.Printf("Received webhook task with prompt: %s", taskReq.Prompt)

		// Generate a new session ID for each webhook task
		sessionID := s.SessionService.GenerateSessionID()

		// We need a separate context for the background task.
		taskCtx := context.Background()

		go func() {
			eventChan, err := s.ChatService.SendMessage(taskCtx, sessionID, taskReq.Prompt)
			if err != nil {
				log.Printf("Error starting webhook task: %v", err)
				return
			}
			s.broadcastEvents(eventChan, sessionID)
		}()

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "task accepted", "sessionId": sessionID})
	}
}
