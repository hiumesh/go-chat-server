package websocket

import (
	"log"

	websocketx "github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocketx.Conn
	manager    *Manager
}

func NewClient(conn *websocketx.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
	}
}

func (c *Client) readMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		messageType, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocketx.IsUnexpectedCloseError(err, websocketx.CloseGoingAway, websocketx.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}
		log.Println("MessageType: ", messageType)
		log.Println("Payload: ", string(payload))
	}
}
