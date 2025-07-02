package learning

import (
	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
)

// Trainer implements the genetic algorithm training functionality
type Trainer struct {
	Name           string
	Models         []EvaluationModel
	BaseModel      evaluation.EvaluationCoefficients
	BestModel      EvaluationModel
	Generation     int
	PopulationSize int
	MutationRate   float64
	NumGames       int
	MaxDepth       int8
}

// TrainerInterface defines the common interface for all trainers
type TrainerInterface interface {
	InitializePopulation()
	StartTraining(generations int)
	LoadModel(filename string) (EvaluationModel, error)
	SaveModel(filename string, model EvaluationModel) error
	SaveGenerationStats(gen int) error
}
