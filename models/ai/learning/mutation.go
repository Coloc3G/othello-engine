package learning

import (
	"math/rand"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
)

// ImprovedMutateArray applies mutations of varying magnitudes to an array of values
func ImprovedMutateArray(arr []int16, minVal, maxVal int) []int16 {
	newArr := make([]int16, len(arr))

	for i, val := range arr {
		// Copy original value by default
		newArr[i] = val

		// Completely reroll the value (rare) - helps with exploration
		if rand.Float64() < RerollRate {
			newArr[i] = int16(minVal + rand.Intn(maxVal-minVal+1))
			continue
		}

		// Apply small mutation (common)
		if rand.Float64() < SmallMutationRate {
			delta := rand.Intn(2*SmallDeltaMax+1) - SmallDeltaMax
			newArr[i] = int16(AdjustValueInRange(int(val)+delta, minVal, maxVal))
			continue
		}

		// Apply medium mutation (occasional)
		if rand.Float64() < MediumMutationRate {
			delta := rand.Intn(2*MediumDeltaMax+1) - MediumDeltaMax
			newArr[i] = int16(AdjustValueInRange(int(val)+delta, minVal, maxVal))
			continue
		}

		// Apply large mutation (rare)
		if rand.Float64() < LargeMutationRate {
			delta := rand.Intn(2*LargeDeltaMax+1) - LargeDeltaMax
			newArr[i] = int16(AdjustValueInRange(int(val)+delta, minVal, maxVal))
		}
	}

	return newArr
}

// AdjustValueInRange keeps a value within the specified range
func AdjustValueInRange(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// MutateCoefficients applies mutations to all coefficient arrays in an evaluation model
func MutateCoefficients(coeffs evaluation.EvaluationCoefficients) evaluation.EvaluationCoefficients {
	mutated := coeffs

	// Apply mutations to all coefficient arrays
	mutated.MaterialCoeffs = ImprovedMutateArray(coeffs.MaterialCoeffs, MaterialMin, MaterialMax)
	mutated.MobilityCoeffs = ImprovedMutateArray(coeffs.MobilityCoeffs, MobilityMin, MobilityMax)
	mutated.CornersCoeffs = ImprovedMutateArray(coeffs.CornersCoeffs, CornersMin, CornersMax)
	mutated.ParityCoeffs = ImprovedMutateArray(coeffs.ParityCoeffs, ParityMin, ParityMax)
	mutated.StabilityCoeffs = ImprovedMutateArray(coeffs.StabilityCoeffs, StabilityMin, StabilityMax)
	mutated.FrontierCoeffs = ImprovedMutateArray(coeffs.FrontierCoeffs, FrontierMin, FrontierMax)

	return mutated
}

// CreateDiverseModel creates a different but not wildly different model for initial population
func CreateDiverseModel(baseModel EvaluationModel) EvaluationModel {
	newModel := EvaluationModel{
		Coeffs: evaluation.EvaluationCoefficients{
			MaterialCoeffs:  make([]int16, 3),
			MobilityCoeffs:  make([]int16, 3),
			CornersCoeffs:   make([]int16, 3),
			ParityCoeffs:    make([]int16, 3),
			StabilityCoeffs: make([]int16, 3),
			FrontierCoeffs:  make([]int16, 3),
			Name:            "Gen1",
		},
	}
	newModel.Generation = baseModel.Generation + 1

	// Apply random scaling factors with more moderate ranges
	materialFactor := 0.8 + rand.Float64()*0.4 // 0.8x to 1.2x
	mobilityFactor := 0.8 + rand.Float64()*0.4
	cornersFactor := 0.8 + rand.Float64()*0.4
	parityFactor := 0.8 + rand.Float64()*0.4
	stabilityFactor := 0.8 + rand.Float64()*0.4
	frontierFactor := 0.8 + rand.Float64()*0.4

	// Apply factors to all coefficients with bounds checking
	for i := 0; i < 3; i++ {
		// Apply the scaling factors with sensible minimum values
		newModel.Coeffs.MaterialCoeffs[i] = int16(max(1, int(float64(baseModel.Coeffs.MaterialCoeffs[i])*materialFactor)))
		newModel.Coeffs.MobilityCoeffs[i] = int16(max(1, int(float64(baseModel.Coeffs.MobilityCoeffs[i])*mobilityFactor)))
		newModel.Coeffs.CornersCoeffs[i] = int16(max(1, int(float64(baseModel.Coeffs.CornersCoeffs[i])*cornersFactor)))
		newModel.Coeffs.ParityCoeffs[i] = int16(max(1, int(float64(baseModel.Coeffs.ParityCoeffs[i])*parityFactor)))
		newModel.Coeffs.StabilityCoeffs[i] = int16(max(1, int(float64(baseModel.Coeffs.StabilityCoeffs[i])*stabilityFactor)))
		newModel.Coeffs.FrontierCoeffs[i] = int16(max(1, int(float64(baseModel.Coeffs.FrontierCoeffs[i])*frontierFactor)))

		// Apply maximum caps to avoid extreme values
		newModel.Coeffs.MaterialCoeffs[i] = int16(min(int(newModel.Coeffs.MaterialCoeffs[i]), MaterialMax))
		newModel.Coeffs.MobilityCoeffs[i] = int16(min(int(newModel.Coeffs.MobilityCoeffs[i]), MobilityMax))
		newModel.Coeffs.CornersCoeffs[i] = int16(min(int(newModel.Coeffs.CornersCoeffs[i]), CornersMax))
		newModel.Coeffs.ParityCoeffs[i] = int16(min(int(newModel.Coeffs.ParityCoeffs[i]), ParityMax))
		newModel.Coeffs.StabilityCoeffs[i] = int16(min(int(newModel.Coeffs.StabilityCoeffs[i]), StabilityMax))
		newModel.Coeffs.FrontierCoeffs[i] = int16(min(int(newModel.Coeffs.FrontierCoeffs[i]), FrontierMax))
	}

	return newModel
}

// Helper function for CreateDiverseModel
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function for CreateDiverseModel
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
