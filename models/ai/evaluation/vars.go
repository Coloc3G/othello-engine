package evaluation

const (
	MAX_EVAL int16 = 20200
	MIN_EVAL int16 = -20200
)

var (
	V1Coeff = EvaluationCoefficients{
		Name:            "V1",
		MaterialCoeffs:  []int16{0, 1, 50},
		MobilityCoeffs:  []int16{0, 2, 10},
		CornersCoeffs:   []int16{100, 100, 100},
		ParityCoeffs:    []int16{0, 10, 50},
		StabilityCoeffs: []int16{0, 0, 0},
		FrontierCoeffs:  []int16{0, 0, 0},
	}

	V2Coeff = EvaluationCoefficients{
		Name:            "V2",
		MaterialCoeffs:  []int16{5, 24, 44},
		MobilityCoeffs:  []int16{7, 18, 17},
		CornersCoeffs:   []int16{112, 76, 47},
		ParityCoeffs:    []int16{10, 0, 97},
		StabilityCoeffs: []int16{0, 2, 8},
		FrontierCoeffs:  []int16{0, 9, 0},
	}

	V3Coeff = EvaluationCoefficients{
		Name:            "V3",
		MaterialCoeffs:  []int16{0, 1, 100},
		MobilityCoeffs:  []int16{5, 25, 50},
		CornersCoeffs:   []int16{0, 10, 50},
		StabilityCoeffs: []int16{0, 10, 20},
		FrontierCoeffs:  []int16{0, 10, 20},
	}

	V4Coeff = EvaluationCoefficients{
		Name:            "V4",
		MaterialCoeffs:  []int16{0, 1, 11},
		MobilityCoeffs:  []int16{6, 20, 39},
		CornersCoeffs:   []int16{100, 100, 100},
		ParityCoeffs:    []int16{7, 29, 47},
		StabilityCoeffs: []int16{3, 9, 27},
		FrontierCoeffs:  []int16{7, 8, 38},
	}

	V5Coeff = EvaluationCoefficients{
		Name:            "V5",
		MaterialCoeffs:  []int16{1, 1, 11},
		MobilityCoeffs:  []int16{6, 11, 39},
		CornersCoeffs:   []int16{100, 100, 100},
		ParityCoeffs:    []int16{96, 1, 44},
		StabilityCoeffs: []int16{55, 9, 19},
		FrontierCoeffs:  []int16{86, 2, 38},
	}

	Models []EvaluationCoefficients = []EvaluationCoefficients{
		V1Coeff,
		V2Coeff,
		V3Coeff,
		V4Coeff,
		V5Coeff,
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
