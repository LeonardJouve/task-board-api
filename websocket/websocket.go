package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	UserId     UserId
	Connection *websocket.Conn
}

type RegistrationMessage struct {
	SessionId  SessionId
	UserId     UserId
	Connection *websocket.Conn
}

type Message struct {
	Channel     Channel
	MessageType MessageType
	Message     WebsocketMessage
	SessionId   SessionId
	Connection  *websocket.Conn
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
var registerChannel = make(chan RegistrationMessage)
var unregisterChannel = make(chan RegistrationMessage)

var websocketConnections = make(map[SessionId]WebsocketConnection)
var websocketChannels = make(map[Channel]WebsocketChannel)

func HandleUpgrade(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	return c.Next()
}

func handlePingPong(connection *websocket.Conn, pongChannel *PongChannel, sessionId string) {
	timeout := time.Duration(dotenv.GetInt("WEBSOCKET_TIMEOUT_IN_SECOND")) * time.Second

	pingTicker := time.NewTicker(6 * timeout)
	defer pingTicker.Stop()

	timeoutTicker := time.NewTicker(timeout)
	timeoutTicker.Stop()
	defer timeoutTicker.Stop()

	hasPong := true

	for {
		if connection == nil || pongChannel == nil {
			return
		}

		select {
		case <-*pongChannel:
			hasPong = true
			timeoutTicker.Stop()
		case <-timeoutTicker.C:
			close(sessionId)
			return
		case <-pingTicker.C:
			if !hasPong {
				continue
			}
			writeMessage(connection, websocket.TextMessage, PING_TYPE, WebsocketMessage{})
			connection.WriteMessage(websocket.PingMessage, nil)
			timeoutTicker.Reset(timeout)
			hasPong = false
		}
	}
}

var HandleSocket = websocket.New(func(connection *websocket.Conn) {
	sessionId, okSessionId := getSessionId(connection)
	user, okUser := getUser(connection)
	if !okSessionId || !okUser {
		connection.Close()
		return
	}

	registerChannel <- RegistrationMessage{
		SessionId:  sessionId,
		UserId:     user.ID,
		Connection: connection,
	}
	defer close(sessionId)

	pongChannel := make(PongChannel, 1)
	go handlePingPong(connection, &pongChannel, sessionId)

	for {
		if connection == nil {
			return
		}

		websocketMessageType, message, err := connection.ReadMessage()
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
				writeMessage(connection, websocket.TextMessage, PONG_TYPE, WebsocketMessage{})
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
					Channel:     channel,
					MessageType: messageType,
					Message:     unmarshaledMessage,
					SessionId:   sessionId,
					Connection:  connection,
				}
			}
		case websocket.PingMessage:
			connection.WriteMessage(websocket.PongMessage, nil)
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
				user, ok := getUser(message.Connection)
				if !ok {
					continue
				}

				if !isAllowedToJoinChannel(message.Channel, user) {
					continue
				}

				websocketChannels[message.Channel][message.SessionId] = struct{}{}

				writeChannelMessage(message.Channel, websocket.TextMessage, message.MessageType, WebsocketMessage{
					"userId": user.ID,
				})
			case LEAVE_TYPE:
				if !isUserInChannel(message.Channel, message.SessionId) {
					continue
				}

				user, ok := getUser(message.Connection)
				if !ok {
					continue
				}

				delete(websocketChannels[message.Channel], message.SessionId)

				writeChannelMessage(message.Channel, websocket.TextMessage, message.MessageType, WebsocketMessage{
					"userId": user.ID,
				})
			}
		case message := <-registerChannel:
			writeGlobalMessage(websocket.TextMessage, REGISTER_TYPE, WebsocketMessage{
				"userId": message.UserId,
			})

			websocketConnections[message.SessionId] = WebsocketConnection{
				UserId:     message.UserId,
				Connection: message.Connection,
			}
		case message := <-unregisterChannel:
			user, ok := getUser(message.Connection)
			if !ok {
				continue
			}

			for channel := range websocketChannels {
				if !isUserInChannel(channel, message.SessionId) {
					continue
				}

				delete(websocketChannels[channel], message.SessionId)

				writeChannelMessage(channel, websocket.TextMessage, LEAVE_TYPE, WebsocketMessage{
					"userId": user.ID,
				})
			}

			message.Connection.Close()

			delete(websocketConnections, message.SessionId)

			writeGlobalMessage(websocket.TextMessage, UNREGISTER_TYPE, WebsocketMessage{
				"userId": message.UserId,
			})
		}
	}
}

func writeMessage(connection *websocket.Conn, websocketType WebsocketType, messageType MessageType, message WebsocketMessage) bool {
	if connection == nil {
		return false
	}

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

func writeGlobalMessage(websocketType WebsocketType, messageType MessageType, message WebsocketMessage) {
	for _, websocketConnection := range websocketConnections {
		writeMessage(websocketConnection.Connection, websocketType, messageType, message)
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

		writeMessage(websocketConnection.Connection, websocketType, messageType, message)
	}
}

func close(sessionId string) {
	websocketConnection, ok := websocketConnections[sessionId]
	if !ok {
		return
	}

	unregisterChannel <- RegistrationMessage{
		SessionId:  sessionId,
		UserId:     websocketConnection.UserId,
		Connection: websocketConnection.Connection,
	}
}

func isAllowedToJoinChannel(channel Channel, user models.User) bool {
	switch {
	case strings.HasPrefix(channel, BOARD_CHANNEL_PREFIX):
		boardIdString := strings.TrimPrefix(channel, BOARD_CHANNEL_PREFIX)
		boardId, err := strconv.ParseUint(boardIdString, 10, 64)
		if err != nil {
			return false
		}

		return isAllowedToJoinBoardChannel(uint(boardId), user)
	default:
		return false
	}
}

func isAllowedToJoinBoardChannel(boardId uint, user models.User) bool {
	boards, ok := getUserBoards(user)
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

func isUserInChannel(channel Channel, sessionId SessionId) bool {
	websocketChannel, ok := websocketChannels[channel]
	if !ok {
		return false
	}

	if _, ok := websocketChannel[sessionId]; !ok {
		return false
	}

	return true
}

func getBoardChannel(boardId uint) Channel {
	return fmt.Sprintf("%s%d", BOARD_CHANNEL_PREFIX, boardId)
}

func getSessionId(connection *websocket.Conn) (SessionId, bool) {
	if connection == nil {
		return "", false
	}

	sessionId, ok := connection.Locals("sessionId").(SessionId)
	if !ok {
		return "", false
	}

	return sessionId, true
}

func getUser(connection *websocket.Conn) (models.User, bool) {
	if connection == nil {
		return models.User{}, false
	}

	user, ok := connection.Locals("user").(models.User)
	if !ok {
		return models.User{}, false
	}

	return user, true
}

func getUserBoards(user models.User) ([]models.Board, bool) {
	if err := store.Database.Model(&user).Preload("Boards").First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return []models.Board{}, false
	}

	return user.Boards, true
}
