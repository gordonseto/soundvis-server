package basecontroller

import (
	"net/http"
	"encoding/json"
	"fmt"
	"time"
	"io/ioutil"
)

func SendResponse(w http.ResponseWriter, response interface{}) {
	responseJSON, err := json.Marshal(&response)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", responseJSON)
}

type RequestError struct {
	msg string
}

func (e *RequestError) Error() string {
	return e.msg
}

func MakeRequest(url string, method string, timeout time.Duration) ([]byte, error) {
	// Make request
	client := http.Client{
		Timeout: time.Second * timeout,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, &RequestError{"Server Request failed"}
	}

	// Read response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}