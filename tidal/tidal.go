package tidal

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	apiToken = "wc8j_yBJd20zOmx0"
	clientID = "ck3zaWMi8Ka_XdI0"
)

type Filter string

const (
	NoneFilter         Filter = ""
	CompilationsFilter Filter = "COMPILATIONS"
)

type SearchType int

const (
	SearchTypeAlbum SearchType = iota
	SearchTypeArtist
	SearchTypePlaylist
	SearchTypeTrack
)

type Tidal struct {
	session         *Session
	client          http.Client
	CredentialsChan chan bool
	cache           *cache.Cache
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
		cache:           cache.New(30*time.Minute, 30*time.Minute),
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
func (t *Tidal) AddSongToPlaylist(playlistUUID string, trackIDs ...int) error {
	etag, err := t.preflight()
	if err != nil {
		return err
	}
	if len(trackIDs) == 0 {
		return nil
	}

	data := url.Values{}
	data.Set("onArtifactNotFound", "FAIL")
	data.Set("onDupes", "FAIL")
	sids := make([]string, len(trackIDs))
	for i, id := range trackIDs {
		sids[i] = strconv.Itoa(id)
	}
	data.Set("trackIds", strings.Join(sids, ","))
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

// preflight is the required GET request for the etag to be sent in modifying a playlist
//	  https://listen.tidal.com/v1/users/175289677/playlists?offset=13&limit=50&order=DATE_UPDATED&orderDirection=DESC&countryCod
//	  harvest etag response header
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
	if result, ok := t.cache.Get(name); ok {
		return result.(ArtistSearch), nil
	}

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
	// fmt.Printf("Search:\n%s\n", b)

	result := ArtistSearch{}
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&result)

	t.cache.Add(name, result, cache.DefaultExpiration)
	return result, err
}

func (t *Tidal) SearchAlbum(name string) (AlbumSearch, error) {
	if result, ok := t.cache.Get(name); ok {
		return result.(AlbumSearch), nil
	}

	data := url.Values{}
	data.Add("query", name)
	data.Add("limit", "25")
	data.Add("countryCode", t.session.CountryCode)

	req, _ := http.NewRequest("GET", "https://api.tidal.com/v1/search/albums?"+data.Encode(), nil)
	req.Header.Set("X-Tidal-Token", apiToken)
	req.Header.Add("X-Tidal-SessionId", t.session.SessionID)

	resp, err := t.client.Do(req)
	if err != nil {
		return AlbumSearch{}, err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	// fmt.Printf("Search:\n%s\n", b)

	result := AlbumSearch{}
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&result)

	t.cache.Add(name, result, cache.DefaultExpiration)
	return result, err
}

func (t *Tidal) GetArtist(id int) (ArtistSearch, error) {
	if result, ok := t.cache.Get(strconv.Itoa(id)); ok {
		return result.(ArtistSearch), nil
	}
	data := url.Values{}
	data.Add("filter", "ALL")
	data.Add("limit", "25")
	data.Add("offset", "0")
	data.Add("countryCode", t.session.CountryCode)

	url := fmt.Sprintf("https://api.tidal.com/v1/artists/%d?%s", id, data.Encode())
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Tidal-Token", apiToken)
	req.Header.Add("X-Tidal-SessionId", t.session.SessionID)

	resp, err := t.client.Do(req)
	if err != nil {
		return ArtistSearch{}, err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	// fmt.Printf("Artist %d:\n%s\n", id, b)

	var result ArtistSearch
	json.NewDecoder(bytes.NewReader(b)).Decode(&result)

	t.cache.Add(strconv.Itoa(id), result, cache.DefaultExpiration)
	return result, err
}

func (t *Tidal) GetAlbumsForArtist(id int, filter Filter) (AlbumSearch, error) {
	cacheKey := fmt.Sprintf("%d_%s", id, filter)
	if result, ok := t.cache.Get(cacheKey); ok {
		return result.(AlbumSearch), nil
	}
	data := url.Values{}
	if filter != NoneFilter {
		data.Add("filter", string(filter))
	}
	data.Add("limit", "25")
	data.Add("offset", "0")
	data.Add("countryCode", t.session.CountryCode)

	url := fmt.Sprintf("https://api.tidal.com/v1/artists/%d/albums?%s", id, data.Encode())
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Tidal-Token", apiToken)
	req.Header.Add("X-Tidal-SessionId", t.session.SessionID)

	resp, err := t.client.Do(req)
	if err != nil {
		return AlbumSearch{}, err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	// fmt.Printf("Albums of %d:\n%s\n", id, b)

	var result AlbumSearch
	json.NewDecoder(bytes.NewReader(b)).Decode(&result)

	t.cache.Add(cacheKey, result, cache.DefaultExpiration)
	return result, err
}

func (t *Tidal) GetTracksForAlbum(id int) (Tracks, error) {
	if result, ok := t.cache.Get(strconv.Itoa(id)); ok {
		return result.(Tracks), nil
	}

	data := url.Values{}
	data.Add("filter", "ALL")
	data.Add("limit", "50")
	data.Add("offset", "0")
	data.Add("countryCode", t.session.CountryCode)

	url := fmt.Sprintf("https://api.tidal.com/v1/albums/%d/tracks?%s", id, data.Encode())
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Tidal-Token", apiToken)
	req.Header.Add("X-Tidal-SessionId", t.session.SessionID)

	resp, err := t.client.Do(req)
	if err != nil {
		return Tracks{}, err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	// fmt.Printf("tracks of %d:\n%s\n", id, b)

	var result Tracks
	json.NewDecoder(bytes.NewReader(b)).Decode(&result)

	t.cache.Add(strconv.Itoa(id), result, cache.DefaultExpiration)
	return result, err
}

// https://listen.tidal.com/v1/search/top-hits?query=arctic%20monkeys%20Fluorescent%20Adolescent&limit=3&offset=0&types=ARTISTS,ALBUMS,TRACKS&includeContributors=true&countryCode=US

func (t *Tidal) TopHits(search string) (TopHits, error) {
	if result, ok := t.cache.Get(search); ok {
		return result.(TopHits), nil
	}

	data := url.Values{}
	data.Set("query", search)
	data.Set("limit", "10")
	data.Set("offset", "0")
	data.Set("types", "ALBUMS,TRACKS")
	data.Set("countryCode", t.session.CountryCode)
	data.Set("includeContributors", "true")

	url := fmt.Sprintf("https://api.tidal.com/v1/search/top-hits?%s", data.Encode())
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Tidal-Token", apiToken)
	req.Header.Add("X-Tidal-SessionId", t.session.SessionID)

	resp, err := t.client.Do(req)
	if err != nil {
		return TopHits{}, err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	// fmt.Printf("tracks of %d:\n%s\n", id, b)

	var result TopHits
	json.NewDecoder(bytes.NewReader(b)).Decode(&result)

	t.cache.Add(search, result, cache.DefaultExpiration)
	return result, err
}
