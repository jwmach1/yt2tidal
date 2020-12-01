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
var tokenFlag = flag.String("token", "", "spotify token printed from prior run")

func main() {
	fmt.Println("spotify load")
	flag.Parse()

	if (*clientIDFlag == "" || *clientSecretFlag == "") && *tokenFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	s := spotify.New(*clientIDFlag, *clientSecretFlag, *tokenFlag)
	if *tokenFlag == "" {
		s.HandleAuthCallback()
		if err := s.Authorize(); err != nil {
			fmt.Println("failed to get authorization: ", err)
			os.Exit(1)
		}

		select {
		case <-s.CredentialsChan:
			// go for it!
		}
	}

	err := s.Search()

	if err != nil {
		fmt.Printf("search response:%v\n", err)
	}
}
