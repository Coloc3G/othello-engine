package main

import (
	"flag"
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
		g.Board, _ = game.GetNewBoardAfterMove(g.Board, move, g.CurrentPlayer.Color)
		g.CurrentPlayer = game.GetOtherPlayer(g.CurrentPlayer.Color)
		if !game.HasAnyMoves(g.Board, g.CurrentPlayer.Color) {
			g.CurrentPlayer = game.GetOtherPlayer(g.CurrentPlayer.Color)
		}
	}
	return
}

func main() {

	debug := flag.Bool("debug", false, "Debug mode")
	depth := flag.Int("depth", 10, "Search depth for AI evaluation")
	mateDepth := flag.Int("mate-depth", 21, "Mate Search depth for AI evaluation")
	flag.Parse()

	evaluator := evaluation.NewMixedEvaluation(evaluation.Models[len(evaluation.Models)-1]) // Use the latest evaluation model

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

			var searchDepth = int8(*depth)
			if len(pos) >= 64-*mateDepth {
				searchDepth = int8(*mateDepth)
			}

			moves, score := evaluation.Solve(g.Board, g.CurrentPlayer.Color, searchDepth, evaluator)
			if len(moves) == 0 || (len(moves) == 1 && moves[0].Row == -1 && moves[0].Col == -1) {
				fmt.Println("No valid moves found")
				continue
			}
			move = moves[0]
			if *debug {
				fmt.Printf("Depth %d (%d move) ; Score %d ; Continuation %s\n", searchDepth, len(pos), score, utils.PositionsToAlgebraic(moves))
			}
		}

		fmt.Println(utils.PositionToAlgebraic(move))
	}
}
