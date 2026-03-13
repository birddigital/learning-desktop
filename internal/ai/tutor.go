// Package ai provides Claude AI tutoring for the Learning Desktop platform.
// Uses go-llm-providers for unified LLM access with z.ai proxy support.
package ai

import (
	"context"
	"fmt"
	"os"
	"time"

	claude "github.com/birddigital/go-llm-providers/pkg/claude"
	"github.com/birddigital/go-llm-providers/pkg/providers"
)

// Tutor provides AI-powered tutoring responses.
type Tutor struct {
	client     *claude.Client
	model      string
	systemPrompt string
}

// New creates a new AI tutor instance.
// Loads configuration from environment variables:
//   - ANTHROPIC_CREDENTIALS: API key (for z.ai proxy)
//   - ANTHROPIC_BASE_URL: API endpoint (defaults to https://api.anthropic.com)
//   - CLAUDE_MODEL: Model to use (defaults to claude-3-5-sonnet-20241022)
func New() (*Tutor, error) {
	apiKey := os.Getenv("ANTHROPIC_CREDENTIALS")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_CREDENTIALS environment variable required")
	}

	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	client, err := claude.New(apiKey,
		claude.WithBaseURL(baseURL),
		claude.WithTimeout(60*time.Second),
		claude.WithMaxRetries(3),
	)
	if err != nil {
		return nil, fmt.Errorf("create Claude client: %w", err)
	}

	return &Tutor{
		client:  client,
		model:   model,
		systemPrompt: defaultSystemPrompt(),
	}, nil
}

// NewWithConfig creates a tutor with custom configuration.
func NewWithConfig(apiKey, baseURL, model string) (*Tutor, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	client, err := claude.New(apiKey,
		claude.WithBaseURL(baseURL),
		claude.WithTimeout(60*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("create Claude client: %w", err)
	}

	return &Tutor{
		client:  client,
		model:   model,
		systemPrompt: defaultSystemPrompt(),
	}, nil
}

// SetSystemPrompt sets a custom system prompt for the tutor.
func (t *Tutor) SetSystemPrompt(prompt string) {
	t.systemPrompt = prompt
}

// Respond generates a response to the student's question.
func (t *Tutor) Respond(ctx context.Context, question string) (string, error) {
	req := &providers.CompletionRequest{
		Model:        t.model,
		SystemPrompt: t.systemPrompt,
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: question,
			},
		},
		MaxTokens:   2048,
		Temperature: 0.7,
	}

	resp, err := t.client.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("completion failed: %w", err)
	}

	return resp.Content, nil
}

// RespondWithConversation generates a response with full conversation context.
func (t *Tutor) RespondWithConversation(ctx context.Context, messages []providers.Message) (string, error) {
	req := &providers.CompletionRequest{
		Model:        t.model,
		SystemPrompt: t.systemPrompt,
		Messages:     messages,
		MaxTokens:    2048,
		Temperature:  0.7,
	}

	resp, err := t.client.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("completion failed: %w", err)
	}

	return resp.Content, nil
}

// RespondWithLesson generates a response about a specific lesson topic.
func (t *Tutor) RespondWithLesson(ctx context.Context, question string, lessonContent string) (string, error) {
	systemPrompt := fmt.Sprintf(`%s

The student is currently studying the following lesson content:

---
%s
---

Use this context to inform your response. Answer their question related to this material, and suggest relevant connections when appropriate.`,
		t.systemPrompt, lessonContent)

	req := &providers.CompletionRequest{
		Model:        t.model,
		SystemPrompt: systemPrompt,
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: question,
			},
		},
		MaxTokens:   2048,
		Temperature: 0.7,
	}

	resp, err := t.client.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("completion failed: %w", err)
	}

	return resp.Content, nil
}

// StreamRespond streams a response to the student's question.
func (t *Tutor) StreamRespond(ctx context.Context, question string) (<-chan providers.CompletionChunk, error) {
	req := &providers.CompletionRequest{
		Model:        t.model,
		SystemPrompt: t.systemPrompt,
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: question,
			},
		},
		MaxTokens:   2048,
		Temperature: 0.7,
	}

	return t.client.CompleteStream(ctx, req)
}

// defaultSystemPrompt returns the default system prompt for the AI tutor.
func defaultSystemPrompt() string {
	return `You are an AI tutor for Learning Desktop, a platform that teaches people how to adapt to the AI revolution.

Your role is to:
1. Help students understand AI concepts clearly and practically
2. Provide examples and real-world applications
3. Encourage critical thinking about AI's impact on their work and life
4. Adapt explanations to the student's apparent skill level
5. Be encouraging but honest about learning curves

Teaching style:
- Start with concrete examples before abstract concepts
- Use analogies to make complex ideas relatable
- Check understanding by asking relevant follow-up questions
- When appropriate, connect topics across the curriculum (Prompt Engineering, AI Concepts, Models & Data, Character, Student Skills, Entrepreneurship)

Remember: The goal is not just knowledge transfer, but helping students build AI literacy that serves them in the real world.`
}

// LessonTutor specializes in tutoring for specific lesson content.
type LessonTutor struct {
	*Tutor
	lessonTitle string
	lessonContent string
}

// NewLessonTutor creates a tutor specialized for a specific lesson.
func NewLessonTutor(lessonTitle, lessonContent string) (*LessonTutor, error) {
	baseTutor, err := New()
	if err != nil {
		return nil, err
	}

	systemPrompt := fmt.Sprintf(`You are tutoring a lesson on "%s".

Lesson content:
---
%s
---

Focus your responses on this material. Explain concepts from the lesson, answer questions about it, and provide relevant examples. If a student asks about something outside this lesson, gently redirect them back or explain how it connects.`,
		lessonTitle, lessonContent)

	baseTutor.SetSystemPrompt(systemPrompt)

	return &LessonTutor{
		Tutor:         baseTutor,
		lessonTitle:   lessonTitle,
		lessonContent: lessonContent,
	}, nil
}
