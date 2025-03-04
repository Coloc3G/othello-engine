package test

import (
	"testing"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
)

func TestMixedEvaluation_InitialBoard(t *testing.T) {
	// ...existing setup code...
	// Prepare a new game and get initial board
	g := game.NewGame()
	evalInstance := evaluation.NewMixedEvaluation()

	// Evaluate initial board for current player
	score := evalInstance.Evaluate(g.Board, g.CurrentPlayer)
	t.Logf("Initial board score for %s: %d", g.CurrentPlayer.Name, score)
}

func TestMixedEvaluation_CustomBoard(t *testing.T) {
	// ...existing setup code...
	// Create a custom board configuration
	var customBoard game.Board
	// Initialize board: set all cells to empty
	for i := range customBoard {
		for j := range customBoard[i] {
			customBoard[i][j] = game.Empty
		}
	}
	// Place pieces manually (example configuration)
	customBoard[0][0] = game.Black
	customBoard[0][1] = game.White
	customBoard[1][0] = game.White
	customBoard[1][1] = game.Black

	// Create a dummy game with the custom board, players and current player (Black)
	g := game.Game{
		Board: customBoard,
		Players: [2]game.Player{
			{Color: game.Black, Name: "Black"},
			{Color: game.White, Name: "White"},
		},
		CurrentPlayer: game.Player{Color: game.Black, Name: "Black"},
	}

	evalInstance := evaluation.NewMixedEvaluation()
	score := evalInstance.Evaluate(g.Board, g.CurrentPlayer)
	t.Logf("Custom board score for %s: %d", g.CurrentPlayer.Name, score)
}
