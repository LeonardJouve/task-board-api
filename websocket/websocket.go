package websocket

import (
	"encoding/json"
	"time"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type SessionId = string
type UserId = uint
type Channel = string
type WebsocketType = int
type MessageType = string
type WebsocketChannel = map[SessionId]struct{}
type WebsocketMessage = map[string]interface{}

type WebsocketConnection = struct {
	UserId     UserId
	Connection *websocket.Conn
}

type RegistrationMessage = struct {
	SessionId  SessionId
	UserId     UserId
	Connection *websocket.Conn
}

type Message = struct {
	WebsocketType WebsocketType
	Channel       Channel
	MessageType   MessageType
	Message       WebsocketMessage
	SessionId     SessionId
}

const (
	JOIN_TYPE       = "join"
	LEAVE_TYPE      = "leave"
	REGISTER_TYPE   = "register"
	UNREGISTER_TYPE = "unregister"
)

var messageChannel = make(chan *Message)
var registerChannel = make(chan RegistrationMessage)
var unregisterChannel = make(chan RegistrationMessage)

var websocketConnections = make(map[SessionId]WebsocketConnection)
var websocketChannels = make(map[Channel]WebsocketChannel)

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

	for {
		if connection == nil {
			break
		}

		websocketMessageType, message, err := connection.ReadMessage()
		if err != nil {
			break
		}

		var unmarshaledMessage WebsocketMessage
		err = json.Unmarshal(message, &unmarshaledMessage)
		messageType, ok := unmarshaledMessage["type"].(string)
		if err != nil || !ok {
			break
		}
		channel, ok := unmarshaledMessage["channel"].(string)
		if err != nil || !ok {
			break
		}

		messageChannel <- &Message{
			WebsocketType: websocketMessageType,
			Channel:       channel,
			MessageType:   messageType,
			Message:       unmarshaledMessage,
			SessionId:     sessionId,
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
		case hookMessage := <-models.HookChannel:
			for _, websocketConnection := range websocketConnections {
				// send channel message
				writeTextMessage(websocketConnection.Connection, hookMessage.Type, hookMessage.Message)
			}
		case message := <-messageChannel:
			switch message.MessageType {
			case JOIN_TYPE:
				// if can join channel

				websocketChannels[message.Channel][message.SessionId] = struct{}{}

				// send channel message
			case LEAVE_TYPE:
				if _, ok := websocketChannels[message.Channel]; !ok {
					continue
				}

				delete(websocketChannels[message.Channel], message.SessionId)

				// send channel message
			}
		case message := <-registerChannel:
			for sessionId, websocketConnection := range websocketConnections {
				if sessionId == message.SessionId {
					continue
				}

				writeTextMessage(websocketConnection.Connection, REGISTER_TYPE, WebsocketMessage{
					"userId": message.UserId,
				})
			}

			websocketConnections[message.SessionId] = WebsocketConnection{
				UserId:     message.UserId,
				Connection: message.Connection,
			}
		case message := <-unregisterChannel:
			for channel, members := range websocketChannels {
				if _, ok := members[message.SessionId]; !ok {
					continue
				}

				messageChannel <- &Message{
					WebsocketType: websocket.TextMessage,
					Channel:       channel,
					MessageType:   LEAVE_TYPE,
					SessionId:     message.SessionId,
				}
			}

			delete(websocketConnections, message.SessionId)

			for _, websocketConnection := range websocketConnections {
				writeTextMessage(websocketConnection.Connection, UNREGISTER_TYPE, WebsocketMessage{
					"userId": message.UserId,
				})
			}
		}
	}
}

func writeTextMessage(connection *websocket.Conn, messageType string, message WebsocketMessage) bool {
	message["type"] = messageType

	marshaledMessage, err := json.Marshal(message)
	if err != nil {
		return false
	}

	if err := connection.WriteMessage(websocket.TextMessage, marshaledMessage); err != nil {
		return false
	}

	return true
}
