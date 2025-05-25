package main

import (
	"fmt"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/utils"
)

// applyOpening applies a predefined opening to a game
func applyPosition(g *game.Game, pos []game.Position) (err error) {
	for _, move := range pos {
		if !game.IsValidMove(g.Board, g.CurrentPlayer.Color, move) {
			return fmt.Errorf("invalid move %s for player %s", utils.PositionToAlgebraic(move), g.CurrentPlayer.Name)
		}
		// Apply the move
		g.Board, _ = game.GetNewBoardAfterMove(g.Board, move, g.CurrentPlayer)
		g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
		if !game.HasAnyMoves(g.Board, g.CurrentPlayer.Color) {
			g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
		}
	}
	return
}

func printBoard(b game.Board) {
	fmt.Println("  a b c d e f g h")
	for i := 0; i < 8; i++ {
		fmt.Printf("%d ", i+1)
		for j := 0; j < 8; j++ {
			switch b[i][j] {
			case game.Black:
				fmt.Print("B ")
			case game.White:
				fmt.Print("W ")
			default:
				fmt.Print(". ")
			}
		}
		fmt.Println()
	}
}

func main() {
	evaluator := evaluation.NewMixedEvaluationWithCoefficients(evaluation.V2Coeff)

	for {

		algebraicPosition := ""

		fmt.Print("Board > ")
		fmt.Scanln(&algebraicPosition)

		g := game.NewGame("Black", "White")
		pos := utils.AlgebraicToPositions(algebraicPosition)
		err := applyPosition(g, pos)
		if err != nil {
			fmt.Println(err)
			continue
		}
		move := evaluation.Solve(*g, g.CurrentPlayer, 5, evaluator)
		fmt.Println(utils.PositionToAlgebraic(move))
	}
}
