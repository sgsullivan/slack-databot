package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func getDilbertPostedDir() string {
	dilbertDir := fmt.Sprintf("%s/dilbertPosted", getHomeEtc())

	err := createDirIfMissing(dilbertDir, 0750)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating [%s]: %s", dilbertDir, err))
	}

	return dilbertDir
}

func getTodaysDilbertFile() string {
	timeNow := time.Now()
	year, month, day := timeNow.Date()
	curDate := fmt.Sprintf("%d-%s-%d", year, month, day)

	return fmt.Sprintf("%s/%s", getDilbertPostedDir(), curDate)
}

func signifyDilbertPostedToday() {
	dilbertSignifyFile := getTodaysDilbertFile()
	err := createFile(dilbertSignifyFile)
	if err != nil {
		log.Fatalf(fmt.Sprintf("##Error creating [%s]: %s", dilbertSignifyFile, err))
	}
}

func dilbertRoutine(wsClient websocketData) {
	// if we arent connected atm give up now
	if !wsClient.ws.IsClientConn() {
		logDebug("dilbertRoutine() wsClient.ws.IsClientConn says we arent connected! Bailing..")
		return
	}

	timeNow := time.Now()

	// bail now if we've already posted today
	todaysDilbert := getTodaysDilbertFile()
	if pathExists(todaysDilbert) {
		return
	}

	// only start looking for new comics at 7am
	if timeNow.Hour() >= 7 {
		// back off now if instructed
		if !store.DilbertBackOffUntil.IsZero() {
			if !timeNow.After(store.DilbertBackOffUntil) {
				return
			}
		}

		// define assumed url for todays comic
		year, month, day := time.Now().Date()
		dateString := fmt.Sprintf("%d-%s-%d", year, month, day)

		// define assumed url for todays comic
		comic := fmt.Sprintf("http://dilbert.com/strip/%s", dateString)

		// verify comic exists or bail
		httpClient := &http.Client{Timeout: time.Duration(time.Duration(3) * time.Second)}
		resp, err := httpClient.Get(comic)
		defer resp.Body.Close()
		if resp.StatusCode != 200 || err != nil {
			// dilbert isnt yet present (or some error); oh noes!
			// Tell future dilbert routines to not try another HTTP request
			// for another 10 minutes
			logDebug(fmt.Sprintf("Error response from dilbert.com: code [%s] err [%s]", resp.StatusCode, err))
			store.DilbertBackOffUntil = timeNow.Add(time.Duration(600) * time.Second)
			return
		}
		wsClient.createSlackPost(fmt.Sprintf("%s\n", comic), config.SlackDilbertChannel)
		logDebug(fmt.Sprintf("Dilbert posted: %s", comic))
		// update records that we posted today
		signifyDilbertPostedToday()
		// reset back to zero time
		store.DilbertBackOffUntil = time.Time{}
	}
}
