package socketmanager

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gorilla/websocket"
	"log"
	"errors"
	"github.com/gordonseto/soundvis-server/stream/IO"
	"fmt"
	"sync"
)

type SocketManager struct {
	connections map[string]*websocket.Conn
}

var instance *SocketManager
var once sync.Once

func Shared() *SocketManager {
	once.Do(func() {
		instance = &SocketManager{make(map[string]*websocket.Conn)}
	})
	return instance
}

func (sm *SocketManager) POSTPath() string {
	return "/sock"
}

func (sm *SocketManager) GETPath() string {
	return "/sock"
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var mutex = &sync.Mutex{}

func (sm *SocketManager) Connect(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := "3"
	fmt.Println("UserId from header: " + r.Header.Get("userId"))
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
		delete(sm.connections, userId)
		}()
	defer conn.Close()
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("UserId: " + userId + " Read: ", err)
			break
		}
		log.Printf("UserId: " + userId + " Recv: %s", message)
		err = conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("UserId: " + userId + " Write:", err)
			break
		}
	}
}

func (sm *SocketManager) SendStreamUpdateMessage(userId string, response streamIO.GetCurrentStreamResponse) error {
	conn, ok := sm.connections[userId]
	if !ok {
		return errors.New("No connection found for userId: " + userId)
	}
	err := conn.WriteJSON(response)
	return err
}



