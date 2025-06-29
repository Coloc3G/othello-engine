package stats

import (
	"sync"
	"time"
)

// PerformanceStats tracks performance statistics for training
type PerformanceStats struct {
	mu         sync.Mutex
	Operations map[string]*struct {
		Count int
		Time  time.Duration
		Cache map[string]int64
	}
}

// NewPerformanceStats creates a new performance stats tracker
func NewPerformanceStats() *PerformanceStats {
	return &PerformanceStats{
		Operations: make(map[string]*struct {
			Count int
			Time  time.Duration
			Cache map[string]int64
		}),
	}
}

// Reset clears all statistics
func (s *PerformanceStats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Operations = make(map[string]*struct {
		Count int
		Time  time.Duration
		Cache map[string]int64
	})
}

// RecordOperation records the time taken for a specific operation
func (s *PerformanceStats) RecordOperation(name string, duration time.Duration, hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.Operations[name]; !exists {
		s.Operations[name] = &struct {
			Count int
			Time  time.Duration
			Cache map[string]int64
		}{
			Count: 0,
			Time:  0,
			Cache: make(map[string]int64),
		}
	}
	s.Operations[name].Count++
	s.Operations[name].Time += duration
	if _, ok := s.Operations[name].Cache[hash]; !ok {
		s.Operations[name].Cache[hash] = 1
	} else {
		s.Operations[name].Cache[hash]++
	}
}
