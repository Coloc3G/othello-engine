package learning

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/ai/stats"
)

// BaseTrainer implements common trainer functionality
type BaseTrainer struct {
	Models         []EvaluationModel
	BaseModel      evaluation.EvaluationCoefficients
	BestModel      EvaluationModel
	Generation     int
	PopulationSize int
	MutationRate   float64
	NumGames       int
	MaxDepth       int
	Stats          *stats.PerformanceStats
}

// NewBaseTrainer creates a new base trainer
func NewBaseTrainer(popSize, numGames, depth int, baseModelCoeffs evaluation.EvaluationCoefficients) *BaseTrainer {
	return &BaseTrainer{
		Models:         make([]EvaluationModel, 0),
		BaseModel:      baseModelCoeffs,
		PopulationSize: popSize,
		MutationRate:   0.3,
		NumGames:       numGames,
		MaxDepth:       depth,
		Generation:     1,
		Stats:          stats.NewPerformanceStats(),
	}
}

// createNextGeneration creates a new generation through selection, crossover and mutation
func (t *BaseTrainer) createNextGeneration() {
	startTime := time.Now()

	newModels := make([]EvaluationModel, t.PopulationSize)

	// Increase elitism to preserve more good models
	eliteCount := t.PopulationSize / 4
	copy(newModels[:eliteCount], t.Models[:eliteCount])

	// Fill the rest with crossover and mutation
	for i := eliteCount; i < t.PopulationSize; i++ {
		crossoverStart := time.Now()

		// Use larger tournament size to focus on better models
		parent1 := t.tournamentSelect(5)
		parent2 := t.tournamentSelect(5)

		// Crossover
		child := t.crossover(parent1, parent2)

		crossoverTime := time.Since(crossoverStart)
		t.Stats.RecordOperation("crossover", crossoverTime)

		mutationStart := time.Now()

		// Mutation
		child = t.mutateModel(child)
		child.Generation = t.Generation + 1

		mutationTime := time.Since(mutationStart)
		t.Stats.RecordOperation("mutation", mutationTime)

		newModels[i] = child
	}

	t.Models = newModels

	totalTime := time.Since(startTime)
	t.Stats.RecordOperation("generation_creation", totalTime)
}

// tournamentSelect selects a model using tournament selection
func (t *BaseTrainer) tournamentSelect(tournamentSize int) EvaluationModel {
	best := t.Models[0]
	bestFitness := t.Models[0].Fitness

	for i := 1; i < tournamentSize; i++ {
		idx := i % len(t.Models) // Ensure we don't go out of bounds
		if t.Models[idx].Fitness > bestFitness {
			best = t.Models[idx]
			bestFitness = t.Models[idx].Fitness
		}
	}

	return best
}

// crossover combines two models to create a child model
func (t *BaseTrainer) crossover(parent1, parent2 EvaluationModel) EvaluationModel {
	child := EvaluationModel{
		Coeffs: evaluation.EvaluationCoefficients{
			MaterialCoeffs:  make([]int, 3),
			MobilityCoeffs:  make([]int, 3),
			CornersCoeffs:   make([]int, 3),
			ParityCoeffs:    make([]int, 3),
			StabilityCoeffs: make([]int, 3),
			FrontierCoeffs:  make([]int, 3),
		},
	}

	// Create crossover patterns that determine which parent each coefficient comes from
	materialPattern := []bool{true, false, true}
	mobilityPattern := []bool{false, true, false}
	cornersPattern := []bool{true, false, true}
	parityPattern := []bool{false, true, false}
	stabilityPattern := []bool{true, true, false}
	frontierPattern := []bool{false, false, true}

	// Apply crossover patterns
	child.Coeffs.MaterialCoeffs = crossoverCoefficients(
		parent1.Coeffs.MaterialCoeffs, parent2.Coeffs.MaterialCoeffs, materialPattern)
	child.Coeffs.MobilityCoeffs = crossoverCoefficients(
		parent1.Coeffs.MobilityCoeffs, parent2.Coeffs.MobilityCoeffs, mobilityPattern)
	child.Coeffs.CornersCoeffs = crossoverCoefficients(
		parent1.Coeffs.CornersCoeffs, parent2.Coeffs.CornersCoeffs, cornersPattern)
	child.Coeffs.ParityCoeffs = crossoverCoefficients(
		parent1.Coeffs.ParityCoeffs, parent2.Coeffs.ParityCoeffs, parityPattern)
	child.Coeffs.StabilityCoeffs = crossoverCoefficients(
		parent1.Coeffs.StabilityCoeffs, parent2.Coeffs.StabilityCoeffs, stabilityPattern)
	child.Coeffs.FrontierCoeffs = crossoverCoefficients(
		parent1.Coeffs.FrontierCoeffs, parent2.Coeffs.FrontierCoeffs, frontierPattern)

	return child
}

// mutateModel applies random mutations to a model
func (t *BaseTrainer) mutateModel(model EvaluationModel) EvaluationModel {
	mutated := model

	// Use the utils package for mutation
	mutated.Coeffs = MutateCoefficients(model.Coeffs)

	// Give the mutated model a name for tracking
	if mutated.Coeffs.Name == "" {
		mutated.Coeffs.Name = fmt.Sprintf("Model_Gen%d", t.Generation)
	}

	return mutated
}

// sortModelsByFitness sorts models by fitness in descending order
func (t *BaseTrainer) sortModelsByFitness() {
	sort.Slice(t.Models, func(i, j int) bool {
		return t.Models[i].Fitness > t.Models[j].Fitness
	})
}

// SaveModel saves a model to a JSON file
func (t *BaseTrainer) SaveModel(filename string, model EvaluationModel) error {
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadModel loads a model from a JSON file
func (t *BaseTrainer) LoadModel(filename string) (EvaluationModel, error) {
	var model EvaluationModel
	data, err := os.ReadFile(filename)
	if err != nil {
		return model, err
	}
	err = json.Unmarshal(data, &model)
	return model, err
}

// SaveModelToFile is a generic helper method to save structs to JSON files
func (t *BaseTrainer) SaveModelToFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

// SaveGenerationStats saves statistics about the current generation
func (t *BaseTrainer) SaveGenerationStats(gen int) error {
	stats := struct {
		Generation     int               `json:"generation"`
		BestFitness    float64           `json:"best_fitness"`
		AvgFitness     float64           `json:"avg_fitness"`
		BestModel      EvaluationModel   `json:"best_model"`
		AllModels      []EvaluationModel `json:"all_models"`
		Timestamp      string            `json:"timestamp"`
		PerformanceLog struct {
			EvaluationTimeMs int `json:"evaluation_time_ms"`
			TournamentTimeMs int `json:"tournament_time_ms"`
			CrossoverTimeMs  int `json:"crossover_time_ms"`
			MutationTimeMs   int `json:"mutation_time_ms"`
			TotalTimeMs      int `json:"total_time_ms"`
		} `json:"performance"`
		Stats       map[string]int    `json:"stats_counts"`
		TimingStats map[string]string `json:"timing_stats"`
	}{
		Generation:  gen,
		BestFitness: t.Models[0].Fitness,
		BestModel:   t.Models[0],
		AllModels:   t.Models,
		Timestamp:   time.Now().Format(time.RFC3339),
		Stats:       t.Stats.Counts,
	}

	// Calculate average fitness
	var sum float64
	for _, model := range t.Models {
		sum += model.Fitness
	}
	stats.AvgFitness = sum / float64(len(t.Models))

	// Add performance statistics
	stats.PerformanceLog.EvaluationTimeMs = int(t.Stats.EvaluationTime.Milliseconds())
	stats.PerformanceLog.TournamentTimeMs = int(t.Stats.TournamentTime.Milliseconds())
	stats.PerformanceLog.CrossoverTimeMs = int(t.Stats.CrossoverTime.Milliseconds())
	stats.PerformanceLog.MutationTimeMs = int(t.Stats.MutationTime.Milliseconds())
	stats.PerformanceLog.TotalTimeMs = int(t.Stats.TotalGenerationTime.Milliseconds())

	// Add detailed timing stats
	stats.TimingStats = make(map[string]string)
	for opName, duration := range t.Stats.OpTimes {
		stats.TimingStats[opName] = duration.String()
	}

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("stats_gen_%d.json", gen)
	return os.WriteFile(filename, data, 0644)
}

// calculateAvgFitness calculates the average fitness of the population
func (t *BaseTrainer) calculateAvgFitness() float64 {
	sum := 0.0
	for _, model := range t.Models {
		sum += model.Fitness
	}
	return sum / float64(len(t.Models))
}
