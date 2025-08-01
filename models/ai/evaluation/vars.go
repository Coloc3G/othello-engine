package evaluation

const (
	MAX_EVAL int16 = 20200
	MIN_EVAL int16 = -20200
)

var (
	V1Coeff = EvaluationCoefficients{
		Name:            "V1",
		MaterialCoeffs:  []int16{0, 0, 1, 1, 50, 50},
		MobilityCoeffs:  []int16{0, 0, 2, 2, 10, 10},
		CornersCoeffs:   []int16{100, 100, 100, 100, 100, 100},
		ParityCoeffs:    []int16{0, 0, 10, 10, 50, 50},
		StabilityCoeffs: []int16{0, 0, 0, 0, 0, 0},
		FrontierCoeffs:  []int16{0, 0, 0, 0, 0, 0},
	}

	V2Coeff = EvaluationCoefficients{
		Name:            "V2",
		MaterialCoeffs:  []int16{5, 5, 24, 24, 44, 44},
		MobilityCoeffs:  []int16{7, 7, 18, 18, 17, 17},
		CornersCoeffs:   []int16{112, 112, 76, 76, 47, 47},
		ParityCoeffs:    []int16{10, 10, 0, 0, 97, 97},
		StabilityCoeffs: []int16{0, 0, 2, 2, 8, 8},
		FrontierCoeffs:  []int16{0, 0, 9, 9, 0, 0},
	}

	V3Coeff = EvaluationCoefficients{
		Name:            "V3",
		MaterialCoeffs:  []int16{0, 0, 1, 1, 100, 100},
		MobilityCoeffs:  []int16{5, 5, 25, 25, 50, 50},
		CornersCoeffs:   []int16{100, 100, 100, 100, 100, 100},
		ParityCoeffs:    []int16{0, 0, 10, 10, 50, 50},
		StabilityCoeffs: []int16{0, 0, 10, 10, 20, 20},
		FrontierCoeffs:  []int16{0, 0, 10, 10, 20, 20},
	}

	V4Coeff = EvaluationCoefficients{
		Name:            "V4",
		MaterialCoeffs:  []int16{0, 0, 1, 1, 11, 11},
		MobilityCoeffs:  []int16{6, 6, 20, 20, 39, 39},
		CornersCoeffs:   []int16{100, 100, 100, 100, 100, 100},
		ParityCoeffs:    []int16{7, 7, 29, 29, 47, 47},
		StabilityCoeffs: []int16{3, 3, 9, 9, 27, 27},
		FrontierCoeffs:  []int16{7, 7, 8, 8, 38, 38},
	}

	V5Coeff = EvaluationCoefficients{
		Name:            "V5",
		MaterialCoeffs:  []int16{1, 1, 1, 1, 13, 13},
		MobilityCoeffs:  []int16{6, 6, 1, 1, 78, 78},
		CornersCoeffs:   []int16{66, 66, 81, 81, 100, 100},
		ParityCoeffs:    []int16{29, 29, 1, 1, 1, 1},
		StabilityCoeffs: []int16{1, 1, 9, 9, 1, 1},
		FrontierCoeffs:  []int16{58, 58, 11, 11, 23, 23},
	}

	V6Coeff = EvaluationCoefficients{
		Name:            "V6",
		MaterialCoeffs:  []int16{2, 2, 1, 1, 12, 12},
		MobilityCoeffs:  []int16{21, 21, 5, 5, 79, 79},
		CornersCoeffs:   []int16{89, 89, 100, 100, 82, 82},
		ParityCoeffs:    []int16{45, 45, 9, 9, 2, 2},
		StabilityCoeffs: []int16{20, 20, 7, 7, 1, 1},
		FrontierCoeffs:  []int16{67, 67, 12, 12, 11, 11},
	}

	V7Coeff = EvaluationCoefficients{
		Name:            "V7",
		MaterialCoeffs:  []int16{1, 1, 1, 1, 10, 14},
		MobilityCoeffs:  []int16{18, 33, 5, 5, 65, 68},
		CornersCoeffs:   []int16{87, 61, 100, 100, 97, 92},
		ParityCoeffs:    []int16{36, 39, 9, 9, 2, 21},
		StabilityCoeffs: []int16{23, 23, 4, 7, 1, 1},
		FrontierCoeffs:  []int16{54, 66, 14, 13, 11, 12},
	}

	Models []EvaluationCoefficients = []EvaluationCoefficients{
		V1Coeff,
		V2Coeff,
		V3Coeff,
		V4Coeff,
		V5Coeff,
		V6Coeff,
		V7Coeff,
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
