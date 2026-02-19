package src

import (
	"fmt"
	"os"
)

// Config contient les configurations de l'application
type Config struct {
	DatabaseURL string
	DeezerAPIEnabled bool
}

// LoadConfig charge la configuration depuis les variables d'environnement
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Configuration de la base de données
	config.DatabaseURL = os.Getenv("DATABASE_URL")
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL est requise")
	}

	// Deezer API est public, pas besoin de clé
	config.DeezerAPIEnabled = true

	return config, nil
}
