package websocket

import (
	"fmt"
	"time"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var MessageChannel = make(chan *Message)
var RegisterChannel = make(chan *websocket.Conn)
var UnregisterChannel = make(chan *websocket.Conn)
var Connections = make(map[string]*websocket.Conn)

type Message struct {
	Type      int
	Value     []byte
	SessionId string
}

func close(connection *websocket.Conn) {
	if connection == nil {
		return
	}

	UnregisterChannel <- connection
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
	RegisterChannel <- connection
	defer close(connection)

	sessionId, ok := getSessionId(connection)
	if !ok {
		close(connection)
		return
	}

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

		MessageChannel <- &Message{
			Type:      messageType,
			Value:     messageValue,
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
			for _, connection := range Connections {
				if err := connection.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					close(connection)
					continue
				}
			}
		case connection := <-RegisterChannel:
			sessionId, ok := getSessionId(connection)
			if !ok {
				close(connection)
				continue
			}

			for _, conn := range Connections {
				if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("registered %s", sessionId))); err != nil {
					close(conn)
				}
			}

			Connections[sessionId] = connection
		case connection := <-UnregisterChannel:
			sessionId, ok := getSessionId(connection)
			if !ok {
				close(connection)
				continue
			}

			delete(Connections, sessionId)

			for _, conn := range Connections {
				if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("unregistered %s", sessionId))); err != nil {
					close(conn)
				}
			}
		}
	}
}
