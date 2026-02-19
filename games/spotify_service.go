package games

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Spotify Credentials (should be environment variables in production)
// Using placeholders as requested
var (
	SpotifyClientID     = "YOUR_SPOTIFY_CLIENT_ID"
	SpotifyClientSecret = "YOUR_SPOTIFY_CLIENT_SECRET"
)

type SpotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type SpotifyPlaylistResponse struct {
	Tracks struct {
		Items []struct {
			Track SpotifyTrack `json:"track"`
		} `json:"items"`
	} `json:"tracks"`
}

type SpotifyTrack struct {
	Name    string        `json:"name"`
	Preview string        `json:"preview_url"`
	Artists []SpotifyArtist `json:"artists"`
}

type SpotifyArtist struct {
	Name string `json:"name"`
}

var spotifyToken string
var tokenExpiry time.Time

func GetSpotifyToken() (string, error) {
	if spotifyToken != "" && time.Now().Before(tokenExpiry) {
		return spotifyToken, nil
	}

	authURL := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(SpotifyClientID + ":" + SpotifyClientSecret))
	req.Header.Set("Authorization", "Basic "+authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("spotify auth failed: %s", resp.Status)
	}

	var result SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	spotifyToken = result.AccessToken
	tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return spotifyToken, nil
}

func LoadSongsFromSpotify(playlistID string) ([]Song, error) {
	token, err := GetSpotifyToken()
	if err != nil {
		return nil, err
	}

	// Spotify API URL for playlist tracks
	apiURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s", playlistID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("spotify api error: %s", resp.Status)
	}

	var result SpotifyPlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var gameSongs []Song
	count := 0

	for _, item := range result.Tracks.Items {
		if item.Track.Preview != "" {
			count++
			artistName := "Unknown"
			if len(item.Track.Artists) > 0 {
				artistName = item.Track.Artists[0].Name
			}
			
			newSong := Song{
				ID:     count,
				Title:  item.Track.Name,
				Artist: artistName,
				File:   item.Track.Preview,
			}
			gameSongs = append(gameSongs, newSong)
		}
	}

	return gameSongs, nil
}
