package learning

import (
	"fmt"
	"math/rand"

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

// mutateCoefficient applies random mutation to a single coefficient
func mutateCoefficient(value, min, max int) int {
	// Apply small mutation most of the time
	if rand.Float64() < SmallMutationRate {
		delta := rand.Intn(SmallDeltaMax*2+1) - SmallDeltaMax
		value += delta
	}

	// Apply medium mutation occasionally
	if rand.Float64() < MediumMutationRate {
		delta := rand.Intn(MediumDeltaMax*2+1) - MediumDeltaMax
		value += delta
	}

	// Apply large mutation rarely
	if rand.Float64() < LargeMutationRate {
		delta := rand.Intn(LargeDeltaMax*2+1) - LargeDeltaMax
		value += delta
	}

	// Completely reroll value with very low probability
	if rand.Float64() < RerollRate {
		value = min + rand.Intn(max-min+1)
	}

	// Ensure value is within bounds
	if value < min {
		value = min
	}
	if value > max {
		value = max
	}

	return value
}

// Generate a random model for evaluation
func GenerateRandomModel() EvaluationModel {
	model := EvaluationModel{
		Coeffs: evaluation.EvaluationCoefficients{
			Name:            fmt.Sprintf("Random_Gen1_%d", rand.Intn(1000)),
			MaterialCoeffs:  make([]int, 3),
			MobilityCoeffs:  make([]int, 3),
			CornersCoeffs:   make([]int, 3),
			ParityCoeffs:    make([]int, 3),
			StabilityCoeffs: make([]int, 3),
			FrontierCoeffs:  make([]int, 3),
		},
		Generation: 1,
	}

	// Generate random coefficients within bounds
	for i := 0; i < 3; i++ {
		model.Coeffs.MaterialCoeffs[i] = MaterialMin + rand.Intn(MaterialMax-MaterialMin+1)
		model.Coeffs.MobilityCoeffs[i] = MobilityMin + rand.Intn(MobilityMax-MobilityMin+1)
		model.Coeffs.CornersCoeffs[i] = CornersMin + rand.Intn(CornersMax-CornersMin+1)
		model.Coeffs.ParityCoeffs[i] = ParityMin + rand.Intn(ParityMax-ParityMin+1)
		model.Coeffs.StabilityCoeffs[i] = StabilityMin + rand.Intn(StabilityMax-StabilityMin+1)
		model.Coeffs.FrontierCoeffs[i] = FrontierMin + rand.Intn(FrontierMax-FrontierMin+1)
	}

	return model
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
