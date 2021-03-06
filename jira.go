package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

type jiraIssueResp struct {
	Fields struct {
		Assignee struct {
			Name string `json:"name,omitempty"`
		}
		Summary     string `json:"summary,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"fields"`
}

func getJiraIssues(readFromSlack []byte) ([]string, string, error) {
	var slackEvent slackRtmEvent
	readFromSlack = bytes.Trim(readFromSlack, "\x00")
	err := json.Unmarshal(readFromSlack, &slackEvent)
	if err != nil {
		return []string{}, slackEvent.Channel, fmt.Errorf("uh, someone cant JSON :/\n%s", err)
	}
	re := regexp.MustCompile("jira#[A-z]+-[0-9]+")
	jiraIssues := re.FindAllString(slackEvent.Text, -1)
	if len(jiraIssues) == 0 {
		return []string{}, slackEvent.Channel, fmt.Errorf("that appears to be an invalid Jira issue :-1:")
	}
	return jiraIssues, slackEvent.Channel, nil
}

func isJiraIssueUrlRequest(readFromSlack []byte) bool {
	var slackEvent slackRtmEvent
	readFromSlack = bytes.Trim(readFromSlack, "\x00")
	err := json.Unmarshal(readFromSlack, &slackEvent)
	// This is here because it will fail to json decode non message type events with the given reference
	if err != nil {
		logDebug(fmt.Sprintf("Failed json decoding: [%s]", readFromSlack))
	} else if len(slackEvent.Text) > 0 {
		logDebug(fmt.Sprintf("Comparing message event text field: [%s]", slackEvent.Text))
		jiraLinkRequested, _ := regexp.MatchString("jira#.*", slackEvent.Text)
		if jiraLinkRequested {
			return true
		}
	}
	return false
}

func jiraIssueDescriptionRequested(readFromSlack []byte) bool {
	var slackEvent slackRtmEvent
	readFromSlack = bytes.Trim(readFromSlack, "\x00")
	json.Unmarshal(readFromSlack, &slackEvent)
	// Only message events have the Text field.
	if len(slackEvent.Text) > 0 {
		r, _ := regexp.Compile(`jira#\S+\.describe`)
		res := r.FindString(slackEvent.Text)
		if len(res) > 0 {
			logDebug("Recieved a request for a jira issue description")
			return true
		}
	}
	return false
}

func getJiraIssueDetails(jiraIssue string) (string, string, error) {
	hClient := &http.Client{}
	jiraReqUrl := fmt.Sprintf("%s/rest/api/latest/issue/%s", config.JiraUrl, jiraIssue)
	logDebug(fmt.Sprintf("JIRA URL: %s", jiraReqUrl))
	req, err := http.NewRequest("GET", jiraReqUrl, nil)
	req.SetBasicAuth(config.JiraUser, config.JiraPass)
	resp, err := hClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf(fmt.Sprintf("got non-200 http code: [%d]", resp.StatusCode))
	}
	bsRb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf(fmt.Sprintf("Error reading response body: %s", err))
	}
	// json decode
	var jr jiraIssueResp
	jsonDecodeErr := json.Unmarshal(bsRb, &jr)
	if jsonDecodeErr != nil {
		return "", "", fmt.Errorf(fmt.Sprintf("Error JSON decoding response body: %s", jsonDecodeErr))
	}
	return jr.Fields.Summary, jr.Fields.Description, nil
}
