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
func NewTrainer(popSize int) *Trainer {
	return &Trainer{
		BaseTrainer: NewBaseTrainer(popSize),
	}
}

// InitializePopulation creates initial random population of models
func (t *Trainer) InitializePopulation() {
	t.Models = make([]EvaluationModel, t.PopulationSize)

	// Initialize with a reasonable default model
	defaultModel := EvaluationModel{
		Coeffs:     evaluation.V1Coeff,
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
		}

		// Display current best fitness
		fmt.Printf("Best fitness: %.2f, Avg fitness: %.2f\n", t.Models[0].Fitness, t.calculateAvgFitness())

		// Save generation statistics
		t.SaveGenerationStats(gen)

		// Create next generation
		if gen < generations {
			t.createNextGeneration()
		}

		genTime := time.Since(genStartTime)
		t.Stats.RecordOperation("generation", genTime)

		// Print performance statistics every 5 generations
		if gen%5 == 0 || gen == generations {
			t.Stats.PrintSummary()
		}
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

	// Define CPU evaluation function creator
	createEvalFunc := func(model EvaluationModel) evaluation.Evaluation {
		return evaluation.NewMixedEvaluationWithCoefficients(model.Coeffs)
	}

	// Evaluate all models in parallel
	evaluateModelsInParallel(modelPtrs, createEvalFunc, t.MaxDepth, t.NumGames, t.Stats)
}

// TournamentTraining runs training using tournaments for evaluation
func (t *Trainer) TournamentTraining(generations int) {
	if len(t.Models) == 0 {
		t.InitializePopulation()
	}

	t.Stats.Reset()

	for gen := 1; gen <= generations; gen++ {
		genStartTime := time.Now()

		t.Generation = gen
		fmt.Printf("\nGeneration %d/%d\n", gen, generations)

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

		// Print performance statistics every 5 generations
		if gen%5 == 0 || gen == generations {
			t.Stats.PrintSummary()
		}
	}

	fmt.Println("\nTournament training completed!")
}

// EvaluateWithTournament evaluates models using tournament play instead of individual games
func (t *Trainer) EvaluateWithTournament() {
	fmt.Println("Evaluating models using tournament system...")

	// Create a tournament with all models plus standard AI
	tournament := NewTournament(t.Models, t.NumGames/2, t.MaxDepth, true)

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
		fmt.Printf("New best model from tournament: fitness %.2f, win rate %.2f%%\n",
			t.BestModel.Fitness,
			float64(t.BestModel.Wins)/float64(t.BestModel.Wins+t.BestModel.Losses+t.BestModel.Draws)*100)
		t.SaveModel("best_model.json", t.BestModel)
	} else if bestIdx == len(t.Models) {
		fmt.Println("Standard AI won the tournament!")
	}
}
