package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
	"yt2tidal/takeout"
	"yt2tidal/text"
	"yt2tidal/tidal"
)

type Config struct {
	Username     string
	Password     string
	PlaylistFile string
	IsDryrun     bool
}

func main() {
	fmt.Println("music library load")
	var usernameFlag = flag.String("username", "", "tidal username")
	var passwordFlag = flag.String("password", "", "tidal password")
	var playlistFileFlag = flag.String("playlist", "", "the path to the playlist file downloaded from extractor.js")
	var dryrunFlag = flag.Bool("dryrun", false, "pass to to parse takeout, but not submit Tidal requestes")
	flag.Parse()

	if *usernameFlag == "" || *passwordFlag == "" || *playlistFileFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	config := Config{
		Username:     *usernameFlag,
		Password:     *passwordFlag,
		PlaylistFile: *playlistFileFlag,
		IsDryrun:     *dryrunFlag,
	}

	process(config)
}

func process(config Config) {
	playlist, err := takeout.Parse(config.PlaylistFile)
	if err != nil {
		fmt.Println("failed to load playlist file: ", err.Error())
		os.Exit(-1)
	}

	t := tidal.New()
	if err := t.Login(config.Username, config.Password); err != nil {
		fmt.Printf("login failed %v", err)
		os.Exit(-1)
	}
	var songIDs []int
	begin := time.Now()

	// playlist = takeout.Playlist{Title: "Test", Songs: []takeout.Song{{Title: "Blue", Artist: "Serj Tankian", Album: "Elect the Dead (Deluxe)"}}}
	songIDs = buildSongList(t, playlist)
	// a, _ := t.SearchArtist("The Beach Boys")
	// fmt.Println("the beach boys: ")
	// for _, a := range a.Items {
	// 	fmt.Printf("\t%d %s\n", a.ID, a.Name)
	// }
	// al, _ := t.GetAlbumsForArtist(a.Items[0].ID, tidal.NoneFilter)
	// fmt.Println("the beach boys albums: ")
	// for _, a := range al.Items {
	// 	fmt.Printf("\t%d %s\n", a.ID, a.Title)
	// }

	fmt.Printf("building tracklist took: %s\n", time.Now().Sub(begin))
	if config.IsDryrun {
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
	for _, song := range playlist.Songs {
		logs := []string{fmt.Sprint("song: ", song.Title, " Album:", song.Album, " Artist:", song.Artist)}
		var artistSearch tidal.ArtistSearch
		artistSearch, err = t.SearchArtist(song.Artist)
		if err != nil {
			fmt.Printf("\tfailed to find playlist artist %s : %v\n", song.Artist, err)
			//TODO search by album?
			//https://listen.tidal.com/v1/search/top-hits?query=arctic%20monkeys%20Fluorescent%20Adolescent&limit=3&offset=0&types=ARTISTS,ALBUMS,TRACKS,VIDEOS,PLAYLISTS&includeContributors=true&countryCode=US
			continue
		}

		hits := []tidal.Track{}

		for _, artist := range artistSearch.Items {
			var albumSearch tidal.AlbumSearch
			albumSearch, err = t.GetAlbumsForArtist(artist.ID, tidal.NoneFilter)
			if err != nil {
				logs = append(logs, fmt.Sprintf("\tfailed to get albums for artist %s (%d) : %v", song.Artist, artist.ID, err))
				continue
			}
			if albumSearch.TotalNumberOfItems == 0 {
				albumSearch, err = t.GetAlbumsForArtist(artist.ID, tidal.CompilationsFilter)
				if err != nil {
					logs = append(logs, fmt.Sprintf("\tfailed to get albums for artist %s (%d) : %v", song.Artist, artist.ID, err))
					continue
				}
			}
			if albumSearch.TotalNumberOfItems == 0 {
				logs = append(logs, fmt.Sprintf("\tfailed to find ANY albums for (%d) %s", artist.ID, artist.Name))
			}

			// fmt.Printf("artist %s (%d) has %d albums\n", artist.Name, artist.ID, albumSearch.TotalNumberOfItems)
			for _, album := range albumSearch.Items {
				var tracksSearch tidal.Tracks
				tracksSearch, err = t.GetTracksForAlbum(album.ID)
				if err != nil {
					logs = append(logs, fmt.Sprintf("\tfailed to get tracks for album %s (%d) : %v", album.Title, album.ID, err))
					continue
				}
				// fmt.Printf("\t album: %s\n", album.Title)
				for _, track := range tracksSearch.Items {
					if text.Matches(song.Title, track.Title) {
						hits = append(hits, track)
					}
				}
			}
		}

		if len(hits) > 0 {
			track := text.Score(song, hits)
			songIDs = append(songIDs, track.ID)
			fmt.Println("have track " + track.Title)
		} else {
			logs = append(logs, fmt.Sprintf("\tfailed to find track among (%d) hits", len(hits)))
			fmt.Println(strings.Join(logs, "\n"))
			for _, h := range hits {
				fmt.Printf("\t\t%s from %s\n", h.Title, h.Album.Title)
			}
		}
	}
	return songIDs
}
