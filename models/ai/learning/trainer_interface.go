package learning

import (
	"fmt"
	"sync"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
)

// EvaluationModel represents an evaluation model with customizable coefficients
type EvaluationModel struct {
	Coeffs     evaluation.EvaluationCoefficients `json:"coefficients"`
	Wins       int                               `json:"wins"`
	Losses     int                               `json:"losses"`
	Draws      int                               `json:"draws"`
	Fitness    float64                           `json:"fitness"`
	Generation int                               `json:"generation"`
}

// TrainerInterface defines the common interface for all trainers
type TrainerInterface interface {
	InitializePopulation()
	StartTraining(generations int)
	TournamentTraining(generations int)
	LoadModel(filename string) (EvaluationModel, error)
	SaveModel(filename string, model EvaluationModel) error
	SaveGenerationStats(gen int) error
}

// PerformanceStats keeps track of performance metrics during training
type PerformanceStats struct {
	EvaluationTime      time.Duration
	CrossoverTime       time.Duration
	MutationTime        time.Duration
	TournamentTime      time.Duration
	TotalGenerationTime time.Duration
	Counts              map[string]int
	OpTimes             map[string]time.Duration
	OpCounts            map[string]int
	mutex               sync.Mutex // Add mutex for thread safety
}

// NewPerformanceStats creates a new performance stats tracker
func NewPerformanceStats() *PerformanceStats {
	return &PerformanceStats{
		Counts:   make(map[string]int),
		OpTimes:  make(map[string]time.Duration),
		OpCounts: make(map[string]int),
	}
}

// Reset clears all performance stats
func (ps *PerformanceStats) Reset() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.EvaluationTime = 0
	ps.CrossoverTime = 0
	ps.MutationTime = 0
	ps.TournamentTime = 0
	ps.TotalGenerationTime = 0
	ps.Counts = make(map[string]int)
	ps.OpTimes = make(map[string]time.Duration)
	ps.OpCounts = make(map[string]int)
}

// RecordOperation records the time taken for a specific operation
func (ps *PerformanceStats) RecordOperation(op string, duration time.Duration) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// Update specific metric if it matches a known operation
	switch op {
	case "evaluation":
		ps.EvaluationTime += duration
	case "crossover":
		ps.CrossoverTime += duration
	case "mutation":
		ps.MutationTime += duration
	case "tournament":
		ps.TournamentTime += duration
	case "generation":
		ps.TotalGenerationTime += duration
	}

	// Update general operation stats
	ps.OpTimes[op] += duration
	ps.OpCounts[op]++
}

// PrintSummary prints a summary of performance statistics
func (ps *PerformanceStats) PrintSummary() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	fmt.Println("\nPerformance Statistics:")
	fmt.Printf("Total generation time: %s\n", ps.TotalGenerationTime.Round(time.Millisecond))
	fmt.Printf("Evaluation time: %s (%.1f%%)\n",
		ps.EvaluationTime.Round(time.Millisecond),
		float64(ps.EvaluationTime)/float64(ps.TotalGenerationTime)*100)

	if ps.TournamentTime > 0 {
		fmt.Printf("Tournament time: %s (%.1f%%)\n",
			ps.TournamentTime.Round(time.Millisecond),
			float64(ps.TournamentTime)/float64(ps.TotalGenerationTime)*100)
	}

	// Print counts for specific operations
	if moveCount, ok := ps.Counts["moves_made"]; ok {
		fmt.Printf("Moves evaluated: %d\n", moveCount)
	}

	if gameCount, ok := ps.Counts["games_played"]; ok {
		fmt.Printf("Games played: %d\n", gameCount)
		if ps.OpTimes["game"] > 0 {
			avgGameTime := ps.OpTimes["game"] / time.Duration(gameCount)
			fmt.Printf("Average game time: %s\n", avgGameTime.Round(time.Millisecond))
		}
	}
}

// BaseTrainerInterface defines common operations for all trainers
type BaseTrainerInterface interface {
	InitializePopulation()
	StartTraining(int)
	TournamentTraining(int)
	LoadModel(string) (EvaluationModel, error)
}
