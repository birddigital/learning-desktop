# Voice Integration Guide

Complete guide for voice capabilities in the Learning Desktop AI tutor platform.

## Overview

The voice system provides:
- **Speech-to-Text (STT)**: Students speak their input instead of typing
- **Text-to-Speech (TTS)**: AI tutor speaks responses aloud
- **Voice Activity Detection (VAD)**: Detects when student is speaking
- **Speech Analysis**: Extracts learning insights from student speech patterns

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Browser Layer                            │
│  ┌────────────────────────────────────────────────────────────┐│
│  │  voice-chat.js                                             ││
│  │  - Web Speech API (SpeechRecognition)                      ││
│  │  - Web Speech API (SpeechSynthesis)                        ││
│  │  - AudioWorklet VAD (optional)                             ││
│  └────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ Fallback for unsupported browsers
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Server Layer (Go)                         │
│  ┌────────────────────────────────────────────────────────────┐│
│  │  VoiceHandler (internal/voice/service.go)                  ││
│  │  ┌─────────────┬─────────────┬──────────────────────────┐ ││
│  │  │ STT Handler │ TTS Handler │ VAD Handler              │ ││
│  │  │ /api/voice  │ /api/voice  │ /api/voice                │ ││
│  │  │   /stt      │   /tts      │   /vad                    │ ││
│  │  └─────────────┴─────────────┴──────────────────────────┘ ││
│  │                                                             ││
│  │  ┌─────────────────────────────────────────────────────┐  ││
│  │  │  Speech Analysis                                     │  ││
│  │  │  - Sentiment detection (confident/uncertain)         │  ││
│  │  │  - Filler word counting (um, uh, like)              │  ││
│  │  │  - Speaking pace calculation                        │  ││
│  │  │  - Understanding score (0-1)                        │  ││
│  │  └─────────────────────────────────────────────────────┘  ││
│  └────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ External APIs (optional)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      External Services                           │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────────┐
│  │ OpenAI       │  │ ElevenLabs   │  │ WebRTC VAD             │
│  │ Whisper STT  │  │ TTS          │  │ (embedded)              │
│  └──────────────┘  └──────────────┘  └─────────────────────────┘
└─────────────────────────────────────────────────────────────────┘
```

---

## Quick Start

### 1. Browser-Native Setup (Zero Dependencies)

The simplest approach uses the Web Speech API built into modern browsers:

```html
<!-- Include the voice chat script -->
<script src="/static/js/voice-chat.js"></script>

<!-- Chat interface with voice support -->
<div data-voice-chat="true">
    <input id="chat-input" type="text">
    <button data-voice-mic>🎤</button>
    <button data-voice-speaker>🔊</button>
</div>
```

**Browser Support:**
- ✅ Chrome/Edge: Full STT + TTS support
- ✅ Safari: TTS only (no STT)
- ❌ Firefox: No Web Speech API

**Fallback:** When browser doesn't support Web Speech API, the system automatically falls back to server-side processing.

### 2. Go Server Setup

```go
package main

import (
    "database/sql"
    "net/http"

    "github.com/birddigital/htmx-r/internal/event"
    "github.com/birddigital/htmx-r/internal/learning"
    "github.com/birddigital/htmx-r/internal/voice"
)

func main() {
    bus := event.NewBus()

    // Create voice service
    stt := voice.NewWhisperSTT("your-api-key")
    tts := voice.NewElevenLabsTTS("your-api-key", "voice-id")
    vad := voice.NewWebRTCVAD(16000, 3)

    voiceService := voice.NewVoiceService(stt, tts, vad)
    voiceHandler := voice.NewVoiceHandler(voiceService)

    // Register routes
    mux := http.NewServeMux()
    voiceHandler.RegisterRoutes(mux)
}
```

---

## Client-Side API

### VoiceChat Class

```javascript
// Initialize
const voice = new VoiceChat({
    inputSelector: '#chat-input',
    micButtonSelector: '[data-voice-mic]',
    speakerSelector: '[data-voice-speaker]',
    onTranscript: (text) => console.log('Heard:', text),
    onSpeakingStart: () => console.log('Listening...'),
    onSpeakingEnd: () => console.log('Stopped'),
    onError: (error) => console.error('Error:', error),
    debug: true
});
```

### Methods

| Method | Description |
|--------|-------------|
| `startListening()` | Start speech recognition |
| `stopListening()` | Stop speech recognition |
| `speak(text, options)` | Speak text aloud |
| `stopSpeaking()` | Stop current speech |
| `toggleMute()` | Toggle voice output |
| `analyzeSpeech(transcript, timing)` | Analyze speech for insights |

### Speech Commands

Students can use voice commands:

| Command | Action |
|---------|--------|
| "Start listening" / "Listen" | Start listening |
| "Stop listening" / "Stop" | Stop listening |
| "Submit" / "Send" | Submit current input |
| "Clear" | Clear input |

### Text-to-Speech Options

```javascript
voice.speak('Hello, student!', {
    voice: 'Google US English',  // Voice name
    rate: 1.0,                   // 0.1 to 10
    pitch: 1.0,                  // 0 to 2
    volume: 1.0,                 // 0 to 1
    emotion: 'excited'           // 'excited' | 'calm' | 'confused'
});
```

---

## Server-Side API

### POST /api/voice/stt

Transcribe audio to text.

**Request:** `multipart/form-data` with audio file
```bash
curl -X POST http://localhost:3000/api/voice/stt \
    -F "audio=@recording.wav"
```

**Response:**
```json
{
    "transcription": {
        "text": "how do llms actually work",
        "confidence": 0.95,
        "duration": "2.5s",
        "language": "en-US",
        "words": [
            {"word": "how", "start": "0.0s", "end": "0.3s"},
            {"word": "do", "start": "0.4s", "end": "0.6s"}
        ]
    },
    "analysis": {
        "sentiment": "engaged",
        "pace": {"words_per_minute": 120, "label": "normal"},
        "fillers": [],
        "understanding_score": 0.6
    }
}
```

### POST /api/voice/tts

Convert text to speech audio.

**Request:**
```json
{
    "text": "Large Language Models work by predicting the next token.",
    "voice": "alloy",
    "speed": 1.0,
    "emotion": "calm"
}
```

**Response:** Audio file (audio/mpeg)

### POST /api/voice/vad

Detect voice activity in audio.

**Request:** Raw audio bytes

**Response:**
```json
{
    "is_speech": true,
    "confidence": 0.87
}
```

---

## Speech Analysis for Learning

The system analyzes student speech to extract learning insights:

### Metrics Tracked

| Metric | Description | Learning Value |
|--------|-------------|----------------|
| **Sentiment** | confident, uncertain, confused, engaged | Indicates comprehension level |
| **Filler Words** | "um", "uh", "like", "you know" | Higher count = uncertainty |
| **Speaking Pace** | words per minute | Fast = excited, Slow = thinking |
| **Questions Asked** | Count of question marks | Engagement indicator |
| **Concepts Mentioned** | Technical terms used | Knowledge demonstration |
| **Understanding Score** | 0-1 aggregate score | Progress tracking |

### Example Analysis

```json
{
    "transcript": "I think I understand, but like, how does the attention mechanism actually work?",
    "sentiment": "engaged",
    "pace": {"words_per_minute": 135, "label": "normal"},
    "fillers": [
        {"word": "like", "position": 1.2, "type": "filler"}
    ],
    "concepts_mentioned": ["attention", "mechanism"],
    "questions_asked": 1,
    "understanding_score": 0.65
}
```

**Interpretation:**
- Student is engaged (asked a question)
- Knows relevant terminology ("attention mechanism")
- Some uncertainty (one filler word)
- Good understanding score (0.65)

---

## Student Trajectory Tracking

Voice analysis feeds into the student's learning trajectory:

### 1. Initial Assessment

```go
// Analyze first voice interaction
analysis := voiceService.AnalyzeSpeech(firstTranscription)

// Update student state
student.SkillLevel = inferSkillLevel(analysis)
student.Goals = extractGoals(analysis.Transcript)
student.Interests = extractInterests(analysis.ConceptsMentioned)
```

### 2. Progress Tracking

```go
// Track understanding over time
type ProgressTrajectory struct {
    SessionID       string
    BaselineScore   float64
    CurrentScore    float64
    Trend           string  // "improving" | "stable" | "declining"
    ConceptMastery  map[string]int
}

// Analyze trend
if analysis.UnderstandingScore > baseline + 0.1 {
    trajectory.Trend = "improving"
}
```

### 3. Adaptive Responses

```go
// Adjust AI tutor behavior based on speech analysis
func (t *Tutor) GenerateResponse(ctx context.Context, analysis *voice.StudentSpeechAnalysis) string {
    switch analysis.Sentiment {
    case "confused":
        return "I sense you're confused. Let me explain this differently..."
    case "confident":
        return "Great! You've got this. Ready for the next concept?"
    case "engaged":
        return "Excellent question! Let's dive deeper..."
    }
}
```

---

## Server-Side Providers

### OpenAI Whisper (STT)

```go
stt := voice.NewWhisperSTT("sk-...")
result, err := stt.Transcribe(ctx, audioReader)
```

### ElevenLabs (TTS)

```go
tts := voice.NewElevenLabsTTS("xi-...", "your-voice-id")
result, err := tts.Synthesize(ctx, &voice.SynthesisRequest{
    Text: "Hello, student!",
    Voice: "alloy",
    Speed: 1.0,
})
```

### WebRTC VAD

```go
vad := voice.NewWebRTCVAD(16000, 3)
isSpeech, confidence := vad.DetectSpeech(audioChunk)
```

---

## CSS Styles

```css
/* Microphone button */
.chat-btn-mic.listening {
    color: #ef4444;
    animation: micPulse 1.5s infinite;
}

@keyframes micPulse {
    0%, 100% { box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.4); }
    50% { box-shadow: 0 0 0 8px rgba(239, 68, 68, 0); }
}

/* Speaker button */
.chat-btn-speaker.speaking {
    color: var(--chat-accent);
}

/* Global listening indicator */
body[data-voice-listening="true"] .chat-interface {
    box-shadow: 0 0 0 2px #ef4444;
}
```

---

## VAD (Voice Activity Detection)

### Browser-Native VAD

Simple energy-based VAD using Web Audio API:

```javascript
// In voice-chat.js
setupVAD: function() {
    this.audioContext = new AudioContext();
    this.analyzer = this.audioContext.createAnalyser();
    this.vadThreshold = 0.02;

    // Analyze audio stream
    // Detect speech vs silence
}
```

### Server-Side VAD

For production use, integrate:

1. **WebRTC VAD** - Lightweight, embedded
2. **Silero VAD** - More accurate, requires model
3. **Azure/AWS VAD** - Cloud-based

---

## Deployment Considerations

### Cost (Server-Side Processing)

| Service | Cost (approx) |
|---------|---------------|
| Whisper STT | $0.006 / minute |
| ElevenLabs TTS | $0.30 / 1K chars |
| WebRTC VAD | Free (embedded) |

**Recommendation:** Start with browser-native (free), upgrade to server-side for better quality.

### Latency

| Method | Latency |
|--------|---------|
| Browser STT | ~100ms |
| Whisper API | ~500ms |
| Browser TTS | ~50ms |
| ElevenLabs | ~200ms |

---

`★ Insight ─────────────────────────────────────`
**Voice-First Learning Patterns:**

1. **Speech Analysis Hidden Gems**: Filler words and speaking pace are stronger indicators of student understanding than transcript content alone. A student saying "I understand... um... the attention mechanism" with slow pace and pauses signals uncertainty despite confident words.

2. **Progressive Enhancement Architecture**: Browser-native Web Speech API works immediately for 80% of users (Chrome/Edge). Server-side fallback covers the rest. No "all or nothing" implementation needed.

3. **Emotion Simulation**: Simple rate/pitch adjustments in TTS can convey emotion (faster + higher pitch = excited, slower + lower pitch = calm) without requiring emotion-specific models.
`─────────────────────────────────────────────────`