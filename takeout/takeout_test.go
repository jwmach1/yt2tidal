package takeout_test

import (
	"testing"
	"yt2tidal/takeout"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Parse(t *testing.T) {

	actual, err := takeout.Parse("./test_data/sample_playlist.json")

	assert.NoError(t, err)
	assert.Equal(t, "7Mary3", actual.Title)
	require.Len(t, actual.Songs, 18)
	expected := takeout.Song{
		Title:  "Water's Edge",
		Artist: "Seven Mary Three",
		Album:  "American Standard",
	}
	assert.Equal(t, expected, actual.Songs[0])
}
