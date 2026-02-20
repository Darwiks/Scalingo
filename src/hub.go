package src

import "fmt"

var hub *Hub

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			fmt.Println(client.Name + " connecté au WebSocket")
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				fmt.Println(client.Name + " déconnecté du WebSocket")
			}
		case message := <-h.Broadcast:
			fmt.Printf("Broadcasting message to %d clients: %s\n", len(h.Clients), string(message))
			for client := range h.Clients {
				select {
				case client.Send <- message:
					// Message envoyé avec succès
				default:
					// Le canal est plein, fermer la connexion
					fmt.Printf("Client %s a un canal Send plein, déconnexion\n", client.Name)
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
