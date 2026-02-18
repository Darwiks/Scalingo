package games

import "time"

var DefaultCategories = []string{
	"Artiste",
	"Album",
	"Groupe de musique",
	"Instrument de musique",
	"Featuring",
}

var Letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Player struct {
	Name    string
	Answers map[string]string
	Score   int
}

type Answer struct {
	PlayerName string
	Category   string
	Answer     string
	Votes      int
	TotalVotes int
	IsValid    bool
	Points     int
	VotedBy    map[string]bool
	Round      int
	Letter     string
}

type PetitBacGame struct {
	Letter                   string
	UsedLetters              []string
	Round                    int
	MaxRounds                int
	TimeLimit                int
	Categories               []string
	Players                  map[string]*Player
	CurrentAnswers           []Answer
	History                  []Answer
	Phase                    string
	RoundStartTime           time.Time
	RoomCode                 string
	CreatorID                int
	IsFinishedByCreator      bool
	PlayersSubmitted         map[string]bool
	PlayersVoted             map[string]bool
	PlayersReadyForNextRound map[string]bool
}

func NewPetitBacGame(maxRounds int, timeLimit int, categories []string) *PetitBacGame {
	if len(categories) == 0 {
		categories = DefaultCategories
	}
	return &PetitBacGame{
		UsedLetters:              []string{},
		Round:                    0,
		MaxRounds:                maxRounds,
		TimeLimit:                timeLimit,
		Categories:               categories,
		Players:                  make(map[string]*Player),
		CurrentAnswers:           []Answer{},
		History:                  []Answer{},
		Phase:                    "playing",
		PlayersSubmitted:         make(map[string]bool),
		PlayersVoted:             make(map[string]bool),
		PlayersReadyForNextRound: make(map[string]bool),
	}
}
