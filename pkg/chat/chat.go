package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"
)

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	ID        string      `json:"id"`
	Role      string      `json:"role"`      // "user" | "assistant" | "system"
	Content   string      `json:"content"`
	Media     *MediaRef   `json:"media,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Metadata  MessageMeta  `json:"metadata"`
}

// MediaRef represents injected media (video, 3D, board)
type MediaRef struct {
	Type   string                 `json:"type"`     // "video" | "3d" | "board" | "code"
	Source string                 `json:"source"`
	Title  string                 `json:"title"`
	Data   map[string]interface{} `json:"data"`
}

// ID returns a unique ID for the media reference
func (m *MediaRef) ID() string {
	return fmt.Sprintf("%s-%s", m.Type, m.Source)
}

// MessageMeta contains additional message metadata
type MessageMeta struct {
	LessonID     *string `json:"lesson_id,omitempty"`
	ConceptID    *string `json:"concept_id,omitempty"`
	IsCheckpoint bool    `json:"is_checkpoint"`
}

// ChatSession represents a student's chat session
type ChatSession struct {
	ID           string         `json:"id"`
	StudentID    string         `json:"student_id"`
	TenantID     string         `json:"tenant_id"`
	StartedAt    time.Time      `json:"started_at"`
	LastActive   time.Time      `json:"last_active"`
	Messages     []ChatMessage  `json:"messages"`
	CurrentState SessionState   `json:"current_state"`
	Context      SessionContext  `json:"context"`
}

// SessionState tracks where the student is in the course
type SessionState struct {
	ModuleID     *int     `json:"module_id,omitempty"`
	LessonID     *int     `json:"lesson_id,omitempty"`
	StepID       *int     `json:"step_id,omitempty"`
	CompletedIDs []int    `json:"completed_ids"`
	SkillLevel   string   `json:"skill_level"` // "beginner" | "intermediate" | "advanced"
	Interests    []string `json:"interests"`
	Goals        []string `json:"goals"`
}

// SessionContext provides AI with context about the student
type SessionContext struct {
	UserName   string    `json:"user_name"`
	Background string    `json:"background"`
	Concerns   []string  `json:"concerns"`
	Progress   float64   `json:"progress"` // 0-1 overall course completion
	LastLogin  time.Time `json:"last_login"`
	TotalTime  int       `json:"total_minutes_spent"`
}

// ChatProps configures the chat interface
type ChatProps struct {
	SessionID     string
	Placeholder   string
	WelcomeTitle  string
	WelcomePrompt string
	SSEEndpoint   string
	Theme         string // "dark" | "light"
	ShowMedia     bool
}

// ChatComponent renders the Claude Desktop-style chat interface
type ChatComponent struct {
	props ChatProps
}

// NewChat creates a new chat component
func NewChat(props ChatProps) *ChatComponent {
	if props.Placeholder == "" {
		props.Placeholder = "Type your message..."
	}
	if props.WelcomeTitle == "" {
		props.WelcomeTitle = "Get Ahead of AI"
	}
	if props.Theme == "" {
		props.Theme = "dark"
	}
	return &ChatComponent{props: props}
}

// Render renders the chat interface
func (c *ChatComponent) Render(w io.Writer) error {
	html, err := c.build()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(html))
	return err
}

// HTML returns the chat as template.HTML
func (c *ChatComponent) HTML() (template.HTML, error) {
	html, err := c.build()
	if err != nil {
		return "", err
	}
	return template.HTML(html), nil
}

func (c *ChatComponent) build() (string, error) {
	themeClass := "chat-dark"
	if c.props.Theme == "light" {
		themeClass = "chat-light"
	}

	return fmt.Sprintf(`
<!-- Learning Desktop Chat Interface -->
<div id="learning-desktop" class="chat-interface %s" data-chat-session="%s" data-voice-chat="true" hx-ext="sse" sse-connect="%s">

	<!-- Header -->
	<header class="chat-header">
		<div class="chat-title">
			<h2>%s</h2>
			<span class="chat-status" data-chat-status="connecting">
				<span class="status-dot"></span>
				<span class="status-text">Connecting...</span>
			</span>
		</div>
		<div class="chat-actions">
			<button class="chat-btn chat-btn-icon" hx-post="/api/chat/clear" hx-target="#learning-desktop" hx-swap="outerHTML" title="Clear conversation">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
				</svg>
			</button>
			<button class="chat-btn chat-btn-icon" hx-get="/api/chat/export" hx-target="_blank" title="Export conversation">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
				</svg>
			</button>
		</div>
	</header>

	<!-- Messages Container -->
	<div class="chat-messages" id="chat-messages" hx-swap-oob="beforeend">
		<!-- Messages will be injected here -->
		<div class="chat-message chat-message-welcome">
			<div class="message-avatar">
				<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"/>
				</svg>
			</div>
			<div class="message-content">
				<p>%s</p>
			</div>
		</div>
	</div>

	<!-- Typing Indicator (hidden by default) -->
	<div class="chat-typing" data-chat-typing="hidden" style="display: none;">
		<span></span><span></span><span></span>
	</div>

	<!-- Input Area -->
	<div class="chat-input-area">
		<form hx-post="/api/chat" hx-target="#chat-messages" hx-swap="beforeend" hx-indicator="#chat-input">
			<div class="chat-input-wrapper">
				<!-- File Upload Button -->
				<button type="button" class="chat-btn chat-btn-attach" onclick="document.getElementById('chat-file-input').click()" title="Attach file">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M15.172 7l-6.586 6.586a2 2 0 112.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"/>
					</svg>
				</button>
				<input type="file" id="chat-file-input" class="hidden" hx-post="/api/chat/upload" hx-encoding="multipart/form-data" hx-target="#chat-messages" hx-swap="beforeend">

				<!-- Voice Input Button (Microphone) -->
				<button type="button" class="chat-btn chat-btn-mic" data-voice-mic title="Voice input">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
						<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
						<line x1="12" y1="19" x2="12" y2="23"/>
						<line x1="8" y1="23" x2="16" y2="23"/>
					</svg>
				</button>

				<!-- Text Input -->
				<textarea
					id="chat-input"
					name="message"
					class="chat-textarea"
					placeholder="%s"
					rows="1"
					data-chat-auto-resize
					data-chat-submit-on-enter="true"
					data-voice-interim=""
					required></textarea>

				<!-- Voice Output Toggle (Speaker) -->
				<button type="button" class="chat-btn chat-btn-speaker" data-voice-speaker title="Voice output">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/>
						<path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"/>
					</svg>
				</button>

				<!-- Send Button -->
				<button type="submit" class="chat-btn chat-btn-send" id="chat-send-btn" disabled title="Send message">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="22" y1="2" x2="11" y2="13"/>
						<polygon points="22 2 15 22 11 13 2 9 22 2 22"/>
					</svg>
				</button>
			</div>

			<!-- Voice interim transcript (hidden by default) -->
			<div data-voice-interim style="display: none; font-size: 12px; color: var(--chat-accent); margin-top: 4px;"></div>
		</form>
	</div>

	<!-- SSE Event Handlers -->
	<script src="/static/js/voice-chat.js"></script>
	<script>
		(function() {
			// Initialize Voice Chat
			if (window.VoiceChat) {
				window.voiceChat = new window.VoiceChat({
					inputSelector: '#chat-input',
					micButtonSelector: '[data-voice-mic]',
					speakerSelector: '[data-voice-speaker]',
					onTranscript: function(text) {
						console.log('Voice transcript:', text);
						// Auto-submit after voice input
						const form = document.querySelector('#chat-input').form;
						if (form) {
							setTimeout(function() {
								form.dispatchEvent(new Event('submit', {cancelable: true}));
							}, 500);
						}
					},
					onSpeakingStart: function() {
						document.body.setAttribute('data-voice-listening', 'true');
					},
					onSpeakingEnd: function() {
						document.body.removeAttribute('data-voice-listening');
					},
					onError: function(error) {
						console.error('Voice error:', error);
					},
					debug: true
				});
			}

			// Auto-resize textarea
			const textarea = document.getElementById('chat-input');
			if (textarea) {
				textarea.addEventListener('input', function() {
					this.style.height = 'auto';
					this.style.height = Math.min(this.scrollHeight, 150) + 'px';
					const sendBtn = document.getElementById('chat-send-btn');
					if (sendBtn) {
						sendBtn.disabled = this.value.trim() === '';
					}
				});

				textarea.addEventListener('keydown', function(e) {
					if (e.key === 'Enter' && !e.shiftKey) {
						e.preventDefault();
						this.form.dispatchEvent(new Event('submit', {cancelable: true}));
					}
				});
			}

			// Listen for SSE messages to trigger TTS
			if (window.EventSource) {
				const desktop = document.getElementById('learning-desktop');
				if (desktop) {
					desktop.addEventListener('htmx:sseMessage', function(e) {
						if (e.detail.type === 'chat.message' || e.detail.type === 'chat.complete') {
							const data = e.detail.data;
							if (data.message && window.voiceChat && !window.voiceChat.isMuted) {
								window.voiceChat.speak(data.message, {
									emotion: data.emotion || 'neutral'
								});
							}
						}
					});
				}
			}
		})();
	</script>
</div>
`, themeClass, c.props.SessionID, c.props.SSEEndpoint, c.props.WelcomeTitle, c.props.WelcomePrompt, c.props.Placeholder), nil
}

// RenderMessageHTML renders a single message directly to writer (for htmx-r integration)
func RenderMessageHTML(w io.Writer, msg ChatMessage) {
	roleClass := "chat-message-assistant"
	if msg.Role == "user" {
		roleClass = "chat-message-user"
	} else if msg.Role == "system" {
		roleClass = "chat-message-system"
	}

	mediaHTML := ""
	if msg.Media != nil {
		mediaHTML = renderMedia(msg.Media)
	}

	html := fmt.Sprintf(`
<div class="chat-message %s" data-message-id="%s">
	<div class="message-avatar">
		%s
	</div>
	<div class="message-content">
		%s
		<div class="message-text">%s</div>
		<div class="message-time">%s</div>
	</div>
</div>
`, roleClass, msg.ID, renderAvatar(msg.Role), mediaHTML, msg.Content, formatTimestamp(msg.Timestamp))

	w.Write([]byte(html))
}

// RenderMessage renders a single message (for server-side rendering)
func RenderMessage(msg ChatMessage) template.HTML {
	roleClass := "chat-message-assistant"
	if msg.Role == "user" {
		roleClass = "chat-message-user"
	} else if msg.Role == "system" {
		roleClass = "chat-message-system"
	}

	mediaHTML := ""
	if msg.Media != nil {
		mediaHTML = renderMedia(msg.Media)
	}

	return template.HTML(fmt.Sprintf(`
<div class="chat-message %s" data-message-id="%s">
	<div class="message-avatar">
		%s
	</div>
	<div class="message-content">
		%s
		<div class="message-text">%s</div>
		<div class="message-time">%s</div>
	</div>
</div>
`, roleClass, msg.ID, renderAvatar(msg.Role), mediaHTML, msg.Content, formatTimestamp(msg.Timestamp)))
}

// renderAvatar renders the avatar icon for a message role
func renderAvatar(role string) string {
	switch role {
	case "user":
		return `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/>
			<circle cx="12" cy="7" r="4"/>
		</svg>`
	case "assistant", "ai":
		return `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"/>
		</svg>`
	default:
		return ``
}

// renderMedia renders injected media based on type
func renderMedia(media *MediaRef) string {
	if media == nil {
		return ""
	}

	switch media.Type {
	case "video":
		return fmt.Sprintf(`
<div class="message-media message-media-video" data-media-type="video">
	<div class="media-player" data-media-source="%s">
		<div class="media-placeholder">
			<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<polygon points="5 3 19 12 5 21 5 3"/>
			</svg>
		</div>
	</div>
	<div class="media-caption">%s</div>
</div>
`, media.Source, media.Title)

	case "3d":
		return fmt.Sprintf(`
<div class="message-media message-media-3d" data-media-type="3d">
	<div class="media-3d-container" data-3d-scene="%s">
		<canvas id="canvas-%s"></canvas>
	</div>
	<div class="media-caption">%s</div>
	<div class="media-controls">
		<button class="media-btn" data-action="reset">Reset View</button>
		<button class="media-btn" data-action="animate">Auto-rotate</button>
	</div>
</div>
`, media.Source, media.ID(), media.Title)

	case "board":
		return fmt.Sprintf(`
<div class="message-media message-media-board" data-media-type="board">
	<div class="media-board-container" data-board-id="%s">
		<iframe src="%s" width="100%%" height="300"></iframe>
	</div>
	<div class="media-caption">%s</div>
</div>
`, media.Source, media.Source, media.Title)

	case "code":
		return fmt.Sprintf(`
<div class="message-media message-media-code">
	<div class="code-block" data-language="%s">
		<pre><code>%s</code></pre>
	</div>
	<div class="media-caption">%s</div>
</div>
`, media.Data["language"], htmlEscape(media.Source), media.Title)

	default:
		return ""
	}
}

// formatTimestamp formats a timestamp for display
func formatTimestamp(t time.Time) string {
	now := time.Now()
	if now.Sub(t) < 24*time.Hour {
		return t.Format("3:04 PM")
	}
	return t.Format("Jan 2")
}

// htmlEscape escapes HTML in strings
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
