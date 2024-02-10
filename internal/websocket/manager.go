package websocket

import (
	"context"
	"encoding/json"
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
	config            *conf.GlobalConfiguration `required:"true"`
	rdb               *redis.Client             `required:"true"`
	db                gocqlx.Session            `required:"true"`
	clients           ClientList
	handlers          map[string]EventHandler
	subscribeHandlers map[string]SubscribeEventHandler
	sync.RWMutex
}

func NewManager(ctx context.Context, config *conf.GlobalConfiguration, redisDb *redis.Client, db gocqlx.Session) *Manager {
	m := &Manager{
		rdb:               redisDb,
		db:                db,
		config:            config,
		clients:           make(ClientList),
		handlers:          make(map[string]EventHandler),
		subscribeHandlers: make(map[string]SubscribeEventHandler),
	}
	m.setupEventHandlers()
	m.setupSubscribeEventHandlers()
	go m.setupAndListenRedisSubscriber()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendDirectMessage] = SendMessageHandler
}

func (m *Manager) setupSubscribeEventHandlers() {
	m.subscribeHandlers[SubscribeEventSendDirectMessage] = SubscribeEventSendDirectMessageHandler
}

func (m *Manager) setupAndListenRedisSubscriber() {
	ctx := context.Background()
	sub := m.rdb.Subscribe(ctx, m.config.SERVER.Id)
	for {
		msg, err := sub.ReceiveMessage(ctx)
		if err != nil {
			logrus.Fatalf("error on receeving subscribe message: %v", err)
		}

		var request SubscribeEvent
		if err := json.Unmarshal([]byte(msg.Payload), &request); err != nil {
			logrus.Errorf("error marshalling message: %v", err)
			continue
		}

		logrus.Debugf("new subscribe event")

		if err := m.routeSubscribeEvent(request); err != nil {
			logrus.Errorf("error handeling subscribe event: %v", err)
		}

	}
}

func (m *Manager) routeEvent(ginCtx *gin.Context, event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(ginCtx, event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}

func (m *Manager) routeSubscribeEvent(event SubscribeEvent) error {
	if handler, ok := m.subscribeHandlers[event.Type]; ok {
		if err := handler(event, m); err != nil {
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

	m.clients[client.connectionId] = client
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client.connectionId]; ok {
		client.connection.Close()
		delete(m.clients, client.connectionId)
	}
}

func (m *Manager) ServeWS(ginCtx *gin.Context) {
	conn, err := websocketUpgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)

	if err != nil {
		conn.Close()
		logrus.Errorf("failed to upgrage the connection: %+v", err)
		utils.HandleHttpError(utils.InternalServerError("Failed to upgrage the connection: %+v", err), ginCtx)
		return
	}

	client, err := NewClient(ginCtx, conn, m)
	if err != nil {
		conn.Close()
		logrus.Errorf("failed to setup the connection: %+v", err)
		utils.HandleHttpError(utils.InternalServerError("Failed to setup the connection: %+v", err), ginCtx)
		return
	}

	m.addClient(client)

	go client.readMessage(ginCtx)
	go client.writeMessages(ginCtx)
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
