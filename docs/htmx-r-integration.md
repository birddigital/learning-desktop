# Learning Desktop - htmx-r Integration Guide

Complete guide for integrating the Learning Desktop AI tutor platform with htmx-r.

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                         Browser (htmx-r)                         │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  <div id="learning-desktop" hx-ext="sse"                    │  │
│  │       sse-connect="/api/chat/events">                       │  │
│  │    <!-- Chat messages injected here via HTMX -->            │  │
│  │    <form hx-post="/api/chat" hx-swap="beforeend">          │  │
│  │  </div>                                                     │  │
│  └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
                              │
                              │ HTTP + SSE
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                       Go Server (htmx-r)                         │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────────┐    │
│  │ChatHandler  │  │Event Bus     │  │MediaInjectionEngine  │    │
│  │- handleChat │  │- Publish     │  │- Decide()            │    │
│  │- handleStream│ │- Subscribe   │  │  - Video             │    │
│  └──────┬──────┘  └──────┬───────┘  │  - 3D                │    │
│         │                │          │  - Board             │    │
│         └────────────────┼──────────│  - Code              │    │
│                          │          │  - Quiz              │    │
│  ┌───────────────────────▼──────────▼──────────────────────┐   │
│  │                    AITutorService                       │   │
│  │  - GenerateResponse()                                   │   │
│  │  - StreamResponse()                                     │   │
│  └───────────────────────┬─────────────────────────────────┘   │
│                          │                                       │
│  ┌───────────────────────▼─────────────────────────────────┐   │
│  │                  ProgressStore                           │   │
│  │  - GetSession() / CreateSession()                       │   │
│  │  - AddMessage() / GetMessages()                         │   │
│  │  - RecordProgress()                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
                              │
                              │ PostgreSQL RLS
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                      Database (PostgreSQL)                       │
│  students | student_sessions | chat_messages | progress_events  │
│  courses | course_modules | course_lessons                     │
└──────────────────────────────────────────────────────────────────┘
```

---

## Quick Start

### 1. File Structure

```
htmx-r/
├── internal/learning/
│   ├── chat/
│   │   ├── chat.go              # Chat component & types
│   │   ├── chat_test.go         # Tests
│   │   └── media.go             # Media rendering
│   ├── media/
│   │   └── injection.go         # Media injection engine
│   ├── reliability/
│   │   ├── circuit.go           # Circuit breaker
│   │   └── pool.go              # Resource pool
│   └── learning.go              # Main package exports
├── static/css/
│   └── chat.css                 # Chat interface styles
├── docs/
│   ├── course-content.md        # Course structure
│   ├── database-schema.md       # Database design
│   └── htmx-r-integration.md    # This file
└── examples/
    └── tenant-aware-jobs.go     # Background jobs pattern
```

### 2. Basic Setup

```go
package main

import (
    "context"
    "database/sql"
    "log"
    "net/http"

    "github.com/birddigital/htmx-r/internal/event"
    "github.com/birddigital/htmx-r/internal/learning"
    _ "github.com/lib/pq"
)

func main() {
    // Setup database
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/learning")
    if err != nil {
        log.Fatal(err)
    }

    // Setup event bus for SSE
    bus := event.NewBus()

    // Setup AI tutor (Claude API integration)
    tutor := NewClaudeTutor("your-api-key")

    // Setup progress store
    progressStore := NewPostgresProgressStore(db)

    // Setup chat handler
    chatHandler := learning.NewChatHandler(bus, tutor, progressStore)

    // Setup router
    mux := http.NewServeMux()

    // Serve static files
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // Register chat routes
    chatHandler.RegisterRoutes(mux)

    // Start server
    log.Println("Server running on :3000")
    log.Fatal(http.ListenAndServe(":3000", mux))
}
```

### 3. HTML Template

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Get Ahead of AI - Learning Desktop</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10/dist/ext/sse.js"></script>
    <link rel="stylesheet" href="/static/css/chat.css">
</head>
<body>
    <!-- Learning Desktop Chat Interface -->
    {{ .ChatComponent }}

    <script>
        // Optional: Add custom behavior
        document.body.addEventListener('htmx:sseMessage', function(e) {
            if (e.detail.type === 'chat.message') {
                // Handle incoming streaming message
                console.log('Message chunk:', e.detail.data);
            }
        });
    </script>
</body>
</html>
```

---

## HTTP API

### POST /api/chat

Send a message and get a response.

**Request:**
```json
{
    "session_id": "uuid",
    "message": "How do LLMs work?",
    "lesson_id": "llm-how-it-works",
    "module_id": "ai-literacy"
}
```

**Response:** HTML fragment for message insertion

### POST /api/chat/stream

Stream a response via SSE.

**Request:** Same as `/api/chat`

**Response:** SSE stream
```
event: start
data: {}

event: chunk
data: {"chunk": "Large "}

event: chunk
data: {"chunk": "Language "}

event: complete
data: {"message": "...", "media": {...}}
```

### GET /api/chat/history

Retrieve conversation history.

**Query Parameters:**
- `session_id`: Session UUID
- `limit`: Max messages (default: 50)

**Response:** HTML fragment of messages

### POST /api/chat/clear

Clear current conversation and start fresh.

**Response:** Fresh chat interface HTML

### GET /api/chat/export

Export conversation as markdown.

**Response:** Markdown file download

---

## Media Injection

The AI tutor automatically injects media based on conversation context:

### How It Works

1. **Student sends message** → `ChatHandler` receives request
2. **AI generates response** → `AITutorService.StreamResponse()`
3. **Injection decision** → `MediaInjectionEngine.Decide()`
4. **Media rendered** → Appropriate component HTML
5. **SSE broadcast** → Client receives complete response

### Injection Rules

| Trigger | Media | Example |
|---------|-------|---------|
| "show me", "visualize" | 2D Board | Student asks for diagram |
| "confused", "don't understand" | 3D Model | After 2+ explanations |
| "code example" | Interactive Code | Student wants implementation |
| "how do I..." | Workflow Board | Process planning |
| "quiz me" | Quiz | Assessment request |
| Beginner skill level | Video | First lesson interaction |

### Custom Rules

```go
// Add custom injection rule
engine.AddRule(&media.InjectionRule{
    ID:   "my-custom-rule",
    Name: "My Custom Media",
    Trigger: &media.TriggerCondition{
        StudentContains: []string{"keyword"},
    },
    MediaType: media.MediaTypeBoard,
    Confidence: media.ConfidenceHigh,
    MediaSource: func(ctx context.Context, req *media.InjectionRequest) (string, map[string]interface{}, error) {
        return "custom-board", map[string]interface{}{
            "type": "my-template",
        }, nil
    },
})
```

---

## SSE Events

### Client-Side

```html
<div id="learning-desktop"
     hx-ext="sse"
     sse-connect="/api/chat/events"
     sse-swap="message">
    <!-- Messages injected here -->
</div>

<div hx-ext="sse" sse-connect="/api/chat/events">
    <!-- Progress updates -->
    <div sse-swap="progress.update" hx-swap="innerHTML">
        <span class="completion-percent">0%</span>
    </div>
</div>
```

### Server-Side Events

| Event Type | Payload | Purpose |
|------------|---------|---------|
| `chat.message` | `{chunk}` | Streaming response |
| `chat.complete` | `{message, media}` | Response finished |
| `progress.update` | `{completion_percent}` | Progress update |
| `media.inject` | `{type, source}` | Media injection |
| `checkpoint.passed` | `{lesson_id, score}` | Checkpoint result |

---

## Progress Tracking

### Recording Progress

```go
// Record lesson started
handler.进度Store.RecordProgress(ctx, &learning.ProgressEvent{
    StudentID: studentID,
    EventType: "lesson_started",
    LessonID:  lessonID,
    Timestamp: time.Now(),
})

// Record lesson completed
handler.进度Store.RecordProgress(ctx, &learning.ProgressEvent{
    StudentID: studentID,
    EventType: "lesson_completed",
    LessonID:  lessonID,
    Timestamp: time.Now(),
})
```

### Querying Progress

```go
progress, err := handler.进度Store.GetStudentProgress(ctx, studentID, courseID)
if err != nil {
    return err
}

fmt.Printf("Progress: %.0f%%\n", progress.CompletionPercent)
fmt.Printf("Current Lesson: %d\n", progress.CurrentLessonIndex)
```

---

## Reliability Patterns

### Circuit Breaker

```go
import "github.com/birddigital/htmx-r/internal/reliability"

circuit := reliability.NewCircuitBreaker(&reliability.CircuitBreakerConfig{
    OpenThreshold:    5,   // 5 consecutive failures
    HalfOpenMaxCalls: 3,   // 3 trial calls
    ResetTimeout:     60 * time.Second,
})

// Execute with circuit breaker protection
err := circuit.Execute(ctx, func() error {
    return tutor.GenerateResponse(ctx, req)
})
```

### Resource Pool

```go
pool := reliability.NewStudentSessionPool(&reliability.ResourcePoolConfig{
    MaxConcurrentStudents: 500,
    MaxConcurrentRequests: 50,
    SessionTimeout:        30 * time.Minute,
    ClaudeRateLimit:       50,  // requests/second
}, circuit)

// Submit request to pool
err := pool.SubmitRequest(ctx, &reliability.StudentRequest{
    StudentID:   studentID,
    TenantID:    tenantID,
    Prompt:      prompt,
    Context:     ctx,
    ResponseCh:  responseCh,
    ErrorCh:     errorCh,
})
```

---

## Tenant-Aware Background Jobs

```go
// Based on bird-m1 pattern
cron := NewLearningDesktopCronJobs(db, nc, logger)
pool := NewTenantJobPool(10, cron) // 10 workers

// Register jobs
pool.Register(&ProgressSyncJob{db: db, nats: nc, logger: logger})
pool.Register(&SessionCleanupJob{db: db, logger: logger})

// Execute job across all active tenants
ctx := context.Background()
err := pool.ExecuteJob(ctx, "progress-sync")
```

---

## Testing

```go
func TestChatHandler(t *testing.T) {
    // Setup
    bus := event.NewBus()
    tutor := &MockTutor{}
    store := &MockProgressStore{}
    handler := learning.NewChatHandler(bus, tutor, store)

    // Test request
    req := ChatRequest{
        SessionID: "test-session",
        Message:   "Hello!",
    }

    // Execute
    w := httptest.NewRecorder()
    r := httptest.NewRequest("POST", "/api/chat", toJSONReader(req))
    handler.handleChat(w, r)

    // Assert
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }
}
```

---

## Deployment

### Using serve-r.ai CLI

```bash
# Build and deploy
serve-r deploy --prod

# Check logs
serve-r logs --tail

# Set environment variables
serve-r env set CLAUDE_API_KEY=sk-xxx
serve-r env set DATABASE_URL=postgres://...
```

### Docker

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o learning-desktop ./cmd/server

FROM alpine:latest
COPY --from=builder /app/learning-desktop /usr/local/bin/
COPY static /static
EXPOSE 3000
CMD ["learning-desktop"]
```

---

## Monitoring

### Prometheus Metrics

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
```

### Key Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `learning_desktop_active_sessions` | Gauge | Current active sessions |
| `learning_desktop_circuit_state` | Gauge | Circuit breaker state (0/1/2) |
| `learning_desktop_rejected_requests_total` | Counter | Requests rejected due to limits |

---

`★ Insight ─────────────────────────────────────`
**htmx-r Integration Patterns Discovered:**

1. **SSE for Streaming**: htmx's `sse-swap` attribute enables perfect server-sent event integration for AI streaming responses without custom JavaScript.

2. **OOB Swap Targets**: Use `hx-swap-oob="beforeend"` to inject messages outside the form, allowing the form to remain pristine for re-submission.

3. **Component Composition**: The chat component is fully self-contained with embedded CSS and minimal inline JS, making it drop-in compatible with any htmx-r app.
`─────────────────────────────────────────────────`