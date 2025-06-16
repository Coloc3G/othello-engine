package learning

import (
	"fmt"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
)

// Trainer manages the training of AI evaluation functions using CPU
type Trainer struct {
	*BaseTrainer
}

// NewTrainer creates a new trainer with default parameters
func NewTrainer(popSize, numGames, depth int, baseModelCoeffs evaluation.EvaluationCoefficients) *Trainer {
	return &Trainer{
		BaseTrainer: NewBaseTrainer(popSize, numGames, depth, baseModelCoeffs),
	}
}

// InitializePopulation creates initial random population of models
func (t *Trainer) InitializePopulation() {
	t.Models = make([]EvaluationModel, t.PopulationSize)

	// Initialize with a reasonable default model
	defaultModel := EvaluationModel{
		Coeffs:     t.BaseModel,
		Generation: 1,
	}

	t.Models[0] = defaultModel
	t.BestModel = defaultModel

	// Create variations of the default model
	for i := 1; i < t.PopulationSize; i++ {
		t.Models[i] = t.mutateModel(defaultModel)
		t.Models[i].Generation = 1
	}
}

// StartTraining begins the genetic algorithm training process
func (t *Trainer) StartTraining(generations int) {
	if len(t.Models) == 0 {
		t.InitializePopulation()
	}

	t.Stats.Reset()

	for gen := 1; gen <= generations; gen++ {
		genStartTime := time.Now()

		t.Generation = gen
		fmt.Printf("\nGeneration %d/%d\n", gen, generations)

		// Evaluate all models using CPU
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

			if t.BestModel.Fitness >= 2*float64(t.NumGames)-1 {
				fmt.Println("Best model reached target fitness, now training on this best model.")
				t.BaseModel = t.BestModel.Coeffs
			}
		}

		// Display current best fitness
		fmt.Printf("Best fitness: %.2f, Avg fitness: %.2f\n", t.Models[0].Fitness, t.calculateAvgFitness())

		// Save generation statistics
		t.SaveGenerationStats(gen)

		// Modify this reinforcement section - use gentler reinforcement
		if gen >= 3 && calculateAverageWins(t.Models) == 0 {
			fmt.Println("No wins detected. Using more aggressive exploration for this generation.")
			// Instead of reinforcing the whole population, just increase mutation rates temporarily
			t.MutationRate += 0.1
			if t.MutationRate > 0.8 {
				t.MutationRate = 0.8
			}
		} else {
			// Return to normal mutation rate
			t.MutationRate = 0.3
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

// evaluatePopulation evaluates all models by playing games
func (t *Trainer) evaluatePopulation() {
	// Get models as pointer slice for parallel evaluation
	modelPtrs := make([]*EvaluationModel, len(t.Models))
	for i := range t.Models {
		modelPtrs[i] = &t.Models[i]
	}

	// Evaluate all models in parallel
	evaluateModelsInParallel(modelPtrs, t.BestModel.Coeffs, t.MaxDepth, t.NumGames, t.Stats)
}
