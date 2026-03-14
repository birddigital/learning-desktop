// Package handlers provides page handlers for Learning Desktop.
package handlers

import (
	"html/template"
	"net/http"

	"github.com/birddigital/htmx-r/components"
	"github.com/birddigital/learning-desktop/internal/ai"
	"github.com/birddigital/learning-desktop/internal/service"
)

// ChatHandler renders the chat page
type ChatHandler struct {
	tutor         *ai.Tutor
	sessionService *service.SessionService
	defaultTenantID string
	defaultStudentID string
}

// NewChatHandler creates a new chat page handler
func NewChatHandler(tutor *ai.Tutor, sessionService *service.SessionService, defaultTenantID, defaultStudentID string) *ChatHandler {
	return &ChatHandler{
		tutor:          tutor,
		sessionService: sessionService,
		defaultTenantID: defaultTenantID,
		defaultStudentID: defaultStudentID,
	}
}

// ServeHTTP handles the chat page request
func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := GetOrCreateSessionID(r)
	isPartial := r.URL.Query().Get("partial") == "true"

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := PageContext{
		Title:             "Chat",
		UserName:          "Alex Student",
		SessionID:         sessionID,
		ActiveNav:         "chat",
		NotificationCount: 0,
	}

	if isPartial {
		// Return just the chat content for HTMX swap
		RenderPartial(w, r, chatHTML)
	} else {
		// Return full layout
		RenderLayout(w, r, ctx, chatHTML)
	}
}

// CoursesHandler renders the courses catalog page
type CoursesHandler struct{}

// NewCoursesHandler creates a new courses page handler
func NewCoursesHandler() *CoursesHandler {
	return &CoursesHandler{}
}

// ServeHTTP handles the courses page request
func (h *CoursesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := GetOrCreateSessionID(r)
	isPartial := r.URL.Query().Get("partial") == "true"

	// Get course page HTML
	contentHTML, err := renderTemplate("courses", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := PageContext{
		Title:             "Courses",
		UserName:          "Alex Student",
		SessionID:         sessionID,
		ActiveNav:         "courses",
		NotificationCount: 0,
	}

	if isPartial {
		RenderPartial(w, r, contentHTML)
	} else {
		RenderLayout(w, r, ctx, contentHTML)
	}
}

// ProgressHandler renders the progress tracking page
type ProgressHandler struct {
	sessionService *service.SessionService
}

// NewProgressHandler creates a new progress page handler
func NewProgressHandler(sessionService *service.SessionService) *ProgressHandler {
	return &ProgressHandler{
		sessionService: sessionService,
	}
}

// ServeHTTP handles the progress page request
func (h *ProgressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := GetOrCreateSessionID(r)
	isPartial := r.URL.Query().Get("partial") == "true"

	// Get progress page HTML
	contentHTML, err := renderTemplate("progress", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := PageContext{
		Title:             "Progress",
		UserName:          "Alex Student",
		SessionID:         sessionID,
		ActiveNav:         "progress",
		NotificationCount: 0,
	}

	if isPartial {
		RenderPartial(w, r, contentHTML)
	} else {
		RenderLayout(w, r, ctx, contentHTML)
	}
}

// SettingsHandler renders the settings page
type SettingsHandler struct{}

// NewSettingsHandler creates a new settings page handler
func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

// ServeHTTP handles the settings page request
func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := GetOrCreateSessionID(r)
	isPartial := r.URL.Query().Get("partial") == "true"

	// Get settings page HTML
	contentHTML, err := renderTemplate("settings", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := PageContext{
		Title:             "Settings",
		UserName:          "Alex Student",
		SessionID:         sessionID,
		ActiveNav:         "settings",
		NotificationCount: 0,
	}

	if isPartial {
		RenderPartial(w, r, contentHTML)
	} else {
		RenderLayout(w, r, ctx, contentHTML)
	}
}

// ToastNotification sends a toast notification
func ToastNotification(w http.ResponseWriter, message string, toastType string) {
	html := `<script>
		(function() {
			const container = document.getElementById('toast-container');
			if (!container) return;

			const toast = document.createElement('div');
			toast.className = 'toast toast-` + toastType + `';
			toast.innerHTML = '<span class="toast-message">' + ` + message + ` + '</span>';

			container.appendChild(toast);

			// Animate in
			requestAnimationFrame(function() {
				toast.classList.add('toast-show');
			});

			// Remove after 3 seconds
			setTimeout(function() {
				toast.classList.remove('toast-show');
				setTimeout(function() {
					toast.remove();
				}, 300);
			}, 3000);
		})();
	</script>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// Modal renders a modal dialog
func RenderModal(w http.ResponseWriter, id, title string, content template.HTML, width string) {
	if width == "" {
		width = "600px"
	}

	modal := components.Modal(components.ModalProps{
		ID:      id,
		Title:   title,
		Content: content,
		Width:   width,
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	modal.Render(w)
}

// Spinner renders a loading spinner
func RenderSpinner(w http.ResponseWriter, size string) {
	html := `<div class="spinner-container" style="display: flex; justify-content: center; align-items: center; padding: 2rem;">
		<div class="spinner spinner-` + size + `">
			<div class="spinner-dot"></div>
			<div class="spinner-dot"></div>
			<div class="spinner-dot"></div>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
