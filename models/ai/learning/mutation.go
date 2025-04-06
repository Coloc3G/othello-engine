package learning

import (
	"math/rand"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
)

const (
	MaterialMin       = 1
	MaterialMax       = 1000
	MobilityMin       = 1
	MobilityMax       = 500
	CornersMin        = 1
	CornersMax        = 2000
	ParityMin         = 1
	ParityMax         = 1000
	StabilityMin      = 1
	StabilityMax      = 300
	FrontierMin       = 1
	FrontierMax       = 200
	mutationIntensity = 0.2 // standard deviation fraction of (max-min)
	bigMutationChance = 0.1 // chance for a full re-roll mutation
)

func adaptiveMutateArray(arr []int, minVal, maxVal int, rate float64) []int {
	newArr := make([]int, len(arr))
	mutationOccurred := false
	for i, val := range arr {
		if rand.Float64() < rate {
			if rand.Float64() < bigMutationChance {
				newArr[i] = minVal + rand.Intn(maxVal-minVal+1)
			} else {
				rangeVal := float64(maxVal - minVal)
				noise := rand.NormFloat64() * rangeVal * mutationIntensity
				newVal := val + int(noise)
				if newVal < minVal {
					newVal = minVal
				} else if newVal > maxVal {
					newVal = maxVal
				}
				newArr[i] = newVal
			}
			mutationOccurred = true
		} else {
			newArr[i] = val
		}
	}
	// If no mutation occurred, force a random change on one element.
	if !mutationOccurred && len(arr) > 0 {
		idx := rand.Intn(len(arr))
		newArr[idx] = minVal + rand.Intn(maxVal-minVal+1)
	}
	return newArr
}

// AdaptiveMutation applies adaptive mutation to all coefficient arrays.
func AdaptiveMutation(coeffs evaluation.EvaluationCoefficients, rate float64) evaluation.EvaluationCoefficients {
	coeffs.MaterialCoeffs = adaptiveMutateArray(coeffs.MaterialCoeffs, MaterialMin, MaterialMax, rate)
	coeffs.MobilityCoeffs = adaptiveMutateArray(coeffs.MobilityCoeffs, MobilityMin, MobilityMax, rate)
	coeffs.CornersCoeffs = adaptiveMutateArray(coeffs.CornersCoeffs, CornersMin, CornersMax, rate)
	coeffs.ParityCoeffs = adaptiveMutateArray(coeffs.ParityCoeffs, ParityMin, ParityMax, rate)
	coeffs.StabilityCoeffs = adaptiveMutateArray(coeffs.StabilityCoeffs, StabilityMin, StabilityMax, rate)
	coeffs.FrontierCoeffs = adaptiveMutateArray(coeffs.FrontierCoeffs, FrontierMin, FrontierMax, rate)
	return coeffs
}
