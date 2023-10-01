package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LeonardJouve/task-board-api/dotenv"
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SessionId = string
type UserId = uint
type Channel = string
type WebsocketType = int
type MessageType = string
type WebsocketChannel = map[SessionId]struct{}
type WebsocketMessage = map[string]interface{}
type PongChannel = chan struct{}

type WebsocketConnection struct {
	SessionId  SessionId
	User       models.User
	Connection *websocket.Conn
	sync.Mutex
}

type Message struct {
	Channel             Channel
	MessageType         MessageType
	Message             WebsocketMessage
	WebsocketConnection *WebsocketConnection
}

const (
	JOIN_TYPE            = "join"
	LEAVE_TYPE           = "leave"
	REGISTER_TYPE        = "register"
	UNREGISTER_TYPE      = "unregister"
	PING_TYPE            = "ping"
	PONG_TYPE            = "pong"
	BOARD_CHANNEL_PREFIX = "board_"
)

var textChannel = make(chan *Message)
var registerChannel = make(chan *WebsocketConnection)
var unregisterChannel = make(chan *WebsocketConnection)

var websocketConnections = make(map[SessionId]*WebsocketConnection)
var websocketChannels = make(map[Channel]WebsocketChannel)

func HandleUpgrade(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	return c.Next()
}

var HandleSocket = websocket.New(func(connection *websocket.Conn) {
	sessionId, ok := connection.Locals("sessionId").(SessionId)
	if !ok {
		connection.Close()
		return
	}

	user, ok := connection.Locals("user").(models.User)
	if !ok {
		connection.Close()
		return
	}

	websocketConnection := &WebsocketConnection{
		SessionId:  sessionId,
		User:       user,
		Connection: connection,
	}

	registerChannel <- websocketConnection
	defer websocketConnection.close()

	pongChannel := make(PongChannel, 1)
	go websocketConnection.handlePingPong(&pongChannel)

	for {
		if websocketConnection.Connection == nil {
			return
		}

		websocketMessageType, message, err := websocketConnection.Connection.ReadMessage()
		if err != nil {
			continue
		}

		switch websocketMessageType {
		case websocket.TextMessage:
			var unmarshaledMessage WebsocketMessage
			if err := json.Unmarshal(message, &unmarshaledMessage); err != nil {
				continue
			}

			messageType, ok := unmarshaledMessage["type"].(string)
			if !ok {
				continue
			}

			switch messageType {
			case PING_TYPE:
				websocketConnection.writeMessage(websocket.TextMessage, PONG_TYPE, WebsocketMessage{})
			case PONG_TYPE:
				select {
				case pongChannel <- struct{}{}:
				default:
				}
			default:
				channel, ok := unmarshaledMessage["channel"].(string)
				if !ok {
					continue
				}

				textChannel <- &Message{
					Channel:             channel,
					MessageType:         messageType,
					Message:             unmarshaledMessage,
					WebsocketConnection: websocketConnection,
				}
			}
		case websocket.PingMessage:
			websocketConnection.Connection.WriteMessage(websocket.PongMessage, nil)
		case websocket.PongMessage:
			select {
			case pongChannel <- struct{}{}:
			default:
			}
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
			writeChannelMessage(getBoardChannel(hookMessage.BoardId), websocket.TextMessage, hookMessage.Type, hookMessage.Message)
		case message := <-textChannel:
			switch message.MessageType {
			case JOIN_TYPE:
				if !message.WebsocketConnection.isAllowedToJoinChannel(message.Channel) {
					continue
				}

				websocketChannels[message.Channel][message.WebsocketConnection.SessionId] = struct{}{}

				writeChannelMessage(message.Channel, websocket.TextMessage, message.MessageType, WebsocketMessage{
					"userId": message.WebsocketConnection.User.ID,
				})
			case LEAVE_TYPE:
				if !message.WebsocketConnection.isUserInChannel(message.Channel) {
					continue
				}

				delete(websocketChannels[message.Channel], message.WebsocketConnection.SessionId)

				writeChannelMessage(message.Channel, websocket.TextMessage, message.MessageType, WebsocketMessage{
					"userId": message.WebsocketConnection.User.ID,
				})
			}
		case websocketConnection := <-registerChannel:
			writeGlobalMessage(websocket.TextMessage, REGISTER_TYPE, WebsocketMessage{
				"userId": websocketConnection.User.ID,
			})

			websocketConnections[websocketConnection.SessionId] = websocketConnection
		case websocketConnection := <-unregisterChannel:
			for channel := range websocketChannels {
				if !websocketConnection.isUserInChannel(channel) {
					continue
				}

				delete(websocketChannels[channel], websocketConnection.SessionId)

				writeChannelMessage(channel, websocket.TextMessage, LEAVE_TYPE, WebsocketMessage{
					"userId": websocketConnection.User.ID,
				})
			}

			websocketConnection.Connection.Close()

			delete(websocketConnections, websocketConnection.SessionId)

			writeGlobalMessage(websocket.TextMessage, UNREGISTER_TYPE, WebsocketMessage{
				"userId": websocketConnection.User.ID,
			})
		}
	}
}

func (websocketConnection *WebsocketConnection) handlePingPong(pongChannel *PongChannel) {
	timeout := time.Duration(dotenv.GetInt("WEBSOCKET_TIMEOUT_IN_SECOND")) * time.Second

	pingTicker := time.NewTicker(6 * timeout)
	defer pingTicker.Stop()

	timeoutTicker := time.NewTicker(timeout)
	timeoutTicker.Stop()
	defer timeoutTicker.Stop()

	hasPong := true

	for {
		if websocketConnection.Connection == nil || pongChannel == nil {
			return
		}

		select {
		case <-*pongChannel:
			hasPong = true
			timeoutTicker.Stop()
		case <-timeoutTicker.C:
			websocketConnection.close()
			return
		case <-pingTicker.C:
			if !hasPong {
				continue
			}
			websocketConnection.writeMessage(websocket.TextMessage, PING_TYPE, WebsocketMessage{})
			websocketConnection.Connection.WriteMessage(websocket.PingMessage, nil)
			timeoutTicker.Reset(timeout)
			hasPong = false
		}
	}
}

func (websocketConnection *WebsocketConnection) close() {
	unregisterChannel <- websocketConnection
}

func (websocketConnection *WebsocketConnection) writeMessage(websocketType WebsocketType, messageType MessageType, message WebsocketMessage) bool {
	if websocketConnection.Connection == nil {
		return false
	}

	message["type"] = messageType

	marshaledMessage, err := json.Marshal(message)
	if err != nil {
		return false
	}

	if err := websocketConnection.Connection.WriteMessage(websocket.TextMessage, marshaledMessage); err != nil {
		return false
	}

	return true
}

func (websocketConnection *WebsocketConnection) isAllowedToJoinChannel(channel Channel) bool {
	switch {
	case strings.HasPrefix(channel, BOARD_CHANNEL_PREFIX):
		boardIdString := strings.TrimPrefix(channel, BOARD_CHANNEL_PREFIX)
		boardId, err := strconv.ParseUint(boardIdString, 10, 64)
		if err != nil {
			return false
		}

		return websocketConnection.isAllowedToJoinBoardChannel(uint(boardId))
	default:
		return false
	}
}

func (websocketConnection *WebsocketConnection) isAllowedToJoinBoardChannel(boardId uint) bool {
	boards, ok := websocketConnection.getUserBoards()
	if !ok {
		return false
	}

	for _, board := range boards {
		if board.ID == boardId {
			return true
		}
	}

	return false
}

func (websocketConnection *WebsocketConnection) isUserInChannel(channel Channel) bool {
	websocketChannel, ok := websocketChannels[channel]
	if !ok {
		return false
	}

	if _, ok := websocketChannel[websocketConnection.SessionId]; !ok {
		return false
	}

	return true
}

func (websocketConnection *WebsocketConnection) getUserBoards() ([]models.Board, bool) {
	var user models.User
	if err := store.Database.Model(&websocketConnection.User).Preload("Boards").First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return []models.Board{}, false
	}

	return user.Boards, true
}

func writeGlobalMessage(websocketType WebsocketType, messageType MessageType, message WebsocketMessage) {
	for _, websocketConnection := range websocketConnections {
		websocketConnection.writeMessage(websocketType, messageType, message)
	}
}

func writeChannelMessage(channel Channel, websocketType WebsocketType, messageType MessageType, message WebsocketMessage) {
	message["channel"] = channel

	websocketChannel, ok := websocketChannels[channel]
	if !ok {
		return
	}

	for member := range websocketChannel {
		websocketConnection, ok := websocketConnections[member]
		if !ok {
			continue
		}

		websocketConnection.writeMessage(websocketType, messageType, message)
	}
}

func getBoardChannel(boardId uint) Channel {
	return fmt.Sprintf("%s%d", BOARD_CHANNEL_PREFIX, boardId)
}
