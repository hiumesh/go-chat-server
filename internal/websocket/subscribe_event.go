package websocket

import (
	"encoding/json"
	"fmt"
)

type SubscribeEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SubscribeEventHandler func(event SubscribeEvent, m *Manager) error

type SendDirectMessageSubscribeEvent struct {
	Sent    string `json:"sent"`
	Message string `json:"message"`
	From    string `json:"from"`
	To      string `json:"to"`
}

const SubscribeEventSendDirectMessage = "direct_message"

func SubscribeEventSendDirectMessageHandler(event SubscribeEvent, m *Manager) error {
	var chatevent SendDirectMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	data, err := json.Marshal(chatevent)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	var outgoingEvent Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventNewMessage

	connection := m.clients[chatevent.To]
	connection.egress <- outgoingEvent

	return nil
}
