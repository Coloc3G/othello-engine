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
	BlackGames map[string]string                 `json:"black_game"`
	WhiteGames map[string]string                 `json:"white_game"`
}
