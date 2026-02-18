package games

import (
	"math/rand"
	"time"
)

// Ajouter un joueur à la partie.
func (g *PetitBacGame) AddPlayer(name string) {
	if _, exists := g.Players[name]; !exists {
		g.Players[name] = &Player{
			Name:    name,
			Answers: make(map[string]string),
			Score:   0,
		}
	}
}

// Tirer une lettre aléatoire non utilisée et avancer la manche.
func (g *PetitBacGame) DrawLetter() string {
	rand.Seed(time.Now().UnixNano())

	available := ""
	for _, letter := range Letters {
		used := false
		for _, usedLetter := range g.UsedLetters {
			if string(letter) == usedLetter {
				used = true
				break
			}
		}
		if !used {
			available += string(letter)
		}
	}

	if len(available) == 0 {
		available = Letters
		g.UsedLetters = []string{}
	}

	letter := string(available[rand.Intn(len(available))])

	g.UsedLetters = append(g.UsedLetters, letter)
	g.Letter = letter
	g.Round++
	g.RoundStartTime = time.Now()
	g.Phase = "playing"

	return letter
}

// Vérifier si la partie est terminée.
func (g *PetitBacGame) IsFinished() bool {
	return g.Round >= g.MaxRounds || g.IsFinishedByCreator
}

// Terminer la partie (par le créateur).
func (g *PetitBacGame) StopGame() {
	g.IsFinishedByCreator = true
}

// Vérifier si le temps de la manche est écoulé.
func (g *PetitBacGame) IsTimeUp() bool {
	if g.RoundStartTime.IsZero() {
		return false
	}
	elapsed := time.Since(g.RoundStartTime).Seconds()
	return elapsed >= float64(g.TimeLimit)
}

// Obtenir le temps restant de la manche en secondes.
func (g *PetitBacGame) GetTimeRemaining() int {
	if g.RoundStartTime.IsZero() {
		return g.TimeLimit
	}
	elapsed := int(time.Since(g.RoundStartTime).Seconds())
	remaining := g.TimeLimit - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Réinitialiser les réponses pour une nouvelle manche.
func (g *PetitBacGame) ResetAnswers() {
	for _, player := range g.Players {
		player.Answers = make(map[string]string)
	}
	g.CurrentAnswers = nil
	g.Phase = "waiting_next"
	g.PlayersSubmitted = make(map[string]bool)
	g.PlayersVoted = make(map[string]bool)
	g.PlayersReadyForNextRound = make(map[string]bool)
}

// Marquer qu'un joueur est prêt pour la prochaine manche.
func (g *PetitBacGame) MarkPlayerReadyForNextRound(playerName string) {
	g.PlayersReadyForNextRound[playerName] = true
}

// Vérifier si tous les joueurs sont prêts pour la prochaine manche.
func (g *PetitBacGame) AllPlayersReadyForNextRound() bool {
	return g.allDone(g.PlayersReadyForNextRound)
}

// Préparer la manche suivante (tire la lettre une seule fois).
func (g *PetitBacGame) PrepareNextRound() {
	if g.Phase == "playing" {
		return
	}
	g.ResetAnswers()
	g.DrawLetter()
	g.Phase = "playing"
}
