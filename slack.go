package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// Once connected to RTM, all messages conform to this format
// {"type":"message","channel":"STRING","user":"STRING","text":"hello","ts":"1467931915.000002","team":"STRING"}.

type slackRtmEvent struct {
	Id      int    `json:"id,omitempty"`
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Channel string `json:"channel,omitempty"`
	Team    string `json:"team,omitempty"`
	User    string `json:"user,omitempty"`
	Url     string `json:"url,omitempty"`
}

// The only output from a rtm.start we care about is the websocket url
type slackRtmStartResp struct {
	Url string
}

type httpClient struct {
	client *http.Client
}

var store struct {
	DilbertBackOffUntil time.Time
	//LastRtmConnectEpoch int64
}

// Per slack docs, this is the maximum size
const slackMsgSizeCapBytes = 16000

var clientConfig = &http.Client{Timeout: time.Duration(time.Duration(30) * time.Second)}
var client = httpClient{clientConfig}

// createSlackPost ...
// Given a message and a channel name, return json suitable for posting to slack
func (wsClient *websocketData) createSlackPost(msg string, channel string) {
	payload := slackRtmEvent{
		Id:      1,
		Type:    "message",
		Channel: channel,
		Text:    msg,
	}
	jPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error encoding slackRtmEvent payload: %s", err))
	}
	wsClient.writeSocket(jPayload)
}

func connAndCreateWsClient() websocketData {
	wssUrl := rtmStart()
	ws := connectWebsocket(wssUrl)
	wsClient := websocketData{ws}
	return wsClient
}

// connectToSlack ...
func connectToSlack() {
	wsClient := connAndCreateWsClient()
	for {
		// slack, you are evil
		slackTimeout := make(chan bool, 1)
		go func() {
			time.Sleep(90 * time.Second)
			slackTimeout <- true
		}()

		message := make(chan []byte)
		go func() {
			message <- wsClient.readSocket()
			close(message)
		}()
		select {
		case readFromSlack := <-message:
			log.Printf("received: %s", readFromSlack)

			var slackEvent slackRtmEvent
			readFromSlack = bytes.Trim(readFromSlack, "\x00")
			if unencodeErr := json.Unmarshal(readFromSlack, &slackEvent); unencodeErr != nil {
				log.Printf("Invalid json received from slack? [%s]", unencodeErr)
			}
			if slackEvent.Type == "team_migration_started" {
				log.Printf("Team migration started. Will need to reconnect!")
				time.Sleep(time.Second * 10)
				wsClient = connAndCreateWsClient()
			} else {
				dilbertRoutine(wsClient)
				processJiraReq(wsClient, readFromSlack)
			}
		case <-slackTimeout:
			// damn you slack
			log.Printf("Hit slackTimeout! Attempting reconnection...")
			wsClient = connAndCreateWsClient()
		}
	}
}

func processJiraReq(wsClient websocketData, readFromSlack []byte) {
	if isJiraIssueUrlRequest(readFromSlack) {
		jiraIssues, slackChannel, err := getJiraIssues(readFromSlack)
		if err == nil {
			for v := range jiraIssues {
				jiraIssue := strings.Replace(jiraIssues[v], "jira#", "", 1)
				subject, description, err := getJiraIssueDetails(jiraIssue)
				if err != nil {
					wsClient.createSlackPost(fmt.Sprintf("Error when fetching jira issue [%s]: %s :rage:", jiraIssue, err), slackChannel)
				} else {
					// Show description if requested
					if jiraIssueDescriptionRequested(readFromSlack) {
						wsClient.createSlackPost(fmt.Sprintf("*[jira#%s] Description:* :point_down:\n%s", jiraIssue, description), slackChannel)
					} else {
						wsClient.createSlackPost(fmt.Sprintf("%s/browse/%s :point_left:\n*Subject:* [%s]", config.JiraUrl, jiraIssue, subject), slackChannel)
					}
				}
			}
		} else {
			wsClient.createSlackPost(fmt.Sprintf("%s", err), slackChannel)
		}
	}
}

// rtmStart ...
func rtmStart() string {
	log.Printf("Attempting rtm.start [%s]...", config.SlackApiUrl)
	// Slack just uses query strings here...
	// https://api.slack.com/methods/rtm.start/test
	var payload []byte
	resp, err := client.client.Post(fmt.Sprintf("%s/rtm.start?token=%s", config.SlackApiUrl, config.SlackApiToken), "application/json;charset=UTF-8", bytes.NewReader(payload))
	if err != nil {
		log.Fatal(fmt.Sprintf("Error in /rtm.start POST request to [%s]: %s", config.SlackApiUrl, err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal(fmt.Sprintf("Got http code [%d] back from [%s]", resp.StatusCode, config.SlackApiUrl))
	}
	logDebug(fmt.Sprintf("Got http code [%d] back from [%s]", resp.StatusCode, config.SlackApiUrl))

	// Read response body into byte slice
	bsRb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body: %s", err)
	}

	// json decode the slice, so we can get the websocket URL
	var rtm slackRtmStartResp
	jsonDecodeErr := json.Unmarshal(bsRb, &rtm)
	if jsonDecodeErr != nil {
		log.Fatal("Error JSON decoding response body: %s", jsonDecodeErr)
	}

	logDebug(fmt.Sprintf("Offered websocket URL: [%s]", rtm.Url))
	return rtm.Url
}
