package main

import (
	"flag"
	"fmt"
	"math/rand"
	"music_load/tidal"
	"os"
	"strconv"
)

var usernameFlag = flag.String("username", "", "tidal username")
var passwordFlag = flag.String("password", "", "tidal password")

func main() {
	fmt.Println("music library load")
	flag.Parse()

	t := tidal.New()

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
