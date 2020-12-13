package text

import (
	"fmt"
	"math"
	"strings"
	"yt2tidal/takeout"
	"yt2tidal/tidal"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

func Matches(song, track string) bool {
	distance := scoreStrings(song, track)
	if distance == 0 || float64(distance) < math.Min(float64(6), float64(len(song)/2)) {
		return true
	}

	if float64(distance) < math.Min(float64(len(song)), float64(len(track)))/2 {
		//only print when it's not obviously wrong song
		fmt.Printf("\tlevenshtein failure (%s, %s) %d\n", song, track, distance)
	}

	return false
}

func scoreStrings(one, two string) int {
	oneNorm := normalize(one)
	twoNorm := normalize(two)
	if twoNorm == oneNorm {
		return 0
	}
	if strings.HasPrefix(twoNorm, oneNorm) || strings.HasPrefix(oneNorm, twoNorm) {
		return 1
	}
	return levenshtein.DistanceForStrings([]rune(oneNorm), []rune(twoNorm), levenshtein.DefaultOptions)
}

func normalize(input string) string {
	return strings.Trim(strings.ToLower(input), " ,()[]/.:'")
}

func Score(song takeout.Song, hits []tidal.Track) tidal.Track {
	var bestTrack tidal.Track
	score := math.MaxInt16
	songCombo := song.Album + song.Title

	for _, hit := range hits {
		// albumScore := scoreStrings(song.Album, hit.Album.Title)
		// trackScore := scoreStrings(song.Title, hit.Title)
		comboScore := scoreStrings(songCombo, hit.Album.Title+hit.Title)
		if comboScore < score { //albumScore+trackScore < score ||
			bestTrack = hit
			score = comboScore //albumScore + trackScore
		}
	}
	// fmt.Printf("best score (%d) for %s %s", score, song.Title, song.Artist)
	return bestTrack
}
