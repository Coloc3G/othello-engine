package learning

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
)

// GPUTrainer enhances the Trainer with GPU capabilities
type GPUTrainer struct {
	*BaseTrainer
	UseGPU bool
}

// NewGPUTrainer creates a new trainer with GPU support
func NewGPUTrainer(popSize int) *GPUTrainer {
	// Initialize CUDA and check availability
	gpuAvailable := evaluation.InitCUDA()

	return &GPUTrainer{
		BaseTrainer: NewBaseTrainer(popSize),
		UseGPU:      gpuAvailable,
	}
}

// InitializePopulation creates initial random population of models
func (t *GPUTrainer) InitializePopulation() {
	t.Models = make([]EvaluationModel, t.PopulationSize)

	// Initialize with a reasonable default model - use V2 coefficients as a starting point
	// as these have been tuned and should be more reasonable
	defaultModel := EvaluationModel{
		Coeffs:     evaluation.V1Coeff,
		Generation: 1,
	}

	// Set the first model to the default
	t.Models[0] = defaultModel
	t.BestModel = defaultModel

	// Create variations of the default model - using more controlled diversity
	for i := 1; i < t.PopulationSize; i++ {
		t.Models[i] = createDiverseModel(defaultModel, i)
		t.Models[i].Generation = 1
		t.Models[i].Coeffs.Name = fmt.Sprintf("Initial Model %d", i)
	}
}

// createDiverseModel creates a different but not wildly different model for initial population
func createDiverseModel(baseModel EvaluationModel, index int) EvaluationModel {
	newModel := baseModel

	// Apply random scaling factors with more moderate ranges
	materialFactor := 0.8 + rand.Float64()*0.4 // 0.8x to 1.2x
	mobilityFactor := 0.8 + rand.Float64()*0.4
	cornersFactor := 0.8 + rand.Float64()*0.4
	parityFactor := 0.8 + rand.Float64()*0.4
	stabilityFactor := 0.8 + rand.Float64()*0.4
	frontierFactor := 0.8 + rand.Float64()*0.4

	// Apply factors to all coefficients with bounds checking
	for i := 0; i < 3; i++ {
		// Apply the scaling factors with sensible minimum values
		newModel.Coeffs.MaterialCoeffs[i] = max(1, int(float64(baseModel.Coeffs.MaterialCoeffs[i])*materialFactor))
		newModel.Coeffs.MobilityCoeffs[i] = max(1, int(float64(baseModel.Coeffs.MobilityCoeffs[i])*mobilityFactor))
		newModel.Coeffs.CornersCoeffs[i] = max(1, int(float64(baseModel.Coeffs.CornersCoeffs[i])*cornersFactor))
		newModel.Coeffs.ParityCoeffs[i] = max(1, int(float64(baseModel.Coeffs.ParityCoeffs[i])*parityFactor))
		newModel.Coeffs.StabilityCoeffs[i] = max(1, int(float64(baseModel.Coeffs.StabilityCoeffs[i])*stabilityFactor))
		newModel.Coeffs.FrontierCoeffs[i] = max(1, int(float64(baseModel.Coeffs.FrontierCoeffs[i])*frontierFactor))

		// Apply maximum caps to avoid extreme values
		newModel.Coeffs.MaterialCoeffs[i] = min(newModel.Coeffs.MaterialCoeffs[i], 1000)
		newModel.Coeffs.MobilityCoeffs[i] = min(newModel.Coeffs.MobilityCoeffs[i], 500)
		newModel.Coeffs.CornersCoeffs[i] = min(newModel.Coeffs.CornersCoeffs[i], 2000)
		newModel.Coeffs.ParityCoeffs[i] = min(newModel.Coeffs.ParityCoeffs[i], 1000)
		newModel.Coeffs.StabilityCoeffs[i] = min(newModel.Coeffs.StabilityCoeffs[i], 300)
		newModel.Coeffs.FrontierCoeffs[i] = min(newModel.Coeffs.FrontierCoeffs[i], 200)
	}

	return newModel
}

// StartTraining begins the genetic algorithm training process
func (t *GPUTrainer) StartTraining(generations int) {
	if len(t.Models) == 0 {
		t.InitializePopulation()
	}

	t.Stats.Reset()

	for gen := 1; gen <= generations; gen++ {
		genStartTime := time.Now()

		t.Generation = gen
		fmt.Printf("\nGeneration %d/%d using %s\n", gen, generations,
			map[bool]string{true: "GPU acceleration", false: "CPU only"}[t.UseGPU])

		// Evaluate all models
		evalStartTime := time.Now()
		t.evaluatePopulation()
		evalTime := time.Since(evalStartTime)
		t.Stats.RecordOperation("evaluation", evalTime)

		// Sort models by fitness
		t.sortModelsByFitness()

		// Update best model
		if t.Models[0].Fitness > t.BestModel.Fitness {
			t.BestModel = t.Models[0]
			t.SaveModel("best_model.json", t.BestModel)
			fmt.Printf("New best model: fitness %.2f, win rate %.2f%%\n",
				t.BestModel.Fitness,
				float64(t.BestModel.Wins)/float64(t.BestModel.Wins+t.BestModel.Losses+t.BestModel.Draws)*100)
		}

		// Display current best fitness
		fmt.Printf("Best fitness: %.2f, Avg fitness: %.2f\n", t.Models[0].Fitness, t.calculateAvgFitness())

		// Save generation statistics
		t.SaveGenerationStats(gen)

		// NEW: If after 3 or more generations no model has any wins, reinforce population aggressively
		if gen >= 3 && calculateAverageWins(t.Models) == 0 {
			fmt.Println("No wins detected over GPU training generations. Reinforcing population!")
			t.reinforcePopulation()
		}

		// Create next generation if not last generation
		if gen < generations {
			t.createNextGeneration()
		}

		genTime := time.Since(genStartTime)
		t.Stats.RecordOperation("generation", genTime)
	}

	fmt.Println("\nTraining completed!")
}

// reinforcePopulation applies a strong mutation to all models to force exploration.
func (t *GPUTrainer) reinforcePopulation() {
	for i := range t.Models {
		t.Models[i].Coeffs.MaterialCoeffs = heavyMutate(t.Models[i].Coeffs.MaterialCoeffs)
		t.Models[i].Coeffs.MobilityCoeffs = heavyMutate(t.Models[i].Coeffs.MobilityCoeffs)
		t.Models[i].Coeffs.CornersCoeffs = heavyMutate(t.Models[i].Coeffs.CornersCoeffs)
		t.Models[i].Coeffs.ParityCoeffs = heavyMutate(t.Models[i].Coeffs.ParityCoeffs)
		t.Models[i].Coeffs.StabilityCoeffs = heavyMutate(t.Models[i].Coeffs.StabilityCoeffs)
		t.Models[i].Coeffs.FrontierCoeffs = heavyMutate(t.Models[i].Coeffs.FrontierCoeffs)
	}
}

// calculateAvgFitness calculates the average fitness of the population
func (t *GPUTrainer) calculateAvgFitness() float64 {
	sum := 0.0
	for _, model := range t.Models {
		sum += model.Fitness
	}
	return sum / float64(len(t.Models))
}

// evaluatePopulation evaluates all models by playing games
func (t *GPUTrainer) evaluatePopulation() {
	// Get models as pointer slice for parallel evaluation
	modelPtrs := make([]*EvaluationModel, len(t.Models))
	for i := range t.Models {
		modelPtrs[i] = &t.Models[i]
	}

	// Define evaluation function creator based on GPU/CPU
	createEvalFunc := func(model EvaluationModel) evaluation.Evaluation {
		if t.UseGPU {
			// Configure GPU evaluation with larger batch size for better GPU utilization
			gpuEval := evaluation.NewGPUMixedEvaluation(model.Coeffs)
			gpuEval.SetBatchSize(1024) // Increase batch size for better GPU utilization

			// Ensure coefficients are set in CUDA
			evaluation.SetCUDACoefficients(model.Coeffs)

			return gpuEval
		}
		return evaluation.NewMixedEvaluationWithCoefficients(model.Coeffs)
	}

	// Evaluate all models in parallel using shared utility
	evaluateModelsInParallel(modelPtrs, createEvalFunc, t.MaxDepth, t.NumGames, t.Stats)
}

// createNextGeneration creates a new generation through selection, crossover and mutation
func (t *GPUTrainer) createNextGeneration() {
	// Use the base implementation for simplicity and consistency
	t.BaseTrainer.createNextGeneration()
}

// SaveGenerationStats overrides the base trainer's method to include GPU information
func (t *GPUTrainer) SaveGenerationStats(gen int) error {
	stats := struct {
		Generation     int               `json:"generation"`
		BestFitness    float64           `json:"best_fitness"`
		AvgFitness     float64           `json:"avg_fitness"`
		BestModel      EvaluationModel   `json:"best_model"`
		AllModels      []EvaluationModel `json:"all_models"`
		Timestamp      string            `json:"timestamp"`
		UsingGPU       bool              `json:"using_gpu"`
		GpuMemoryFree  uint64            `json:"gpu_memory_free"`
		GpuMemoryTotal uint64            `json:"gpu_memory_total"`
		PerformanceLog struct {
			EvaluationTimeMs int `json:"evaluation_time_ms"`
			TournamentTimeMs int `json:"tournament_time_ms"`
			CrossoverTimeMs  int `json:"crossover_time_ms"`
			MutationTimeMs   int `json:"mutation_time_ms"`
			TotalTimeMs      int `json:"total_time_ms"`
		} `json:"performance"`
		Stats       map[string]int    `json:"stats_counts"`
		TimingStats map[string]string `json:"timing_stats"`
		Wins        []int             `json:"wins"`
		Losses      []int             `json:"losses"`
		Draws       []int             `json:"draws"`
		Fitness     []float64         `json:"fitness"`
	}{
		Generation:  gen,
		BestFitness: t.Models[0].Fitness,
		BestModel:   t.Models[0],
		AllModels:   t.Models,
		Timestamp:   time.Now().Format(time.RFC3339),
		UsingGPU:    t.UseGPU,
		Stats:       t.Stats.Counts,
	}

	// Calculate average fitness
	var sum float64
	stats.Wins = make([]int, len(t.Models))
	stats.Losses = make([]int, len(t.Models))
	stats.Draws = make([]int, len(t.Models))
	stats.Fitness = make([]float64, len(t.Models))

	for i, model := range t.Models {
		sum += model.Fitness
		stats.Wins[i] = model.Wins
		stats.Losses[i] = model.Losses
		stats.Draws[i] = model.Draws
		stats.Fitness[i] = model.Fitness
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

	// Save to file using the base trainer's helper methods
	return t.BaseTrainer.SaveModelToFile(fmt.Sprintf("stats_gen_%d.json", gen), stats)
}

// TournamentTraining runs training using tournaments for evaluation
func (t *GPUTrainer) TournamentTraining(generations int) {
	if len(t.Models) == 0 {
		t.InitializePopulation()
	}

	t.Stats.Reset()

	for gen := 1; gen <= generations; gen++ {
		genStartTime := time.Now()

		t.Generation = gen

		// Evaluate all models using tournament
		tournamentStart := time.Now()
		t.EvaluateWithTournament()
		tournamentTime := time.Since(tournamentStart)
		t.Stats.RecordOperation("tournament", tournamentTime)

		// Save generation statistics
		t.SaveGenerationStats(gen)

		// Create next generation
		if gen < generations {
			t.createNextGeneration()
		}

		genTime := time.Since(genStartTime)
		t.Stats.RecordOperation("generation", genTime)

	}
}

// EvaluateWithTournament evaluates models using tournament play
func (t *GPUTrainer) EvaluateWithTournament() {
	fmt.Printf("\nRunning tournament for generation %d\n", t.Generation)

	// Create progress bar description
	progressDesc := fmt.Sprintf("Tournament Gen %d", t.Generation)

	// Create a tournament
	tournament := NewTournament(t.Models, t.NumGames/2, t.MaxDepth, true)

	// Configure GPU usage for tournament
	tournament.UseGPU = t.UseGPU

	// Set the progress bar description
	tournament.ProgressDescription = progressDesc

	// Run the tournament
	tournament.RunTournament()

	// Display tournament results
	tournament.PrintResults()

	// Update model fitness based on tournament results
	for _, result := range tournament.Results {
		// Skip the standard AI entry
		if result.ModelIndex >= len(t.Models) {
			continue
		}

		// Update model statistics and fitness
		t.Models[result.ModelIndex].Wins = result.Wins
		t.Models[result.ModelIndex].Losses = result.Losses
		t.Models[result.ModelIndex].Draws = result.Draws
		t.Models[result.ModelIndex].Fitness = result.Score
	}

	// Sort models by fitness
	t.sortModelsByFitness()

	// Update best model if needed
	bestIdx, bestModel := tournament.GetBestModel()
	if bestIdx != -1 && bestModel != nil && bestModel.Fitness > t.BestModel.Fitness {
		t.BestModel = *bestModel
		t.SaveModel("best_model.json", t.BestModel)
	}
}

// CleanupGPU cleans up GPU resources
func (t *GPUTrainer) CleanupGPU() {
	if t.UseGPU {
		evaluation.CleanupCUDA()
	}
}
