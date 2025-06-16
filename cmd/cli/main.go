package main

import (
	"fmt"
	"strings"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
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

func main() {
	evaluator := evaluation.NewMixedEvaluationWithCoefficients(evaluation.V3Coeff)

	for {
		algebraicPosition := ""

		fmt.Print("Board > ")
		fmt.Scanln(&algebraicPosition)
		algebraicPosition = strings.ToLower(algebraicPosition)

		g := game.NewGame("Black", "White")
		pos := utils.AlgebraicToPositions(algebraicPosition)
		err := applyPosition(g, pos)
		if err != nil {
			fmt.Println(err)
			continue
		}

		var move game.Position
		found := false
		if openings := opening.MatchOpening(algebraicPosition); len(openings) > 0 {
			best := opening.Opening{}
			for _, opening := range openings {
				if len(opening.Transcript) > len(best.Transcript) {
					best = opening
				}
			}

			if len(best.Transcript) > len(algebraicPosition) {
				found = true
				nextMove := utils.AlgebraicToPosition(best.Transcript[len(algebraicPosition) : len(algebraicPosition)+2])

				move = nextMove
			}

		}
		if !found {
			move, _ = evaluation.Solve(*g, g.CurrentPlayer, 5, evaluator)
		}

		fmt.Println(utils.PositionToAlgebraic(move))
	}
}
