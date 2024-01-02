package websocket

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	_websocket "github.com/gorilla/websocket"
)

var websocketUpgrader = _websocket.Upgrader{
	CheckOrigin:     checkOrigin,
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var ErrEventNotSupported = errors.New("this event type is not supported")

type Manager struct {
	clients ClientList
	sync.RWMutex
	handlers map[string]EventHandler
}

func NewManager() *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler
	m.handlers[EventChangeRoom] = ChatRoomHandler
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}

func (m *Manager) ServeWS(ctx *gin.Context) {

	log.Println("New Connection...")

	conn, err := websocketUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)

	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn, m)

	m.addClient(client)

	go client.readMessage()
	go client.writeMessages()
}

func checkOrigin(r *http.Request) bool {

	// Grab the request origin
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:8080":
		return true
	default:
		return false
	}
}
