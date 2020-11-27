package main

/*
start with setting up spotify developer account: https://developer.spotify.com/documentation/web-api/quick-start/
 * https://developer.spotify.com/dashboard/applications

*/

import (
	"flag"
	"fmt"
	"os"
	"spotify_load/spotify"
)

var clientIDFlag = flag.String("clientID", "", "spotify application client id")
var clientSecretFlag = flag.String("clientSecret", "", "spotify application client secret")

func main() {
	fmt.Println("spotify load")
	flag.Parse()

	if *clientIDFlag == "" || *clientSecretFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	s := spotify.New(*clientIDFlag, *clientSecretFlag)
	s.HandleAuthCallback()
	err := s.Authorize()

	select {
	case <-s.CredentialsChan:
		// go for it!
	}

	err = s.Search()

	fmt.Printf("search response:%s\n", err)
}
