package games

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type DeezerPlaylistResponse struct {
	Tracks struct {
		Data []DeezerTrack `json:"data"`
	} `json:"tracks"`
}

type DeezerTrack struct {
	Title   string       `json:"title"`
	Preview string       `json:"preview"`
	Artist  DeezerArtist `json:"artist"`
}

type DeezerArtist struct {
	Name string `json:"name"`
}

func LoadSongsFromDeezer(playlistID string) ([]Song, error) {
	url := fmt.Sprintf("https://api.deezer.com/playlist/%s", playlistID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("erreur API Deezer")
	}

	var result DeezerPlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var gameSongs []Song
	count := 0

	for _, track := range result.Tracks.Data {
		if track.Preview != "" {
			count++
			newSong := Song{
				ID:     count,
				Title:  track.Title,
				Artist: track.Artist.Name,
				File:   track.Preview,
			}
			gameSongs = append(gameSongs, newSong)
		}
	}

	return gameSongs, nil
}
