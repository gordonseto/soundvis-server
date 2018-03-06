package stream

import (
	"gopkg.in/mgo.v2"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gordonseto/soundvis-server/authentication"
	"github.com/gordonseto/soundvis-server/stream/IO"
	"github.com/gordonseto/soundvis-server/general"
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
	user, err := authentication.CheckAuthentication(r, sc.session)
	if err != nil {
		panic(err)
	}

	response := streamIO.GetCurrentStreamResponse{}
	response.IsPlaying = user.IsPlaying

	basecontroller.SendResponse(w, response)
}