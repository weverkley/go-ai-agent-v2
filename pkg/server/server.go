package server

import (
	"log"
	"net/http"

	"go-ai-agent-v2/go-cli/pkg/services"
	"github.com/gorilla/mux"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	Router        *mux.Router
	Broadcaster   *Broadcaster
	ChatService   *services.ChatService
	SessionService *services.SessionService
}

// NewServer creates a new Server instance.
func NewServer(chatService *services.ChatService, sessionService *services.SessionService) *Server {
	s := &Server{
		Router:        mux.NewRouter(),
		Broadcaster:   NewBroadcaster(),
		ChatService:   chatService,
		SessionService: sessionService,
	}
	s.routes()
	return s
}

// Start starts the HTTP server on the specified address.
func (s *Server) Start(addr string) {
	log.Printf("Agent server listening on %s", addr)
	if err := http.ListenAndServe(addr, s.Router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// routes sets up the routes for the server.
func (s *Server) routes() {
	s.Router.HandleFunc("/ws", s.handleWebSocket())
	s.Router.HandleFunc("/api/v1/tasks", s.handleTaskWebhook()).Methods("POST")
}
