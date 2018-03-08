package notifications

import (
	"github.com/NaySoftware/go-fcm"
	"fmt"
	"github.com/gordonseto/soundvis-server/config"
	"github.com/gordonseto/soundvis-server/stream/IO"
)

func sendNotification(deviceTokens []string, data interface{}) error {
	serverKey := config.FIREBASE_SERVER_KEY

	c := fcm.NewFcmClient(serverKey)
	c.NewFcmRegIdsMsg(deviceTokens, data)

	status, err := c.Send()

	if err == nil {
		status.PrintResults()
	} else {
		fmt.Println(err)
	}

	return err
}

func SendStreamUpdateNotification(deviceTokens []string, streamResponse streamIO.GetCurrentStreamResponse) error {
	return sendNotification(deviceTokens, streamResponse)
}