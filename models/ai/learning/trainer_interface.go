package learning

import (
	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
)

// EvaluationModel represents an evaluation model with customizable coefficients
type EvaluationModel struct {
	Coeffs     evaluation.EvaluationCoefficients `json:"coefficients"`
	Wins       int                               `json:"wins"`
	Losses     int                               `json:"losses"`
	Draws      int                               `json:"draws"`
	Fitness    float64                           `json:"fitness"`
	Generation int                               `json:"generation"`
}

// TrainerInterface defines the common interface for all trainers
type TrainerInterface interface {
	InitializePopulation()
	StartTraining(generations int)
	TournamentTraining(generations int)
	LoadModel(filename string) (EvaluationModel, error)
	SaveModel(filename string, model EvaluationModel) error
	SaveGenerationStats(gen int) error
}

// BaseTrainerInterface defines common operations for all trainers
type BaseTrainerInterface interface {
	InitializePopulation()
	StartTraining(int)
	TournamentTraining(int)
	LoadModel(string) (EvaluationModel, error)
}
