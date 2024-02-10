package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hiumesh/go-chat-server/internal/models"
	"github.com/hiumesh/go-chat-server/internal/utils"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(ginCtx *gin.Context, event Event, c *Client) error

const EventSendDirectMessage = "direct_message"
const EventNewMessage = "new_message"
const EventChangeRoom = "change_room"

type SendDirectMessageEvent struct {
	Body string `json:"body"`
	From string `json:"from"`
	To   string `json:"to"`
}

type NewMessageEvent struct {
	SendDirectMessageEvent
	Sent time.Time `json:"sent"`
}

func SendMessageHandler(ctx *gin.Context, event Event, c *Client) error {
	manager := c.manager
	var chatevent SendDirectMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return errors.New("bad payload in request")
	}

	claims := utils.GetClaims(ctx)

	if claims == nil {
		return errors.New("claims not found")
	}

	senderUserActiveConnections, err := manager.rdb.ZRange(ctx, claims.Subject, 0, -1).Result()
	if err != nil {
		return err
	}
	reciverUserActiveConnections, err := manager.rdb.ZRange(ctx, chatevent.To, 0, -1).Result()
	if err != nil {
		return err
	}

	dbMessage := models.Message{
		UserId: claims.Subject,
		Body:   chatevent.Body,
	}

	if err := models.InsertMessage(manager.db, &dbMessage); err != nil {
		return err
	}

	var broadMessage NewMessageEvent

	broadMessage.Sent = time.Now()
	broadMessage.Body = chatevent.Body
	broadMessage.From = claims.Id
	broadMessage.To = chatevent.To

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return err
	}

	var outgoingEvent Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventNewMessage

	for _, connectionStr := range reciverUserActiveConnections {
		split := strings.Split(connectionStr, " ")
		serverId := split[0]
		connectionId := split[1]

		if manager.config.SERVER.Id == serverId {
			connection := c.manager.clients[connectionId]
			connection.egress <- outgoingEvent
		} else {
			manager.rdb.Publish(ctx, serverId, event)
		}
	}

	for _, connectionStr := range senderUserActiveConnections {
		split := strings.Split(connectionStr, " ")
		serverId := split[0]
		connectionId := split[1]

		println(connectionId, serverId)
	}

	return nil

}

type ChangeRoomEvent struct {
	Name string `json:"name"`
}

func ChatRoomHandler(ginCtx *gin.Context, event Event, c *Client) error {

	var changeRoomEvent ChangeRoomEvent
	if err := json.Unmarshal(event.Payload, &changeRoomEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	c.chatroom = changeRoomEvent.Name

	return nil
}
