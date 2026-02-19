package src

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func Home(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := false
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value == "connection_ok" {
		isLoggedIn = true
	}

	data := map[string]interface{}{
		"IsLoggedIn": isLoggedIn,
	}

	template, err := template.ParseFiles("./index.html")
	if err != nil {
		log.Fatal(err)
	}
	template.Execute(w, data)
}

func Register(w http.ResponseWriter, r *http.Request) {
	errorMessage := r.URL.Query().Get("error")

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Redirect(w, r, "/register?error=Erreur+lors+de+la+lecture+du+formulaire", http.StatusSeeOther)
			return
		}

		pseudo := r.FormValue("pseudo")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirmPassword")

		if EmailExists(email) {
			http.Redirect(w, r, "/register?error=Cet+email+est+déjà+utilisé", http.StatusSeeOther)
			return
		}
		if PseudoExists(pseudo) {
			http.Redirect(w, r, "/register?error=Ce+pseudo+est+déjà+utilisé", http.StatusSeeOther)
			return
		}

		if password != confirmPassword {
			http.Redirect(w, r, "/register?error=Les+mots+de+passe+ne+correspondent+pas", http.StatusSeeOther)
			return
		}

		// strength := zxcvbn.PasswordStrength(password, inputs)

		// if strength.Score < 2 {
		// 	http.Redirect(w, r, "/register?error=Mot+de+passe+trop+faible", http.StatusSeeOther)
		// 	return
		// }

		if pseudo == "" || email == "" || password == "" {
			http.Redirect(w, r, "/register?error=Tous+les+champs+sont+requis", http.StatusSeeOther)
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			http.Redirect(w, r, "/register?error=Erreur+lors+du+hashage+du+mot+de+passe", http.StatusSeeOther)
			return
		}

		if err := InitDatabase(); err != nil {
			log.Println("db init:", err)
			http.Redirect(w, r, "/register?error=Erreur+de+serveur", http.StatusSeeOther)
			return
		}

		stmt, err := DB.Prepare("INSERT INTO users(pseudo, email, password) VALUES($1, $2, $3)")
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/register?error=Erreur+de+base+de+données", http.StatusSeeOther)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(pseudo, email, string(hashed))
		if err != nil {
			log.Println("Error creating user:", err)
			http.Redirect(w, r, "/register?error=Erreur+lors+de+la+création+du+compte", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/login?success=Compte+créé+avec+succès", http.StatusSeeOther)
		return

	}

	data := map[string]interface{}{
		"ErrorMessage": errorMessage,
	}

	template, err := template.ParseFiles("./pages/register.html")
	if err != nil {
		log.Fatal(err)
	}
	template.Execute(w, data)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var hashedPassword string

	errorMessage := r.URL.Query().Get("error")

	if r.Method == http.MethodPost {
		identifiant := r.FormValue("identifiant")
		password := r.FormValue("password")

		if identifiant == "" || password == "" {
			http.Redirect(w, r, "/login?error=Tous+les+champs+sont+requis", http.StatusSeeOther)
			return
		}

		var pseudo string
	err := DB.QueryRow("SELECT password, pseudo FROM users WHERE email = $1 OR pseudo = $2", identifiant, identifiant).Scan(&hashedPassword, &pseudo)
		if err == sql.ErrNoRows {
			http.Redirect(w, r, "/login?error=Identifiant+ou+mot+de+passe+incorrect", http.StatusSeeOther)
			return
		} else if err != nil {
			http.Redirect(w, r, "/login?error=Erreur+de+serveur", http.StatusSeeOther)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

		if err != nil {
			http.Redirect(w, r, "/login?error=Identifiant+ou+mot+de+passe+incorrect", http.StatusSeeOther)
			return
		}

		fmt.Println("Connexion réussie")

		http.SetCookie(w, &http.Cookie{
			Name:  "session_token",
			Value: "connection_ok",
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "Pseudo",
			Value: pseudo,
			Path:  "/",
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"ErrorMessage": errorMessage,
	}

	template, err := template.ParseFiles("./pages/login.html")
	if err != nil {
		log.Fatal(err)
	}
	template.Execute(w, data)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatal(err)
	}

	pseudo := "None"
	c, err := r.Cookie("Pseudo")
	if err == nil {
		pseudo = c.Value
	}

	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
		Name: pseudo,
	}

	hub.Register <- client

	if CurrentGame != nil && len(CurrentGame.Playlist) > 0 {
		song := CurrentGame.GetCurrentSong()
		if song != nil {
			msg := fmt.Sprintf(`{"type": "audio", "url": "%s", "msg": "Bienvenue ! La musique joue déjà."}`, song.File)
			client.Send <- []byte(msg)

			if !CurrentGame.IsTimerRunning && !CurrentGame.IsFinished {
				CurrentGame.StartRoundTimer(func() {
					StartNewRound()
				})
			}
		}
	}
	go client.readPump()
	go client.writePump()
}

func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
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

	http.Redirect(w, r, "/petit-bac/attente?room="+roomCode, http.StatusSeeOther)
}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
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
	} else if hub != nil {
		// Inform room members only when a user is newly added
		msg := fmt.Sprintf(`{"type":"room_update","room":"%s"}`, roomCode)
		hub.Broadcast <- []byte(msg)
	}

	http.Redirect(w, r, "/petit-bac/attente?room="+roomCode, http.StatusSeeOther)
}

func AttenteHandler(w http.ResponseWriter, r *http.Request) {
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
		var gameType string
		DB.QueryRow("SELECT jeu FROM room WHERE id = $1", roomID).Scan(&gameType)
		if gameType == "PetitBac" {
			http.Redirect(w, r, "/petit-bac/game", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/blindgame", http.StatusSeeOther)
		}
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

	tpl, err := template.ParseFiles("pages/attente.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tpl.Execute(w, data)
}

func StartGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomCode := r.FormValue("room_code")
	roomID, err := GetRoomID(DB, roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	cookie, _ := r.Cookie("Pseudo")
	pseudo := cookie.Value
	userID, _ := GetUserIDByPseudo(DB, pseudo)
	_, _, creatorID, _, _ := GetRoomDetails(DB, roomID)

	if userID != creatorID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	UpdateRoomStatus(DB, roomID, "playing")

	// Notifier les clients en salle d'attente via websocket
	if hub != nil {
		msg := fmt.Sprintf(`{"type":"redirect","room":"%s","url":"/petit-bac/game"}`, roomCode)
		hub.Broadcast <- []byte(msg)
	}

	// Redirect creator back to waiting room, auto-refresh will take care of redirecting everyone
	http.Redirect(w, r, "/attente?room="+roomCode, http.StatusSeeOther)
}
