# Learning Desktop

An AI-powered learning platform with a Claude Desktop-style interface for teaching "how to not fall behind in the oncoming AI onslaught."

## Overview

Learning Desktop is a web-based AI tutor platform featuring:

- **Voice-First Learning**: Speech-to-text for student input, text-to-speech for AI responses
- **Speech Analysis**: Extract learning insights from student speech patterns
- **Student Trajectory Tracking**: Monitor progress over time through voice interactions
- **Media Injection**: AI-driven content injection (video, 3D, whiteboards, code)
- **Multi-Tenant Architecture**: Built for educational organizations

## Tech Stack

- **Backend**: Go 1.23 with htmx-r framework
- **Frontend**: HTMX + vanilla JavaScript
- **Voice**: Web Speech API (browser-native) with server-side fallback
- **Database**: PostgreSQL with Row-Level Security
- **Real-time**: Server-Sent Events (SSE)

## Quick Start

```bash
# Clone the repository
git clone https://github.com/birddigital/learning-desktop.git
cd learning-desktop

# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go
```

Visit `http://localhost:3000` to access the Learning Desktop.

## Voice Features

### Browser-Native (Zero Dependencies)

The simplest approach uses the Web Speech API built into modern browsers:

- **Chrome/Edge**: Full STT + TTS support
- **Safari**: TTS only (no STT)
- **Firefox**: No Web Speech API (falls back to server-side)

### Speech Analysis for Learning

The system analyzes student speech to extract learning insights:

| Metric | Learning Value |
|--------|----------------|
| **Sentiment** | confident, uncertain, confused, engaged |
| **Filler Words** | "um", "uh", "like", "you know" - indicates uncertainty |
| **Speaking Pace** | Fast = excited, Slow = thinking |
| **Concepts Mentioned** | Technical terms used = knowledge demonstration |
| **Understanding Score** | 0-1 aggregate for progress tracking |

### Example Analysis

```json
{
    "transcript": "I think I understand, but like, how does the attention mechanism actually work?",
    "sentiment": "engaged",
    "pace": {"words_per_minute": 135, "label": "normal"},
    "fillers": [{"word": "like", "type": "filler"}],
    "concepts_mentioned": ["attention", "mechanism"],
    "questions_asked": 1,
    "understanding_score": 0.65
}
```

## Project Structure

```
learning-desktop/
├── cmd/
│   └── server/
│       └── main.go          # Server entry point
├── internal/
│   └── voice/
│       └── service.go       # Voice service (STT/TTS/VAD)
├── pkg/
│   └── chat/
│       └── chat.go          # Chat component
├── static/
│   ├── css/
│   │   └── chat.css         # Chat interface styles
│   └── js/
│       └── voice-chat.js    # Voice client
└── docs/
    ├── voice-integration.md # Voice integration guide
    ├── database-schema.md   # Database design
    └── course-content.md    # Course structure
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat` | POST | Send chat message |
| `/api/chat/clear` | POST | Clear conversation |
| `/api/chat/export` | GET | Export conversation |
| `/api/voice/stt` | POST | Speech-to-text |
| `/api/voice/tts` | POST | Text-to-speech |
| `/api/voice/analyze` | POST | Analyze speech for insights |
| `/api/voice/vad` | POST | Voice activity detection |

## Documentation

- [Voice Integration Guide](docs/voice-integration.md) - Complete voice capabilities
- [Database Schema](docs/database-schema.md) - Multi-tenant PostgreSQL design
- [Course Content](docs/course-content.md) - Learning module structure

## License

MIT License - see LICENSE file for details.
