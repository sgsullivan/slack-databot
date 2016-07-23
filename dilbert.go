package main

import (
	"fmt"
	"time"
)

func dilbertRoutine(wsClient websocketData) {
	hour, minute, _ := time.Now().Clock()
	// post at 07:30 AM
	if hour == 07 && minute == 30 {
		year, month, day := time.Now().Date()
		dateString := fmt.Sprintf("%d-%s-%d\n", year, month, day)
		// only post once per given hour/minute
		if store.DilbertLastPosted != dateString {
			store.DilbertLastPosted = dateString
			comic := fmt.Sprintf("http://dilbert.com/strip/%s\n", dateString)
			wsClient.createSlackPost(comic, config.SlackDilbertChannel)
			logDebug(fmt.Sprintf("Posted %s", comic))
		}
	}
}
