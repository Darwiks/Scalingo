package main

import (
	"groupie/src"
)

func main() {
	// src.InitDatabase()

	// Initialisation du jeu vide au départ
	src.CurrentGame = nil

	src.Server()
}
