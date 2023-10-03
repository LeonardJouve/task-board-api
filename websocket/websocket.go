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
type Channel = string
type WebsocketType = int
type MessageType = string
type WebsocketChannel = map[SessionId]struct{}
type WebsocketMessage = map[string]interface{}
type PongChannel = chan struct{}
type CloseChannel = chan struct{}

type WebsocketConnection struct {
	SessionId    SessionId
	User         models.User
	Connection   *websocket.Conn
	PongChannel  *PongChannel
	CloseChannel *CloseChannel
	WaitGroup    sync.WaitGroup
	sync.Mutex
}

type Message struct {
	Channel             Channel
	MessageType         MessageType
	Message             WebsocketMessage
	WebsocketConnection *WebsocketConnection
}

type WebsocketChannels struct {
	Channels map[Channel]WebsocketChannel
	sync.Mutex
}

type WebsocketConnections struct {
	Connections map[SessionId]*WebsocketConnection
	sync.Mutex
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

var websocketConnections = WebsocketConnections{
	Connections: make(map[SessionId]*WebsocketConnection),
}
var websocketChannels = WebsocketChannels{
	Channels: make(map[Channel]WebsocketChannel),
}

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

	pongChannel := make(PongChannel, 1)
	closeChannel := make(CloseChannel, 1)

	websocketConnection := &WebsocketConnection{
		SessionId:    sessionId,
		User:         user,
		Connection:   connection,
		PongChannel:  &pongChannel,
		CloseChannel: &closeChannel,
	}

	registerChannel <- websocketConnection
	defer func() {
		websocketConnection.WaitGroup.Wait()
		websocketConnection.close()
	}()

	go websocketConnection.handlePingPong()

	for {
		websocketMessageType, message, err := websocketConnection.Connection.ReadMessage()
		if err != nil {
			break
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
				case *websocketConnection.PongChannel <- struct{}{}:
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

				websocketChannels.add(message.WebsocketConnection, message.Channel)

				writeChannelMessage(message.Channel, websocket.TextMessage, message.MessageType, WebsocketMessage{
					"userId": message.WebsocketConnection.User.ID,
				})
			case LEAVE_TYPE:
				if !message.WebsocketConnection.isInChannel(message.Channel) {
					continue
				}

				websocketChannels.remove(message.WebsocketConnection, message.Channel)

				writeChannelMessage(message.Channel, websocket.TextMessage, message.MessageType, WebsocketMessage{
					"userId": message.WebsocketConnection.User.ID,
				})
			}
		case websocketConnection := <-registerChannel:
			writeGlobalMessage(websocket.TextMessage, REGISTER_TYPE, WebsocketMessage{
				"userId": websocketConnection.User.ID,
			})

			websocketConnections.add(websocketConnection)
		case websocketConnection := <-unregisterChannel:
			for channel := range websocketChannels.Channels {
				if !websocketConnection.isInChannel(channel) {
					continue
				}

				websocketChannels.remove(websocketConnection, channel)

				writeChannelMessage(channel, websocket.TextMessage, LEAVE_TYPE, WebsocketMessage{
					"userId": websocketConnection.User.ID,
				})
			}

			websocketConnection.Connection.Close()

			websocketConnections.remove(websocketConnection)

			writeGlobalMessage(websocket.TextMessage, UNREGISTER_TYPE, WebsocketMessage{
				"userId": websocketConnection.User.ID,
			})
		}
	}
}

func (websocketConnection *WebsocketConnection) handlePingPong() {
	websocketConnection.WaitGroup.Add(1)
	defer websocketConnection.WaitGroup.Done()

	timeout := time.Duration(dotenv.GetInt("WEBSOCKET_TIMEOUT_IN_SECOND")) * time.Second

	pingTicker := time.NewTicker(6 * timeout)
	defer pingTicker.Stop()

	timeoutTicker := time.NewTicker(timeout)
	timeoutTicker.Stop()
	defer timeoutTicker.Stop()

	hasPong := true

	for {
		select {
		case <-*websocketConnection.CloseChannel:
			return
		case <-*websocketConnection.PongChannel:
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
			timeoutTicker.Reset(timeout)
			hasPong = false
		}
	}
}

func (websocketConnection *WebsocketConnection) close() {
	select {
	case _, ok := <-*websocketConnection.CloseChannel:
		if ok {
			close(*websocketConnection.CloseChannel)
		}
	default:
	}
	websocketConnection.Connection.SetReadDeadline(time.Now())
	unregisterChannel <- websocketConnection
}

func (websocketConnection *WebsocketConnection) writeMessage(websocketType WebsocketType, messageType MessageType, message WebsocketMessage) bool {
	websocketConnection.Lock()
	defer websocketConnection.Unlock()

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
	boards, ok := websocketConnection.getBoards()
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

func (websocketConnection *WebsocketConnection) isInChannel(channel Channel) bool {
	websocketChannel, ok := websocketChannels.get(channel)
	if !ok {
		return false
	}

	if _, ok := websocketChannel[websocketConnection.SessionId]; !ok {
		return false
	}

	return true
}

func (websocketConnection *WebsocketConnection) getBoards() ([]models.Board, bool) {
	var user models.User
	if err := store.Database.Model(&websocketConnection.User).Preload("Boards").First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return []models.Board{}, false
	}

	return user.Boards, true
}

func writeGlobalMessage(websocketType WebsocketType, messageType MessageType, message WebsocketMessage) {
	for _, websocketConnection := range websocketConnections.Connections {
		websocketConnection.writeMessage(websocketType, messageType, message)
	}
}

func writeChannelMessage(channel Channel, websocketType WebsocketType, messageType MessageType, message WebsocketMessage) {
	message["channel"] = channel

	websocketChannel, ok := websocketChannels.get(channel)
	if !ok {
		return
	}

	for sessionId := range websocketChannel {
		websocketConnection, ok := websocketConnections.get(sessionId)
		if !ok {
			continue
		}

		websocketConnection.writeMessage(websocketType, messageType, message)
	}
}

func getBoardChannel(boardId uint) Channel {
	return fmt.Sprintf("%s%d", BOARD_CHANNEL_PREFIX, boardId)
}

func (websocketConnections *WebsocketConnections) add(websocketConnection *WebsocketConnection) {
	websocketConnections.Lock()
	defer websocketConnections.Unlock()

	websocketConnections.Connections[websocketConnection.SessionId] = websocketConnection
}

func (websocketConnections *WebsocketConnections) remove(websocketConnection *WebsocketConnection) {
	websocketConnections.Lock()
	defer websocketConnections.Unlock()

	delete(websocketConnections.Connections, websocketConnection.SessionId)
}

func (websocketConnections *WebsocketConnections) get(sessionId SessionId) (*WebsocketConnection, bool) {
	websocketConnections.Lock()
	defer websocketConnections.Unlock()

	websocketConnection, ok := websocketConnections.Connections[sessionId]

	return websocketConnection, ok
}

func (websocketChannels *WebsocketChannels) add(websocketConnection *WebsocketConnection, channel Channel) {
	websocketChannels.Lock()
	defer websocketChannels.Unlock()

	websocketChannels.Channels[channel][websocketConnection.SessionId] = struct{}{}
}

func (websocketChannels *WebsocketChannels) remove(websocketConnection *WebsocketConnection, channel Channel) {
	websocketChannels.Lock()
	defer websocketChannels.Unlock()

	delete(websocketChannels.Channels[channel], websocketConnection.SessionId)
}
func (websocketChannels *WebsocketChannels) get(channel Channel) (WebsocketChannel, bool) {
	websocketChannels.Lock()
	defer websocketChannels.Unlock()

	websocketChannel, ok := websocketChannels.Channels[channel]

	return websocketChannel, ok
}
