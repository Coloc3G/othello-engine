package opening

import (
	"math/rand"
	"strings"
)

func MatchOpening(transcript string) Opening {
	for _, opening := range KNOWN_OPENINGS {
		if strings.HasPrefix(transcript, opening.Transcript) {
			return opening
		}
	}
	return Opening{}
}

func SelectRandomOpening() Opening {
	return KNOWN_OPENINGS[rand.Intn(len(KNOWN_OPENINGS))]
}
