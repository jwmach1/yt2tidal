package tidal

import (
	"os/exec"

	"github.com/godsic/tidalapi"
)

type Tidal struct {
	session         *tidalapi.Session
	CredentialsChan chan bool
}

func New() *Tidal {
	return &Tidal{
		session:         tidalapi.NewSession(tidalapi.HIGH),
		CredentialsChan: make(chan bool),
	}
}

func (t *Tidal) Authorize() error {
	url := t.session.GetOauth2URL()
	//url := fmt.Sprintf("https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=user-read-private", s.clientID, url.QueryEscape("http://localhost:8080/"))
	cmd := exec.Command("open", url)
	err := cmd.Start()
	return err
}

func (t *Tidal) Login(username, password string) error {
	return t.session.Login(username, password)
}

func (t *Tidal) SearchArtist() error {

}
