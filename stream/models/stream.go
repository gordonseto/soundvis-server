package stream

import "strings"

type Song struct {
	Name string		`json:"name"`
	Title string 	`json:"title"`
	ImageURL string	`json:"imageUrl"`	// this will be empty unless GetAlbumArtworkForSong is called
}

// converts shoutcast now playing format to a song
// example format: 119,1,328,10000,117,128,Luis Miguel Del Amargue - Sombra Perdida
func SongFromString(str string) *Song {
	song := &Song{}
	// the now playing info is after the 6th comma
	strArray := strings.Split(str, ",")
	if len(strArray) >= 7 {
		nowPlayingStr := strArray[6]
		// seperate artist and title
		nowPlayingStrArray := strings.Split(nowPlayingStr, "-")
		if len(nowPlayingStrArray) >= 2 {
			song.Name = strings.TrimSpace(nowPlayingStrArray[0])
			song.Title = strings.TrimSpace(nowPlayingStrArray[1])
		}
	}
	return song
}