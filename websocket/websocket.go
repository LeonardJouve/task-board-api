package websocket

import (
	"encoding/json"
	"time"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var messageChannel = make(chan *Message)
var registerChannel = make(chan RegistrationMessage)
var unregisterChannel = make(chan RegistrationMessage)

var connections = make(map[string]*websocket.Conn)
var websocketChannels = make(map[string]map[string]struct{})

type WebsocketMessage = map[string]interface{}

type RegistrationMessage = struct {
	SessionId  string
	UserId     uint
	Connection *websocket.Conn
}

type Message = struct {
	Type      int
	Value     WebsocketMessage
	SessionId string
}

const (
	JOIN_TYPE       = "join"
	LEAVE_TYPE      = "leave"
	REGISTER_TYPE   = "register"
	UNREGISTER_TYPE = "unregister"
)

func close(connection *websocket.Conn) {
	if connection == nil {
		return
	}

	sessionId, ok := getSessionId(connection)
	if ok {
		unregisterChannel <- RegistrationMessage{
			SessionId:  sessionId,
			Connection: connection,
		}
	}

	connection.Close()
}

func getSessionId(connection *websocket.Conn) (string, bool) {
	sessionId, ok := connection.Locals("sessionId").(string)
	if !ok {
		return "", false
	}

	return sessionId, true
}

func HandleUpgrade(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	return c.Next()
}

var HandleSocket = websocket.New(func(connection *websocket.Conn) {
	sessionId, ok := getSessionId(connection)
	if !ok {
		close(connection)
		return
	}

	registerChannel <- RegistrationMessage{
		SessionId:  sessionId,
		Connection: connection,
	}
	defer close(connection)

	var (
		messageType  int
		messageValue []byte
		err          error
	)
	for {
		if connection == nil {
			break
		}

		if messageType, messageValue, err = connection.ReadMessage(); err != nil {
			break
		}

		var unmarshaledMessage WebsocketMessage
		err := json.Unmarshal(messageValue, &unmarshaledMessage)
		_, ok := unmarshaledMessage["type"]
		if err != nil || !ok {
			break
		}

		messageChannel <- &Message{
			Type:      messageType,
			Value:     unmarshaledMessage,
			SessionId: sessionId,
		}
	}

}, websocket.Config{
	HandshakeTimeout: 10 * time.Second,
	ReadBufferSize:   2048,
	WriteBufferSize:  2048,
})

func Process() {
	for {
		select {
		case message := <-models.HookChannel:
			for _, connection := range connections {
				writeTextMessage(connection, message)
			}
		case message := <-messageChannel:
			switch message.Value["type"] {
			case JOIN_TYPE:
				channel, ok := getWebsocketMessageString(message.Value, "channel")
				if !ok {
					continue
				}

				websocketChannels[channel][message.SessionId] = struct{}{}
			case LEAVE_TYPE:
				channel, ok := getWebsocketMessageString(message.Value, "channel")
				if !ok {
					continue
				}

				if _, ok := websocketChannels[channel]; !ok {
					continue
				}

				delete(websocketChannels[channel], message.SessionId)
			}
		case message := <-registerChannel:
			for sessionId, conn := range connections {
				if sessionId == message.SessionId {
					continue
				}

				writeTextMessage(conn, WebsocketMessage{
					"type":   REGISTER_TYPE,
					"userId": message.UserId,
				})
			}

			connections[message.SessionId] = message.Connection
		case message := <-unregisterChannel:
			for channel := range websocketChannels {
				messageChannel <- &Message{
					Type: websocket.TextMessage,
					Value: WebsocketMessage{
						"type":    LEAVE_TYPE,
						"channel": channel,
					},
				}
			}

			delete(connections, message.SessionId)

			for _, conn := range connections {
				writeTextMessage(conn, WebsocketMessage{
					"type":   UNREGISTER_TYPE,
					"userId": message.UserId,
				})
			}
		}
	}
}

func writeTextMessage(connection *websocket.Conn, message WebsocketMessage) bool {
	marshaledMessage, err := json.Marshal(message)
	if err != nil {
		return false
	}

	if err := connection.WriteMessage(websocket.TextMessage, marshaledMessage); err != nil {
		return false
	}

	return true
}

func getWebsocketMessageString(websocketMessage WebsocketMessage, key string) (string, bool) {
	value, ok := websocketMessage[key]
	if !ok {
		return "", false
	}

	channel, ok := value.(string)
	if !ok {
		return "", false
	}

	return channel, true
}
