package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/gordonseto/soundvis-server/stations/controllers"
)

var PORT string = ":8080"

func main() {
	r := httprouter.New()
	sc := stations.NewStationsController()

	r.GET(sc.GETPath(), sc.GetStations)

	http.ListenAndServe(PORT, r)
}