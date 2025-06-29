package learning

import (
	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
)

// EvaluationModel represents a model for othello evaluation
type EvaluationModel struct {
	Coeffs     evaluation.EvaluationCoefficients `json:"coeffs"`
	Generation int                               `json:"generation"`
	Fitness    float64                           `json:"fitness"`
	Wins       int                               `json:"wins"`
	Losses     int                               `json:"losses"`
	Draws      int                               `json:"draws"`
}

// Function to get average wins across models
func calculateAverageWins(models []EvaluationModel) float64 {
	if len(models) == 0 {
		return 0
	}

	total := 0
	for _, model := range models {
		total += model.Wins
	}

	return float64(total) / float64(len(models))
}
