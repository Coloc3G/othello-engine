package evaluation

var (
	V1Coeff = EvaluationCoefficients{
		Name:            "V1",
		MaterialCoeffs:  []int{0, 10, 500},
		MobilityCoeffs:  []int{50, 20, 100},
		CornersCoeffs:   []int{1000, 1000, 1000},
		ParityCoeffs:    []int{0, 100, 500},
		StabilityCoeffs: []int{0, 0, 0},
		FrontierCoeffs:  []int{0, 0, 0},
	}

	V2Coeff = EvaluationCoefficients{
		Name:            "V2",
		MaterialCoeffs:  []int{51, 242, 440},
		MobilityCoeffs:  []int{69, 177, 167},
		CornersCoeffs:   []int{1123, 759, 467},
		ParityCoeffs:    []int{100, 3, 964},
		StabilityCoeffs: []int{0, 21, 85},
		FrontierCoeffs:  []int{0, 86, 0},
	}

	V3Coeff = EvaluationCoefficients{
		Name:            "V3",
		MaterialCoeffs:  []int{0, 10, 1000},
		MobilityCoeffs:  []int{50, 250, 500},
		CornersCoeffs:   []int{1000, 1000, 1000},
		ParityCoeffs:    []int{0, 100, 500},
		StabilityCoeffs: []int{0, 100, 200},
		FrontierCoeffs:  []int{0, 100, 200},
	}

	V4Coeff = EvaluationCoefficients{
		Name:            "V4",
		MaterialCoeffs:  []int{1, 9, 114},
		MobilityCoeffs:  []int{57, 195, 390},
		CornersCoeffs:   []int{1000, 1000, 1000},
		ParityCoeffs:    []int{67, 287, 473},
		StabilityCoeffs: []int{28, 94, 270},
		FrontierCoeffs:  []int{70, 81, 376},
	}

	Models []EvaluationCoefficients = []EvaluationCoefficients{
		V1Coeff,
		V2Coeff,
		V3Coeff,
		V4Coeff,
	}
)

func GetCoefficientsByName(name string) (EvaluationCoefficients, bool) {
	for _, coeff := range Models {
		if coeff.Name == name {
			return coeff, true
		}
	}
	return EvaluationCoefficients{}, false
}
