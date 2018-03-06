package basecontroller

import (
	"net/http"
	"encoding/json"
	"fmt"
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