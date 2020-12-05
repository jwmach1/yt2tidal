package takeout

import (
	"encoding/json"
	"os"
)

type Playlist struct {
	Title string `json:"title"`
	Songs []Song `json:"songs"`
}
type Song struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
}

func Parse(path string) (Playlist, error) {
	f, err := os.Open(path)
	if err != nil {
		return Playlist{}, err
	}
	playlist := Playlist{}
	err = json.NewDecoder(f).Decode(&playlist)

	return playlist, err
}
