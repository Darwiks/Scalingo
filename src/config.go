package src

import (
	"fmt"
	"os"
	"strings"
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

	// Correction pour lib/pq: remplacer sslmode=prefer par sslmode=require
	// car lib/pq ne supporte pas "prefer", seulement: disable, require, verify-full, verify-ca
	config.DatabaseURL = strings.Replace(config.DatabaseURL, "sslmode=prefer", "sslmode=require", 1)

	// Deezer API est public, pas besoin de clé
	config.DeezerAPIEnabled = true

	return config, nil
}
