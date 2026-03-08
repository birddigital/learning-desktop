// Package main provides the Learning Desktop server.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/birddigital/learning-desktop/internal/voice"
	"github.com/birddigital/learning-desktop/pkg/chat"
	"github.com/google/uuid"
)

func main() {
	// Configuration
	port := flag.String("port", "3000", "Server port")
	flag.Parse()

	// Create HTTP multiplexer
	mux := http.NewServeMux()

	// Initialize services
	voiceService := voice.NewVoiceService(nil, nil, nil)
	voiceHandler := voice.NewVoiceHandler(voiceService)
	voiceHandler.RegisterRoutes(mux)

	// Static file server
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Chat endpoints
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/chat", handleChat)
	mux.HandleFunc("/api/chat/clear", handleClear)
	mux.HandleFunc("/api/chat/export", handleExport)
	mux.HandleFunc("/api/events", handleSSE)

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
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

// handleIndex serves the main page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	sessionID := getOrCreateSessionID(r)
	chatComponent := chat.NewChat(chat.ChatProps{
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

	// Create user message
	userMsg := chat.ChatMessage{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	}

	// Render user message
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	chat.RenderMessageHTML(w, userMsg)

	// TODO: Generate AI response
	// For now, echo a simple response
	assistantMsg := chat.ChatMessage{
		ID:        uuid.New().String(),
		Role:      "assistant",
		Content:   "I understand you're interested in: \"" + message + "\". Let me help you learn more about this topic.",
		Timestamp: time.Now(),
	}

	chat.RenderMessageHTML(w, assistantMsg)
}

// handleClear clears the chat session
func handleClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Re-render the main page with new session
	sessionID := uuid.New().String()
	chatComponent := chat.NewChat(chat.ChatProps{
		SessionID:     sessionID,
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
	}
}

// handleExport exports the conversation
func handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=conversation.txt")
	w.Write([]byte("Learning Desktop Conversation Export\n\n"))
	w.Write([]byte("Session ID: " + getOrCreateSessionID(r) + "\n"))
	w.Write([]byte("Date: " + time.Now().Format(time.RFC3339) + "\n"))
	w.Write([]byte("\n---\n\n"))
	w.Write([]byte("(Conversation history would be exported here)\n"))
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
	fmt.Fprintf(w, "event: connected\ndata: {\"session_id\":\"%s\"}\n\n", getOrCreateSessionID(r))
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
