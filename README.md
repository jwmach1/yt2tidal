# Youtube Playlist Conversion to Tidal

The other playist conversion tools soundiiz, tunemymusic were not working for me.  They missed songs, picked the wrong bands, so I'm trying out reverse engineering an API that will load playlists into Tidal sourced from a Google Takeout.

* Google Takeout from google Music doesn't give me enough data
* Google Takeout from Youtube Music only gives a _Video ID_ which I can't find a way to turn into song title, artist, album

Using the extractor.js we can download a playlist json which is then feed that into this program

## Process
* open youtube music tab in chrome
* navigate to your Playlist
  * scroll the browser to the bottom so YT loads the whole thing
* paste the contents of _extractor.js_ into the chrome developer console
* run this program, passing `-playlist ~/Downloads/_____playlist.json` among the other arguments
  * I recommend passing `-dryrun` the first time to see if all the songs are found on Tidal.  I've had to tweak song names.
 
