package websocket

import (
	"encoding/json"
	"log"
	"time"

	_websocket "github.com/gorilla/websocket"
)

var pongWait = 10 * time.Second
var pingInterval = (pongWait * 9) / 10

type ClientList map[*Client]bool

type Client struct {
	connection *_websocket.Conn
	manager    *Manager
	egress     chan Event
	chatroom   string
}

func NewClient(conn *_websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
	}
}

func (c *Client) readMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()

	c.connection.SetReadLimit(512)

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}

	c.connection.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if _websocket.IsUnexpectedCloseError(err, _websocket.CloseGoingAway, _websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling message: %v", err)
			break
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Error handeling Message: ", err)
		}
	}
}

func (c *Client) writeMessages() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(_websocket.CloseMessage, nil); err != nil {

					log.Println("connection closed: ", err)
				}

				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return
			}

			if err := c.connection.WriteMessage(_websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
			log.Println("sent message")
		case <-ticker.C:
			log.Println("ping")
			if err := c.connection.WriteMessage(_websocket.PingMessage, []byte{}); err != nil {
				log.Println("writemsg: ", err)
				return
			}
		}

	}
}

func (c *Client) pongHandler(pongMsg string) error {
	log.Println("pong")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
