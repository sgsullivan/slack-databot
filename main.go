package main

import (
	"fmt"
)

type configData struct {
	SlackApiUrl         string
	SlackApiToken       string
	HttpTimeout         int
	JiraUrl             string
	JiraUser            string
	JiraPass            string
	SlackDilbertChannel string
}

var config configData

// main ...
func main() {
	populateConfig()
	logDebug(fmt.Sprintf("Starting up with Slack API url [%s] token [%s]", config.SlackApiUrl, config.SlackApiToken))
	connectToSlack()
}
