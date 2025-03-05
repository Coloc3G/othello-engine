package opening

import (
	"fmt"
	"math/rand"
	"strings"
)

func MatchOpening(transcript string) []Opening {
	fmt.Println("Matching opening for transcript: ", transcript)
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
