package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

func Solve(g game.Game, player game.Player, depth int, eval Evaluation) game.Position {
	bestScore := -1 << 31
	var bestMove game.Position
	for _, move := range game.ValidMoves(g.Board, player.Color) {
		newNode, _ := game.GetNewBoardAfterMove(g.Board, move, player)
		childScore := MMAB(g, newNode, player, depth-1, false, -1<<31, 1<<31-1, eval)
		if childScore > bestScore {
			bestScore = childScore
			bestMove = move
		}
	}
	return bestMove
}

func MMAB(g game.Game, node game.Board, player game.Player, depth int, max bool, alpha, beta int, eval Evaluation) int {
	if depth == 0 || game.IsGameFinished(node) {
		return eval.Evaluate(g, node, player)
	}
	oplayer := game.GetOtherPlayer(g.Players, player.Color)
	if (max && !game.HasAnyMoves(node, player.Color)) || (!max && !game.HasAnyMoves(node, oplayer.Color)) {
		return MMAB(g, node, player, depth-1, !max, alpha, beta, eval)
	}
	var score int
	if max {
		score = -1 << 31
		for _, move := range game.ValidMoves(node, player.Color) {
			newNode, _ := game.GetNewBoardAfterMove(node, move, player)
			childScore := MMAB(g, newNode, player, depth-1, false, alpha, beta, eval)
			if childScore > score {
				score = childScore
			}
			if score > alpha {
				alpha = score
			}
			if beta <= alpha {
				break
			}
		}
	} else {
		score = 1<<31 - 1
		for _, move := range game.ValidMoves(node, oplayer.Color) {
			newNode, _ := game.GetNewBoardAfterMove(node, move, oplayer)
			childScore := MMAB(g, newNode, player, depth-1, true, alpha, beta, eval)
			if childScore < score {
				score = childScore
			}
			if score < beta {
				beta = score
			}
			if beta <= alpha {
				break
			}
		}
	}
	return score
}
