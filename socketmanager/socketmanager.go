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
	"strings"
	"strconv"
	"github.com/gordonseto/soundvis-server/authentication"
	"github.com/gordonseto/soundvis-server/stream/helpers"
	"github.com/gordonseto/soundvis-server/notifications"
	"github.com/gordonseto/soundvis-server/config"
	"github.com/gordonseto/soundvis-server/streamjobs"
	"github.com/gordonseto/soundvis-server/users/repositories"
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

func (sm *SocketManager) Connect(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.Println("New socket connection!")
	fmt.Println("UserId from header: " + r.Header.Get("userId"))
	// if from DE1, userId will be userId + "DE1", if from raspberry pi only userId
	userId := r.Header.Get("userId")
	if userId == "" {
		// TODO: Remove this
		userId = config.DEFAULT_USER
	}
	if userId == "" {
		panic(errors.New("UserId is required for socket connection"))
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(errors.New("Error connecting socket"))
	}

	// register connection with userId
	sm.connections[userId] = conn

	// send update message
	// TODO: Fix this
	uid := userId
	if userId[len(userId)-3:] == "DE1" {
		uid = userId[:len(userId)-3]
	}
	user, err := usersrepository.Shared().FindUserById(uid)
	if err != nil {
		log.Println("User not found")
	} else {
		streamjobmanager.Shared().CheckNowPlayingForUser(*user)
	}

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
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("UserId: " + userId + " Read: ", err)
			break
		}
		log.Println("UserId: " + userId + " Msg: ", string(msg))
		//err = conn.WriteMessage(mt, msg)
		//if err != nil {
		//	log.Println("UserId: " + userId + " Write:", err)
		//	break
		//}

		uid := userId
		if userId[len(userId)-3:] == "DE1" {
			uid = userId[:len(userId)-3]
		}

		user, err := authentication.FindUser(uid)
		if err != nil {
			log.Println("UserId: " + userId + " does not exist")
			break
		}
		request := socketUpdateMessageToSetCurrentStreamRequest(string(msg))
		if request != nil {
			log.Println("UserId: " + userId + "Msg JSON: ", request)
			response, err := streamcontrollerhelper.UpdateUsersStream(request, user)
			if err != nil {
				log.Println("UserId: " + userId + " UpdateUsersStream Error:", err)
			} else {
				log.Println("UserId: " + userId + " UpdateUsersStream Response: ", response)

				// send to PI
				err = sm.SendStreamUpdateMessage(uid, *response)
				if err != nil {
					log.Println("UserId: " + userId + " Write:", err)
					break
				}
				// send to DE1
				err = sm.SendStreamUpdateMessage(uid + "DE1", *response)
				if err != nil {
					log.Println("UserId: " + userId + " Write:", err)
					break
				}

				err = notifications.SendStreamUpdateNotification([]string{user.DeviceToken}, *response)
				log.Println(err)
			}
		} else {
			log.Println("Socket message not in correct format")
		}
	}
}

func socketUpdateMessageToSetCurrentStreamRequest(message string) *streamIO.SetCurrentStreamRequest {
	request := &streamIO.SetCurrentStreamRequest{}
	messageArray := strings.Split(message, ",")
	if len(messageArray) == 3 {
		if messageArray[0] == "1" {
			request.IsPlaying = true
		} else {
			request.IsPlaying = false
		}
		request.CurrentStream = messageArray[1]
		request.CurrentVolume, _ = strconv.Atoi(messageArray[2])
		return request
	}
	return nil
}

func (sm *SocketManager) SendStreamUpdateMessage(userId string, response streamIO.GetCurrentStreamResponse) error {
	log.Println("SendStreamUpdateMessageUserId: ", userId)
	conn, ok := sm.connections[userId]
	if !ok {
		return errors.New("No connection found for userId: " + userId)
	}
	//message := streamUpdateResponseToSocketMessage(&response)
	//log.Println(message)

	// remove unnecessary fields when sending to DE1
	//stationCopy := models.Station{
	//	Id: response.CurrentStation.Id,
	//	Name: response.CurrentStation.Name,
	//	Genre: response.CurrentStation.Genre,
	//	Country: response.CurrentStation.Country,
	//}
	//responseCopy := streamIO.GetCurrentStreamResponse{
	//	IsPlaying: response.IsPlaying,
	//	CurrentPlaying: response.CurrentPlaying,
	//	CurrentVolume: response.CurrentVolume,
	//	CurrentStation: &stationCopy,
	//	CurrentStreamURL: response.CurrentStreamURL,
	//	CurrentSong: response.CurrentSong,
	//}

	err := conn.WriteJSON(response)
	//err := conn.WriteMessage(websocket.TextMessage, []byte(message))

	return err
}

func streamUpdateResponseToSocketMessage(response *streamIO.GetCurrentStreamResponse) string {
	message := ""
	if response.IsPlaying {
		message += "1,"
	} else {
		message += "0,"
	}
	message += response.CurrentStreamURL
	return message
}


