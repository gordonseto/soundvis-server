package socketmanager

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gorilla/websocket"
	"log"
	"errors"
	"github.com/gordonseto/soundvis-server/stream/IO"
	"fmt"
)

type SocketManager struct {
	connections map[string]*websocket.Conn
}

func NewSocketManager() *SocketManager {
	return &SocketManager{connections: make(map[string]*websocket.Conn)}
}

func (sm *SocketManager) POSTPath() string {
	return "/sock"
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (sm *SocketManager) Connect(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := "3"
	fmt.Println("UserId from header: + " + r.Header.Get("userId"))
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(errors.New("Error connecting socket"))
	}

	// register connection with userId
	sm.connections[userId] = conn
	// listen at connection
	go sm.Listen(userId, conn)
}

func (sm *SocketManager) Listen(userId string, conn *websocket.Conn) {
	defer func(){
		// remove connection from table if disconnects
		sm.connections[userId] = nil
		}()
	defer conn.Close()
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (sm *SocketManager) SendStreamUpdateMessage(userId string, response streamIO.GetCurrentStreamResponse) {
	conn, ok := sm.connections[userId]
	if !ok {
		fmt.Println("No connection found for userId: " + userId)
	}
	err := conn.WriteJSON(response)
	if err != nil {
		fmt.Println("Error writing message to connection for userId: " + userId)
	}
}



