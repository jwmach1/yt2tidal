package main

/*
start with setting up spotify developer account: https://developer.spotify.com/documentation/web-api/quick-start/
 * https://developer.spotify.com/dashboard/applications

*/

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"spotify_load/tidal"
	"strconv"
)

var usernameFlag = flag.String("username", "", "tidal username")
var passwordFlag = flag.String("password", "", "tidal password")

func main() {
	fmt.Println("music library load")
	flag.Parse()

	t := tidal.New()

	// t.HandleAuthCallback()
	// if err := t.Authorize(); err != nil {
	// 	fmt.Println("failed to get authorization: ", err)
	// 	os.Exit(1)
	// }

	// select {
	// case <-t.CredentialsChan:
	// 	// go for it!
	// }

	if *usernameFlag == "" || *passwordFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := t.Login(*usernameFlag, *passwordFlag); err != nil {
		fmt.Printf("login failed %v", err)
		os.Exit(-1)
	}

	result, err := t.SearchArtist("Seven Mary Three")
	fmt.Println("search error? ", err)
	fmt.Printf("%+v\n", result)

	pl, err := t.CreatePlaylist("7Mary3_" + strconv.Itoa(rand.Int()))
	fmt.Println("create playlist error? ", err)

	err = t.AddSongToPlaylist(pl.UUID, "3029318", "3029322")
	fmt.Println("add song error? ", err)
}
