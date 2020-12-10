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
	songIDs = buildSongList(t, playlist)
	// songIDs = buildSongListFromAlbums(t, playlist)

	if *dryrunFlag {
		fmt.Printf("dry run complete\n\tfound %d songs of %d for playlist %s\n", len(songIDs), len(playlist.Songs), playlist.Title)
		return
	}

	pl, err := t.CreatePlaylist(playlist.Title)
	if err != nil {
		fmt.Println("failed to create playlist ", err)
		os.Exit(-2)
	}

	err = t.AddSongToPlaylist(pl.UUID, songIDs...)
	if err != nil {
		fmt.Printf("failed to add songs %d: %s\n", len(songIDs), err)
	} else {
		fmt.Printf("created playlist %s with %d of %d songs\n\n", playlist.Title, len(songIDs), len(playlist.Songs))
	}
}

func buildSongList(t *tidal.Tidal, playlist takeout.Playlist) []int {
	var songIDs []int
	var err error
	artistsMap := make(map[string]tidal.ArtistSearch)
	albumsMap := make(map[int]tidal.AlbumSearch)
	tracksMap := make(map[int]tidal.Tracks)
	for _, song := range playlist.Songs {
		var haveSong bool
		var artistSearch tidal.ArtistSearch
		artistSearch, ok := artistsMap[song.Artist]
		if !ok {
			artistSearch, err = t.SearchArtist(song.Artist)
			if err != nil {
				fmt.Printf("failed to find playlist artist %s : %v\n", song.Artist, err)
				continue
			}
			artistsMap[song.Artist] = artistSearch
		}

		for _, artist := range artistSearch.Items {
			var albumSearch tidal.AlbumSearch
			albumSearch, ok = albumsMap[artist.ID]
			if !ok {
				albumSearch, err = t.GetAlbumsForArtist(artist.ID, tidal.NoneFilter)
				if err != nil {
					fmt.Printf("failed to get albums for artist %s (%d) : %v\n", song.Artist, artist.ID, err)
					continue
				}
				if albumSearch.TotalNumberOfItems == 0 {
					albumSearch, err = t.GetAlbumsForArtist(artist.ID, tidal.CompilationsFilter)
					if err != nil {
						fmt.Printf("failed to get albums for artist %s (%d) : %v\n", song.Artist, artist.ID, err)
						continue
					}
				}
				albumsMap[artist.ID] = albumSearch
			}

			// fmt.Printf("artist %s (%d) has %d albums\n", artist.Name, artist.ID, albumSearch.TotalNumberOfItems)
			for _, album := range albumSearch.Items {
				var tracksSearch tidal.Tracks
				tracksSearch, ok = tracksMap[album.ID]
				if !ok {
					tracksSearch, err = t.GetTracksForAlbum(album.ID)
					if err != nil {
						fmt.Printf("failed to get tracks for album %s (%d) : %v\n", album.Title, album.ID, err)
						continue
					}
					tracksMap[album.ID] = tracksSearch
				}
				for _, track := range tracksSearch.Items {
					if strings.ToLower(track.Title) == strings.ToLower(song.Title) {
						songIDs = append(songIDs, track.ID)
						fmt.Println("have track " + track.Title)
						haveSong = true
						break
					}
				}
				if haveSong {
					break
				} else {
					// fmt.Printf("song %s not on %s\n", song.Title, album.Title)
				}
			}
			if haveSong {
				break
			}
		}

		if !haveSong {
			fmt.Printf("failed to find Artist:%s Album:%s Song:%s\n", song.Artist, song.Album, song.Title)
		}
	}
	return songIDs
}

func buildSongListFromAlbums(t *tidal.Tidal, playlist takeout.Playlist) []int {
	var songIDs []int
	var err error

	albumsMap := make(map[string]tidal.AlbumSearch)
	tracksMap := make(map[int]tidal.Tracks)
	for _, song := range playlist.Songs {
		var haveSong bool
		var albumSearch tidal.AlbumSearch
		albumSearch, ok := albumsMap[song.Album]
		if !ok {
			albumSearch, err = t.SearchAlbum(song.Album)
			if err != nil {
				fmt.Printf("failed to search albums %s : %v\n", song.Album, err)
				continue
			}
			albumsMap[song.Album] = albumSearch
		}

		// fmt.Printf("artist %s (%d) has %d albums\n", artist.Name, artist.ID, albumSearch.TotalNumberOfItems)
		for _, album := range albumSearch.Items {
			var tracksSearch tidal.Tracks
			tracksSearch, ok = tracksMap[album.ID]
			if !ok {
				tracksSearch, err = t.GetTracksForAlbum(album.ID)
				if err != nil {
					fmt.Printf("failed to get tracks for album %s (%d) : %v\n", album.Title, album.ID, err)
					continue
				}
				tracksMap[album.ID] = tracksSearch
			}
			for _, track := range tracksSearch.Items {
				if strings.ToLower(track.Title) == strings.ToLower(song.Title) {
					songIDs = append(songIDs, track.ID)
					fmt.Println("have track " + track.Title)
					haveSong = true
					break
				}
			}
			if haveSong {
				break
			} else {
				// fmt.Printf("song %s not on %s\n", song.Title, album.Title)
			}
		}
		if !haveSong {
			fmt.Printf("failed to find Artist:%s Album:%s Song:%s\n", song.Artist, song.Album, song.Title)
		}
	}
	return songIDs
}
