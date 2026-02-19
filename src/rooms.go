package src

import (
	"database/sql"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
)

func CreateTableRoom(DB *sql.DB) (sql.Result, error) {

	query := `CREATE TABLE IF NOT EXISTS room (
		id SERIAL PRIMARY KEY,
		nom TEXT UNIQUE NOT NULL,
		jeu TEXT NOT NULL,
		creator_id INTEGER,
		status TEXT DEFAULT 'waiting',
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);`

	return DB.Exec(query)
}

func CreateTableRoomUsers(DB *sql.DB) (sql.Result, error) {

	query := `CREATE TABLE IF NOT EXISTS roomusers (
		room_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		PRIMARY KEY (room_id, user_id),
		FOREIGN KEY (room_id) REFERENCES room(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	return DB.Exec(query)
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRoomCode() string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func IsRoomCodeUnique(db *sql.DB, code string) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM room WHERE nom = $1)"
	err := db.QueryRow(query, code).Scan(&exists)
	if err != nil {
		return false
	}
	return !exists
}

func GetUniqueRoomCode(db *sql.DB) string {
	for {
		code := GenerateRoomCode()
		if IsRoomCodeUnique(db, code) {
			return code
		}
	}
}

func CreateRoom(db *sql.DB, name string, game string, creatorID int) error {
	query := "INSERT INTO room (nom, jeu, creator_id, status) VALUES ($1, $2, $3, 'waiting')"
	_, err := db.Exec(query, name, game, creatorID)
	return err
}

func GetRoomID(db *sql.DB, name string) (int, error) {
	var id int
	query := "SELECT id FROM room WHERE nom = $1"
	err := db.QueryRow(query, name).Scan(&id)
	return id, err
}

func GetRoomDetails(db *sql.DB, roomID int) (string, string, int, string, error) {
	var name, game, status string
	var creatorID int
	query := "SELECT nom, jeu, creator_id, status FROM room WHERE id = $1"
	err := db.QueryRow(query, roomID).Scan(&name, &game, &creatorID, &status)
	return name, game, creatorID, status, err
}

func GetRoomUsers(db *sql.DB, roomID int) ([]string, error) {
	query := `
		SELECT u.pseudo 
		FROM users u 
		JOIN roomusers ru ON u.id = ru.user_id 
		WHERE ru.room_id = $1
		ORDER BY ru.ctid ASC
	`
	rows, err := db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var pseudo string
		if err := rows.Scan(&pseudo); err != nil {
			return nil, err
		}
		users = append(users, pseudo)
	}
	return users, nil
}

func UpdateRoomStatus(db *sql.DB, roomID int, status string) error {
	query := "UPDATE room SET status = $1 WHERE id = $2"
	_, err := db.Exec(query, status, roomID)
	return err
}

func AddUserToRoom(db *sql.DB, roomID int, userID int) error {
	query := "INSERT INTO roomusers (room_id, user_id) VALUES ($1, $2)"
	_, err := db.Exec(query, roomID, userID)
	return err
}
