package rag

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
)

type Chunk struct {
	ID        string    `json:"id"`
	FilePath  string    `json:"file_path"`
	Content   string    `json:"content"`
	Embedding []float32 `json:"embedding"`
}

type Store struct {
	Chunks []Chunk `json:"chunks"`
}

func GetStorePath(repoRoot string) string {
	return filepath.Join(repoRoot, ".git", "ai-git-embeddings.json")
}

func LoadStore(repoRoot string) (*Store, error) {
	path := GetStorePath(repoRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{Chunks: []Chunk{}}, nil
		}
		return nil, err
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *Store) Save(repoRoot string) error {
	path := GetStorePath(repoRoot)
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Store) AddChunk(chunk Chunk) {
	s.Chunks = append(s.Chunks, chunk)
}

type SearchResult struct {
	Chunk Chunk
	Score float32
}

func (s *Store) Search(queryEmbedding []float32, topK int) []SearchResult {
	var results []SearchResult

	for _, chunk := range s.Chunks {
		if len(chunk.Embedding) == 0 {
			continue
		}
		score := CosineSimilarity(queryEmbedding, chunk.Embedding)
		results = append(results, SearchResult{Chunk: chunk, Score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score // Descending order
	})

	if len(results) > topK {
		return results[:topK]
	}
	return results
}

func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}
