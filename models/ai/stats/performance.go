package stats

import (
	"sync"
	"time"
)

// PerformanceStats tracks performance statistics for training
type PerformanceStats struct {
	// Operation timing
	EvaluationTime      time.Duration
	TournamentTime      time.Duration
	CrossoverTime       time.Duration
	MutationTime        time.Duration
	TotalGenerationTime time.Duration

	// GPU/CPU specific timings
	GPUEvaluationTime time.Duration
	CPUEvaluationTime time.Duration
	GPUFallbackTime   time.Duration // Time spent in CPU after GPU failed
	CacheHitTime      time.Duration

	// Operation counts
	Counts map[string]int

	// GPU/CPU specific counts
	GPUEvaluations   int
	CPUEvaluations   int
	GPUFallbacks     int
	GPUSuccesses     int
	CacheHits        int
	CPUCacheHits     int
	GPUCacheHits     int
	CacheMisses      int
	TotalEvaluations int

	// Thread safety
	mu sync.Mutex

	// Operation timings by name
	OpTimes map[string]time.Duration
}

// NewPerformanceStats creates a new performance stats tracker
func NewPerformanceStats() *PerformanceStats {
	return &PerformanceStats{
		Counts:  make(map[string]int),
		OpTimes: make(map[string]time.Duration),
	}
}

// Reset clears all statistics
func (s *PerformanceStats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.EvaluationTime = 0
	s.TournamentTime = 0
	s.CrossoverTime = 0
	s.MutationTime = 0
	s.TotalGenerationTime = 0

	s.GPUEvaluationTime = 0
	s.CPUEvaluationTime = 0
	s.GPUFallbackTime = 0
	s.CacheHitTime = 0

	s.Counts = make(map[string]int)

	s.GPUEvaluations = 0
	s.CPUEvaluations = 0
	s.GPUFallbacks = 0
	s.GPUSuccesses = 0
	s.CacheHits = 0
	s.CPUCacheHits = 0
	s.GPUCacheHits = 0
	s.CacheMisses = 0
	s.TotalEvaluations = 0

	s.OpTimes = make(map[string]time.Duration)
}

// RecordOperation records the time taken for a specific operation
func (s *PerformanceStats) RecordOperation(name string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store in the general operation times map
	s.OpTimes[name] = s.OpTimes[name] + duration

	// Also update specific fields based on the operation name
	switch name {
	case "evaluation":
		s.EvaluationTime += duration
	case "tournament":
		s.TournamentTime += duration
	case "crossover":
		s.CrossoverTime += duration
	case "mutation":
		s.MutationTime += duration
	case "generation":
		s.TotalGenerationTime += duration
	case "gpu_eval":
		s.GPUEvaluationTime += duration
		s.GPUEvaluations++
		s.TotalEvaluations++
	case "cpu_eval":
		s.CPUEvaluationTime += duration
		s.CPUEvaluations++
		s.TotalEvaluations++
	case "gpu_fallback":
		s.GPUFallbackTime += duration
		s.GPUFallbacks++
	case "cache_hit":
		s.CacheHitTime += duration
		s.CacheHits++
	case "cpu_cache_hit":
		s.CPUCacheHits++
	case "gpu_cache_hit":
		s.GPUCacheHits++
	case "cache_miss":
		s.CacheMisses++
	}
}

// IncrementCount increments a named counter
func (s *PerformanceStats) IncrementCount(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Counts[name]++
}
