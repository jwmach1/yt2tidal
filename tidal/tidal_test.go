package tidal_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"spotify_load/tidal"
	"spotify_load/tidal/mocks"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type RoundTripFunc func(r *http.Request) (*http.Response, error)

//go:generate  mockery --name=RoundTripFunc

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
func Test_AddSongToPlaylist(t *testing.T) {
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
		err = testObject.AddSongToPlaylist("megalist")
		assert.NoError(t, err)
	})
}

type RecordingTransport struct {
	Tripper RoundTripFunc
}

func (rt *RecordingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt.Tripper(r)
}