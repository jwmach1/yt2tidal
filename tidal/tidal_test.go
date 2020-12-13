package tidal_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"yt2tidal/tidal"
	"yt2tidal/tidal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

//go:generate  mockery --name=RoundTripFunc
// this seemed like a good idea to try, I'm not sure if I like how verbose it got
// on the upside, there is a lot of possiblity to further assert the request.
type RoundTripFunc func(r *http.Request) (*http.Response, error)

func Test_UnmarshalSession(t *testing.T) {
	sample := `{"userId":123456789,"sessionId":"fb49e2f6-529c-470a-90dd-f15b612ab8d4","countryCode":"US"}`

	testObject := tidal.Session{}
	err := json.NewDecoder(strings.NewReader(sample)).Decode(&testObject)

	assert.NoError(t, err)
	expected := tidal.Session{
		UserID:      123456789,
		SessionID:   "fb49e2f6-529c-470a-90dd-f15b612ab8d4",
		CountryCode: "US",
	}
	assert.Equal(t, expected, testObject)

}

func Test_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tripper := new(mocks.RoundTripFunc)
		transport := &RecordingTransport{
			Tripper: tripper.Execute,
		}
		response := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(`{"userId":123,"sessionId":"thesessionid","countryCode":"US"}`)),
			StatusCode: http.StatusOK,
		}
		tripper.On("Execute", mock.Anything).Return(response, nil)
		testObject := tidal.NewClient(http.Client{
			Transport: transport,
		})

		err := testObject.Login("fred", "yabba-dabba-do")
		assert.NoError(t, err)
	})
	t.Run("error status", func(t *testing.T) {
		tripper := new(mocks.RoundTripFunc)
		transport := &RecordingTransport{
			Tripper: tripper.Execute,
		}
		response := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(`body content`)),
			StatusCode: http.StatusUnauthorized,
		}
		tripper.On("Execute", mock.Anything).Return(response, nil)
		testObject := tidal.NewClient(http.Client{
			Transport: transport,
		})

		err := testObject.Login("fred", "yabba-dabba-do")
		assert.EqualError(t, err, "failed to login: body content")
	})
	t.Run("transport error", func(t *testing.T) {
		tripper := new(mocks.RoundTripFunc)
		transport := &RecordingTransport{
			Tripper: tripper.Execute,
		}
		tripper.On("Execute", mock.Anything).Return(nil, errors.New("simulate transport error"))
		testObject := tidal.NewClient(http.Client{
			Transport: transport,
		})

		err := testObject.Login("fred", "yabba-dabba-do")
		assert.EqualError(t, err, "Post \"https://api.tidal.com/v1/login/username\": simulate transport error")
	})

}

func Test_CreatePlaylist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tripper := new(mocks.RoundTripFunc)
		transport := &RecordingTransport{
			Tripper: tripper.Execute,
		}
		loginResponse := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(`{"userId":123,"sessionId":"thesessionid","countryCode":"FR"}`)),
			StatusCode: http.StatusOK,
		}
		b, err := ioutil.ReadFile("./test_data/create_playlist_response.json")
		require.NoError(t, err)
		createResponse := &http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			StatusCode: http.StatusOK,
		}

		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			return r.URL.String() == "https://api.tidal.com/v1/login/username"
		})).Return(loginResponse, nil).Once()
		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			matches := r.URL.String() == "https://api.tidal.com/v1/users/123/playlists?countryCode=FR"
			if !matches {
				t.Logf("matching %+v\n", r.URL.String())
			}
			return matches
		})).Return(createResponse, nil).Once()

		testObject := tidal.NewClient(http.Client{
			Transport: transport,
		})

		require.NoError(t, testObject.Login("fred", "yabba-dabba-do"))
		actual, err := testObject.CreatePlaylist("megalist")
		assert.NoError(t, err)
		assert.Equal(t, "created from loader", actual.Description)
	})
}

func Test_SearchArtist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tripper := new(mocks.RoundTripFunc)
		transport := &RecordingTransport{
			Tripper: tripper.Execute,
		}
		loginResponse := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(`{"userId":123,"sessionId":"thesessionid","countryCode":"FR"}`)),
			StatusCode: http.StatusOK,
		}
		b, err := ioutil.ReadFile("./test_data/search_artist_sample.json")
		require.NoError(t, err)
		searchResponse := &http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			StatusCode: http.StatusOK,
		}

		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			return r.URL.String() == "https://api.tidal.com/v1/login/username"
		})).Return(loginResponse, nil).Once()
		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			matches := r.URL.String() == "https://api.tidal.com/v1/search/artists?countryCode=FR&limit=25&query=u2"
			if !matches {
				t.Logf("matching %+v\n", r.URL.String())
			}
			return matches
		})).Return(searchResponse, nil).Once()

		testObject := tidal.NewClient(http.Client{
			Transport: transport,
		})

		require.NoError(t, testObject.Login("fred", "yabba-dabba-do"))
		actual, err := testObject.SearchArtist("u2")
		assert.NoError(t, err)
		assert.Equal(t, "Seven Mary Three", actual.Items[0].Name)
		assert.Equal(t, 14873, actual.Items[0].ID)
	})
}

func Test_GetAlbumsForArtist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tripper := new(mocks.RoundTripFunc)
		transport := &RecordingTransport{
			Tripper: tripper.Execute,
		}
		loginResponse := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(`{"userId":123,"sessionId":"thesessionid","countryCode":"FR"}`)),
			StatusCode: http.StatusOK,
		}
		b, err := ioutil.ReadFile("./test_data/artist_albums.json")
		require.NoError(t, err)
		searchResponse := &http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			StatusCode: http.StatusOK,
		}

		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			return r.URL.String() == "https://api.tidal.com/v1/login/username"
		})).Return(loginResponse, nil).Once()
		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			matches := r.URL.String() == "https://api.tidal.com/v1/artists/123/albums?countryCode=FR&limit=25&offset=0"
			if !matches {
				t.Logf("matching failed %+v\n", r.URL.String())
			}
			return matches
		})).Return(searchResponse, nil).Once()

		testObject := tidal.NewClient(http.Client{
			Transport: transport,
		})

		require.NoError(t, testObject.Login("fred", "yabba-dabba-do"))
		actual, err := testObject.GetAlbumsForArtist(123, tidal.NoneFilter)
		assert.NoError(t, err)
		assert.Equal(t, "Cumbersome", actual.Items[0].Title)
		assert.Equal(t, 88932695, actual.Items[0].ID)
		assert.Equal(t, 7, actual.TotalNumberOfItems)
	})
}

func Test_GetTracksForAlbums(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tripper := new(mocks.RoundTripFunc)
		transport := &RecordingTransport{
			Tripper: tripper.Execute,
		}
		loginResponse := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(`{"userId":123,"sessionId":"thesessionid","countryCode":"FR"}`)),
			StatusCode: http.StatusOK,
		}
		b, err := ioutil.ReadFile("./test_data/getTracksForAlbum.json")
		require.NoError(t, err)
		searchResponse := &http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			StatusCode: http.StatusOK,
		}

		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			return r.URL.String() == "https://api.tidal.com/v1/login/username"
		})).Return(loginResponse, nil).Once()
		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			matches := r.URL.String() == "https://api.tidal.com/v1/albums/456/tracks?countryCode=FR&filter=ALL&limit=50&offset=0"
			if !matches {
				t.Logf("matching failed %+v\n", r.URL.String())
			}
			return matches
		})).Return(searchResponse, nil).Once()

		testObject := tidal.NewClient(http.Client{
			Transport: transport,
		})

		require.NoError(t, testObject.Login("fred", "yabba-dabba-do"))
		actual, err := testObject.GetTracksForAlbum(456)
		assert.NoError(t, err)
		require.Len(t, actual.Items, 17)
		assert.Equal(t, "Last Kiss", actual.Items[1].Title)
		assert.Equal(t, 11896647, actual.Items[1].ID)
		assert.Equal(t, 17, actual.TotalNumberOfItems)
	})
}

func Test_AddSongToPlaylist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tripper := new(mocks.RoundTripFunc)
		transport := &RecordingTransport{
			Tripper: tripper.Execute,
		}
		loginResponse := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(`{"userId":123,"sessionId":"thesessionid","countryCode":"GR"}`)),
			StatusCode: http.StatusOK,
		}
		preflight := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader("")),
			StatusCode: http.StatusOK,
			Header:     http.Header{"etag": []string{"your_it"}},
		}
		b, err := ioutil.ReadFile("./test_data/create_playlist_response.json")
		require.NoError(t, err)
		createResponse := &http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			StatusCode: http.StatusOK,
		}

		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			return r.URL.String() == "https://api.tidal.com/v1/login/username"
		})).Return(loginResponse, nil).Once()
		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			return r.URL.String() == "https://api.tidal.com/v1/users/123/playlists?offset=0&limit=50&order=DATE_UPDATED&orderDirection=DESC&countryCode=GR"
		})).Return(preflight, nil).Once()
		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			matches := r.URL.String() == "https://listen.tidal.com/v1/playlists/megalist_uuid/items?countryCode=GR"
			if !matches {
				t.Logf("matching %+v\n", r.URL.String())
			}
			return matches
		})).Return(createResponse, nil).Once()

		testObject := tidal.NewClient(http.Client{
			Transport: transport,
		})

		require.NoError(t, testObject.Login("fred", "yabba-dabba-do"))
		err = testObject.AddSongToPlaylist("megalist_uuid")
		assert.NoError(t, err)
	})
}

func Test_TopHits(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tripper := new(mocks.RoundTripFunc)
		transport := &RecordingTransport{
			Tripper: tripper.Execute,
		}
		loginResponse := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(`{"userId":123,"sessionId":"thesessionid","countryCode":"FR"}`)),
			StatusCode: http.StatusOK,
		}
		b, err := ioutil.ReadFile("./test_data/top_hits_response.json")
		require.NoError(t, err)
		searchResponse := &http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			StatusCode: http.StatusOK,
		}

		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			return r.URL.String() == "https://api.tidal.com/v1/login/username"
		})).Return(loginResponse, nil).Once()
		tripper.On("Execute", mock.MatchedBy(func(r *http.Request) bool {
			matches := r.URL.String() == "https://api.tidal.com/v1/search/top-hits?countryCode=FR&includeContributors=true&limit=10&offset=0&query=Metallica+Load&types=ALBUMS,TRACKS"
			if !matches {
				t.Logf("matching failed %+v\n", r.URL.String())
			}
			return matches
		})).Return(searchResponse, nil).Once()

		testObject := tidal.NewClient(http.Client{
			Transport: transport,
		})

		require.NoError(t, testObject.Login("fred", "yabba-dabba-do"))
		actual, err := testObject.TopHits("Metallica Load")
		assert.NoError(t, err)
		require.Len(t, actual.Albums.Items, 1)
	})
}

type RecordingTransport struct {
	Tripper RoundTripFunc
}

func (rt *RecordingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt.Tripper(r)
}
