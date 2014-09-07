package rhood

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

type WSMessage struct {
	messageType int
	body        []byte
	err         error
}

type WSServer struct {
	upgrader    websocket.Upgrader
	connections map[*websocket.Conn]bool

	chanAddConn    chan (*websocket.Conn)
	chanRemoveConn chan (*websocket.Conn)

	chanBroadcast chan (string)
}

func NewWSServer() *WSServer {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(c *http.Request) bool { return true }

	connections := make(map[*websocket.Conn]bool)

	chanAdd := make(chan (*websocket.Conn))
	chanRemove := make(chan (*websocket.Conn))

	chanBroadcast := make(chan (string))

	server := &WSServer{upgrader, connections, chanAdd, chanRemove, chanBroadcast}

	return server
}

func (wsServer *WSServer) Send(message string) {
	wsServer.chanBroadcast <- message
}

func (wsServer *WSServer) Start() {
	// 1. Send messages
	// 2. Add conn
	// 3. Remove conn

	for {

		select {
		case connToAdd := <-wsServer.chanAddConn:
			if _, alreadyExists := wsServer.connections[connToAdd]; alreadyExists {
				logErr("Connection already contained in the map")
			}
			wsServer.connections[connToAdd] = true

		case connToRemove := <-wsServer.chanRemoveConn:
			if _, alreadyExists := wsServer.connections[connToRemove]; alreadyExists {
				delete(wsServer.connections, connToRemove)
			} else {
				logErr("Attempt to delete unknown connection")
			}

		case messageToBroadcast := <-wsServer.chanBroadcast:
			message := NewMessageWebUIFromString(messageToBroadcast)

			globalCacheLastMessages.AddMessage(message)

			body, err := message.SerializeJson()
			if err != nil {
				logErr("Unable to serialize string %s", messageToBroadcast)
				continue
			}

			for conn, _ := range wsServer.connections {
				err := conn.WriteMessage(websocket.TextMessage, body)
				if err != nil {
					logErr("Error sending message: %s", err.Error())
				}
			}

		}
	}
}

func (wsServer *WSServer) Serve(w http.ResponseWriter, r *http.Request) {
	conn, err := wsServer.upgrader.Upgrade(w, r, nil)

	if err != nil {

		logErr(fmt.Sprintf("Websocket error: (%s) for request with header (%#v)", err.Error, r.Header))
		return
	}

	wsServer.chanAddConn <- conn
	defer func() {
		logDebug("Connection is removed")
		wsServer.chanRemoveConn <- conn
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			logErr("Error reading message, conn is discarded: %s", err.Error())
			return
		}
		if err = conn.WriteMessage(messageType, p); err != nil {
			logErr("Error writing message, conn is discarded: %s ", err.Error())
			return
		}
	}
}
