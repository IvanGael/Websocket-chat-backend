package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ChatRoom struct {
	id         string
	clients    map[*Client]bool
	broadcast  chan Message
	join       chan *Client
	leave      chan *Client
	expiration time.Time
	messages   []Message
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func (cr *ChatRoom) run() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-cr.ctx.Done():
			return
		case client := <-cr.join:
			cr.mutex.Lock()
			cr.clients[client] = true
			cr.mutex.Unlock()
			cr.broadcastUserCount()
		case client := <-cr.leave:
			cr.removeClient(client)
			cr.broadcastUserCount()
			if client != nil {
				cr.broadcastTypingStatus(client)
			}
		case message := <-cr.broadcast:
			encryptedMsg, err := xorEncrypt(message.Message)
			if err != nil {
				logger.Println("Error encrypting message:", err)
				continue
			}
			message.Message = encryptedMsg
			cr.mutex.Lock()
			cr.messages = append(cr.messages, message)
			for client := range cr.clients {
				select {
				case client.send <- message:
				default:
					cr.mutex.Unlock()
					cr.removeClient(client)
					cr.mutex.Lock()
				}
			}
			cr.mutex.Unlock()
		case <-ticker.C:
			cr.deleteOldMessages()
		}
	}
}

func (cr *ChatRoom) broadcastUserCount() {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	userCount := len(cr.clients)
	userCountStr := fmt.Sprintf("%d", userCount)

	encryptedCount, err := xorEncrypt(userCountStr)
	if err != nil {
		logger.Println("Error encrypting user count:", err)
		return
	}

	message := Message{
		Type:    "user_count",
		Message: encryptedCount,
	}

	for client := range cr.clients {
		select {
		case client.send <- message:
		default:
			cr.mutex.RUnlock()
			cr.removeClient(client)
			cr.mutex.RLock()
		}
	}
}

func (cr *ChatRoom) broadcastTypingStatus(typingClient *Client) {
	if typingClient == nil {
		return
	}

	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	typingMessage := Message{
		Type:     "typing",
		Username: typingClient.nickname,
		Typing:   typingClient.typing,
	}

	for client := range cr.clients {
		if client != typingClient {
			select {
			case client.send <- typingMessage:
			default:
				cr.mutex.RUnlock()
				cr.removeClient(client)
				cr.mutex.RLock()
			}
		}
	}
}

func (cr *ChatRoom) deleteOldMessages() {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	now := time.Now()
	messageRetentionPeriod := 24 * time.Hour

	var messagesToKeep []Message
	for _, msg := range cr.messages {
		if now.Sub(msg.Timestamp) < messageRetentionPeriod {
			messagesToKeep = append(messagesToKeep, msg)
		}
	}

	cr.messages = messagesToKeep
	logger.Printf("Deleted %d old messages", len(cr.messages)-len(messagesToKeep))
}

func (cr *ChatRoom) removeClient(client *Client) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	if _, ok := cr.clients[client]; ok {
		delete(cr.clients, client)
		close(client.send)
		client.cancel()
	}
}
