package evaluation

type Evaluation interface {
	// Evaluate the given board state and return a score
	Evaluate(board [8][8]int) int
}
