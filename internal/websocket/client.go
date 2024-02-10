package websocket

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_websocket "github.com/gorilla/websocket"
	"github.com/hiumesh/go-chat-server/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var pongWait = 10 * time.Second
var pingInterval = (pongWait * 9) / 10
var redisPingInterval = 60000000000

type ClientList map[string]*Client

type Client struct {
	claims       *utils.AccessTokenClaims
	connectionId string
	connection   *_websocket.Conn
	manager      *Manager
	egress       chan Event
	chatroom     string
}

func NewClient(ctx *gin.Context, conn *_websocket.Conn, m *Manager) (*Client, error) {
	uniqueConnectionId := utils.GetRequestID(ctx)
	if uniqueConnectionId == "" {
		return nil, errors.New("unique id not found")
	}
	claims := utils.GetClaims(ctx)
	if claims == nil {
		return nil, errors.New("claims not found")
	}

	key := claims.Subject
	value := m.config.SERVER.Id + " " + uniqueConnectionId

	if err := m.rdb.ZRemRangeByScore(ctx, key, "-inf", strconv.Itoa(int(time.Now().UnixMilli()-300000))).Err(); err != nil {
		return nil, err
	}

	count, err := m.rdb.ZCount(ctx, key, "-inf", "+inf").Result()
	if err != nil {
		return nil, err
	}

	mspu, err := strconv.Atoi(m.config.SERVER.MaxPerUserConnection)

	if err != nil {
		return nil, err
	}
	if count >= int64(mspu) {
		return nil, errors.New("maximum connection limit reached")
	}
	if err := m.rdb.ZAdd(ctx, key, redis.Z{Score: float64(time.Now().UnixMilli()), Member: value}).Err(); err != nil {
		return nil, err
	}

	return &Client{
		claims:       claims,
		connectionId: uniqueConnectionId,
		connection:   conn,
		manager:      m,
		egress:       make(chan Event),
	}, nil
}

func (c *Client) readMessage(ctx *gin.Context) {
	defer func() {
		c.manager.removeClient(c)

		logrus.Debugf("exiting reader: %v", c.connectionId)
	}()

	c.connection.SetReadLimit(512)
	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logrus.Errorf("error configuring the connection: %v", err)
		return
	}
	c.connection.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if _websocket.IsUnexpectedCloseError(err, _websocket.CloseGoingAway, _websocket.CloseAbnormalClosure) {
				logrus.Errorf("error reading message: %v", err)
			}
			return
		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			logrus.Errorf("error marshalling message: %v", err)
			continue
		}

		if err := c.manager.routeEvent(ctx, request, c); err != nil {
			logrus.Errorf("error handeling message: %v", err)
		}
	}
}

func (c *Client) writeMessages(ctx *gin.Context) {
	claims := utils.GetClaims(ctx)
	value := c.manager.config.SERVER.Id + " " + c.connectionId
	ticker := time.NewTicker(pingInterval)
	redisPingTicker := time.NewTicker(time.Duration(redisPingInterval))
	defer func() {
		ticker.Stop()
		redisPingTicker.Stop()
		c.manager.rdb.SRem(ctx, claims.Subject, value).Err()
		c.manager.removeClient(c)

		logrus.Debugf("exiting writer: %v", c.connectionId)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(_websocket.CloseMessage, nil); err != nil {
					logrus.Errorf("exiting the writer: %v", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				logrus.Errorf("error marshaling the socket message: %v", err)
				continue
			}

			if err := c.connection.WriteMessage(_websocket.TextMessage, data); err != nil {
				logrus.Errorf("error writing the socket message: %v", err)
			}
			logrus.Debugf("message sent")
		case <-ticker.C:
			if err := c.connection.WriteMessage(_websocket.PingMessage, []byte{}); err != nil {
				logrus.Errorf("ping message fail: %v", err)
				return
			}

		case <-redisPingTicker.C:
			if err := c.manager.rdb.ZIncrBy(ctx, claims.Subject, 60000, value).Err(); err != nil {
				logrus.Errorf("redis ping fail: %v", err)
			}
		}

	}
}

func (c *Client) pongHandler(pongMsg string) error {
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
