// Package main provides the Learning Desktop server.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/birddigital/htmx-r/components"
	"github.com/birddigital/htmx-r/pkg/voice"
	"github.com/birddigital/learning-desktop/internal/ai"
	"github.com/birddigital/learning-desktop/internal/auth"
	"github.com/birddigital/learning-desktop/internal/db"
	pagehandlers "github.com/birddigital/learning-desktop/internal/handlers"
	"github.com/birddigital/learning-desktop/internal/models"
	"github.com/birddigital/learning-desktop/internal/repository"
	"github.com/birddigital/learning-desktop/internal/service"
	"github.com/birddigital/go-llm-providers/pkg/providers"
	"github.com/google/uuid"
)

var (
	tutor         *ai.Tutor
	sessionService *service.SessionService
	authMiddleware *auth.Middleware
	chatHandler    *pagehandlers.ChatHandler
	coursesHandler *pagehandlers.CoursesHandler
	progressHandler *pagehandlers.ProgressHandler
	settingsHandler *pagehandlers.SettingsHandler
)

// defaultTenantID is used for demo purposes. In production, this comes from auth.
var defaultTenantID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

// defaultStudentID is the demo student for testing.
var defaultStudentID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

func getStudentID(w http.ResponseWriter, r *http.Request) uuid.UUID {
	// For demo, use the default student ID
	// In production, this would come from JWT auth
	return defaultStudentID
}

// renderChatLayout renders the chat interface with navigation
func renderChatLayout(w http.ResponseWriter, r *http.Request, sessionID string) error {
	// Create chat component
	chatComponent := components.NewChat(components.ChatProps{
		SessionID:     sessionID,
		Placeholder:   "Type your message or use voice input...",
		WelcomeTitle:  "Get Ahead of AI",
		WelcomePrompt: "Welcome! I'm your AI tutor. Ask me anything about AI, or click the microphone to start talking.",
		SSEEndpoint:   "/api/events",
		Theme:         "dark",
		ShowMedia:     true,
	})

	// Get chat HTML
	chatHTML, err := chatComponent.HTML()
	if err != nil {
		return err
	}

	// Build sidebar navigation items
	sidebarItems := []components.NavItemProps{
		{
			Label:      "Chat",
			Icon:       "M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z",
			HXGet:      "/?partial=true",
			HXTarget:   "#main-content",
			StateValue: "chat",
		},
		{
			Label:      "Courses",
			Icon:       "M4 19.5A2.5 2.5 0 0 1 6.5 17H20",
			HXGet:      "/courses?partial=true",
			HXTarget:   "#main-content",
			StateValue: "courses",
		},
		{
			Label:      "Progress",
			Icon:       "M12 20V10M18 20V4M6 20v-6",
			HXGet:      "/progress?partial=true",
			HXTarget:   "#main-content",
			StateValue: "progress",
		},
		{
			Label:      "Settings",
			Icon:       "M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M10.5 12.5l-2.5 2.5",
			HXGet:      "/settings?partial=true",
			HXTarget:   "#main-content",
			StateValue: "settings",
		},
	}

	layout := components.Layout(components.LayoutProps{
		Sidebar: components.SidebarProps{
			BrandName:  "Learning Desktop",
			BrandShort: "LD",
			Items:      sidebarItems,
			StateKey:   "sidebar",
		},
		PageTitle:         "Chat",
		UserName:          "Alex Student",
		NotificationCount: 0,
		ActiveNav:         "chat",
		ContentHTML:       template.HTML(chatHTML),
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return layout.Render(w)
}

// renderPartialChat renders just the chat component (for HTMX navigation)
func renderPartialChat(w http.ResponseWriter, sessionID string) error {
	chatComponent := components.NewChat(components.ChatProps{
		SessionID:     sessionID,
		Placeholder:   "Type your message or use voice input...",
		WelcomeTitle:  "Get Ahead of AI",
		WelcomePrompt: "Welcome! I'm your AI tutor. Ask me anything about AI, or click the microphone to start talking.",
		SSEEndpoint:   "/api/events",
		Theme:         "dark",
		ShowMedia:     true,
	})

	chatHTML, err := chatComponent.HTML()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(chatHTML))
	return nil
}

func main() {
	// Configuration
	port := flag.String("port", "3000", "Server port")
	flag.Parse()

	// Initialize database
	database, err := db.OpenFromEnv()
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Printf("Chat will use in-memory sessions. Set DATABASE_URL to enable persistence.")
		database = nil
	} else {
		log.Printf("Database connected successfully")
		defer database.Close()

		// Run health check
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := database.Health(ctx); err != nil {
			log.Printf("Warning: Database health check failed: %v", err)
		}
	}

	// Initialize repositories
	var chatRepo *repository.ChatRepository
	var sessionRepo *repository.SessionRepository

	if database != nil {
		// Convert *db.DB to *repository.DB for repository layer
		repoDB := &repository.DB{DB: database.DB}
		chatRepo = repository.NewChatRepository(repoDB)
		sessionRepo = repository.NewSessionRepository(repoDB)
	}

	// Initialize session service
	if chatRepo != nil && sessionRepo != nil {
		sessionService = service.NewSessionService(chatRepo, sessionRepo, defaultTenantID)

		// Start background cleanup goroutine
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				sessionService.CleanupExpiredSessions()
			}
		}()
	}

	// Initialize AI tutor
	tutor, err = ai.New()
	if err != nil {
		log.Printf("Warning: AI tutor initialization failed: %v", err)
		log.Printf("Chat will use fallback responses. Set ANTHROPIC_CREDENTIALS environment variable to enable AI.")
		tutor = nil // Will use fallback
	} else {
		log.Printf("AI tutor initialized successfully")
	}

	// Initialize auth middleware
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production" // Default for development
		log.Printf("Warning: Using default JWT secret. Set JWT_SECRET for production.")
	}
	authMiddleware = auth.NewMiddleware(&auth.Config{
		JWTSecret:      jwtSecret,
		JWTIssuer:      "learning-desktop",
		RequireAuth:    false, // Set to true in production
		DefaultTenantID: defaultTenantID,
	})
	log.Printf("Auth middleware initialized")

	// Initialize page handlers
	chatHandler = pagehandlers.NewChatHandler(tutor, sessionService, defaultTenantID.String(), defaultStudentID.String())
	coursesHandler = pagehandlers.NewCoursesHandler()
	progressHandler = pagehandlers.NewProgressHandler(sessionService)
	settingsHandler = pagehandlers.NewSettingsHandler()

	// Create HTTP multiplexer
	mux := http.NewServeMux()

	// Initialize voice services
	voiceService := voice.NewVoiceService(nil, nil, nil)
	voiceHandler := voice.NewVoiceHandler(voiceService)
	voiceHandler.RegisterRoutes(mux)

	// Static file server (serve from htmx-r)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../htmx-r/static"))))

	// Page routes (with navigation)
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/courses", handleCourses)
	mux.HandleFunc("/progress", handleProgress)
	mux.HandleFunc("/settings", handleSettings)

	// Chat endpoints
	mux.HandleFunc("/api/chat", handleChat)
	mux.HandleFunc("/api/chat/clear", handleClear)
	mux.HandleFunc("/api/chat/export", handleExport)
	mux.HandleFunc("/api/events", handleSSE)

	// Settings endpoints (with toast notifications)
	mux.HandleFunc("/api/settings/profile", handleSettingsProfile)
	mux.HandleFunc("/api/settings/preferences", handleSettingsPreferences)
	mux.HandleFunc("/api/settings/voice", handleSettingsVoice)
	mux.HandleFunc("/api/settings/appearance", handleSettingsAppearance)

	// Health endpoint
	mux.HandleFunc("/health", handleHealth)

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      authMiddleware.TenantContext(mux),
		ReadTimeout:  30 * time.Second,
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
		"status":       "ok",
		"tutor":        "enabled",
		"database":     "disconnected",
		"session_type": "memory",
		"version":      "0.1.0",
	}

	if tutor != nil {
		status["tutor"] = "enabled"
	} else {
		status["tutor"] = "fallback"
		status["warning"] = "AI tutor not configured - set ANTHROPIC_CREDENTIALS"
	}

	if sessionService != nil {
		status["session_type"] = "persistent"
		status["database"] = "connected"
	}

	json.NewEncoder(w).Encode(status)
}

// handleIndex serves the main page with navigation
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	sessionID := getOrCreateSessionID(r)
	isPartial := r.URL.Query().Get("partial") == "true"

	if isPartial {
		// Return just the chat component for HTMX navigation
		if err := renderPartialChat(w, sessionID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		// Return full layout with navigation
		if err := renderChatLayout(w, r, sessionID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// handleCourses serves the courses catalog page
func handleCourses(w http.ResponseWriter, r *http.Request) {
	coursesHandler.ServeHTTP(w, r)
}

// handleProgress serves the progress tracking page
func handleProgress(w http.ResponseWriter, r *http.Request) {
	progressHandler.ServeHTTP(w, r)
}

// handleSettings serves the settings page
func handleSettings(w http.ResponseWriter, r *http.Request) {
	settingsHandler.ServeHTTP(w, r)
}

// handleSettingsProfile handles profile settings updates
func handleSettingsProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// In a real app, would save to database here
	// For now, just show success notification
	pagehandlers.ToastNotification(w, "Profile updated successfully!", "success")
}

// handleSettingsPreferences handles learning preferences updates
func handleSettingsPreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pagehandlers.ToastNotification(w, "Preferences saved!", "success")
}

// handleSettingsVoice handles voice settings updates
func handleSettingsVoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pagehandlers.ToastNotification(w, "Voice settings updated!", "success")
}

// handleSettingsAppearance handles appearance settings updates
func handleSettingsAppearance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pagehandlers.ToastNotification(w, "Theme changed successfully!", "success")
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

	// Get student ID from auth/cookie
	studentID := getStudentID(w, r)
	sessionIDStr := getOrCreateSessionID(r)
	var sessionID uuid.UUID
	if sessionIDStr == "" {
		// Create new session in database
		if sessionService != nil {
			session, err := sessionService.GetOrCreateSession(r.Context(), studentID)
			if err != nil {
				log.Printf("Warning: failed to create session: %v", err)
			} else {
				sessionID = session.ID
			}
		}
		// Fallback to random UUID if session service fails
		if sessionID == uuid.Nil {
			sessionID = uuid.New()
		}
		// Set cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID.String(),
			Path:     "/",
			MaxAge:   int(365 * 24 * time.Hour / time.Second),
			SameSite: http.SameSiteLaxMode,
		})
	} else {
		sessionID = uuid.MustParse(sessionIDStr)
	}

	ctx := r.Context()

	// Persist user message if session service is available
	if sessionService != nil {
		_, err := sessionService.AddMessage(ctx, sessionID, studentID, defaultTenantID, models.RoleUser, message)
		if err != nil {
			log.Printf("Warning: failed to save message: %v", err)
			// Continue anyway - don't block user experience
		}
	}

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
		responseContent = fallbackResponse(message)
	} else {
		// Build conversation context
		var messages []providers.Message

		if sessionService != nil {
			// Load message history from database
			dbMessages, err := sessionService.GetMessages(ctx, sessionID, 10)
			if err == nil && len(dbMessages) > 0 {
				for _, msg := range dbMessages {
					messages = append(messages, providers.Message{
						Role:    string(msg.Role),
						Content: msg.Content,
					})
				}
			}
		}

		// Add current message if not in history
		if len(messages) == 0 || messages[len(messages)-1].Content != message {
			messages = append(messages, providers.Message{
				Role:    "user",
				Content: message,
			})
		}

		// Get response from AI tutor
		reqCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
		defer cancel()

		var resp string
		var err error
		if len(messages) > 1 {
			resp, err = tutor.RespondWithConversation(reqCtx, messages)
		} else {
			resp, err = tutor.Respond(reqCtx, message)
		}

		if err != nil {
			log.Printf("AI tutor error: %v", err)
			responseContent = fmt.Sprintf("I'm having trouble connecting right now. Error: %v\n\nPlease try again.", err)
		} else {
			responseContent = resp
		}
	}

	// Persist assistant message if session service is available
	if sessionService != nil {
		_, err := sessionService.AddMessage(ctx, sessionID, studentID, defaultTenantID, models.RoleAssistant, responseContent)
		if err != nil {
			log.Printf("Warning: failed to save assistant message: %v", err)
		}
	}

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

	// End the current session
	sessionIDStr := getOrCreateSessionID(r)
	if sessionService != nil && sessionIDStr != "" {
		sessionID := uuid.MustParse(sessionIDStr)
		if err := sessionService.EndSession(r.Context(), sessionID); err != nil {
			log.Printf("Warning: failed to end session: %v", err)
		}
	}

	// Generate new session ID
	newSessionID := uuid.New().String()

	// Update the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    newSessionID,
		Path:     "/",
		MaxAge:   int(365 * 24 * time.Hour / time.Second),
		SameSite: http.SameSiteLaxMode,
	})

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

	sessionIDStr := getOrCreateSessionID(r)
	var sessionID uuid.UUID
	if sessionIDStr != "" {
		sessionID = uuid.MustParse(sessionIDStr)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=conversation.txt")
	w.Write([]byte("Learning Desktop Conversation Export\n\n"))
	w.Write([]byte(fmt.Sprintf("Session ID: %s\n", sessionIDStr)))
	w.Write([]byte(fmt.Sprintf("Date: %s\n", time.Now().Format(time.RFC3339))))

	if sessionService != nil {
		messages, err := sessionService.GetMessages(r.Context(), sessionID, 1000)
		if err == nil {
			w.Write([]byte(fmt.Sprintf("Messages: %d\n", len(messages))))
			w.Write([]byte("\n---\n\n"))

			for i, msg := range messages {
				w.Write([]byte(fmt.Sprintf("[%d] %s:\n%s\n\n", i+1, msg.Role, msg.Content)))
			}
			return
		}
	}

	// Fallback if no session service
	w.Write([]byte("Messages: 0\n"))
	w.Write([]byte("\n---\n\n"))
	w.Write([]byte("(No messages available - session service not running)\n"))
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
	// Return empty string - handler will generate new ID
	return ""
}
