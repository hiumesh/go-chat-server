package websocket

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	_websocket "github.com/gorilla/websocket"
	"github.com/hiumesh/go-chat-server/internal/conf"
	"github.com/hiumesh/go-chat-server/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/scylladb/gocqlx/v2"
	"github.com/sirupsen/logrus"
)

var websocketUpgrader = _websocket.Upgrader{
	CheckOrigin:     checkOrigin,
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var ErrEventNotSupported = errors.New("this event type is not supported")

type Manager struct {
	clients  ClientList
	handlers map[string]EventHandler
	sync.RWMutex
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

func (m *Manager) ServeWS(ginCtx *gin.Context, config *conf.GlobalConfiguration, db gocqlx.Session, redisDb *redis.Client) {
	conn, err := websocketUpgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)

	if err != nil {
		conn.Close()
		logrus.Fatalf("Failed to upgrage the connection: %+v", err)
		utils.HandleHttpError(utils.InternalServerError("Failed to upgrage the connection: %+v", err), ginCtx)
		return
	}

	client, err := NewClient(ginCtx, conn, m, config, redisDb)
	if err != nil {
		conn.Close()
		logrus.Fatalf("Failed to setup the connection: %+v", err)
		utils.HandleHttpError(utils.InternalServerError("Failed to setup the connection: %+v", err), ginCtx)
		return
	}

	m.addClient(client)

	go client.readMessage(ginCtx, config, redisDb)
	go client.writeMessages(ginCtx, config, redisDb)
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:8080":
		return true
	default:
		return false
	}
}
