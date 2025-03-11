package learning

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/Coloc3G/othello-engine/models/utils"
	"github.com/schollz/progressbar/v3"
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

// Trainer manages the training of AI evaluation functions
type Trainer struct {
	Models         []EvaluationModel
	BestModel      EvaluationModel
	Generation     int
	PopulationSize int
	MutationRate   float64
	NumGames       int
	MaxDepth       int
}

// NewTrainer creates a new trainer with default parameters
func NewTrainer(popSize int) *Trainer {
	return &Trainer{
		Models:         make([]EvaluationModel, 0),
		PopulationSize: popSize,
		MutationRate:   0.2,
		NumGames:       100,
		MaxDepth:       5,
		Generation:     1,
	}
}

// InitializePopulation creates initial random population of models
func (t *Trainer) InitializePopulation() {
	t.Models = make([]EvaluationModel, t.PopulationSize)

	// Initialize with a reasonable default model
	defaultModel := EvaluationModel{
		Coeffs:     evaluation.V2Coeff,
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

	// Create progress bar for generations
	genBar := progressbar.NewOptions(generations,
		progressbar.OptionSetDescription("Training progress"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	genBar.RenderBlank()
	for gen := 1; gen <= generations; gen++ {
		t.Generation = gen
		fmt.Printf("\nGeneration %d/%d\n", gen, generations)

		// Evaluate all models using CPU
		t.evaluatePopulation()

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

		// Update progress bar for generations
		genBar.Add(1)
	}
}

// calculateAvgFitness calculates the average fitness of the population
func (t *Trainer) calculateAvgFitness() float64 {
	sum := 0.0
	for _, model := range t.Models {
		sum += model.Fitness
	}
	return sum / float64(len(t.Models))
}

// evaluatePopulation evaluates all models by playing games
func (t *Trainer) evaluatePopulation() {
	var wg sync.WaitGroup
	modelCount := len(t.Models)

	// Create progress bar for model evaluation
	evalBar := progressbar.NewOptions(modelCount,
		progressbar.OptionSetDescription("Evaluating models"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	// Mutex for updating progress bar safely
	var mutex sync.Mutex
	evalBar.RenderBlank()

	for i := range t.Models {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			t.evaluateModel(&t.Models[index], t.NumGames)

			// Safely update progress bar
			mutex.Lock()
			evalBar.Add(1)
			mutex.Unlock()
		}(i)
	}

	wg.Wait()
	fmt.Println() // Add newline after progress bar completes
}

// applyRandomOpening selects a random opening and applies it to the game
func applyOpening(g *game.Game, o opening.Opening) {
	transcript := o.Transcript

	// Apply the moves from the opening transcript
	for i := 0; i < len(transcript); i += 2 {
		if i+1 >= len(transcript) {
			break // Ensure we have a complete move (row and column)
		}

		move := utils.AlgebraicToPosition(transcript[i : i+2])

		// Apply the move
		g.Board, _ = game.GetNewBoardAfterMove(g.Board, move, g.CurrentPlayer)
		g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
	}
}

// evaluateModel evaluates a single model by playing games against reference players
func (t *Trainer) evaluateModel(model *EvaluationModel, numGames int) {
	// Reset statistics
	model.Wins = 0
	model.Losses = 0
	model.Draws = 0

	// Create evaluation function with model coefficients
	evalFunc := t.createEvaluationFromModel(*model)

	// Use a standard evaluation for opponent
	standardEval := evaluation.NewMixedEvaluationWithCoefficients(evaluation.V1Coeff)

	for _, o := range opening.KNOWN_OPENINGS {
		for i := 0; i < 2; i++ {

			// Create a new game
			g := game.NewGame("AI", "AI")

			// Apply opening
			applyOpening(g, o)

			// Alternate who plays first
			playerModel := &g.Players[i%2]

			// Play the game
			for !game.IsGameFinished(g.Board) {
				if g.CurrentPlayer.Color == playerModel.Color {
					// Model player's turn
					if len(game.ValidMoves(g.Board, g.CurrentPlayer.Color)) > 0 {
						pos := evaluation.Solve(*g, g.CurrentPlayer, t.MaxDepth, evalFunc)
						g.ApplyMove(pos)
					} else {
						// Skip turn if no valid moves
						g.CurrentPlayer = g.GetOtherPlayerMethod()
					}
				} else {
					// Standard player's turn
					if len(game.ValidMoves(g.Board, g.CurrentPlayer.Color)) > 0 {
						pos := evaluation.Solve(*g, g.CurrentPlayer, t.MaxDepth, standardEval)
						g.ApplyMove(pos)
					} else {
						// Skip turn if no valid moves
						g.CurrentPlayer = g.GetOtherPlayerMethod()
					}
				}

			}

			// Determine winner
			blackCount, whiteCount := game.CountPieces(g.Board)
			if playerModel.Color == game.Black {
				if blackCount > whiteCount {
					model.Wins++
				} else if blackCount < whiteCount {
					model.Losses++
				} else {
					model.Draws++
				}
			} else {
				if whiteCount > blackCount {
					model.Wins++
				} else if whiteCount < blackCount {
					model.Losses++
				} else {
					model.Draws++
				}
			}
		}

	}

	// Calculate fitness
	model.Fitness = float64(model.Wins) + float64(model.Draws)*0.5
	model.Generation = t.Generation
}

// createEvaluationFromModel creates a custom MixedEvaluation using model coefficients
func (t *Trainer) createEvaluationFromModel(model EvaluationModel) evaluation.Evaluation {
	return evaluation.NewMixedEvaluationWithCoefficients(model.Coeffs)
}

// createNextGeneration creates a new generation through selection, crossover and mutation
func (t *Trainer) createNextGeneration() {
	newModels := make([]EvaluationModel, t.PopulationSize)

	// Keep the best models (elitism)
	eliteCount := t.PopulationSize / 5
	copy(newModels[:eliteCount], t.Models[:eliteCount])

	// Fill the rest with crossover and mutation
	for i := eliteCount; i < t.PopulationSize; i++ {
		// Select parents using tournament selection
		parent1 := t.tournamentSelect(3)
		parent2 := t.tournamentSelect(3)

		// Crossover
		child := t.crossover(parent1, parent2)

		// Mutation
		child = t.mutateModel(child)
		child.Generation = t.Generation + 1

		newModels[i] = child
	}

	t.Models = newModels
}

// tournamentSelect selects a model using tournament selection
func (t *Trainer) tournamentSelect(tournamentSize int) EvaluationModel {
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
func (t *Trainer) crossover(parent1, parent2 EvaluationModel) EvaluationModel {
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

	// Crossover each coefficient with 50% chance from each parent
	for i := 0; i < 3; i++ {
		// Material coefficients
		if i%2 == 0 {
			child.Coeffs.MaterialCoeffs[i] = parent1.Coeffs.MaterialCoeffs[i]
		} else {
			child.Coeffs.MaterialCoeffs[i] = parent2.Coeffs.MaterialCoeffs[i]
		}

		// Mobility coefficients
		if (i+1)%2 == 0 {
			child.Coeffs.MobilityCoeffs[i] = parent1.Coeffs.MobilityCoeffs[i]
		} else {
			child.Coeffs.MobilityCoeffs[i] = parent2.Coeffs.MobilityCoeffs[i]
		}

		// Corners coefficients
		if i%2 == 0 {
			child.Coeffs.CornersCoeffs[i] = parent1.Coeffs.CornersCoeffs[i]
		} else {
			child.Coeffs.CornersCoeffs[i] = parent2.Coeffs.CornersCoeffs[i]
		}

		// Parity coefficients
		if (i+1)%2 == 0 {
			child.Coeffs.ParityCoeffs[i] = parent1.Coeffs.ParityCoeffs[i]
		} else {
			child.Coeffs.ParityCoeffs[i] = parent2.Coeffs.ParityCoeffs[i]
		}

		// Stability coefficients
		if i%3 == 0 {
			child.Coeffs.StabilityCoeffs[i] = parent1.Coeffs.StabilityCoeffs[i]
		} else {
			child.Coeffs.StabilityCoeffs[i] = parent2.Coeffs.StabilityCoeffs[i]
		}

		// Frontier coefficients
		if (i+2)%3 == 0 {
			child.Coeffs.FrontierCoeffs[i] = parent1.Coeffs.FrontierCoeffs[i]
		} else {
			child.Coeffs.FrontierCoeffs[i] = parent2.Coeffs.FrontierCoeffs[i]
		}
	}

	return child
}

// mutateModel applies random mutations to a model
func (t *Trainer) mutateModel(model EvaluationModel) EvaluationModel {
	mutated := model

	// Mutate each coefficient with probability = mutation rate
	for i := 0; i < 3; i++ {
		if rand.Float64() < t.MutationRate {
			mutated.Coeffs.MaterialCoeffs[i] = clampMutation(model.Coeffs.MaterialCoeffs[i], 100)
		}
		if rand.Float64() < t.MutationRate {
			mutated.Coeffs.MobilityCoeffs[i] = clampMutation(model.Coeffs.MobilityCoeffs[i], 50)
		}
		if rand.Float64() < t.MutationRate {
			mutated.Coeffs.CornersCoeffs[i] = clampMutation(model.Coeffs.CornersCoeffs[i], 200)
		}
		if rand.Float64() < t.MutationRate {
			mutated.Coeffs.ParityCoeffs[i] = clampMutation(model.Coeffs.ParityCoeffs[i], 100)
		}
		if rand.Float64() < t.MutationRate {
			mutated.Coeffs.StabilityCoeffs[i] = clampMutation(model.Coeffs.StabilityCoeffs[i], 50)
		}
		if rand.Float64() < t.MutationRate {
			mutated.Coeffs.FrontierCoeffs[i] = clampMutation(model.Coeffs.FrontierCoeffs[i], 30)
		}
	}

	return mutated
}

// clampMutation applies a random mutation and ensures the value stays positive
func clampMutation(value, maxDelta int) int {
	// Apply random delta between -maxDelta and +maxDelta
	delta := rand.Intn(maxDelta*2+1) - maxDelta
	result := value + delta

	// Ensure value remains positive
	if result < 0 {
		return 0
	}
	return result
}

// sortModelsByFitness sorts models by fitness in descending order
func (t *Trainer) sortModelsByFitness() {
	sort.Slice(t.Models, func(i, j int) bool {
		return t.Models[i].Fitness > t.Models[j].Fitness
	})
}

// SaveModel saves a model to a JSON file
func (t *Trainer) SaveModel(filename string, model EvaluationModel) error {
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadModel loads a model from a JSON file
func (t *Trainer) LoadModel(filename string) (EvaluationModel, error) {
	var model EvaluationModel
	data, err := os.ReadFile(filename)
	if err != nil {
		return model, err
	}
	err = json.Unmarshal(data, &model)
	return model, err
}

// SaveGenerationStats saves statistics about the current generation
func (t *Trainer) SaveGenerationStats(gen int) error {
	stats := struct {
		Generation  int             `json:"generation"`
		BestFitness float64         `json:"best_fitness"`
		AvgFitness  float64         `json:"avg_fitness"`
		BestModel   EvaluationModel `json:"best_model"`
		Timestamp   string          `json:"timestamp"`
	}{
		Generation:  gen,
		BestFitness: t.Models[0].Fitness,
		BestModel:   t.Models[0],
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	// Calculate average fitness
	var sum float64
	for _, model := range t.Models {
		sum += model.Fitness
	}
	stats.AvgFitness = sum / float64(len(t.Models))

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("stats_gen_%d.json", gen)
	return os.WriteFile(filename, data, 0644)
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

// TournamentTraining runs training using tournaments for evaluation
func (t *Trainer) TournamentTraining(generations int) {
	if len(t.Models) == 0 {
		t.InitializePopulation()
	}

	// Create progress bar for generations
	genBar := progressbar.NewOptions(generations,
		progressbar.OptionSetDescription("Tournament training progress"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	genBar.RenderBlank()
	for gen := 1; gen <= generations; gen++ {
		t.Generation = gen
		fmt.Printf("\nGeneration %d/%d\n", gen, generations)

		// Evaluate all models using tournament
		t.EvaluateWithTournament()

		// Save generation statistics
		t.SaveGenerationStats(gen)

		// Create next generation
		if gen < generations {
			t.createNextGeneration()
		}

		// Update progress bar for generations
		genBar.Add(1)
	}
}
