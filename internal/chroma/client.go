// Package chroma provides ChromaDB vector database integration for RAG.
package chroma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client provides access to ChromaDB for vector storage and retrieval.
type Client struct {
	baseURL    string
	httpClient *http.Client
	collection string // Default collection ID
}

// Config holds ChromaDB client configuration.
type Config struct {
	BaseURL    string        // Default: http://localhost:8000
	Collection string        // Collection ID/name
	Timeout    time.Duration // Request timeout
}

// New creates a new ChromaDB client.
// Loads configuration from environment variables:
//   - CHROMA_BASE_URL: ChromaDB server URL (defaults to http://localhost:8000)
//   - CHROMA_COLLECTION: Default collection ID
func New() (*Client, error) {
	baseURL := os.Getenv("CHROMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	collection := os.Getenv("CHROMA_COLLECTION")
	if collection == "" {
		collection = "learning_desktop"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		collection: collection,
	}, nil
}

// NewWithConfig creates a client with custom configuration.
func NewWithConfig(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8000"
	}
	if cfg.Collection == "" {
		cfg.Collection = "learning_desktop"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL:    cfg.BaseURL,
		httpClient: &http.Client{Timeout: cfg.Timeout},
		collection: cfg.Collection,
	}, nil
}

// EnsureCollection creates the collection if it doesn't exist.
func (c *Client) EnsureCollection(ctx context.Context, metadata map[string]string) error {
	req := collectionCreateRequest{
		Name:     c.collection,
		Metadata: metadata,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/collections", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Collection might already exist - that's okay
	if resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusOK {
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("create collection failed: status %d: %s", resp.StatusCode, string(respBody))
}

// AddDocuments adds documents to the collection.
func (c *Client) AddDocuments(ctx context.Context, docs []Document) error {
	if len(docs) == 0 {
		return nil
	}

	// Batch documents into groups of 100 (ChromaDB recommended batch size)
	const batchSize = 100
	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}
		batch := docs[i:end]

		if err := c.addDocumentsBatch(ctx, batch); err != nil {
			return fmt.Errorf("add batch %d-%d: %w", i, end, err)
		}
	}

	return nil
}

// addDocumentsBatch adds a batch of documents.
func (c *Client) addDocumentsBatch(ctx context.Context, docs []Document) error {
	// Build request
	req := addRequest{
		CollectionID: c.collection,
		Documents:    make([]string, len(docs)),
		Metadatas:    make([]map[string]string, len(docs)),
		IDs:          make([]string, len(docs)),
	}

	for i, doc := range docs {
		req.Documents[i] = doc.Content
		req.Metadatas[i] = doc.Metadata
		req.IDs[i] = doc.ID

		if doc.Embedding != nil {
			// In a real implementation, you'd generate embeddings
			// For now, we'll let ChromaDB handle embedding generation
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, c.collection)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add documents failed: status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Query performs semantic search on the collection.
func (c *Client) Query(ctx context.Context, queryText string, nResults int, where map[string]string) ([]QueryResult, error) {
	req := queryRequest{
		QueryTexts:  []string{queryText},
		NResults:    nResults,
		Where:       where,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/collections/%s/query", c.baseURL, c.collection)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed: status %d: %s", resp.StatusCode, string(respBody))
	}

	var queryResp queryResponse
	if err := json.Unmarshal(respBody, &queryResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Convert to results
	results := make([]QueryResult, 0)
	if len(queryResp.Documents) > 0 {
		for i := 0; i < len(queryResp.Documents[0]); i++ {
			result := QueryResult{
				Content:   queryResp.Documents[0][i],
				Metadata:  queryResp.Metadatas[0][i],
				ID:        queryResp.IDs[0][i],
			}
			if len(queryResp.Distances) > 0 && i < len(queryResp.Distances[0]) {
				result.Distance = queryResp.Distances[0][i]
				result.Similarity = 1 - result.Distance
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// GetHealth checks if ChromaDB is accessible.
func (c *Client) GetHealth(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/heartbeat", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}

	return nil
}

// ============================================================================
// Types
// ============================================================================

// Document represents a document to be stored in ChromaDB.
type Document struct {
	ID        string                 // Unique identifier
	Content   string                 // Text content
	Metadata  map[string]string      // Metadata (topic, difficulty, skill tree, etc.)
	Embedding []float32              // Optional pre-computed embedding
}

// QueryResult represents a search result from ChromaDB.
type QueryResult struct {
	ID         string            // Document ID
	Content    string            // Document content
	Metadata   map[string]string // Document metadata
	Distance   float64           // Distance score (lower is better)
	Similarity float64           // Similarity score (higher is better)
}

// ============================================================================
// API Request/Response Types
// ============================================================================

type collectionCreateRequest struct {
	Name     string                 `json:"name"`
	Metadata map[string]string      `json:"metadata,omitempty"`
}

type addRequest struct {
	CollectionID string            `json:"-"`
	Documents    []string         `json:"documents"`
	Metadatas    []map[string]string `json:"metadatas,omitempty"`
	IDs          []string         `json:"ids"`
	Embeddings   [][]float32      `json:"embeddings,omitempty"`
}

type queryRequest struct {
	QueryTexts   []string          `json:"query_texts"`
	NResults     int               `json:"n_results"`
	Where        map[string]string `json:"where,omitempty"`
	WhereDocument map[string]string `json:"where_document,omitempty"`
}

type queryResponse struct {
	Distances [][]float64        `json:"distances,omitempty"`
	Documents [][]string        `json:"documents,omitempty"`
	Metadatas [][]map[string]string `json:"metadatas,omitempty"`
	IDs       [][]string        `json:"ids,omitempty"`
	Embeddings [][]float32      `json:"embeddings,omitempty"`
}

// ============================================================================
// Content Importer
// ============================================================================

// Importer imports research content into ChromaDB.
type Importer struct {
	client *Client
}

// NewImporter creates a new content importer.
func NewImporter(client *Client) *Importer {
	return &Importer{client: client}
}

// ImportFromDirectory imports all topics from a research directory.
func (i *Importer) ImportFromDirectory(ctx context.Context, researchDir string) error {
	// Walk the research directory and import each topics.json file
	// This is a placeholder - implement based on actual research content structure

	// Ensure collection exists
	metadata := map[string]string{
		"description": "Learning Desktop research content",
		"created_at":  time.Now().Format(time.RFC3339),
	}
	if err := i.client.EnsureCollection(ctx, metadata); err != nil {
		return fmt.Errorf("ensure collection: %w", err)
	}

	// TODO: Walk directory and import files
	// For now, this is a placeholder that would:
	// 1. Read all topics.json files from ~/.learning-desktop/research/content/
	// 2. Create Document structs from each topic
	// 3. Call AddDocuments to vectorize and store

	return nil
}

// ImportTopics imports a list of topics directly.
func (i *Importer) ImportTopics(ctx context.Context, topics []Topic) error {
	// Ensure collection exists
	metadata := map[string]string{
		"description": "Learning Desktop research content",
		"source":      "learning-desktop",
		"created_at":  time.Now().Format(time.RFC3339),
	}
	if err := i.client.EnsureCollection(ctx, metadata); err != nil {
		return fmt.Errorf("ensure collection: %w", err)
	}

	// Convert topics to documents
	docs := make([]Document, 0, len(topics))
	for _, topic := range topics {
		doc := Document{
			ID:      topic.ID,
			Content: topic.Content,
			Metadata: map[string]string{
				"title":      topic.Title,
				"tree":       topic.SkillTree,
				"node":       topic.Node,
				"difficulty": topic.Difficulty,
			},
		}
		docs = append(docs, doc)
	}

	return i.client.AddDocuments(ctx, docs)
}

// Topic represents a research topic.
type Topic struct {
	ID         string
	Title      string
	Content    string
	SkillTree  string
	Node       string
	Difficulty string
}
