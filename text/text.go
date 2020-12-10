package text

import (
	"fmt"
	"math"
	"strings"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

func Matches(song, track string) bool {
	songTitle := normalize(song)
	trackTitle := normalize(track)
	if trackTitle == songTitle || strings.HasPrefix(trackTitle, songTitle) {
		return true
	}
	distance := levenshtein.DistanceForStrings([]rune(songTitle), []rune(trackTitle), levenshtein.DefaultOptions)
	if float64(distance) < math.Min(float64(6), float64(len(songTitle)/2)) {
		return true
	}

	if float64(distance) < math.Min(float64(len(songTitle)), float64(len(trackTitle)))/2 {
		fmt.Printf("\tlevenshtein failure (%s, %s) %d\n", songTitle, trackTitle, distance)
	}

	return distance < 3
}

func normalize(input string) string {
	return strings.Trim(strings.ToLower(input), " ,()[]/.:'")
}
