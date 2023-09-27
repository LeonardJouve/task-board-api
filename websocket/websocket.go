package websocket

import (
	"fmt"
	"time"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

var MessageChannel = make(chan *Message)
var RegisterChannel = make(chan *websocket.Conn)
var UnregisterChannel = make(chan *websocket.Conn)
var Connections = make(map[uint]*websocket.Conn)

type Message struct {
	Type       int
	Value      []byte
	Connection *websocket.Conn
}

func close(connection *websocket.Conn) {
	if connection == nil {
		return
	}

	UnregisterChannel <- connection
	connection.Close()
}

func getUser(connection *websocket.Conn) (models.User, bool) {
	user, ok := connection.Locals("user").(models.User)
	if !ok {
		return models.User{}, false
	}

	return user, true
}

func HandleUpgrade(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	c.Locals("sessionId", utils.UUIDv4())

	return c.Next()
}

var HandleSocket = websocket.New(func(connection *websocket.Conn) {
	RegisterChannel <- connection
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

		MessageChannel <- &Message{
			Type:       messageType,
			Value:      messageValue,
			Connection: connection,
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
		case message := <-MessageChannel:
			if err := message.Connection.WriteMessage(message.Type, message.Value); err != nil {
				close(message.Connection)
			}
		case connection := <-RegisterChannel:
			user, ok := getUser(connection)
			if !ok {
				close(connection)
			}

			for _, conn := range Connections {
				if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("registered %d", user.ID))); err != nil {
					close(conn)
				}
			}

			Connections[user.ID] = connection
		case connection := <-UnregisterChannel:
			user, ok := getUser(connection)
			if !ok {
				close(connection)
			}

			delete(Connections, user.ID)

			for _, conn := range Connections {
				if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("unregistered %d", user.ID))); err != nil {
					close(conn)
				}
			}
		}
	}
}
