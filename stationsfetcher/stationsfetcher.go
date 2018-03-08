package stationsfetcher

import (
	"github.com/gordonseto/soundvis-server/general"
	"net/http"
	"encoding/xml"
	"fmt"
	"strings"
	"golang.org/x/net/html"
	"errors"
	"bytes"
	"io"
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

type TuneInResponse struct {
	XMLName xml.Name	`xml:"playlist"`
	Tracklist	TrackList	`xml:"trackList"`
}

type TrackList struct {
	XMLName xml.Name	`xml:"trackList"`
	Tracks	[]Track		`xml:"track"`
}

type Track struct {
	XMLName xml.Name	`xml:"track"`
	Location string		`xml:"location"`
}

func FetchAndStoreStations() {
	FETCH_STATIONS_URL := "http://api.shoutcast.com/legacy/Top500?k=t7kGHgPoxtvtuUOc"

	// Fetch all the stations
	body, err := basecontroller.MakeRequest(FETCH_STATIONS_URL, http.MethodGet, 10)
	if err != nil {
		fmt.Println(err)
	}
	var response ShoutCastStationResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
	}

	// For each station, check tune in for their streamURL
	TUNE_IN_URL := "http://yp.shoutcast.com/sbin/tunein-station.xspf?id="
	for i := 0; i < 20; i++ {
		shoutcastStation := response.StationList[i]
		fmt.Println(shoutcastStation.Name + " " + shoutcastStation.Genre + " " + shoutcastStation.Id)

		// Make request for streamURL
		body, err := basecontroller.MakeRequest(TUNE_IN_URL + shoutcastStation.Id , http.MethodGet, 2)
		if err != nil {
			fmt.Println(err)
		}
		var tuneInResponse TuneInResponse
		err = xml.Unmarshal(body, &tuneInResponse)
		if err != nil {
			fmt.Println(err)
		}
		if len(tuneInResponse.Tracklist.Tracks) > 0 {
			// Get first streamURL out of array
			track := tuneInResponse.Tracklist.Tracks[0]
			fmt.Println(track.Location)
			streamBaseArray := strings.Split(track.Location, "/")
			streamBase := ""
			for j := 0; j < len(streamBaseArray) - 1; j++ {
				streamBase += streamBaseArray[j] + "/"
			}
			fmt.Println(streamBase)
			res, err := basecontroller.MakeRequest(streamBase + "7", http.MethodGet, 5)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(res)
				doc, err := html.Parse(strings.NewReader(string(res)))
				if err != nil {
					fmt.Println(err)
				}
				bn, err := getBody(doc)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(renderNode(bn))
			}
		}

	}

	//for _, station := range response.StationList {
	//	fmt.Println(station.Name + " " + station.Genre + " " + station.Id)
	//}
}

func getBody(doc *html.Node) (*html.Node, error) {
	var b *html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			b = n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	if b != nil {
		return b, nil
	}
	return nil, errors.New("Missing <body> in the node tree")
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}