package src

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)


func InitDatabase() error {
    if DB !=nil{
        return nil
    }

	// Charger la configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("Erreur de configuration:", err)
		return err
	}

	// Connection à la base PostgreSQL
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Vérifier la connexion
	err = db.Ping()
	if err != nil {
		fmt.Println("Impossible de se connecter à la base de données:", err)
		db.Close()
		return err
	}
	_, err = CreateTableUsers(db)
	if err != nil {
        db.Close()
		return err
	}

	_, err = CreateTableRoom(db)
	if err != nil {
        db.Close()
		return err
	}

	_, err = CreateTableRoomUsers(db)
	if err != nil {
        db.Close()
		return err
	}

	
    DB = db
	fmt.Println("Connected to the PostgreSQL database successfully.")
    return nil
}
