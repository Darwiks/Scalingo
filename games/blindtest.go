package games

import (
	"log"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"
)

type PlayerScore struct {
	Name  string
	Score int
}

type Song struct {
	ID     int
	Title  string
	Artist string
	File   string
}

type PlaylistInfo struct {
	ID   string
	Name string
}

var AvailablePlaylists = []PlaylistInfo{
	{ID: "37i9dQZF1DXcBWIGoYBM5M", Name: "Top Hits (Spotify)"},
	{ID: "37i9dQZF1DX0XUsuxWHRQd", Name: "Rap (Spotify)"},
	{ID: "37i9dQZF1DX4o1oenSJRJd", Name: "Années 2000 (Spotify)"},
	{ID: "37i9dQZF1DWXRqgorJj26U", Name: "Rock (Spotify)"},
}

func ShufflePlaylist(songs []Song) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(songs), func(i, j int) {
		songs[i], songs[j] = songs[j], songs[i]
	})
}

type GameState struct {
	CurrentSongIndex  int
	Playlist          []Song
	Scores            map[string]int
	IsFinished        bool
	TitleFound        bool
	ArtistFound       bool
	Timer             *time.Timer
	IsTimerRunning    bool
	CurrentPlaylistID string
	TimeLimit         int
	MaxRounds         int
}

func CleanString(input string) string {
	s := strings.ToLower(input)
	reg, _ := regexp.Compile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "")

	return s
}

func (g *GameState) CheckAnswer(userAnswer string, playerName string) (int, string) {
	if g.IsFinished {
		return 0, ""
	}

	if g.Scores == nil {
		g.Scores = make(map[string]int)
	}

	currentSong := g.Playlist[g.CurrentSongIndex]
	cleanedInput := CleanString(userAnswer)

	points := 0
	foundType := ""

	if !g.ArtistFound && cleanedInput == CleanString(currentSong.Artist) {
		g.ArtistFound = true
		points += 1
		if foundType == "" {
			foundType = "artist"
		}
	}

	if !g.TitleFound && cleanedInput == CleanString(currentSong.Title) {
		g.TitleFound = true
		points += 2
		if foundType == "" {
			foundType = "title"
		} else {
			foundType = "both"
		}
	}

	if !g.ArtistFound || !g.TitleFound {
		fullAnswer := CleanString(currentSong.Artist + " " + currentSong.Title)
		fullAnswerReverse := CleanString(currentSong.Title + " " + currentSong.Artist)

		if cleanedInput == fullAnswer || cleanedInput == fullAnswerReverse {
			if !g.ArtistFound {
				g.ArtistFound = true
				points += 1
			}
			if !g.TitleFound {
				g.TitleFound = true
				points += 2
			}
			foundType = "both"
		}
	}

	if points > 0 {
		g.Scores[playerName] += points
	}

	return points, foundType
}

func (g *GameState) NextSong() {
	if g.Timer != nil {
		g.Timer.Stop()
	}
	g.IsTimerRunning = false // Reset du flag

	g.CurrentSongIndex++

	g.TitleFound = false
	g.ArtistFound = false

	maxRounds := g.MaxRounds
	if maxRounds == 0 {
		maxRounds = 5 // Default fallback
	}

	if g.CurrentSongIndex >= maxRounds || g.CurrentSongIndex >= len(g.Playlist) {
		g.IsFinished = true
		g.CurrentSongIndex = -1
	} else {
		nextSong := g.Playlist[g.CurrentSongIndex]
		log.Printf("NOUVELLE CHANSON : %s - %s", nextSong.Artist, nextSong.Title)
	}
}

func (g *GameState) StartRoundTimer(callback func()) {
	if g.Timer != nil {
		g.Timer.Stop()
	}
	g.IsTimerRunning = true

	duration := time.Duration(g.TimeLimit) * time.Second
	if duration == 0 {
		duration = 30 * time.Second
	}

	g.Timer = time.AfterFunc(duration, func() {
		g.IsTimerRunning = false
		callback()
	})
}

func (g *GameState) GetCurrentSong() *Song {
	if len(g.Playlist) == 0 || g.CurrentSongIndex < 0 || g.CurrentSongIndex >= len(g.Playlist) {
		return nil
	}
	return &g.Playlist[g.CurrentSongIndex]
}

func (g *GameState) GetScoreboard() []PlayerScore {
	var scoreboard []PlayerScore
	for name, score := range g.Scores {
		scoreboard = append(scoreboard, PlayerScore{Name: name, Score: score})
	}
	sort.Slice(scoreboard, func(i, j int) bool {
		return scoreboard[i].Score > scoreboard[j].Score
	})
	return scoreboard
}
