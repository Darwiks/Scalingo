[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/ZND1-UNW)

# 🎮 Groupie Tracker - La Salle de Jeux

Salut ! Voici notre projet. C'est un site web super simple pour jouer à des mini-jeux entre amis ou en famille.

## C'est quoi ce projet ?

C'est une application qui regroupe deux jeux multijoueurs en temps réel :

1.  🎵 **Le BlindTest** : Un quiz musical. Une musique se lance, et vous devez taper le nom de l'artiste ou le titre de la chanson le plus vite possible pour gagner des points.
2.  🎓 **Le Petit Bac** : Un jeu de réflexion. Une lettre est tirée au sort, et vous devez trouver des mots commençant par cette lettre dans différentes catégories (Prénom, Animal, Métier...).

## 🛠️ Comment lancer le jeu ? (Guide simple)

Pas besoin d'être un expert en informatique, suivez juste ces étapes :

1.  **Ouvrez le dossier** du projet sur votre ordinateur.
2.  Ouvrez un **Terminal** (l'invite de commande) dans ce dossier.
3.  Tapez simplement la commande suivante et appuyez sur la touche `Entrée` :
    ```bash
    go run .
    ```
4.  Attendez quelques secondes. Quand vous voyez le message `Serveur démarré sur http://localhost:8080`, c'est bon !
5.  Ouvrez votre navigateur internet (Chrome, Firefox, Safari...) et tapez l'adresse :
    👉 **http://localhost:8080**

## 🕹️ Comment jouer ?

1.  **Inscrivez-vous** (c'est rapide) et connectez-vous.
2.  Choisissez votre jeu (**BlindTest** ou **Petit Bac**).
3.  **Créez une salle** : Vous obtiendrez un code (par exemple `ABCD`).
4.  Donnez ce code à vos amis pour qu'ils rejoignent votre salle via le bouton "Rejoindre".
5.  Une fois tout le monde prêt, le créateur de la salle clique sur **Configuration** pour choisir les règles (playlist, temps, nombre de tours) et lance la partie !

Amusez-vous bien ! 🚀

## 📂 Architecture du projet

Pour les curieux, voici comment est rangé notre code :

*   **`main.go`** : C'est la porte d'entrée, le fichier qui démarre tout.
*   **`games/`** : C'est le cerveau des jeux.
    *   `blindtest.go` : Les règles du BlindTest.
    *   `petitBac.go ` & ` petibac_resultats.go ` : Les règles du Petit Bac.
    *   `petitbac_gestion.go` : La gestion de la partie.
    *   `deezer_service.go` : Le lien avec Deezer pour récupérer les musiques.
*   **`src/`** : La machinerie du serveur.
    *   `server.go` & `handler.go` : Ils gèrent les pages web et les actions des joueurs.
    *   `client.go` & `hub.go` : Ils gèrent la communication en direct (WebSockets) pour que tout le monde joue en même temps.
    *   `database.go` : La gestion de la base de données (utilisateurs, scores).
    *   `middleware.go` : La gestion de la connexion.
*   **`pages/`** : Les fichiers HTML (ce que vous voyez à l'écran).
*   **`static/`** : Les fichiers CSS (pour faire joli) et les images.
*   **`script/`** : Le code JavaScript qui tourne sur votre navigateur.

---
*Projet réalisé par des étudiants en Bachelor 1 à Ynov Campus.*
# Scalingo
