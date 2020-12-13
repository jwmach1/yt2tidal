package tidal

var (
	IMGPATH = "https://resources.tidal.com/images/%s/%dx%d.jpg"
)

type Error struct {
	userMessage string
}

type Artist struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Types      []string          `json:"artistTypes"`
	Popularity int               `json:"popularity"`
	URL        string            `json:"url"`
	PictureID  string            `json:"picture"`
	Roles      []ArtistRole      `json:"artistRoles"`
	Mixes      map[string]string `json:"mixes"`
}
type ArtistRole struct {
	CategoryID int    `json:"categoryId"`
	Category   string `json:"category"`
}

type Album struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	Cover          string `json:"cover"`
	NumberOfTracks int    `json:"numberOfTracks"`
	Duration       int    `json:"duration"`
	Artist         Artist `json:"artist"`
	ReleaseDate    string `json:"releaseDate"`
	Copyright      string `json:"copyright"`
	Upc            string `json:"ups"`
	Explicit       bool   `json:"explicit"`
	Tracks         *Tracks
}

type TrackPath struct {
	URL                   string `json:"Url"`
	TrackID               int    `json:"trackId"`
	PlayTimeLeftInMinutes int    `json:"playTimeLeftInMinutes"`
	SoundQuality          string `json:"soundQuality"`
	EncryptionKey         string `json:"encryptionKey"`
	Codec                 string `json:"codec"`
}

type Track struct {
	Duration        int           `json:"duration"`
	ReplayGain      float32       `json:"replayGain"`
	Copyright       string        `json:"copyright"`
	Artists         []Artist      `json:"artists"`
	URL             string        `json:"url"`
	ISRC            string        `json:"isrc"`
	Editable        bool          `json:"editable"`
	SurroundTypes   []interface{} `json:"surroundTypes"`
	Artist          Artist        `json:"artist"`
	Explicit        bool          `json:"explicit"`
	AudioQuality    string        `json:"audioQuality"`
	ID              int           `json:"id"`
	Peak            float32       `json:"peak"`
	StreamReady     bool          `json:"streamReady"`
	StreamStartDate string        `json:"streamStartDate"`
	Popularity      int           `json:"popularity"`
	Album           Album         `json:"album"`
	Title           string        `json:"title"`
	AllowStreaming  bool          `json:"allowStreaming"`
	TrackNumber     int           `json:"trackNumber"`
	VolumeNumber    int           `json:"volumeNumber"`
	Version         string        `json:"version"`
	Path            TrackPath
}
type Item struct {
	Created string `json:"created"`
	Item    Track  `json:"item"`
}

type ResultHeader struct {
	Limit              int `json:"limit"`
	Offset             int `json:"offset"`
	TotalNumberOfItems int `json:"totalNumberOfItems"`
}

type Tracks struct {
	*ResultHeader
	Items []Track `json:"items"`
}

type TracksFavorite struct {
	*ResultHeader
	Items []Item `json:"items"`
}

type Creator struct {
	ID int `json:"id"`
}

type Playlist struct {
	UUID            string  `json:"uuid"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Creator         Creator `json:"creator"`
	Type            string  `json:"type"`
	IsPublic        bool    `json:"publicPlaylist"`
	Created         string  `json:"created"`
	LastUpdated     string  `json:"lastUpdated"`
	NumberOfTracks  int     `json:"numberOfTracks"`
	NumberOfVideos  int     `json:"numberOfVideos"`
	Duration        int     `json:"duration"`
	URL             string  `json:"url"`
	ImageUUID       string  `json:"image"`
	SquareImageUUID string  `json:"squareImage"`
	Popularity      int     `json:"popularity"`
	Tracks          *Tracks
}

type ArtistSearch struct {
	*ResultHeader
	Items []Artist `json:"items"`
}
type AlbumSearch struct {
	*ResultHeader
	Items []Album `json:"items"`
}

type TopHits struct {
	Albums  AlbumSearch
	Artists ArtistSearch
	Tracks  Tracks
}
