package spotify

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
)

type Spotify struct {
	clientID        string
	clientSecret    string
	CredentialsChan chan bool
	Credentials     TokenResponse
}

func New(clientID, clientSecret string) *Spotify {

	return &Spotify{
		clientID:        clientID,
		clientSecret:    clientSecret,
		CredentialsChan: make(chan bool),
	}
}

func (s *Spotify) Authorize() error {
	url := fmt.Sprintf("https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s", s.clientID, url.QueryEscape("http://localhost:8080/"))
	cmd := exec.Command("open", url)
	err := cmd.Start()
	return err
}

// Search will call Spotify https://developer.spotify.com/documentation/web-api/reference/search/search/
func (s *Spotify) Search() error {
	url := fmt.Sprintf("https://api.spotify.com/v1/search?type=track&q=%s", url.QueryEscape("friends in low places"))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", s.Credentials.TokenType, s.Credentials.AccessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("search response: ", resp.Status)
	fmt.Printf("%s\n", b)

	return nil
}
