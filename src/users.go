package src

import (
	"database/sql"

	_ "github.com/glebarez/go-sqlite"
)

func CreateTableUsers(DB *sql.DB) (sql.Result, error) {
	sql := `CREATE TABLE IF NOT EXISTS users (
        id          INTEGER PRIMARY KEY,
        pseudo    TEXT UNIQUE NOT NULL,
		email       TEXT UNIQUE NOT NULL,
        password    TEXT NOT NULL
    );`

	return DB.Exec(sql)
}

func EmailExists(email string) bool {
	var id int
	sqlStmt := `SELECT id FROM users WHERE email = ?`

	err := DB.QueryRow(sqlStmt, email).Scan(&id)

	if err != nil {
		return err != sql.ErrNoRows
	}
	return true
}

func PseudoExists(pseudo string) bool {
	var id int
	sqlStmt := `SELECT id FROM users WHERE pseudo = ?`

	err := DB.QueryRow(sqlStmt, pseudo).Scan(&id)

	if err != nil {
		return err != sql.ErrNoRows
	}
	return true
}

func GetUserIDByPseudo(db *sql.DB, pseudo string) (int, error) {
	var id int
	query := "SELECT id FROM users WHERE pseudo = ?"
	err := db.QueryRow(query, pseudo).Scan(&id)
	return id, err
}
