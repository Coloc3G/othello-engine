package learning

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
