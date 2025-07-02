package opening

import (
	"math/rand"
	"strings"
)

func MatchOpening(transcript string) []Opening {
	matches := make([]Opening, 0)
	for _, opening := range KNOWN_OPENINGS {
		if strings.HasPrefix(opening.Transcript, transcript) {
			matches = append(matches, opening)
		}
	}
	return matches
}

func SelectRandomOpening() Opening {
	return KNOWN_OPENINGS[rand.Intn(len(KNOWN_OPENINGS))]
}

func SelectRandomOpenings(numGames int) []Opening {
	openingCount := len(KNOWN_OPENINGS)
	if numGames > openingCount {
		numGames = openingCount
	}

	shuffled := make([]Opening, len(KNOWN_OPENINGS))
	copy(shuffled, KNOWN_OPENINGS)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:numGames]
}
