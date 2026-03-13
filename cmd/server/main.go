// Package main provides the Learning Desktop server.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/birddigital/htmx-r/components"
	"github.com/birddigital/htmx-r/pkg/voice"
	"github.com/birddigital/learning-desktop/internal/ai"
	"github.com/birddigital/go-llm-providers/pkg/providers"
	"github.com/google/uuid"
)

// sessionStore holds active chat sessions in memory.
// In production, this would use Redis or a database.
type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*chatSession
}

type chatSession struct {
	ID        string
	Messages  []providers.Message
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	tutor    *ai.Tutor
	sessions = &sessionStore{
		sessions: make(map[string]*chatSession),
	}
)

func main() {
	// Configuration
	port := flag.String("port", "3000", "Server port")
	flag.Parse()

	// Initialize AI tutor
	var err error
	tutor, err = ai.New()
	if err != nil {
		log.Printf("Warning: AI tutor initialization failed: %v", err)
		log.Printf("Chat will use fallback responses. Set ANTHROPIC_CREDENTIALS environment variable to enable AI.")
		tutor = nil // Will use fallback
	} else {
		log.Printf("AI tutor initialized successfully")
	}

	// Create HTTP multiplexer
	mux := http.NewServeMux()

	// Initialize services
	voiceService := voice.NewVoiceService(nil, nil, nil)
	voiceHandler := voice.NewVoiceHandler(voiceService)
	voiceHandler.RegisterRoutes(mux)

	// Static file server (serve from htmx-r)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../htmx-r/static"))))

	// Chat endpoints
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/chat", handleChat)
	mux.HandleFunc("/api/chat/clear", handleClear)
	mux.HandleFunc("/api/chat/export", handleExport)
	mux.HandleFunc("/api/events", handleSSE)
	mux.HandleFunc("/health", handleHealth)

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second, // Increased for AI responses
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Learning Desktop server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}

// handleHealth returns the health status of the server.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := map[string]interface{}{
		"status":  "ok",
		"tutor":   "enabled",
		"version": "0.1.0",
	}
	if tutor == nil {
		status["tutor"] = "fallback"
		status["warning"] = "AI tutor not configured - set ANTHROPIC_CREDENTIALS"
	}
	json.NewEncoder(w).Encode(status)
}

// handleIndex serves the main page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	sessionID := getOrCreateSessionID(r)
	chatComponent := components.NewChat(components.ChatProps{
		SessionID:     sessionID,
		Placeholder:   "Type your message or use voice input...",
		WelcomeTitle:  "Get Ahead of AI",
		WelcomePrompt: "Welcome! I'm your AI tutor. Ask me anything about AI, or click the microphone to start talking.",
		SSEEndpoint:   "/api/events",
		Theme:         "dark",
		ShowMedia:     true,
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := chatComponent.Render(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleChat processes chat messages
func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	sessionID := getOrCreateSessionID(r)

	// Get or create session
	session := getOrCreateSession(sessionID)

	// Create user message
	userMsg := providers.Message{
		Role:    "user",
		Content: message,
	}
	session.Messages = append(session.Messages, userMsg)

	// Render user message
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	components.RenderMessageHTML(w, components.ChatMessage{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	})

	// Generate AI response
	var responseContent string

	if tutor == nil {
		// Fallback response when AI is not configured
		responseContent = fallbackResponse(message)
	} else {
		// Get response from AI tutor
		ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
		defer cancel()

		resp, err := tutor.RespondWithConversation(ctx, session.Messages)
		if err != nil {
			log.Printf("AI tutor error: %v", err)
			responseContent = fmt.Sprintf("I'm having trouble connecting right now. Error: %v\n\nPlease try again or check that the AI service is configured.", err)
		} else {
			responseContent = resp
		}
	}

	// Add assistant message to session
	assistantMsg := providers.Message{
		Role:    "assistant",
		Content: responseContent,
	}
	session.Messages = append(session.Messages, assistantMsg)

	// Render assistant message
	components.RenderMessageHTML(w, components.ChatMessage{
		ID:        uuid.New().String(),
		Role:      "assistant",
		Content:   responseContent,
		Timestamp: time.Now(),
	})
}

// fallbackResponse provides a simple response when AI is not configured.
func fallbackResponse(message string) string {
	return fmt.Sprintf("I understand you're asking about: \"%s\"\n\n**Note:** The AI tutor is not currently configured. To enable full AI responses, set the ANTHROPIC_CREDENTIALS environment variable.\n\nWhen configured, I can help you learn about:\n- Prompt Engineering\n- AI Concepts\n- Models & Data\n- Character & Manhood\n- Student Skills\n- Entrepreneurship", message)
}

// handleClear clears the chat session
func handleClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := getOrCreateSessionID(r)

	// Clear session
	sessions.mu.Lock()
	delete(sessions.sessions, sessionID)
	sessions.mu.Unlock()

	// Re-render the main page with new session
	newSessionID := uuid.New().String()
	chatComponent := components.NewChat(components.ChatProps{
		SessionID:     newSessionID,
		Placeholder:   "Type your message or use voice input...",
		WelcomeTitle:  "Get Ahead of AI",
		WelcomePrompt: "Conversation cleared. Ready for a fresh start!",
		SSEEndpoint:   "/api/events",
		Theme:         "dark",
		ShowMedia:     true,
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := chatComponent.Render(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleExport exports the conversation
func handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := getOrCreateSessionID(r)
	session := getSession(sessionID)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=conversation.txt")
	w.Write([]byte("Learning Desktop Conversation Export\n\n"))
	w.Write([]byte(fmt.Sprintf("Session ID: %s\n", sessionID)))
	w.Write([]byte(fmt.Sprintf("Date: %s\n", time.Now().Format(time.RFC3339))))
	w.Write([]byte(fmt.Sprintf("Messages: %d\n", len(session.Messages))))
	w.Write([]byte("\n---\n\n"))

	for i, msg := range session.Messages {
		w.Write([]byte(fmt.Sprintf("[%d] %s:\n%s\n\n", i+1, msg.Role, msg.Content)))
	}
}

// handleSSE provides Server-Sent Events for real-time updates
func handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send initial connection event
	sessionID := getOrCreateSessionID(r)
	fmt.Fprintf(w, "event: connected\ndata: {\"session_id\":\"%s\"}\n\n", sessionID)
	flusher.Flush()

	// Keep connection alive
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// getOrCreateSessionID gets or creates a session ID from cookie
func getOrCreateSessionID(r *http.Request) string {
	if cookie, err := r.Cookie("session_id"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return uuid.New().String()
}

// getOrCreateSession gets an existing session or creates a new one
func getOrCreateSession(id string) *chatSession {
	sessions.mu.RLock()
	session, exists := sessions.sessions[id]
	sessions.mu.RUnlock()

	if exists {
		session.UpdatedAt = time.Now()
		return session
	}

	sessions.mu.Lock()
	defer sessions.mu.Unlock()

	// Check again in case another goroutine created it
	if session, exists := sessions.sessions[id]; exists {
		return session
	}

	session = &chatSession{
		ID:        id,
		Messages:  make([]providers.Message, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	sessions.sessions[id] = session
	return session
}

// getSession retrieves an existing session
func getSession(id string) *chatSession {
	sessions.mu.RLock()
	defer sessions.mu.RUnlock()
	if session, exists := sessions.sessions[id]; exists {
		return session
	}
	return &chatSession{
		ID:       id,
		Messages: make([]providers.Message, 0),
	}
}
