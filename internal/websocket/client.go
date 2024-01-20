package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_websocket "github.com/gorilla/websocket"
	"github.com/hiumesh/go-chat-server/internal/conf"
	"github.com/hiumesh/go-chat-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

var pongWait = 10 * time.Second
var pingInterval = (pongWait * 9) / 10
var redisPingInterval = 60000000000

type ClientList map[*Client]bool

type Client struct {
	connectionId string
	connection   *_websocket.Conn
	manager      *Manager
	egress       chan Event
	chatroom     string
}

func NewClient(ctx *gin.Context, conn *_websocket.Conn, manager *Manager, config *conf.GlobalConfiguration, redisDb *redis.Client) (*Client, error) {
	uniqueConnectionId := utils.GetRequestID(ctx)
	if uniqueConnectionId == "" {
		return nil, errors.New("unique id not found")
	}
	claims := utils.GetClaims(ctx)
	if claims == nil {
		return nil, errors.New("claims not found")
	}

	log.Println(uniqueConnectionId, claims)

	key := claims.Id
	value := config.SERVER.Id + " " + uniqueConnectionId

	if err := redisDb.ZRemRangeByScore(ctx, key, "-inf", strconv.Itoa(int(time.Now().UnixMilli()-300000))).Err(); err != nil {
		return nil, err
	}

	count, err := redisDb.ZCount(ctx, key, "-inf", "+inf").Result()
	if err != nil {
		return nil, err
	}

	mspu, err := strconv.Atoi(config.SERVER.MaxPerUserConnection)

	if err != nil {
		return nil, err
	}
	if count >= int64(mspu) {
		return nil, errors.New("maximum_connection_limit_reached")
	}
	if err := redisDb.ZAdd(ctx, key, redis.Z{Score: float64(time.Now().UnixMilli()), Member: value}).Err(); err != nil {
		return nil, err
	}

	return &Client{
		connectionId: uniqueConnectionId,
		connection:   conn,
		manager:      manager,
		egress:       make(chan Event),
	}, nil
}

func (c *Client) readMessage(ctx *gin.Context, config *conf.GlobalConfiguration, redisDb *redis.Client) {
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

func (c *Client) writeMessages(ctx *gin.Context, config *conf.GlobalConfiguration, redisDb *redis.Client) {
	key := "userId"
	value := config.SERVER.Id + " " + c.connectionId
	ticker := time.NewTicker(pingInterval)
	redisPingTicker := time.NewTicker(time.Duration(redisPingInterval))
	defer func() {
		ticker.Stop()
		redisPingTicker.Stop()
		redisDb.SRem(ctx, key, value).Err()
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
			// log.Println("ping")
			if err := c.connection.WriteMessage(_websocket.PingMessage, []byte{}); err != nil {
				log.Println("writemsg: ", err)
				return
			}

		case <-redisPingTicker.C:
			log.Println("redis ping")
			if err := redisDb.ZIncrBy(ctx, key, 60000, value).Err(); err != nil {
				log.Println("redis ping error: ", err)
			}
		}

	}
}

func (c *Client) pongHandler(pongMsg string) error {
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
