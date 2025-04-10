package learning

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/ai/stats"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/schollz/progressbar/v3"
)

// TournamentResult stores the results of a tournament match
type TournamentResult struct {
	ModelIndex int     // Index of the model in the tournament
	Name       string  // Name of the model
	Wins       int     // Number of wins
	Losses     int     // Number of losses
	Draws      int     // Number of draws
	Score      float64 // Total score (2 points for win, 1 for draw)
}

// Tournament manages tournament play between AI models
type Tournament struct {
	Models              []EvaluationModel
	Results             []TournamentResult
	NumGames            int // Number of games per match
	MaxDepth            int // Search depth for AI
	StandardAI          evaluation.Evaluation
	UseStandard         bool                    // Whether to include standard AI in tournament
	UseGPU              bool                    // Whether to use GPU acceleration
	Stats               *stats.PerformanceStats // Performance statistics
	ProgressDescription string                  // Description for progress bar
}

// NewTournament creates a new tournament with specified parameters
func NewTournament(models []EvaluationModel, numGames, maxDepth int, useStandard bool) *Tournament {
	// Check GPU availability
	gpuAvailable := evaluation.IsGPUAvailable()

	// Create standardAI with GPU if available
	var standardAI evaluation.Evaluation
	if gpuAvailable {
		standardAI = evaluation.NewGPUMixedEvaluation(evaluation.V1Coeff)
	} else {
		standardAI = evaluation.NewMixedEvaluation()
	}

	return &Tournament{
		Models:      models,
		Results:     make([]TournamentResult, 0),
		NumGames:    numGames,
		MaxDepth:    maxDepth,
		StandardAI:  standardAI,
		UseStandard: useStandard,
		UseGPU:      gpuAvailable,
		Stats:       stats.NewPerformanceStats(),
	}
}

// RunTournament runs a tournament between all models
func (t *Tournament) RunTournament() {
	tournamentStart := time.Now()

	// Calculate total number of competitors
	numCompetitors := len(t.Models)
	if t.UseStandard {
		numCompetitors++
	}

	// Initialize results array
	t.Results = make([]TournamentResult, numCompetitors)
	for i := 0; i < len(t.Models); i++ {
		t.Results[i] = TournamentResult{
			ModelIndex: i,
			Name:       fmt.Sprintf("Model %d", i),
		}
	}

	// Add standard AI if needed
	if t.UseStandard {
		t.Results[numCompetitors-1] = TournamentResult{
			ModelIndex: numCompetitors - 1,
			Name:       "Standard AI",
		}
	}

	// Calculate total number of matches
	totalMatches := numCompetitors * (numCompetitors - 1) / 2
	totalGames := totalMatches * t.NumGames

	// Use custom description if provided, otherwise use default
	barDescription := "Tournament progress"
	if t.ProgressDescription != "" {
		barDescription = t.ProgressDescription
	}

	// Create progress bar
	bar := progressbar.NewOptions(
		totalGames,
		progressbar.OptionSetDescription(barDescription),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	bar.RenderBlank()

	// Set up channels for progress tracking
	type matchResult struct {
		model1     int
		model2     int
		wins1      int
		wins2      int
		draws      int
		gamesCount int
		duration   time.Duration
	}

	results := make(chan matchResult, totalMatches)
	var wg sync.WaitGroup

	// Play games between all pairs of models
	for i := 0; i < numCompetitors; i++ {
		for j := i + 1; j < numCompetitors; j++ {
			wg.Add(1)

			go func(model1, model2 int) {
				defer wg.Done()

				// Play games between model1 and model2
				matchStart := time.Now()
				wins1, wins2, draws := t.playMatch(model1, model2)
				matchDuration := time.Since(matchStart)

				// Send results back through channel
				results <- matchResult{
					model1:     model1,
					model2:     model2,
					wins1:      wins1,
					wins2:      wins2,
					draws:      draws,
					gamesCount: t.NumGames,
					duration:   matchDuration,
				}
			}(i, j)
		}
	}

	// Create a goroutine to collect results and update progress
	go func() {
		gamesProcessed := 0
		var totalMatchTime time.Duration
		matchesCompleted := 0

		for result := range results {
			// Update results
			t.Results[result.model1].Wins += result.wins1
			t.Results[result.model1].Losses += result.wins2
			t.Results[result.model1].Draws += result.draws

			t.Results[result.model2].Wins += result.wins2
			t.Results[result.model2].Losses += result.wins1
			t.Results[result.model2].Draws += result.draws

			// Update progress bar
			gamesProcessed += result.gamesCount
			bar.Set(gamesProcessed)
			bar.RenderBlank()

			// Update statistics
			totalMatchTime += result.duration
			matchesCompleted++
		}

		// Record stats
		if t.Stats != nil {
			t.Stats.Counts["tournament_matches"] = matchesCompleted
			t.Stats.Counts["tournament_games"] = gamesProcessed
		}
	}()

	// Wait for all matches to complete
	wg.Wait()
	close(results)

	fmt.Println() // Add newline after progress bar completes

	// Calculate final scores (2 points for win, 1 for draw)
	for i := range t.Results {
		t.Results[i].Score = float64(2*t.Results[i].Wins + t.Results[i].Draws)
	}

	// Sort results by score
	sort.Slice(t.Results, func(i, j int) bool {
		return t.Results[i].Score > t.Results[j].Score
	})

	// Record tournament time
	if t.Stats != nil {
		tournamentTime := time.Since(tournamentStart)
		t.Stats.RecordOperation("tournament_total", tournamentTime)
	}
}

// playMatch plays a match between two models and returns the results
func (t *Tournament) playMatch(model1Idx, model2Idx int) (wins1, wins2, draws int) {
	// Get evaluators for each model
	var eval1, eval2 evaluation.Evaluation

	startEval := time.Now()

	// Check if either model is the standard AI
	if t.UseStandard && model1Idx == len(t.Models) {
		eval1 = t.StandardAI
	} else {
		eval1 = t.createEvaluationFromModel(t.Models[model1Idx])
	}

	if t.UseStandard && model2Idx == len(t.Models) {
		eval2 = t.StandardAI
	} else {
		eval2 = t.createEvaluationFromModel(t.Models[model2Idx])
	}

	evalTime := time.Since(startEval)
	if t.Stats != nil {
		t.Stats.RecordOperation("evaluation_creation", evalTime)
	}

	// Use a limited number of games based on tournament settings
	// Ensure we don't exceed the available openings
	gamesToPlay := t.NumGames
	if gamesToPlay > len(opening.KNOWN_OPENINGS) {
		gamesToPlay = len(opening.KNOWN_OPENINGS)
	}

	// Select random openings
	selectedOpenings := SelectRandomOpenings(gamesToPlay)

	// Play games with selected openings, alternating who starts first
	for i, op := range selectedOpenings {
		gameStart := time.Now()

		// Alternate who plays black (goes first)
		var blackEval, whiteEval evaluation.Evaluation
		var blackIdx, whiteIdx int

		if i%2 == 0 {
			blackEval = eval1
			whiteEval = eval2
			blackIdx = model1Idx
			whiteIdx = model2Idx
		} else {
			blackEval = eval2
			whiteEval = eval1
			blackIdx = model2Idx
			whiteIdx = model1Idx
		}

		// Create a new game with the opening
		g := game.NewGame("Black AI", "White AI")
		applyOpening(g, op)

		// Play the game until completion
		for !game.IsGameFinished(g.Board) {
			var evalFunc evaluation.Evaluation

			// Select the appropriate evaluation function for the current player
			if g.CurrentPlayer.Color == game.Black {
				evalFunc = blackEval
			} else {
				evalFunc = whiteEval
			}

			// Make a move if possible
			if len(game.ValidMoves(g.Board, g.CurrentPlayer.Color)) > 0 {
				moveStart := time.Now()
				pos := evaluation.Solve(*g, g.CurrentPlayer, t.MaxDepth, evalFunc)
				moveTime := time.Since(moveStart)

				if t.Stats != nil {
					t.Stats.RecordOperation("move_decision", moveTime)
					t.Stats.Counts["moves_made"]++
				}

				g.ApplyMove(pos)
			} else {
				// Skip turn if no valid moves
				g.CurrentPlayer = g.GetOtherPlayerMethod()
			}
		}

		// Determine the winner
		blackCount, whiteCount := game.CountPieces(g.Board)
		if blackCount > whiteCount {
			// Black wins
			if blackIdx == model1Idx {
				wins1++
			} else {
				wins2++
			}
		} else if whiteCount > blackCount {
			// White wins
			if whiteIdx == model1Idx {
				wins1++
			} else {
				wins2++
			}
		} else {
			// Draw
			draws++
		}

		gameTime := time.Since(gameStart)
		if t.Stats != nil {
			t.Stats.RecordOperation("game", gameTime)
			t.Stats.Counts["games_played"]++
		}
	}

	return wins1, wins2, draws
}

// playGame plays a single game between two evaluators and returns the winner
func (t *Tournament) playGame(blackEval, whiteEval evaluation.Evaluation) game.Piece {
	g := game.NewGame("Black AI", "White AI")

	// Play the game until completion
	for !game.IsGameFinished(g.Board) {
		var evalFunc evaluation.Evaluation

		// Select the appropriate evaluation function for the current player
		if g.CurrentPlayer.Color == game.Black {
			evalFunc = blackEval
		} else {
			evalFunc = whiteEval
		}

		// Make a move if possible
		if len(game.ValidMoves(g.Board, g.CurrentPlayer.Color)) > 0 {
			moveStart := time.Now()

			var pos game.Position
			if t.UseGPU {
				// Try GPU solve first for both players
				gpuPos, success := evaluation.GPUSolve(*g, g.CurrentPlayer, t.MaxDepth)
				if success {
					pos = gpuPos
				} else {
					// Fall back to regular solve
					pos = evaluation.Solve(*g, g.CurrentPlayer, t.MaxDepth, evalFunc)
				}
			} else {
				pos = evaluation.Solve(*g, g.CurrentPlayer, t.MaxDepth, evalFunc)
			}

			moveTime := time.Since(moveStart)

			if t.Stats != nil {
				t.Stats.RecordOperation("move_decision", moveTime)
				t.Stats.Counts["moves_made"]++
			}

			g.ApplyMove(pos)
		} else {
			// Skip turn if no valid moves
			g.CurrentPlayer = g.GetOtherPlayerMethod()
		}
	}

	// Determine the winner
	blackCount, whiteCount := game.CountPieces(g.Board)
	if blackCount > whiteCount {
		return game.Black
	} else if whiteCount > blackCount {
		return game.White
	} else {
		return game.Empty // Draw
	}
}

// createEvaluationFromModel creates a custom evaluation from model coefficients
func (t *Tournament) createEvaluationFromModel(model EvaluationModel) evaluation.Evaluation {
	if t.UseGPU {
		// Ensure GPU coefficients are set properly for this model
		gpuEval := evaluation.NewGPUMixedEvaluation(model.Coeffs)
		evaluation.SetCUDACoefficients(model.Coeffs)
		return gpuEval
	}
	return evaluation.NewMixedEvaluationWithCoefficients(model.Coeffs)
}

// PrintResults prints tournament results in a formatted table
func (t *Tournament) PrintResults() {
	fmt.Println("\nTournament Results:")
	fmt.Printf("%-15s %-8s %-8s %-8s %-8s %-8s\n", "Model", "Wins", "Losses", "Draws", "Score", "Win %")
	fmt.Println("------------------------------------------------------------------")

	for _, result := range t.Results {
		total := float64(result.Wins + result.Losses + result.Draws)
		winPercent := 0.0
		if total > 0 {
			winPercent = float64(result.Wins) / total * 100
		}

		fmt.Printf("%-15s %-8d %-8d %-8d %-8.1f %-8.1f\n",
			result.Name,
			result.Wins,
			result.Losses,
			result.Draws,
			result.Score,
			winPercent,
		)
	}

	// Print stats if available
	if t.Stats != nil && t.Stats.Counts["games_played"] > 0 {
		fmt.Println("\nTournament Statistics:")
		fmt.Printf("Games played: %d\n", t.Stats.Counts["games_played"])
		fmt.Printf("Average game time: %s\n",
			(t.Stats.TotalGenerationTime / time.Duration(t.Stats.Counts["games_played"])).Round(time.Millisecond))
	}
}

// GetBestModel returns the best model from the tournament
func (t *Tournament) GetBestModel() (int, *EvaluationModel) {
	if len(t.Results) == 0 {
		return -1, nil
	}

	bestIdx := t.Results[0].ModelIndex
	if t.UseStandard && bestIdx == len(t.Models) {
		// Standard AI won
		return bestIdx, nil
	}

	return bestIdx, &t.Models[bestIdx]
}
