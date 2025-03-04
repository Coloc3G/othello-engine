package evaluation

// MaterialEvaluation is an evaluation function that scores a board based on the number of pieces difference between the players
type MaterialEvaluation struct {
}

// Evaluate the given board state and return a score
func (e *MaterialEvaluation) Evaluate(board [8][8]int) int {
	sum := 0
	for _, row := range board {
		for _, piece := range row {
			sum += piece
		}
	}
	return sum
}
