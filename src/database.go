package src

import (
	"database/sql"
	"fmt"

	_ "github.com/glebarez/go-sqlite"
)


func InitDatabase() error {
    if DB !=nil{
        return nil
    }

	// Connection a  la base
	db, err := sql.Open("sqlite", "groupietracker.db")
	if err != nil {
		fmt.Println(err)
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
	fmt.Println("Connected to the SQLite database successfully.")
    return nil
}
