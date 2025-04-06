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

// GPUTrainer enhances the Trainer with GPU capabilities
type GPUTrainer struct {
	Models         []EvaluationModel
	BestModel      EvaluationModel
	Generation     int
	PopulationSize int
	MutationRate   float64
	NumGames       int
	MaxDepth       int
	UseGPU         bool
}

// NewGPUTrainer creates a new trainer with GPU support
func NewGPUTrainer(popSize int) *GPUTrainer {
	// Initialize CUDA and check availability
	gpuAvailable := evaluation.InitCUDA()

	return &GPUTrainer{
		Models:         make([]EvaluationModel, 0),
		PopulationSize: popSize,
		MutationRate:   0.2,
		NumGames:       100,
		MaxDepth:       5,
		Generation:     1,
		UseGPU:         gpuAvailable,
	}
}

// InitializePopulation creates initial random population of models
func (t *GPUTrainer) InitializePopulation() {
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
func (t *GPUTrainer) StartTraining(generations int) {
	if len(t.Models) == 0 {
		t.InitializePopulation()
	}

	for gen := 1; gen <= generations; gen++ {
		t.Generation = gen
		fmt.Printf("\nGeneration %d/%d using %s\n", gen, generations,
			map[bool]string{true: "GPU acceleration", false: "CPU only"}[t.UseGPU])

		// Evaluate all models
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
	}

	fmt.Println("\nTraining completed!")
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
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Calculate total number of matches to play
	totalMatches := t.getTotalMatchCount()

	fmt.Printf("Evaluating models - playing %d matches total\n", totalMatches)

	// Create a single progress bar for all matches
	bar := progressbar.NewOptions(totalMatches,
		progressbar.OptionSetDescription("Match progress"),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	bar.RenderBlank()

	incrementBarFunc := func() {
		mutex.Lock()
		defer mutex.Unlock()
		bar.Add(1)
		bar.RenderBlank()
	}

	for i := range t.Models {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			t.evaluateModel(&t.Models[index], incrementBarFunc)
		}(i)
	}

	wg.Wait()
	fmt.Println() // Add newline after progress bar completes
}

// getTotalMatchCount calculates the total number of matches to be played in evaluation
func (t *GPUTrainer) getTotalMatchCount() int {
	// Each model plays against the standard model
	// Each game is played twice (each player plays once as black and once as white)
	return len(t.Models) * len(opening.KNOWN_OPENINGS) * 2
}

// evaluateModel evaluates a single model by playing games against reference players
// Returns the number of matches played
func (t *GPUTrainer) evaluateModel(model *EvaluationModel, incrementBarFunc func()) {
	// Reset statistics
	model.Wins = 0
	model.Losses = 0
	model.Draws = 0

	// Create evaluation function with model coefficients
	evalFunc := t.createEvaluationFromModel(*model)

	// Use a standard evaluation for opponent
	standardEval := evaluation.NewMixedEvaluationWithCoefficients(evaluation.V1Coeff)

	for _, o := range opening.KNOWN_OPENINGS {
		// Play each opening twice, alternating who plays first
		for i := range 2 {
			// Create a new game
			g := game.NewGame("AI", "AI")

			// Apply opening
			t.applyOpening(g, o)

			// Determine player model (alternate between games)
			playerModelIdx := i % 2
			playerModel := &g.Players[playerModelIdx]

			// Play the game until completion
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

			// Record game result based on player model's perspective
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

			// Increment progress bar for each game played
			incrementBarFunc()
		}
	}

	// Calculate fitness
	model.Fitness = float64(model.Wins) + float64(model.Draws)*0.5
	model.Generation = t.Generation
}

// applyOpening selects a random opening and applies it to the game
func (t *GPUTrainer) applyOpening(g *game.Game, o opening.Opening) {
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

// createEvaluationFromModel creates a custom evaluation using model coefficients
func (t *GPUTrainer) createEvaluationFromModel(model EvaluationModel) evaluation.Evaluation {
	if t.UseGPU {
		return evaluation.NewGPUMixedEvaluation(model.Coeffs)
	}
	return evaluation.NewMixedEvaluationWithCoefficients(model.Coeffs)
}

// createNextGeneration creates a new generation through selection, crossover and mutation
func (t *GPUTrainer) createNextGeneration() {
	// Reuse the Trainer's implementation
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
func (t *GPUTrainer) tournamentSelect(tournamentSize int) EvaluationModel {
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
func (t *GPUTrainer) crossover(parent1, parent2 EvaluationModel) EvaluationModel {
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
func (t *GPUTrainer) mutateModel(model EvaluationModel) EvaluationModel {
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

// sortModelsByFitness sorts models by fitness in descending order
func (t *GPUTrainer) sortModelsByFitness() {
	sort.Slice(t.Models, func(i, j int) bool {
		return t.Models[i].Fitness > t.Models[j].Fitness
	})
}

// SaveModel saves a model to a JSON file
func (t *GPUTrainer) SaveModel(filename string, model EvaluationModel) error {
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadModel loads a model from a JSON file
func (t *GPUTrainer) LoadModel(filename string) (EvaluationModel, error) {
	var model EvaluationModel
	data, err := os.ReadFile(filename)
	if err != nil {
		return model, err
	}
	err = json.Unmarshal(data, &model)
	return model, err
}

// SaveGenerationStats saves statistics about the current generation
func (t *GPUTrainer) SaveGenerationStats(gen int) error {
	stats := struct {
		Generation  int             `json:"generation"`
		BestFitness float64         `json:"best_fitness"`
		AvgFitness  float64         `json:"avg_fitness"`
		BestModel   EvaluationModel `json:"best_model"`
		Timestamp   string          `json:"timestamp"`
		UsingGPU    bool            `json:"using_gpu"`
	}{
		Generation:  gen,
		BestFitness: t.Models[0].Fitness,
		BestModel:   t.Models[0],
		Timestamp:   time.Now().Format(time.RFC3339),
		UsingGPU:    t.UseGPU,
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

// TournamentTraining runs training using tournaments for evaluation
func (t *GPUTrainer) TournamentTraining(generations int) {
	if len(t.Models) == 0 {
		t.InitializePopulation()
	}

	for gen := 1; gen <= generations; gen++ {
		t.Generation = gen
		fmt.Printf("\nGeneration %d/%d using %s\n", gen, generations,
			map[bool]string{true: "GPU acceleration", false: "CPU only"}[t.UseGPU])

		// Evaluate all models using tournament
		t.EvaluateWithTournament()

		// Save generation statistics
		t.SaveGenerationStats(gen)

		// Create next generation
		if gen < generations {
			t.createNextGeneration()
		}
	}

	fmt.Println("\nTournament training completed!")
}

// EvaluateWithTournament evaluates models using tournament play
func (t *GPUTrainer) EvaluateWithTournament() {
	fmt.Println("Evaluating models using tournament system...")

	// Create a tournament
	tournament := NewTournament(t.Models, t.NumGames/2, t.MaxDepth, true)

	// Configure GPU usage for tournament
	tournament.UseGPU = t.UseGPU

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

// CleanupGPU cleans up GPU resources
func (t *GPUTrainer) CleanupGPU() {
	if t.UseGPU {
		evaluation.CleanupCUDA()
	}
}
