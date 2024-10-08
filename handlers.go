package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	logger = log.New(os.Stdout, "CHAT: ", log.Ldate|log.Ltime|log.Lshortfile)
)

type Message struct {
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Color     string    `json:"color"`
	Typing    bool      `json:"typing"`
}

var (
	rooms      = make(map[string]*ChatRoom)
	roomsMutex sync.RWMutex
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	fullRoomID := r.URL.Query().Get("room")
	if fullRoomID == "" {
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	if !isValidRoomID(fullRoomID) {
		http.Error(w, "Invalid Room ID", http.StatusBadRequest)
		return
	}

	baseRoomID := extractBaseRoomID(fullRoomID)

	roomsMutex.RLock()
	room, exists := rooms[baseRoomID]
	roomsMutex.RUnlock()
	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Println("Error upgrading connection:", err)
		return
	}

	ctx, cancel := context.WithCancel(room.ctx)
	client := &Client{
		conn:     conn,
		nickname: generateNickname(),
		color:    generateColor(),
		room:     room,
		send:     make(chan Message, 256),
		ctx:      ctx,
		cancel:   cancel,
	}

	room.join <- client

	go client.writePump()
	go client.readPump()
}

func createRoom(w http.ResponseWriter, r *http.Request) {
	fullRoomID := generateUniqueRoomID()
	baseRoomID := extractBaseRoomID(fullRoomID)
	ctx, cancel := context.WithCancel(context.Background())
	newRoom := &ChatRoom{
		id:         baseRoomID,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message, 256),
		join:       make(chan *Client, 10),
		leave:      make(chan *Client, 10),
		expiration: time.Now().Add(30 * time.Minute),
		mutex:      sync.RWMutex{},
		ctx:        ctx,
		cancel:     cancel,
	}

	roomsMutex.Lock()
	rooms[baseRoomID] = newRoom
	roomsMutex.Unlock()

	go newRoom.run()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"roomID": fullRoomID})
}

func handleEncrypt(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	encrypted, err := xorEncrypt(data.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"encrypted": encrypted})
}

func handleDecrypt(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	decrypted, err := xorDecrypt(data.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"decrypted": decrypted})
}
