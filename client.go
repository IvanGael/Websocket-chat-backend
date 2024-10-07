package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	nickname string
	color    string
	room     *ChatRoom
	typing   bool
	send     chan Message
	ctx      context.Context
	cancel   context.CancelFunc
}

func (c *Client) readPump() {
	defer func() {
		c.room.leave <- c
		c.cancel()
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Printf("Error reading message: %v", err)
				}
				return
			}

			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				logger.Println("Error unmarshalling message:", err)
				continue
			}

			if msg.Type == "typing" {
				c.typing = msg.Typing
				c.room.broadcastTypingStatus(c)
				continue
			}

			decryptedMsg, err := xorDecrypt(msg.Message)
			if err != nil {
				logger.Println("Error decrypting message:", err)
				continue
			}
			msg.Message = decryptedMsg
			msg.Username = c.nickname
			msg.Color = c.color
			msg.Timestamp = time.Now()

			select {
			case c.room.broadcast <- msg:
			case <-c.ctx.Done():
				return
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			json.NewEncoder(w).Encode(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				json.NewEncoder(w).Encode(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}
