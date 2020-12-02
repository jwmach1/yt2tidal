package tidal

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	apiToken = "wc8j_yBJd20zOmx0"
	clientID = "ck3zaWMi8Ka_XdI0"
)

type Tidal struct {
	session         *Session
	client          http.Client
	CredentialsChan chan bool
}

type Session struct {
	UserID      int    `json:"userId"`
	SessionID   string `json:"sessionId"`
	CountryCode string `json:"countryCode"`
}

func New() *Tidal {
	return NewClient(http.Client{})
}

func NewClient(client http.Client) *Tidal {
	return &Tidal{
		client:          client,
		CredentialsChan: make(chan bool),
	}
}

func (t *Tidal) Login(username, password string) error {

	cc := make([]byte, 16)
	rand.Read(cc)
	clientUniqueKey := fmt.Sprintf("%02x", cc)

	data := url.Values{}
	data.Add("clientUniqueKey", clientUniqueKey)
	data.Add("username", username)
	data.Add("password", password)

	req, _ := http.NewRequest("POST", "https://api.tidal.com/v1/login/username", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Tidal-Token", apiToken)

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to login: %s", b)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("login response:\n%s\n", b)

	t.session = new(Session)
	// json.NewDecoder(resp.Body).Decode(t.session)
	err = json.NewDecoder(bytes.NewReader(b)).Decode(t.session)
	return err
}

func (t *Tidal) CreatePlaylist(title string) (Playlist, error) {
	data := url.Values{}
	data.Set("title", title)
	data.Set("description", "created from loader")
	url := fmt.Sprintf("https://api.tidal.com/v1/users/%d/playlists?countryCode=%s", t.session.UserID, t.session.CountryCode)
	req, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Tidal-Token", apiToken)
	req.Header.Add("X-Tidal-SessionId", t.session.SessionID)

	resp, err := t.client.Do(req)
	if err != nil {
		return Playlist{}, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("playlist response: (%s)\n%s\n", resp.Status, body)
	result := Playlist{}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&result)
	return result, err
}

//Request URL: https://listen.tidal.com/v1/playlists/d1455749-1cb0-4da8-aa77-4d9a8fde0a78/items?countryCode=US
//onArtifactNotFound=FAIL&onDupes=FAIL&trackIds=3029318%2C3029322
func (t *Tidal) AddSongToPlaylist(playlistUUID string, trackIDs ...string) error {
	/*
	   requires preflight GET request with nil body
	   --->  https://listen.tidal.com/v1/users/175289677/playlists?offset=13&limit=50&order=DATE_UPDATED&orderDirection=DESC&countryCod
	   --> harvest etag response header

	*/
	etag, err := t.preflight()
	if err != nil {
		return err
	}

	data := url.Values{}
	data.Set("onArtifactNotFound", "FAIL")
	data.Set("onDupes", "FAIL")
	for _, id := range trackIDs {
		data.Add("trackIds", id)
	}
	url := fmt.Sprintf("https://listen.tidal.com/v1/playlists/%s/items?countryCode=%s", playlistUUID, t.session.CountryCode)
	req, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Tidal-Token", apiToken)
	req.Header.Add("X-Tidal-SessionId", t.session.SessionID)
	req.Header.Set("If-None-Match", etag)

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to add song(s): %s", b)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("add song response: (%s)\n%s\n", resp.Status, body)
	// result := Playlist{}
	// err = json.NewDecoder(bytes.NewReader(body)).Decode(&result)
	return err
}

func (t *Tidal) preflight() (string, error) {

	url := fmt.Sprintf("https://api.tidal.com/v1/users/%d/playlists?offset=0&limit=50&order=DATE_UPDATED&orderDirection=DESC&countryCode=%s", t.session.UserID, t.session.CountryCode)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Tidal-Token", apiToken)
	req.Header.Add("X-Tidal-SessionId", t.session.SessionID)

	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to preflight: %s", b)
	}

	return resp.Header.Get("etag"), nil
}

func (t *Tidal) SearchArtist(name string) (ArtistSearch, error) {
	//req.Header.Add("X-Tidal-SessionId", s.SessionID)
	data := url.Values{}
	data.Add("query", name)
	data.Add("limit", "25")
	data.Add("countryCode", t.session.CountryCode)

	req, _ := http.NewRequest("GET", "https://api.tidal.com/v1/search/artists?"+data.Encode(), nil)
	req.Header.Set("X-Tidal-Token", apiToken)
	req.Header.Add("X-Tidal-SessionId", t.session.SessionID)

	resp, err := t.client.Do(req)
	if err != nil {
		return ArtistSearch{}, err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Search:\n%s\n", b)

	result := ArtistSearch{}
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&result)

	return result, err
}
