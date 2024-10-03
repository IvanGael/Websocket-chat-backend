package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool) // Connected clients
var broadcast = make(chan Message)           // Channel for broadcasting messages

// Message defines the structure of the messages exchanged
type Message struct {
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Typing    bool   `json:"typing"` // Indicates if the user is typing
}

func main() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", HandleConnections)

	// Start the goroutine for handling messages
	go HandleMessages()

	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// HandleConnections handles new WebSocket requests from clients
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil) // Upgrade HTTP to WebSocket
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close() // Close the WebSocket connection when the function returns

	clients[ws] = true // Register the new client

	for {
		var msg Message
		// Read a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws) // Remove the client from the list if there is an error
			break
		}

		// Add timestamp to the message
		msg.Timestamp = time.Now().Format(time.RFC3339)

		// Send the message to the broadcast channel
		broadcast <- msg
	}
}

// HandleMessages broadcasts incoming messages to all clients
func HandleMessages() {
	for {
		// Get the next message from the broadcast channel
		msg := <-broadcast

		// Send it to every connected client
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
