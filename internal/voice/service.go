// Package voice provides speech-to-text and text-to-speech services for the Learning Desktop.
// Supports both browser-native Web Speech API (client) and server-side processing.
package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MediaType represents the type of media to inject
type MediaType string

const (
	MediaTypeVideo  MediaType = "video"
	MediaType3D     MediaType = "3d"
	MediaTypeBoard  MediaType = "board"
	MediaTypeCode   MediaType = "code"
	MediaTypeQuiz   MediaType = "quiz"
)

// STTProvider defines speech-to-text capabilities
type STTProvider interface {
	Transcribe(ctx context.Context, audio io.Reader) (*TranscriptionResult, error)
}

// TTSProvider defines text-to-speech capabilities
type TTSProvider interface {
	Synthesize(ctx context.Context, req *SynthesisRequest) (*SynthesisResult, error)
}

// VADProvider defines voice activity detection
type VADProvider interface {
	DetectSpeech(audio []byte) (bool, float64)
}

// TranscriptionResult is the output from speech-to-text
type TranscriptionResult struct {
	Text       string        `json:"text"`
	Confidence float64       `json:"confidence"`
	Duration   time.Duration `json:"duration"`
	Language   string        `json:"language"`
	Words      []Word        `json:"words,omitempty"`
}

// Word represents a single word with timing
type Word struct {
	Word       string        `json:"word"`
	Start      time.Duration `json:"start"`
	End        time.Duration `json:"end"`
	Confidence float64       `json:"confidence"`
}

// SynthesisRequest is input for text-to-speech
type SynthesisRequest struct {
	Text    string `json:"text"`
	Voice   string `json:"voice"`
	Speed   float64 `json:"speed"`
	Pitch   float64 `json:"pitch"`
	Emotion string `json:"emotion"`
}

// SynthesisResult is the output from text-to-speech
type SynthesisResult struct {
	Audio    []byte      `json:"-"`
	Duration time.Duration `json:"duration"`
}

// StudentSpeechAnalysis contains insights from analyzing student speech
type StudentSpeechAnalysis struct {
	SessionID          string    `json:"session_id"`
	Timestamp          time.Time `json:"timestamp"`
	Transcript         string    `json:"transcript"`
	Topics             []string  `json:"topics"`
	Sentiment          string    `json:"sentiment"`
	Fillers            []FillerWord `json:"fillers"`
	Pace               SpeechPace    `json:"pace"`
	ConceptsMentioned  []string  `json:"concepts_mentioned"`
	QuestionsAsked     int       `json:"questions_asked"`
	UnderstandingScore  float64   `json:"understanding_score"`
}

// FillerWord represents a filler word detection
type FillerWord struct {
	Word     string        `json:"word"`
	Position time.Duration `json:"position"`
	Type     string        `json:"type"`
}

// SpeechPace describes speaking speed
type SpeechPace struct {
	WordsPerMinute int    `json:"words_per_minute"`
	Label          string `json:"label"`
}

// VoiceService orchestrates STT, TTS, and VAD
type VoiceService struct {
	stt STTProvider
	tts TTSProvider
	vad VADProvider
}

// NewVoiceService creates a new voice service
func NewVoiceService(stt STTProvider, tts TTSProvider, vad VADProvider) *VoiceService {
	return &VoiceService{stt: stt, tts: tts, vad: vad}
}

// ProcessStudentInput handles speech-to-text with VAD
func (s *VoiceService) ProcessStudentInput(ctx context.Context, audio []byte) (*TranscriptionResult, *StudentSpeechAnalysis, error) {
	isSpeech, _ := s.vad.DetectSpeech(audio)
	if !isSpeech {
		return nil, nil, fmt.Errorf("no speech detected")
	}

	result, err := s.stt.Transcribe(ctx, audio)
	if err != nil {
		return nil, nil, fmt.Errorf("transcription failed: %w", err)
	}

	analysis := s.AnalyzeSpeech(result)
	return result, analysis, nil
}

// GenerateTutorResponse converts text to speech
func (s *VoiceService) GenerateTutorResponse(ctx context.Context, text string, emotion string) (*SynthesisResult, error) {
	req := &SynthesisRequest{
		Text:    text,
		Voice:   "alloy",
		Speed:   1.0,
		Pitch:   1.0,
		Emotion: emotion,
	}
	return s.tts.Synthesize(ctx, req)
}

// AnalyzeSpeech extracts learning insights from transcription
func (s *VoiceService) AnalyzeSpeech(result *TranscriptionResult) *StudentSpeechAnalysis {
	analysis := &StudentSpeechAnalysis{
		Timestamp:         time.Now(),
		Transcript:        result.Text,
		Topics:            s.extractTopics(result.Text),
		Sentiment:         s.detectSentiment(result.Text),
		Fillers:           s.detectFillers(result.Text),
		Pace:              s.calculatePace(result.Text, result.Duration),
		ConceptsMentioned: s.extractConcepts(result.Text),
		QuestionsAsked:    s.countQuestions(result.Text),
	}
	analysis.UnderstandingScore = s.calculateUnderstandingScore(analysis)
	return analysis
}

func (s *VoiceService) extractTopics(text string) []string {
	topics := []string{}
	keywords := map[string]bool{
		"ai": true, "machine learning": true, "llm": true,
		"prompt": true, "chatgpt": true, "claude": true,
	}
	textLower := strings.ToLower(text)
	for keyword := range keywords {
		if strings.Contains(textLower, keyword) {
			topics = append(topics, keyword)
		}
	}
	return topics
}

func (s *VoiceService) detectSentiment(text string) string {
	textLower := strings.ToLower(text)

	confusedWords := []string{"confused", "don't understand", "lost", "not sure"}
	for _, word := range confusedWords {
		if strings.Contains(textLower, word) {
			return "confused"
		}
	}

	confidentWords := []string{"understand", "got it", "clear", "makes sense"}
	for _, word := range confidentWords {
		if strings.Contains(textLower, word) {
			return "confident"
		}
	}

	if strings.Contains(textLower, "?") || strings.Contains(textLower, "how") ||
	   strings.Contains(textLower, "what") || strings.Contains(textLower, "why") {
		return "engaged"
	}

	return "neutral"
}

func (s *VoiceService) detectFillers(text string) []FillerWord {
	fillers := []FillerWord{}
	fillerTypes := []string{" um ", " uh ", " like ", " you know "}
	words := strings.Split(text, " ")

	for i, word := range words {
		wordLower := strings.ToLower(word)
		for _, filler := range fillerTypes {
			if strings.Contains(" "+wordLower+" ", filler) {
				fillers = append(fillers, FillerWord{
					Word: word,
					Type: "filler",
				})
			}
		}
	}
	return fillers
}

func (s *VoiceService) calculatePace(text string, duration time.Duration) SpeechPace {
	words := strings.Fields(text)
	if len(words) == 0 || duration.Seconds() == 0 {
		return SpeechPace{WordsPerMinute: 0, Label: "unknown"}
	}

	wpm := int(float64(len(words)) / duration.Minutes())

	label := "normal"
	if wpm < 120 {
		label = "slow"
	} else if wpm > 160 {
		label = "fast"
	}

	return SpeechPace{WordsPerMinute: wpm, Label: label}
}

func (s *VoiceService) extractConcepts(text string) []string {
	concepts := []string{}
	terms := []string{
		"transformer", "attention", "token", "embedding", "prompt",
		"finetuning", "rag", "temperature",
	}
	textLower := strings.ToLower(text)
	for _, term := range terms {
		if strings.Contains(textLower, term) {
			concepts = append(concepts, term)
		}
	}
	return concepts
}

func (s *VoiceService) countQuestions(text string) int {
	count := 0
	for _, r := range text {
		if r == '?' {
			count++
		}
	}
	return count
}

func (s *VoiceService) calculateUnderstandingScore(analysis *StudentSpeechAnalysis) float64 {
	score := 0.5

	switch analysis.Sentiment {
	case "confident":
		score += 0.3
	case "confused":
		score -= 0.2
	case "engaged":
		score += 0.1
	}

	score += float64(len(analysis.ConceptsMentioned)) * 0.05
	score += float64(analysis.QuestionsAsked) * 0.02

	if len(analysis.Fillers) > 3 {
		score -= 0.1
	}

	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// WhisperSTT implements STT using OpenAI Whisper
type WhisperSTT struct {
	apiKey string
}

// NewWhisperSTT creates a new Whisper STT provider
func NewWhisperSTT(apiKey string) *WhisperSTT {
	return &WhisperSTT{apiKey: apiKey}
}

// Transcribe transcribes audio using Whisper API
func (w *WhisperSTT) Transcribe(ctx context.Context, audio io.Reader) (*TranscriptionResult, error) {
	// TODO: Implement OpenAI Whisper API call
	return &TranscriptionResult{}, nil
}

// ElevenLabsTTS implements TTS using ElevenLabs
type ElevenLabsTTS struct {
	apiKey  string
	voiceID string
}

// NewElevenLabsTTS creates a new ElevenLabs TTS provider
func NewElevenLabsTTS(apiKey, voiceID string) *ElevenLabsTTS {
	return &ElevenLabsTTS{apiKey: apiKey, voiceID: voiceID}
}

// Synthesize generates speech using ElevenLabs API
func (e *ElevenLabsTTS) Synthesize(ctx context.Context, req *SynthesisRequest) (*SynthesisResult, error) {
	// TODO: Implement ElevenLabs API call
	return &SynthesisResult{}, nil
}

// WebRTCVAD implements voice activity detection
type WebRTCVAD struct{}

// NewWebRTCVAD creates a new WebRTC VAD
func NewWebRTCVAD() *WebRTCVAD {
	return &WebRTCVAD{}
}

// DetectSpeech detects speech in audio
func (v *WebRTCVAD) DetectSpeech(audio []byte) (bool, float64) {
	// TODO: Implement WebRTC VAD
	return true, 0.8
}

// VoiceHandler handles voice API endpoints
type VoiceHandler struct {
	voice *VoiceService
}

// NewVoiceHandler creates a new voice handler
func NewVoiceHandler(voice *VoiceService) *VoiceHandler {
	return &VoiceHandler{voice: voice}
}

// RegisterRoutes registers voice HTTP routes
func (h *VoiceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/voice/tts", h.handleTTS)
	mux.HandleFunc("/api/voice/analyze", h.handleAnalyze)
}

func (h *VoiceHandler) handleTTS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SynthesisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, err := h.voice.GenerateTutorResponse(ctx, req.Text, req.Emotion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Write(result.Audio)
}

func (h *VoiceHandler) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Transcript string `json:"transcript"`
		Duration   int    `json:"duration_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result := &TranscriptionResult{
		Text:     req.Transcript,
		Duration: time.Duration(req.Duration) * time.Millisecond,
	}

	analysis := h.voice.AnalyzeSpeech(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}
