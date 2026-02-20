package src

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
	"groupie/games"

	_ "github.com/glebarez/go-sqlite"
)

var DB *sql.DB

// Variable globale pour stocker la partie Petit Bac en cours
var currentGame *games.PetitBacGame

// Variable globale pour stocker la partie Blind Test en cours
var CurrentGame *games.GameState

// Template functions
var templateFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
}

// Handler pour le jeu Petit Bac
func PetitBacGame(w http.ResponseWriter, r *http.Request) {
	if currentGame == nil {
		http.Redirect(w, r, "/petit-bac", http.StatusSeeOther)
		return
	}

	// Si la partie est déjà marquée terminée (par le créateur ou par les manches), aller directement aux résultats
	if currentGame.IsFinished() {
		http.Redirect(w, r, "/petit-bac/results", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {

		// Vérifier si le temps est écoulé côté serveur
		if currentGame.IsTimeUp() {
			// Préparer la validation avant de rediriger si pas déjà fait
			if currentGame.Phase != "validation" {
				currentGame.PrepareValidation()
			}
			// Rediriger directement vers la validation si le temps est écoulé
			if hub != nil {
				msg := fmt.Sprintf(`{"type":"redirect","room":"%s","url":"/petit-bac/validation"}`, currentGame.RoomCode)
				hub.Broadcast <- []byte(msg)
			}
			http.Redirect(w, r, "/petit-bac/validation", http.StatusSeeOther)
			return
		}

		// Parse form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Erreur formulaire", http.StatusBadRequest)
			return
		}

		// Récupérer le pseudo du joueur
		cookie, err := r.Cookie("Pseudo")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		playerName := cookie.Value

		// Récupérer les réponses
		for _, category := range currentGame.Categories {
			answer := r.FormValue(category)
			currentGame.SubmitAnswer(playerName, category, answer)
		}

		// Marquer que le joueur a soumis
		currentGame.MarkPlayerSubmitted(playerName)

		// Dès qu'un joueur soumet, passer TOUT LE MONDE à la validation
		// Préparer la validation avant de rediriger
		if currentGame.Phase != "validation" {
			currentGame.PrepareValidation()
		}
		
		// Envoyer tous les joueurs à la page de validation
		if hub != nil {
			msg := fmt.Sprintf(`{"type":"redirect","room":"%s","url":"/petit-bac/validation"}`, currentGame.RoomCode)
			hub.Broadcast <- []byte(msg)
		}
		http.Redirect(w, r, "/petit-bac/validation", http.StatusSeeOther)
		return
	}

	// GET - Afficher le jeu
	// Récupérer le pseudo du joueur connecté
	cookie, err := r.Cookie("Pseudo")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pseudo := cookie.Value
	userID, _ := GetUserIDByPseudo(DB, pseudo)

	// Si la phase est déjà en validation ou le temps est écoulé, rediriger vers la validation
	if currentGame.Phase == "validation" || currentGame.IsTimeUp() {
		// S'assurer que la phase est bien en validation avant de quitter la page de jeu
		if currentGame.Phase != "validation" {
			currentGame.PrepareValidation()
			if hub != nil {
				msg := fmt.Sprintf(`{"type":"redirect","room":"%s","url":"/petit-bac/validation"}`, currentGame.RoomCode)
				hub.Broadcast <- []byte(msg)
			}
		}
		http.Redirect(w, r, "/petit-bac/validation", http.StatusSeeOther)
		return
	}

	// Vérifier si le joueur est en attente (a déjà soumis mais la phase n'est pas encore en validation)
	isWaiting := currentGame.PlayersSubmitted[pseudo] && currentGame.Phase != "validation"

	// Récupérer le score du joueur actuel
	playerScore := 0
	if player, exists := currentGame.Players[pseudo]; exists {
		playerScore = player.Score
	}

	// Compter les joueurs qui ont soumis
	submittedCount := 0
	for range currentGame.PlayersSubmitted {
		submittedCount++
	}

	data := map[string]interface{}{
		"Letter":         currentGame.Letter,
		"Round":          currentGame.Round,
		"MaxRounds":      currentGame.MaxRounds,
		"TimeLimit":      currentGame.TimeLimit,
		"TimeRemaining":  currentGame.GetTimeRemaining(),
		"Categories":     currentGame.Categories,
		"Score":          playerScore,
		"IsCreator":      userID == currentGame.CreatorID,
		"RoomCode":       currentGame.RoomCode,
		"IsWaiting":      isWaiting,
		"SubmittedCount": submittedCount,
		"TotalPlayers":   len(currentGame.Players),
	}

	tpl, err := template.ParseFiles("pages/petitbac_game.html")
	if err != nil {
		log.Println("Erreur template:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	tpl.Execute(w, data)
}

// Handler pour la validation des réponses
func PetitBacValidation(w http.ResponseWriter, r *http.Request) {
	if currentGame == nil {
		http.Redirect(w, r, "/petit-bac", http.StatusSeeOther)
		return
	}

	// Si la partie est terminée (créateur ou manches), rediriger vers les résultats
	if currentGame.IsFinished() {
		http.Redirect(w, r, "/petit-bac/results", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// Traiter les votes
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Erreur formulaire", http.StatusBadRequest)
			return
		}

		// Récupérer le pseudo du joueur
		cookie, err := r.Cookie("Pseudo")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		playerName := cookie.Value

		// Vérifier si le joueur a déjà voté
		if currentGame.PlayersVoted[playerName] {
			// Le joueur a déjà voté, rediriger vers la page de validation en mode attente
			http.Redirect(w, r, "/petit-bac/validation?waiting=true", http.StatusSeeOther)
			return
		}

		// Récupérer les votes depuis le formulaire (checkbox checked = valid)
		for i := range currentGame.CurrentAnswers {
			vote := r.FormValue(fmt.Sprintf("vote_%d", i))
			// Si le checkbox est coché, vote == "valid", sinon il n'est pas envoyé
			// Utiliser VoteAnswerByPlayer pour éviter les votes multiples
			currentGame.VoteAnswerByPlayer(i, playerName, vote == "valid")
		}

		// Marquer que le joueur a voté
		currentGame.MarkPlayerVoted(playerName)

		// Vérifier si tous les joueurs ont voté
		if currentGame.AllPlayersVoted() {
			// Calculer les scores
			currentGame.CalculateScores()

			if currentGame.IsFinished() {
				if hub != nil {
					msg := fmt.Sprintf(`{"type":"redirect","room":"%s","url":"/petit-bac/results"}`, currentGame.RoomCode)
					hub.Broadcast <- []byte(msg)
				}
				http.Redirect(w, r, "/petit-bac/results", http.StatusSeeOther)
			} else {
				currentGame.PrepareNextRound()
				if hub != nil {
					msg := fmt.Sprintf(`{"type":"redirect","room":"%s","url":"/petit-bac/game"}`, currentGame.RoomCode)
					hub.Broadcast <- []byte(msg)
				}
				http.Redirect(w, r, "/petit-bac/game", http.StatusSeeOther)
			}
		} else {
			// Rester sur la page de validation avec état "waiting"
			http.Redirect(w, r, "/petit-bac/validation?waiting=true", http.StatusSeeOther)
		}
		return
	}

	// GET - Afficher la page de validation
	// Récupérer le pseudo du joueur
	cookie, err := r.Cookie("Pseudo")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	playerName := cookie.Value

	// Si la phase est en cours de jeu (playing), rediriger vers la page de jeu
	if currentGame.Phase == "playing" && !currentGame.IsTimeUp() {
		http.Redirect(w, r, "/petit-bac/game", http.StatusSeeOther)
		return
	}

	// Préparer les réponses seulement si pas déjà fait
	if len(currentGame.CurrentAnswers) == 0 {
		currentGame.PrepareValidation()
	}

	// Vérifier si le joueur est en attente
	isWaiting := (r.URL.Query().Get("waiting") == "true" || currentGame.PlayersVoted[playerName]) && !currentGame.AllPlayersVoted()

	// Compter les joueurs qui ont voté
	votedCount := 0
	for range currentGame.PlayersVoted {
		votedCount++
	}

	// Organiser les réponses par catégorie
	type AnswerWithIndex struct {
		Index      int
		PlayerName string
		Answer     string
		Category   string
	}

	categoryAnswers := make(map[string][]AnswerWithIndex)
	for i, answer := range currentGame.CurrentAnswers {
		categoryAnswers[answer.Category] = append(categoryAnswers[answer.Category], AnswerWithIndex{
			Index:      i,
			PlayerName: answer.PlayerName,
			Answer:     answer.Answer,
			Category:   answer.Category,
		})
	}

	data := map[string]interface{}{
		"Answers":         currentGame.CurrentAnswers,
		"CategoryAnswers": categoryAnswers,
		"Categories":      currentGame.Categories,
		"Letter":          currentGame.Letter,
		"Round":           currentGame.Round,
		"MaxRounds":       currentGame.MaxRounds,
		"Scoreboard":      currentGame.GetScoreboard(),
		"IsWaiting":       isWaiting,
		"VotedCount":      votedCount,
		"TotalPlayers":    len(currentGame.Players),
		"RoomCode":        currentGame.RoomCode,
	}

	tpl, err := template.ParseFiles("pages/petitbac_validation.html")
	if err != nil {
		log.Println("Erreur template:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	tpl.Execute(w, data)
}

// Handler pour les résultats d'une manche
func PetitBacRoundResults(w http.ResponseWriter, r *http.Request) {
	if currentGame == nil {
		http.Redirect(w, r, "/petit-bac", http.StatusSeeOther)
		return
	}

	// Si la partie est terminée, rediriger tout le monde vers les résultats
	if currentGame.IsFinished() {
		http.Redirect(w, r, "/petit-bac/results", http.StatusSeeOther)
		return
	}

	// Récupérer le pseudo du joueur
	cookie, err := r.Cookie("Pseudo")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	playerName := cookie.Value

	// Vérifier si le joueur est en attente
	isWaiting := r.URL.Query().Get("waiting") == "true" || currentGame.PlayersReadyForNextRound[playerName]

	// Si tous sont prêts et phase == playing, rediriger vers le jeu
	if currentGame.AllPlayersReadyForNextRound() && currentGame.Phase == "playing" {
		http.Redirect(w, r, "/petit-bac/game", http.StatusSeeOther)
		return
	}

	// Compter les joueurs prêts
	readyCount := 0
	for range currentGame.PlayersReadyForNextRound {
		readyCount++
	}

	data := map[string]interface{}{
		"Answers":      currentGame.CurrentAnswers,
		"Round":        currentGame.Round,
		"MaxRounds":    currentGame.MaxRounds,
		"Scoreboard":   currentGame.GetScoreboard(),
		"IsWaiting":    isWaiting,
		"ReadyCount":   readyCount,
		"TotalPlayers": len(currentGame.Players),
		"RoomCode":     currentGame.RoomCode,
	}

	tpl, err := template.New("round_results").Funcs(templateFuncs).ParseFiles("pages/petitbac_round_results.html")
	if err != nil {
		log.Println("Erreur template:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "petitbac_round_results.html", data)
}

// Handler pour passer à la manche suivante
func PetitBacNextRound(w http.ResponseWriter, r *http.Request) {
	if currentGame == nil || r.Method != http.MethodPost {
		http.Redirect(w, r, "/petit-bac", http.StatusSeeOther)
		return
	}

	if currentGame.IsFinished() {
		http.Redirect(w, r, "/petit-bac/results", http.StatusSeeOther)
		return
	}

	// Récupérer le pseudo du joueur
	cookie, err := r.Cookie("Pseudo")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	playerName := cookie.Value

	// Si la phase est déjà "playing", c'est que la nouvelle manche a déjà été préparée
	if currentGame.Phase == "playing" {
		http.Redirect(w, r, "/petit-bac/game", http.StatusSeeOther)
		return
	}

	// Marquer le joueur comme prêt
	currentGame.MarkPlayerReadyForNextRound(playerName)

	// Vérifier si tous les joueurs sont prêts
	if currentGame.AllPlayersReadyForNextRound() {
		// Tous sont prêts, préparer la nouvelle manche (une seule fois)
		currentGame.ResetAnswers()
		currentGame.DrawLetter() // Ceci met Phase = "playing" dans DrawLetter
		if hub != nil {
			msg := fmt.Sprintf(`{"type":"redirect","room":"%s","url":"/petit-bac/game"}`, currentGame.RoomCode)
			hub.Broadcast <- []byte(msg)
		}
		http.Redirect(w, r, "/petit-bac/game", http.StatusSeeOther)
	} else {
		// Rester sur la page des résultats avec état "waiting"
		http.Redirect(w, r, "/petit-bac/round-results?waiting=true", http.StatusSeeOther)
	}
}

// Handler pour arrêter la partie (créateur uniquement)
func PetitBacStop(w http.ResponseWriter, r *http.Request) {
	if currentGame == nil || r.Method != http.MethodPost {
		http.Redirect(w, r, "/petit-bac", http.StatusSeeOther)
		return
	}

	// Vérifier que c'est le créateur
	cookie, err := r.Cookie("Pseudo")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userID, _ := GetUserIDByPseudo(DB, cookie.Value)

	if userID != currentGame.CreatorID {
		http.Error(w, "Seul le créateur peut arrêter la partie", http.StatusForbidden)
		return
	}

	// Arrêter la partie
	currentGame.StopGame()

	// Marquer la room comme terminée si elle existe
	if currentGame.RoomCode != "" {
		if roomID, err := GetRoomID(DB, currentGame.RoomCode); err == nil {
			UpdateRoomStatus(DB, roomID, "finished")
		}
	}

	// Notifier tous les clients connectés via websocket (avec room pour filtrer côté client)
	if hub != nil {
		room := currentGame.RoomCode
		msg := fmt.Sprintf(`{"type":"redirect","room":"%s","url":"/petit-bac/results"}`, room)
		hub.Broadcast <- []byte(msg)
	}

	// Rediriger vers les résultats finaux
	http.Redirect(w, r, "/petit-bac/results", http.StatusSeeOther)
}

// Handler pour les résultats
func PetitBacResults(w http.ResponseWriter, r *http.Request) {
	if currentGame == nil {
		http.Redirect(w, r, "/petit-bac", http.StatusSeeOther)
		return
	}

	// Group history by round
	var historyByRound []map[string]interface{}
	currentRound := 0

	for _, answer := range currentGame.History {
		if answer.Round != currentRound {
			currentRound = answer.Round
			historyByRound = append(historyByRound, map[string]interface{}{
				"Round":   currentRound,
				"Letter":  answer.Letter,
				"Answers": []games.Answer{},
			})
		}
		// Append to the last round
		lastIdx := len(historyByRound) - 1
		if lastIdx >= 0 {
			answers := historyByRound[lastIdx]["Answers"].([]games.Answer)
			historyByRound[lastIdx]["Answers"] = append(answers, answer)
		}
	}

	data := map[string]interface{}{
		"Scoreboard":   currentGame.GetScoreboard(),
		"RoundsPlayed": len(historyByRound),
		"History":      historyByRound,
	}

	tpl, err := template.New("results").Funcs(templateFuncs).ParseFiles("pages/petitbac_results.html")
	if err != nil {
		log.Println("Erreur template:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "petitbac_results.html", data)
}

// SecretForcePanel affiche la page secrète avec le bouton Force
func SecretForcePanel(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("pages/secret_force.html")
	if err != nil {
		log.Println("Erreur template secret_force.html:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	tpl.Execute(w, nil)
}

// SecretForceExecute exécute 25000 itérations avec des logs pour saturer le système
func SecretForceExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("🔥 [SECRET FORCE] Début de l'exécution - 25000 itérations")
	
	// Exécuter les 25000 itérations avec des logs inutiles
	for i := 1; i <= 25000; i++ {
		log.Printf("[FORCE %d/25000] Iteration inutile - Timestamp: %d - Random data: %f %s %d %v %t",
			i, time.Now().UnixNano(), 
			float64(i)*3.14159, 
			strings.Repeat("X", i%10),
			i*i, 
			map[string]int{"iteration": i, "total": 25000},
			i%2 == 0)
		
		// Logs supplémentaires pour chaque tranche de 1000
		if i%1000 == 0 {
			log.Printf("🚀 [CHECKPOINT] %d iterations complétées - Mémoire utilisée: ~%dMB - Status: RUNNING", i, i/100)
			log.Printf("📊 [STATS] Progress: %.2f%% - Remaining: %d - Elapsed: ~%ds", float64(i)/250.0, 25000-i, i/1000)
			log.Printf("💾 [DATA] Generating useless data: %s", strings.Repeat(fmt.Sprintf("[%d]", i), 5))
		}
		
		// Logs additionnels tous les 100 itérations
		if i%100 == 0 {
			log.Printf("⚡ [FORCE-ITER-%d] Saturating logs with meaningless information", i)
		}
	}
	
	log.Println("✅ [SECRET FORCE] Terminé - 25000 itérations complétées avec succès")
	log.Println("📈 [METRICS] Total logs générés: ~175000 lignes")
	log.Println("🔥 [IMPACT] Saturation des logs réussie - RPM spike attendu")
	
	// Retourner une réponse de succès
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Force execution completed"))
}

// SecretForceExecute2 exécute 500000 calculs exponentiels en 10 secondes
func SecretForceExecute2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("⚡ [SECRET FORCE 2.0] Début de l'exécution - 500000 calculs exponentiels à pleine vitesse")
	
	startTime := time.Now()
	iterations := 500000
	
	var result float64
	
	for i := 1; i <= iterations; i++ {
		// Calculs exponentiels lourds
		x := float64(i)
		
		// Série de calculs exponentiels et mathématiques complexes
		result = math.Pow(x, 2.5) * math.Exp(math.Sin(x/1000.0))
		result += math.Log(x+1) * math.Sqrt(x) * math.Cos(x/500.0)
		result *= math.Tanh(x / 10000.0)
		result += math.Pow(math.E, x/100000.0) * math.Sinh(x/50000.0)
		
		// Logs tous les 1000 itérations uniquement
		if i%1000 == 0 {
			elapsed := time.Since(startTime).Seconds()
			log.Printf("💥 [FORCE2.0-%d/500000] Exponential calc: x^2.5 * e^sin(x) = %.6e | Elapsed: %.2fs", 
				i, result, elapsed)
		}
	}
	
	totalElapsed := time.Since(startTime).Seconds()
	
	log.Printf("✅ [SECRET FORCE 2.0] Terminé - 500000 calculs exponentiels complétés en %.2fs", totalElapsed)
	log.Printf("📊 [METRICS] Vitesse: %.0f calculs/seconde", float64(iterations)/totalElapsed)
	log.Printf("🎯 [RESULT] Dernier résultat: %.6e", result)
	log.Printf("📈 [METRICS] Total logs générés: ~%d lignes", iterations/1000)
	log.Println("🔥 [IMPACT] CPU spike maximum + Saturation des logs réussie")
	
	// Retourner une réponse de succès
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Force 2.0 execution completed"))
}

// GourmandHandler affiche la page CPU Gourmand
func GourmandHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("pages/gourmand.html")
	if err != nil {
		log.Println("Erreur template gourmand.html:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	tpl.Execute(w, nil)
}

// GourmandExecute exécute des calculs 20x plus intensifs avec parallélisation
func GourmandExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("🔥🔥🔥 [CPU GOURMAND] DÉMARRAGE - 10,000,000 calculs avec 8 goroutines parallèles")
	
	startTime := time.Now()
	totalIterations := 10000000
	numGoroutines := 8
	iterationsPerGoroutine := totalIterations / numGoroutines
	
	// Channel pour synchroniser les goroutines
	done := make(chan bool, numGoroutines)
	
	// Lancer 8 goroutines en parallèle pour saturer le CPU
	for g := 0; g < numGoroutines; g++ {
		goroutineID := g
		go func() {
			var result float64
			startIter := goroutineID * iterationsPerGoroutine
			endIter := startIter + iterationsPerGoroutine
			
			for i := startIter; i < endIter; i++ {
				x := float64(i + 1)
				
				// Calculs exponentiels ULTRA lourds
				result = math.Pow(x, 3.5) * math.Exp(math.Sin(x/500.0))
				result += math.Log(x+1) * math.Sqrt(x) * math.Cos(x/250.0)
				result *= math.Tanh(x / 5000.0) * math.Sinh(x / 25000.0)
				result += math.Pow(math.E, x/50000.0) * math.Cosh(x/30000.0)
				result *= math.Atan(x/10000.0) * math.Asin(math.Sin(x/100000.0))
				result += math.Pow(x, 2.3) * math.Log10(x+1) * math.Sqrt(math.Abs(result)+1)
				
				// Logs tous les 1000 itérations par goroutine
				if (i-startIter)%1000 == 0 && i > startIter {
					elapsed := time.Since(startTime).Seconds()
					log.Printf("💥 [GOURMAND-G%d] Iter %d/%d (Total: %d/%d) | Result: %.4e | Elapsed: %.2fs",
						goroutineID, i-startIter, iterationsPerGoroutine, i, totalIterations, result, elapsed)
				}
			}
			
			log.Printf("✅ [GOURMAND-G%d] Goroutine terminée - %d itérations complétées", goroutineID, iterationsPerGoroutine)
			done <- true
		}()
	}
	
	// Attendre que toutes les goroutines se terminent
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	totalElapsed := time.Since(startTime).Seconds()
	
	log.Printf("✅✅✅ [CPU GOURMAND] TERMINÉ - %d calculs exponentiels complétés en %.2fs", totalIterations, totalElapsed)
	log.Printf("📊 [METRICS] Vitesse moyenne: %.0f calculs/seconde", float64(totalIterations)/totalElapsed)
	log.Printf("🔥 [METRICS] Goroutines utilisées: %d en parallèle", numGoroutines)
	log.Printf("💻 [IMPACT] CPU saturé à 100%% pendant %.2fs", totalElapsed)
	log.Printf("📈 [METRICS] Total logs générés: ~%d lignes", (totalIterations/1000)+numGoroutines+10)
	log.Println("🚀 [IMPACT] CPU spike MAXIMUM atteint - Saturation complète réussie")
	
	// Retourner une réponse de succès
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("CPU Gourmand execution completed"))
}

func Server() {
	hub = NewHub()
	go hub.Run()

	// Routes publiques
	http.HandleFunc("/", Home)
	http.HandleFunc("/login", RedirectIfAuth(Login))
	http.HandleFunc("/register", RedirectIfAuth(Register))
	http.HandleFunc("/logout", Logout)

	http.HandleFunc("/start-game", RequireAuth(StartGameHandler))

	// Routes protégées - Petit Bac
	http.HandleFunc("/petit-bac/attente", RequireAuth(AttenteHandler))
	http.HandleFunc("/petitbac", RequireAuth(PetitBacHome))
	http.HandleFunc("/petit-bac", RequireAuth(PetitBacHomeExplain))
	http.HandleFunc("/petit-bac/config", RequireAuth(PetitBacConfig))
	http.HandleFunc("/petit-bac/launch", RequireAuth(PetitBacLaunchHandler))
	http.HandleFunc("/petit-bac/game", RequireAuth(PetitBacGame))
	http.HandleFunc("/petit-bac/validation", RequireAuth(PetitBacValidation))
	http.HandleFunc("/petit-bac/round-results", RequireAuth(PetitBacRoundResults))
	http.HandleFunc("/petit-bac/next-round", RequireAuth(PetitBacNextRound))
	http.HandleFunc("/petit-bac/stop", RequireAuth(PetitBacStop))
	http.HandleFunc("/petit-bac/results", RequireAuth(PetitBacResults))
	http.HandleFunc("/petit-bac/create", RequireAuth(CreateRoomHandler))
	http.HandleFunc("/petit-bac/join", RequireAuth(JoinRoomHandler))

	// Routes protégées - Blind Test
	http.HandleFunc("/blind-test/attente", RequireAuth(AttenteBTHandler))
	http.HandleFunc("/blind-test/config", RequireAuth(BlindTestConfig))
	http.HandleFunc("/blind-test", RequireAuth(BlindTestHome))
	http.HandleFunc("/blindgame", RequireAuth(BlindTest))
	http.HandleFunc("/blind-test/results", RequireAuth(BlindTestResults))
	http.HandleFunc("/blind-test/create", RequireAuth(CreateRoomBTHandler))
	http.HandleFunc("/blind-test/join", RequireAuth(JoinRoomBTHandler))
	http.HandleFunc("/ws", RequireAuth(wsHandler))

	// Route secrète
	http.HandleFunc("/secret-force", SecretForcePanel)
	http.HandleFunc("/secret-force/execute", SecretForceExecute)
	http.HandleFunc("/secret-force/execute2", SecretForceExecute2)

	// Route CPU Gourmand
	http.HandleFunc("/gourmand", GourmandHandler)
	http.HandleFunc("/gourmand/execute", GourmandExecute)

	// Fichiers statiques
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	scriptFs := http.FileServer(http.Dir("script/"))
	http.Handle("/script/", http.StripPrefix("/script/", scriptFs))

	log.Println("Serveur démarré sur http://localhost:8080")
	ip := ":"
	port := os.Getenv("PORT")
	if err := http.ListenAndServe(ip+port, nil); err != nil {
		log.Fatal("Erreur serveur:", err)
	}
}
