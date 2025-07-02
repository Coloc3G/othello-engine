package learning

import (
	"fmt"
	"sort"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
)

// NewTrainer creates a new trainer with default parameters
func NewTrainer(name string, popSize, numGames int, depth int8, baseModelCoeffs evaluation.EvaluationCoefficients) *Trainer {
	return &Trainer{
		Name:           name,
		Models:         make([]EvaluationModel, 0),
		BaseModel:      baseModelCoeffs,
		PopulationSize: popSize,
		MutationRate:   0.3,
		NumGames:       numGames,
		MaxDepth:       depth,
		Generation:     1,
	}
}

// StartTraining begins the genetic algorithm training process
func (t *Trainer) StartTraining(generations int) {

	if t.createModelDirectory() != nil {
		fmt.Println("Error creating model directory")
		return
	}

	trainingStart := time.Now()
	if len(t.Models) == 0 {
		t.InitializePopulation()
	}

	for gen := 1; gen <= generations; gen++ {
		genStartTime := time.Now()

		t.Generation = gen
		fmt.Printf("\nGeneration %d/%d\n", gen, generations)

		// Evaluate all models
		t.evaluatePopulation()
		t.sortModelsByFitness()

		fmt.Println("Generation time:", time.Since(genStartTime))

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

		// Create next generation if not last generation
		if gen < generations {
			t.createNextGeneration()
		}
	}

	fmt.Printf("\nTraining completed in %s\n", time.Since(trainingStart))
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
		t.Models[i] = CreateDiverseModel(defaultModel)
		t.Models[i].Generation = 1
	}
}

// createNextGeneration creates a new generation through selection, crossover and mutation
func (t *Trainer) createNextGeneration() {

	newModels := make([]EvaluationModel, t.PopulationSize)

	// Increase elitism to preserve more good models
	eliteCount := t.PopulationSize / 4
	copy(newModels[:eliteCount], t.Models[:eliteCount])

	// Fill the rest with crossover and mutation
	for i := eliteCount; i < t.PopulationSize; i++ {

		// Use larger tournament size to focus on better models
		parent1 := t.tournamentSelect(5)
		parent2 := t.tournamentSelect(5)

		// Crossover
		child := t.crossover(parent1, parent2)

		// Mutation
		child = t.mutateModel(child)
		child.Generation = t.Generation + 1

		newModels[i] = child
	}

	t.Models = newModels
}

// evaluatePopulation evaluates all models by playing games
func (t *Trainer) evaluatePopulation() {
	// Get models as pointer slice for parallel evaluation
	modelPtrs := make([]*EvaluationModel, len(t.Models))
	for i := range t.Models {
		modelPtrs[i] = &t.Models[i]
	}

	// Evaluate all models in parallel
	evaluateModelsInParallel(modelPtrs, t.BaseModel, t.MaxDepth, t.NumGames)
}

// sortModelsByFitness sorts models by fitness in descending order
func (t *Trainer) sortModelsByFitness() {
	sort.Slice(t.Models, func(i, j int) bool {
		return t.Models[i].Fitness > t.Models[j].Fitness
	})
}

// calculateAvgFitness calculates the average fitness of the population
func (t *Trainer) calculateAvgFitness() float64 {
	sum := 0.0
	for _, model := range t.Models {
		sum += model.Fitness
	}
	return sum / float64(len(t.Models))
}
