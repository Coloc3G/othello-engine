package evaluation

var (
	V1Coeff = EvaluationCoefficients{
		MaterialCoeffs:  []int{0, 10, 500},
		MobilityCoeffs:  []int{50, 20, 100},
		CornersCoeffs:   []int{1000, 1000, 1000},
		ParityCoeffs:    []int{0, 100, 500},
		StabilityCoeffs: []int{0, 0, 0},
		FrontierCoeffs:  []int{0, 0, 0},
	}

	V2Coeff = EvaluationCoefficients{
		MaterialCoeffs:  []int{30, 118, 466},
		MobilityCoeffs:  []int{13, 94, 158},
		CornersCoeffs:   []int{1201, 805, 865},
		ParityCoeffs:    []int{0, 28, 754},
		StabilityCoeffs: []int{0, 0, 0},
		FrontierCoeffs:  []int{0, 0, 0},
	}
)
