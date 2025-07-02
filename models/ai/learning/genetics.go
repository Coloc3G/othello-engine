package learning

import (
	"fmt"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
)

// tournamentSelect selects a model using tournament selection
func (t *Trainer) tournamentSelect(tournamentSize int) EvaluationModel {
	best := t.Models[0]
	bestFitness := t.Models[0].Fitness

	for i := 1; i < tournamentSize; i++ {
		idx := i % len(t.Models) // Ensure we don't go out of bounds
		if t.Models[idx].Fitness > bestFitness {
			best = t.Models[idx]
			bestFitness = t.Models[idx].Fitness
		}
	}

	return best
}

// crossover combines two models to create a child model
func (t *Trainer) crossover(parent1, parent2 EvaluationModel) EvaluationModel {
	child := EvaluationModel{
		Coeffs: evaluation.EvaluationCoefficients{
			MaterialCoeffs:  make([]int16, 3),
			MobilityCoeffs:  make([]int16, 3),
			CornersCoeffs:   make([]int16, 3),
			ParityCoeffs:    make([]int16, 3),
			StabilityCoeffs: make([]int16, 3),
			FrontierCoeffs:  make([]int16, 3),
		},
	}

	// Create crossover patterns that determine which parent each coefficient comes from
	materialPattern := []bool{true, false, true}
	mobilityPattern := []bool{false, true, false}
	cornersPattern := []bool{true, false, true}
	parityPattern := []bool{false, true, false}
	stabilityPattern := []bool{true, true, false}
	frontierPattern := []bool{false, false, true}

	// Apply crossover patterns
	child.Coeffs.MaterialCoeffs = crossoverCoefficients(
		parent1.Coeffs.MaterialCoeffs, parent2.Coeffs.MaterialCoeffs, materialPattern)
	child.Coeffs.MobilityCoeffs = crossoverCoefficients(
		parent1.Coeffs.MobilityCoeffs, parent2.Coeffs.MobilityCoeffs, mobilityPattern)
	child.Coeffs.CornersCoeffs = crossoverCoefficients(
		parent1.Coeffs.CornersCoeffs, parent2.Coeffs.CornersCoeffs, cornersPattern)
	child.Coeffs.ParityCoeffs = crossoverCoefficients(
		parent1.Coeffs.ParityCoeffs, parent2.Coeffs.ParityCoeffs, parityPattern)
	child.Coeffs.StabilityCoeffs = crossoverCoefficients(
		parent1.Coeffs.StabilityCoeffs, parent2.Coeffs.StabilityCoeffs, stabilityPattern)
	child.Coeffs.FrontierCoeffs = crossoverCoefficients(
		parent1.Coeffs.FrontierCoeffs, parent2.Coeffs.FrontierCoeffs, frontierPattern)

	return child
}

// mutateModel applies random mutations to a model
func (t *Trainer) mutateModel(model EvaluationModel) EvaluationModel {
	mutated := model

	// Use the mutation package for mutation
	mutated.Coeffs = MutateCoefficients(model.Coeffs)

	// Give the mutated model a name for tracking
	if mutated.Coeffs.Name == "" {
		mutated.Coeffs.Name = fmt.Sprintf("Model_Gen%d", t.Generation)
	}

	return mutated
}

// crossoverCoefficients performs crossover on a specific coefficient array
func crossoverCoefficients(parent1, parent2 []int16, pattern []bool) []int16 {
	result := make([]int16, len(parent1))
	for i := range parent1 {
		if pattern[i%len(pattern)] {
			result[i] = parent1[i]
		} else {
			result[i] = parent2[i]
		}
	}
	return result
}
