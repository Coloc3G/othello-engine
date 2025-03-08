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
		MaterialCoeffs:  []int{51, 242, 440},
		MobilityCoeffs:  []int{69, 177, 167},
		CornersCoeffs:   []int{1123, 759, 467},
		ParityCoeffs:    []int{100, 3, 964},
		StabilityCoeffs: []int{0, 21, 85},
		FrontierCoeffs:  []int{0, 86, 0},
	}
)
