package stationsfetcher

import (
	"github.com/gordonseto/soundvis-server/general"
	"net/http"
	"encoding/xml"
	"fmt"
)

type ShoutCastStationResponse struct {
	XMLName	xml.Name	`xml:"stationlist"`
	StationList []ShoutcastStation	`xml:"station"`
}

type ShoutcastStation struct {
	XMLName xml.Name	`xml:"station"`
	Id string `xml:"id,attr"`
	Name string `xml:"name,attr"`
	Genre string `xml:"genre,attr"`
}

func FetchAndStoreStations() {
	body, err := basecontroller.MakeRequest("http://api.shoutcast.com/legacy/Top500?k=t7kGHgPoxtvtuUOc", http.MethodGet)
	if err != nil {
		fmt.Println(err)
	}
	var response ShoutCastStationResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
	}

	for _, station := range response.StationList {
		fmt.Println(station.Name + " " + station.Genre + " " + station.Id)
	}
}