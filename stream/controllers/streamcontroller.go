package stream

import (
	"gopkg.in/mgo.v2"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"fmt"
	"github.com/gordonseto/soundvis-server/authentication"
)

type (
	StreamController struct {
		session *mgo.Session
	}
)

func (sc StreamController) GETPath() string {
	return "/stream"
}

func NewStreamController(s *mgo.Session) *StreamController {
	return &StreamController{s}
}

func (sc StreamController) GetCurrentStream(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	_, err := authentication.CheckAuthentication(r, sc.session)
	if err != nil {
		panic(err)
	}
	fmt.Println("We got here!")
}