package spotify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
)

type Spotify struct {
	clientID        string
	clientSecret    string
	token           string
	CredentialsChan chan bool
	Credentials     TokenResponse
}

func New(clientID, clientSecret, token string) *Spotify {
	creds := TokenResponse{}
	if token != "" {
		creds.AccessToken = token
		creds.TokenType = "Bearer"
	}
	return &Spotify{
		clientID:        clientID,
		clientSecret:    clientSecret,
		token:           token,
		Credentials:     creds,
		CredentialsChan: make(chan bool),
	}
}

func (s *Spotify) Authorize() error {
	url := fmt.Sprintf("https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=user-read-private", s.clientID, url.QueryEscape("http://localhost:8080/"))
	cmd := exec.Command("open", url)
	err := cmd.Start()
	return err
}

// Search will call Spotify https://developer.spotify.com/documentation/web-api/reference/search/search/
//    http://web.archive.org/web/20120704131650/http://www.spotify.com/us/about/features/advanced-search-syntax/
func (s *Spotify) Search() error {
	url := fmt.Sprintf("https://api.spotify.com/v1/search?market=from_token&type=track&q=%s", url.QueryEscape("friends in low places garth brooks"))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", s.Credentials.TokenType, s.Credentials.AccessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("status %s from search\n%s", resp.Status, b)
	}
	searchResult := new(SearchResult)
	err = json.NewDecoder(resp.Body).Decode(searchResult)
	if err != nil {
		return fmt.Errorf("failed to read search response: %s", err)
	}
	fmt.Printf("track count:  %d\n", len(searchResult.Tracks.Tracks))
	for _, t := range searchResult.Tracks.Tracks {
		fmt.Printf("\t%s %s\n", t.Album.Name, t.Artists[0].Name)
	}

	return nil
}
