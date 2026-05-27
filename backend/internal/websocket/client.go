package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", c.Username, err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error unmarshaling WS message from %s: %v", c.Username, err)
			continue
		}

		c.handleMessage(wsMsg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg WSMessage) {
	switch msg.Type {
	case TypeTyping:
		msg.SenderID = c.UserID
		msg.SenderUsername = c.Username
		data, _ := json.Marshal(msg)
		c.Hub.BroadcastToRoom(msg.ConversationID, data, c.UserID)

	case TypeStopTyping:
		msg.SenderID = c.UserID
		msg.SenderUsername = c.Username
		data, _ := json.Marshal(msg)
		c.Hub.BroadcastToRoom(msg.ConversationID, data, c.UserID)

	case TypeReadReceipt:
		msg.SenderID = c.UserID
		msg.SenderUsername = c.Username
		// Persist the read receipt
		if c.Hub.messageReadRepo != nil {
			if err := c.Hub.messageReadRepo.Upsert(context.Background(), msg.ConversationID, c.UserID); err != nil {
				log.Printf("Error persisting read receipt: %v", err)
			}
		}
		data, _ := json.Marshal(msg)
		c.Hub.BroadcastToRoom(msg.ConversationID, data, c.UserID)

	default:
		log.Printf("Unknown WS message type from %s: %s", c.Username, msg.Type)
	}
}
