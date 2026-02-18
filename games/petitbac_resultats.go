package games

import (
	"sort"
	"strings"
)

// Vérifier si une réponse commence par la bonne lettre.
func (g *PetitBacGame) ValidateAnswer(answer string) bool {
	if len(answer) == 0 {
		return false
	}

	firstLetter := strings.ToUpper(string(answer[0]))
	return firstLetter == g.Letter
}

// Soumettre les réponses d'un joueur.
func (g *PetitBacGame) SubmitPlayerAnswers(playerName string, answers map[string]string) {
	if player, exists := g.Players[playerName]; exists {
		player.Answers = answers
		g.PlayersSubmitted[playerName] = true
	}
}

// Soumettre une réponse pour une catégorie spécifique.
func (g *PetitBacGame) SubmitAnswer(playerName string, category string, answer string) {
	if player, exists := g.Players[playerName]; exists {
		player.Answers[category] = strings.TrimSpace(answer)
	}
}

// Marquer qu'un joueur a soumis toutes ses réponses.
func (g *PetitBacGame) MarkPlayerSubmitted(playerName string) {
	g.PlayersSubmitted[playerName] = true
}

// Vérifier si tous les joueurs ont soumis leurs réponses.
func (g *PetitBacGame) AllPlayersSubmitted() bool {
	return g.allDone(g.PlayersSubmitted)
}

// Marquer qu'un joueur a voté.
func (g *PetitBacGame) MarkPlayerVoted(playerName string) {
	g.PlayersVoted[playerName] = true
}

// Vérifier si tous les joueurs ont voté.
func (g *PetitBacGame) AllPlayersVoted() bool {
	return g.allDone(g.PlayersVoted)
}

// Préparer les réponses pour la validation.
func (g *PetitBacGame) PrepareValidation() []Answer {
	g.CurrentAnswers = nil
	for _, player := range g.Players {
		for _, category := range g.Categories {
			if answer, ok := player.Answers[category]; ok {
				trimmed := strings.TrimSpace(answer)
				if trimmed == "" {
					continue
				}
				g.CurrentAnswers = append(g.CurrentAnswers, Answer{
					PlayerName: player.Name,
					Category:   category,
					Answer:     trimmed,
					Votes:      0,
					TotalVotes: 0,
					IsValid:    false,
					Points:     0,
					VotedBy:    make(map[string]bool),
					Round:      g.Round,
					Letter:     g.Letter,
				})
			}
		}
	}
	g.Phase = "validation"
	return g.CurrentAnswers
}

// Valider une réponse.
func (g *PetitBacGame) VoteAnswer(answerIndex int, isValid bool) {
	if answerIndex >= 0 && answerIndex < len(g.CurrentAnswers) {
		g.CurrentAnswers[answerIndex].TotalVotes++
		if isValid {
			g.CurrentAnswers[answerIndex].Votes++
		}
	}
}

// Valider une réponse avec le nom du votant (pour éviter les votes multiples).
func (g *PetitBacGame) VoteAnswerByPlayer(answerIndex int, playerName string, isValid bool) bool {
	if answerIndex >= 0 && answerIndex < len(g.CurrentAnswers) {
		if _, hasVoted := g.CurrentAnswers[answerIndex].VotedBy[playerName]; hasVoted {
			return false
		}

		g.CurrentAnswers[answerIndex].VotedBy[playerName] = isValid
		g.CurrentAnswers[answerIndex].TotalVotes++
		if isValid {
			g.CurrentAnswers[answerIndex].Votes++
		}
		return true
	}
	return false
}

// Calculer les scores après validation.
func (g *PetitBacGame) CalculateScores() {
	if g.Phase == "results" {
		return
	}
	if len(g.Players) == 0 {
		return
	}

	threshold := float64(len(g.Players)) * 2.0 / 3.0

	for i := range g.CurrentAnswers {
		g.CurrentAnswers[i].IsValid = float64(g.CurrentAnswers[i].Votes) >= threshold
	}

	answerCounts := make(map[string]map[string]int)
	for _, answer := range g.CurrentAnswers {
		if !answer.IsValid {
			continue
		}
		catCounts := answerCounts[answer.Category]
		if catCounts == nil {
			catCounts = make(map[string]int)
			answerCounts[answer.Category] = catCounts
		}
		catCounts[strings.ToLower(answer.Answer)]++
	}

	for i := range g.CurrentAnswers {
		ans := &g.CurrentAnswers[i]
		if !ans.IsValid {
			continue
		}
		count := answerCounts[ans.Category][strings.ToLower(ans.Answer)]
		ans.Points = 2
		if count > 1 {
			ans.Points = 1
		}
		if player, ok := g.Players[ans.PlayerName]; ok {
			player.Score += ans.Points
		}
	}

	g.History = append(g.History, g.CurrentAnswers...)
	g.Phase = "results"
}

// Obtenir le scoreboard.
func (g *PetitBacGame) GetScoreboard() []Player {
	scoreboard := make([]Player, 0, len(g.Players))
	for _, player := range g.Players {
		scoreboard = append(scoreboard, *player)
	}
	sort.Slice(scoreboard, func(i, j int) bool { return scoreboard[i].Score > scoreboard[j].Score })
	return scoreboard
}

// Utilitaire partagé pour vérifier qu'une action est complétée par tous les joueurs.
func (g *PetitBacGame) allDone(check map[string]bool) bool {
	if len(g.Players) == 0 {
		return false
	}
	for playerName := range g.Players {
		if !check[playerName] {
			return false
		}
	}
	return true
}
