package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"yt2tidal/takeout"
	"yt2tidal/tidal"
)

var usernameFlag = flag.String("username", "", "tidal username")
var passwordFlag = flag.String("password", "", "tidal password")
var playlistFileFlag = flag.String("playlist", "", "the path to the playlist file downloaded from extractor.js")
var dryrunFlag = flag.Bool("dryrun", false, "pass to to parse takeout, but not submit Tidal requestes")

func main() {
	fmt.Println("music library load")
	flag.Parse()

	if *usernameFlag == "" || *passwordFlag == "" || *playlistFileFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	playlist, err := takeout.Parse(*playlistFileFlag)
	if err != nil {
		fmt.Println("failed to load playlist file: ", err.Error())
		os.Exit(-1)
	}

	t := tidal.New()
	if err := t.Login(*usernameFlag, *passwordFlag); err != nil {
		fmt.Printf("login failed %v", err)
		os.Exit(-1)
	}

	var songIDs []int
	artistsMap := make(map[string]tidal.Artist)
	albumsMap := make(map[int]tidal.AlbumSearch)
	tracksmap := make(map[int]tidal.Tracks)
	for _, song := range playlist.Songs {
		artistFromMap, ok := artistsMap[song.Artist]
		if !ok {
			artistSearch, err := t.SearchArtist(song.Artist)
			if err != nil {
				fmt.Printf("failed to find playlist artist %s : %v\n", song.Artist, err)
				continue
			}
			if len(artistSearch.Items) != 1 {
				fmt.Printf("found not only one artist %s but %d\n", song.Artist, len(artistSearch.Items))
				continue
			}
			artistFromMap = artistSearch.Items[0]
			artistsMap[song.Artist] = artistFromMap
		}
		albumsFromMap, ok := albumsMap[artistFromMap.ID]
		if !ok {
			albumSearch, err := t.GetAlbumsForArtist(artistFromMap.ID)
			if err != nil {
				fmt.Printf("failed to get albums for artist %s (%d) : %v\n", song.Artist, artistFromMap.ID, err)
				continue
			}
			albumsMap[artistFromMap.ID] = albumSearch
			albumsFromMap = albumSearch
		}
		var thisSongsAlbum tidal.Album
		for _, album := range albumsFromMap.Items {
			if album.Title == song.Album {
				thisSongsAlbum = album
				break
			}
		}
		if thisSongsAlbum.ID == 0 {
			titles := make([]string, len(albumsFromMap.Items))
			for i, t := range albumsFromMap.Items {
				titles[i] = t.Title
			}
			fmt.Printf("failed to find album for playlist song (%s) among %s\n", song.Title, strings.Join(titles, ","))
			continue
		}

		tracksFromMap, ok := tracksmap[thisSongsAlbum.ID]
		if !ok {
			fmt.Println("getting tracks for album ", thisSongsAlbum.Title)
			tracksSearch, err := t.GetTracksForAlbum(thisSongsAlbum.ID)
			if err != nil {
				fmt.Printf("failed to get tracks for album %s (%d) : %v\n", thisSongsAlbum.Title, thisSongsAlbum.ID, err)
				continue
			}
			tracksFromMap = tracksSearch
			tracksmap[thisSongsAlbum.ID] = tracksFromMap
		}
		var found bool
		for _, track := range tracksFromMap.Items {
			if strings.ToLower(track.Title) == strings.ToLower(song.Title) {
				songIDs = append(songIDs, track.ID)
				found = true
				break
			}
		}
		if !found {
			titles := make([]string, len(tracksFromMap.Items))
			for i, track := range tracksFromMap.Items {
				titles[i] = track.Title
			}
			fmt.Printf("failed to find %s among %s\n", song.Title, strings.Join(titles, ","))
		}
	}

	if *dryrunFlag {
		fmt.Printf("dry run complete\n\tfound %d songs of %d in playlist\n", len(songIDs), len(playlist.Songs))
		return
	}

	pl, err := t.CreatePlaylist(playlist.Title)
	fmt.Println("create playlist error? ", err)

	err = t.AddSongToPlaylist(pl.UUID, songIDs...)
	fmt.Println("add song error? ", err)
}
