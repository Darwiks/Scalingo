package src

import (
	"database/sql"
	"fmt"
	"groupie/games"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func PetitBacLaunchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomCode := r.FormValue("room_code")

	// Récupérer/Créer la room si nécessaire
	roomID, err := GetRoomID(DB, roomCode)
	if err == sql.ErrNoRows {
		// Créer la room à la volée si elle n'existe pas encore
		cookie, err := r.Cookie("Pseudo")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		pseudo := cookie.Value
		creatorID, _ := GetUserIDByPseudo(DB, pseudo)
		if err := CreateRoom(DB, roomCode, "PetitBac", creatorID); err != nil {
			log.Println("Error auto-creating room:", err)
			http.Error(w, "Erreur lors de la création de la room", http.StatusInternalServerError)
			return
		}
		roomID, _ = GetRoomID(DB, roomCode)
		AddUserToRoom(DB, roomID, creatorID)
	}
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify creator
	cookie, err := r.Cookie("Pseudo")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pseudo := cookie.Value
	userID, _ := GetUserIDByPseudo(DB, pseudo)
	_, _, creatorID, _, _ := GetRoomDetails(DB, roomID)

	if userID != creatorID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse configuration from form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erreur formulaire", http.StatusBadRequest)
		return
	}

	maxRounds := 5
	timeLimit := 60

	if val := r.FormValue("maxRounds"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			maxRounds = n
		}
	}
	if val := r.FormValue("timeLimit"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			timeLimit = n
		}
	}

	categories := r.Form["categories[]"]
	if len(categories) == 0 {
		categories = []string{"Artiste", "Album", "Groupe de musique", "Instrument de musique", "Featuring"}
	}

	// Create the game with the room configuration
	currentGame = games.NewPetitBacGame(maxRounds, timeLimit, categories)
	currentGame.RoomCode = roomCode
	currentGame.CreatorID = creatorID

	// Ajouter tous les joueurs de la room
	users, err := GetRoomUsers(DB, roomID)
	if err == nil {
		for _, userPseudo := range users {
			currentGame.AddPlayer(userPseudo)
		}
	}

	currentGame.DrawLetter()

	// Update room status to playing
	UpdateRoomStatus(DB, roomID, "playing")

	// Notifier les clients en salle d'attente via websocket
	if hub != nil {
		msg := fmt.Sprintf(`{"type":"redirect","room":"%s","url":"/petit-bac/game"}`, roomCode)
		hub.Broadcast <- []byte(msg)
	}

	// Redirect creator to the game
	http.Redirect(w, r, "/petit-bac/game?room="+roomCode, http.StatusSeeOther)
}

func PetitBacHome(w http.ResponseWriter, r *http.Request) {
	roomCode := GetUniqueRoomCode(DB)
	data := map[string]interface{}{
		"RoomCode": roomCode,
	}

	template, err := template.ParseFiles("pages/ACCUEILPETITBAC.html")
	if err != nil {
		log.Fatal(err)
	}
	template.Execute(w, data)
}

func PetitBacHomeExplain(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("pages/AccueilPetitBac.html")
	if err != nil {
		log.Fatal(err)
	}
	template.Execute(w, nil)
}

func PetitBacConfig(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("room")
	data := map[string]interface{}{
		"RoomCode": roomCode,
	}
	tpl, err := template.ParseFiles("pages/petitbac_config.html")
	if err != nil {
		log.Fatal(err)
	}
	tpl.Execute(w, data)
}
