package src

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
	Name string
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		c.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func StartNewRound() {
	go func() {
		time.Sleep(5 * time.Second)

		CurrentGame.NextSong()

		if CurrentGame.IsFinished {
			hub.Broadcast <- []byte(`{"type": "redirect", "url": "/blind-test/results"}`)
		} else {
			nextSong := CurrentGame.GetCurrentSong()
			if nextSong != nil {
				updateMsg := fmt.Sprintf(`{"type": "audio", "url": "%s", "msg": "C'est parti !"}`, nextSong.File)
				hub.Broadcast <- []byte(updateMsg)

				CurrentGame.StartRoundTimer(func() {
					timeoutMsg := fmt.Sprintf(`{"type": "msg", "text": "Temps écoulé ! La réponse était : %s - %s. Prochaine dans 5s..."}`,
						nextSong.Artist, nextSong.Title)
					hub.Broadcast <- []byte(timeoutMsg)

					StartNewRound()
				})
			}
		}
	}()
}

func (c *Client) readPump() {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		userGuess := string(msg)

		if CurrentGame != nil && len(CurrentGame.Playlist) > 0 {
			points, foundType := CurrentGame.CheckAnswer(userGuess, c.Name)

			if points > 0 {
				currentSong := CurrentGame.GetCurrentSong()

				var msgText string
				switch foundType {
				case "artist":
					if CurrentGame.TitleFound {
						msgText = fmt.Sprintf("%s a trouvé l'artiste (+1pt) ! C'est fini pour cette chanson.", c.Name)
					} else {
						msgText = fmt.Sprintf("%s a trouvé l'artiste (+1pt) ! Il reste le titre.", c.Name)
					}
				case "title":
					if CurrentGame.ArtistFound {
						msgText = fmt.Sprintf("%s a trouvé le titre (+2pts) ! C'est fini pour cette chanson.", c.Name)
					} else {
						msgText = fmt.Sprintf("%s a trouvé le titre (+2pts) ! Il reste l'artiste.", c.Name)
					}
				case "both":
					msgText = fmt.Sprintf("%s a tout trouvé (+3pts) !", c.Name)
				}

				hub.Broadcast <- []byte(fmt.Sprintf(`{"type": "msg", "text": "%s"}`, msgText))

				if CurrentGame.ArtistFound && CurrentGame.TitleFound {
					announceMsg := fmt.Sprintf(`{"type": "msg", "text": "La chanson était : %s - %s. Prochaine dans 5s..."}`,
						currentSong.Artist, currentSong.Title)
					hub.Broadcast <- []byte(announceMsg)

					// Utilisation de la fonction centralisée
					StartNewRound()
				}

			} else {
				chatMsg := fmt.Sprintf(`{"type": "msg", "text": "%s: %s"}`, c.Name, userGuess)
				hub.Broadcast <- []byte(chatMsg)
			}
		}
	}
}
