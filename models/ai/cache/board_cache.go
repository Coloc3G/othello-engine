package cache

import (
	"fmt"
	"sync"

	"github.com/Coloc3G/othello-engine/models/game"
)

// BoardHashKey represents a hash of a board state
type BoardHashKey string

// CacheSource represents the source of a cache entry
type CacheSource int

const (
	SourceUnknown CacheSource = iota
	SourceCPU
	SourceGPU
)

// EvalCacheEntry represents a cached evaluation result
type EvalCacheEntry struct {
	Score  int
	Depth  int
	Move   game.Position
	Source CacheSource // Tracks whether this entry came from CPU or GPU
}

// EvaluationCache caches board evaluations to avoid redundant calculations
type EvaluationCache struct {
	cache      map[BoardHashKey]EvalCacheEntry
	mutex      sync.RWMutex
	hits       int64
	cpuHits    int64
	gpuHits    int64
	misses     int64
	capacity   int
	gpuBuffer  []game.Board   // Buffer for GPU batch processing
	gpuPlayers []game.Piece   // Corresponding players for GPU batch
	gpuKeys    []BoardHashKey // Keys for GPU batch results
	batchSize  int            // Size of GPU processing batch
	batchMutex sync.Mutex
}

// Global evaluation cache
var (
	globalCache     *EvaluationCache
	globalCacheOnce sync.Once
)

// GetGlobalCache returns the singleton global evaluation cache
func GetGlobalCache() *EvaluationCache {
	globalCacheOnce.Do(func() {
		globalCache = NewEvaluationCache(500000) // 500k entries default
	})
	return globalCache
}

// NewEvaluationCache creates a new evaluation cache with the specified capacity
func NewEvaluationCache(capacity int) *EvaluationCache {
	return &EvaluationCache{
		cache:      make(map[BoardHashKey]EvalCacheEntry, capacity),
		capacity:   capacity,
		gpuBuffer:  make([]game.Board, 0, 1024),
		gpuPlayers: make([]game.Piece, 0, 1024),
		gpuKeys:    make([]BoardHashKey, 0, 1024),
		batchSize:  1024, // Process 1024 positions at once on GPU
	}
}

// Store adds a board evaluation to the cache
func (c *EvaluationCache) Store(board game.Board, player game.Piece, depth int, score int, bestMove game.Position, source CacheSource) {
	key := GenerateBoardHashKey(board, player)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Only add if we have capacity
	if len(c.cache) >= c.capacity {
		// Simple strategy: just don't add new entries when full
		return
	}

	c.cache[key] = EvalCacheEntry{
		Score:  score,
		Depth:  depth,
		Move:   bestMove,
		Source: source,
	}
}

// Lookup tries to retrieve a cached evaluation
// Returns the score, best move, source, and whether the result was found
func (c *EvaluationCache) Lookup(board game.Board, player game.Piece, depth int) (int, game.Position, CacheSource, bool) {
	key := GenerateBoardHashKey(board, player)

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, found := c.cache[key]
	if found && entry.Depth >= depth {
		c.hits++

		// Track hit source
		switch entry.Source {
		case SourceCPU:
			c.cpuHits++
		case SourceGPU:
			c.gpuHits++
		}

		return entry.Score, entry.Move, entry.Source, true
	}

	c.misses++
	return 0, game.Position{Row: -1, Col: -1}, SourceUnknown, false
}

// GenerateBoardHashKey creates a unique string hash for a board state
func GenerateBoardHashKey(board game.Board, player game.Piece) BoardHashKey {
	// Use Zobrist hashing or similar technique for more efficiency in a real implementation
	// For now, we'll use a simple string representation
	var hash string
	hash = fmt.Sprintf("%v-%d", board, player)
	return BoardHashKey(hash)
}

// GetStats returns cache statistics
func (c *EvaluationCache) GetStats() (size int, hits int64, cpuHits int64, gpuHits int64, misses int64) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	size = len(c.cache)
	return size, c.hits, c.cpuHits, c.gpuHits, c.misses
}

// ClearCache resets the cache
func (c *EvaluationCache) ClearCache() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[BoardHashKey]EvalCacheEntry, c.capacity)
	c.hits = 0
	c.cpuHits = 0
	c.gpuHits = 0
	c.misses = 0
}

// AddToBatch adds a board to the GPU evaluation batch
// Returns true if the batch is full and should be processed
func (c *EvaluationCache) AddToBatch(board game.Board, player game.Piece) (bool, BoardHashKey) {
	key := GenerateBoardHashKey(board, player)

	c.batchMutex.Lock()
	defer c.batchMutex.Unlock()

	// Add to batch
	c.gpuBuffer = append(c.gpuBuffer, board)
	c.gpuPlayers = append(c.gpuPlayers, player)
	c.gpuKeys = append(c.gpuKeys, key)

	// Check if batch is full
	return len(c.gpuBuffer) >= c.batchSize, key
}

// GetBatch retrieves the current batch for GPU processing
// Also clears the batch
func (c *EvaluationCache) GetBatch() ([]game.Board, []game.Piece, []BoardHashKey) {
	c.batchMutex.Lock()
	defer c.batchMutex.Unlock()

	// Create copies of current batch
	boards := make([]game.Board, len(c.gpuBuffer))
	players := make([]game.Piece, len(c.gpuPlayers))
	keys := make([]BoardHashKey, len(c.gpuKeys))

	copy(boards, c.gpuBuffer)
	copy(players, c.gpuPlayers)
	copy(keys, c.gpuKeys)

	// Clear batch
	c.gpuBuffer = c.gpuBuffer[:0]
	c.gpuPlayers = c.gpuPlayers[:0]
	c.gpuKeys = c.gpuKeys[:0]

	return boards, players, keys
}

// SetBatchSize sets the size of GPU processing batches
func (c *EvaluationCache) SetBatchSize(size int) {
	c.batchMutex.Lock()
	defer c.batchMutex.Unlock()

	if size < 1 {
		size = 1
	} else if size > 16384 {
		size = 16384 // Upper limit for safety
	}

	c.batchSize = size
}

// StoreBatchResults stores multiple evaluation results at once
func (c *EvaluationCache) StoreBatchResults(keys []BoardHashKey, scores []int, depth int) {
	if len(keys) != len(scores) {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, key := range keys {
		// Skip if cache is full
		if len(c.cache) >= c.capacity {
			return
		}

		// Use empty position since we're just caching scores
		c.cache[key] = EvalCacheEntry{
			Score:  scores[i],
			Depth:  depth,
			Move:   game.Position{Row: -1, Col: -1},
			Source: SourceGPU,
		}
	}
}
