package websocket

import (
	"fmt"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var MessageChannel = make(chan *Message)
var RegisterChannel = make(chan *websocket.Conn)
var UnregisterChannel = make(chan *websocket.Conn)
var Connections = make(map[string]*websocket.Conn)

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

func HandleUpgrade(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

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
			connectionId := connection.Params("id")

			for _, conn := range Connections {
				if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("registered %s", connectionId))); err != nil {
					close(conn)
				}
			}

			Connections[connectionId] = connection
		case connection := <-UnregisterChannel:
			connectionId := connection.Params("id")

			delete(Connections, connectionId)

			for _, conn := range Connections {
				if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("unregistered %s", connectionId))); err != nil {
					close(conn)
				}
			}
		}
	}
}
