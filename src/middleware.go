package src

import (
	"net/http"
)

// Middleware pour vérifier si l'utilisateur est connecté
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifier si le cookie de session existe
		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value != "connection_ok" {
			// L'utilisateur n'est pas connecté, rediriger vers la page de connexion
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// L'utilisateur est connecté, continuer vers le handler
		next(w, r)
	}
}

// Middleware pour rediriger les utilisateurs connectés vers l'accueil
func RedirectIfAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifier si le cookie de session existe
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie.Value == "connection_ok" {
			// L'utilisateur est déjà connecté, rediriger vers l'accueil
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}
