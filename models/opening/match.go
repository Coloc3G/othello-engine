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
