package src

import (
	"database/sql"
	"groupie/games"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func BlindTestHome(w http.ResponseWriter, r *http.Request) {
	roomCode := GetUniqueRoomCode(DB)
	data := map[string]interface{}{
		"RoomCode": roomCode,
	}

	template, err := template.ParseFiles("pages/AccueilBlindTest.html")
	if err != nil {
		log.Fatal(err)
	}
	template.Execute(w, data)
}

func BlindTestConfig(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("room")
	data := map[string]interface{}{
		"RoomCode":  roomCode,
		"Playlists": games.AvailablePlaylists,
	}
	tpl, err := template.ParseFiles("pages/blindtest_config.html")
	if err != nil {
		log.Fatal(err)
	}
	tpl.Execute(w, data)
}

func AttenteBTHandler(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("room")
	if roomCode == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	roomID, err := GetRoomID(DB, roomCode)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("Pseudo")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pseudo := cookie.Value
	userID, _ := GetUserIDByPseudo(DB, pseudo)

	_, _, creatorID, status, err := GetRoomDetails(DB, roomID)
	if err != nil {
		http.Error(w, "Error getting room details", http.StatusInternalServerError)
		return
	}

	if status == "playing" {
		http.Redirect(w, r, "/blindgame", http.StatusSeeOther)
		return
	}

	users, err := GetRoomUsers(DB, roomID)
	if err != nil {
		http.Error(w, "Error getting users", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"RoomCode":  roomCode,
		"Users":     users,
		"IsCreator": userID == creatorID,
	}

	tpl, err := template.ParseFiles("pages/attenteBT.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tpl.Execute(w, data)
}

func BlindTest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Erreur formulaire", http.StatusBadRequest)
			return
		}

		playlistID := r.FormValue("playlist")
		maxRoundsStr := r.FormValue("maxRounds")
		timeLimitStr := r.FormValue("timeLimit")
		roomCode := r.FormValue("room_code")

		maxRounds, _ := strconv.Atoi(maxRoundsStr)
		timeLimit, _ := strconv.Atoi(timeLimitStr)

		if maxRounds <= 0 {
			maxRounds = 10
		}
		if timeLimit <= 0 {
			timeLimit = 30
		}

		if roomCode != "" {
			roomID, err := GetRoomID(DB, roomCode)
			if err == nil {
				UpdateRoomStatus(DB, roomID, "playing")
			} else {
				log.Println("Error getting room ID:", err)
			}
		}

		songs, err := games.LoadSongsFromDeezer(playlistID)
		if err != nil {
			log.Println("Erreur Deezer :", err)
			songs = []games.Song{}
		}

		games.ShufflePlaylist(songs)

		if len(songs) > maxRounds {
			songs = songs[:maxRounds]
		}

		CurrentGame = &games.GameState{
			Playlist:          songs,
			CurrentSongIndex:  0,
			Scores:            make(map[string]int),
			CurrentPlaylistID: playlistID,
			TimeLimit:         timeLimit,
			MaxRounds:         maxRounds,
		}

		log.Printf("Nouvelle partie configurée : Playlist=%s, Tours=%d, Timer=%ds", playlistID, maxRounds, timeLimit)

		http.Redirect(w, r, "/blindgame", http.StatusSeeOther)
		return
	}

	if CurrentGame == nil {
		http.Redirect(w, r, "/blind-test/config", http.StatusSeeOther)
		return
	}

	template, err := template.ParseFiles("pages/blindgame.html")
	if err != nil {
		log.Fatal(err)
	}

	data := map[string]interface{}{
		"TimeLimit": CurrentGame.TimeLimit,
	}
	template.Execute(w, data)
}

func BlindTestResults(w http.ResponseWriter, r *http.Request) {
	if CurrentGame == nil {
		http.Redirect(w, r, "/blind-test/config", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Scoreboard": CurrentGame.GetScoreboard(),
		"MaxRounds":  CurrentGame.MaxRounds,
	}

	tpl, err := template.New("results").Funcs(templateFuncs).ParseFiles("pages/blindtest_results.html")
	if err != nil {
		log.Println("Erreur template:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "blindtest_results.html", data)
}

func CreateRoomBTHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomCode := r.FormValue("room_code")
	gameType := r.FormValue("game_type")

	if roomCode == "" || gameType == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("Pseudo")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pseudo := cookie.Value

	userID, err := GetUserIDByPseudo(DB, pseudo)
	if err != nil {
		log.Println("Error getting user ID:", err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	err = CreateRoom(DB, roomCode, gameType, userID)
	if err != nil {
		log.Println("Error creating room:", err)
		http.Error(w, "Error creating room", http.StatusInternalServerError)
		return
	}

	roomID, _ := GetRoomID(DB, roomCode)
	AddUserToRoom(DB, roomID, userID)

	http.Redirect(w, r, "/blind-test/attente?room="+roomCode, http.StatusSeeOther)
}

func JoinRoomBTHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomCode := r.FormValue("room_code")
	if roomCode == "" {
		http.Error(w, "Missing room code", http.StatusBadRequest)
		return
	}

	roomID, err := GetRoomID(DB, roomCode)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Room not found", http.StatusNotFound)
		} else {
			log.Println("Error getting room:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	cookie, err := r.Cookie("Pseudo")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pseudo := cookie.Value

	userID, err := GetUserIDByPseudo(DB, pseudo)
	if err != nil {
		log.Println("Error getting user ID:", err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	err = AddUserToRoom(DB, roomID, userID)
	if err != nil {
		log.Println("Error adding user to room:", err)
	}

	http.Redirect(w, r, "/blind-test/attente?room="+roomCode, http.StatusSeeOther)
}
