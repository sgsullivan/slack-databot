package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
)

// populateConfig ...
func populateConfig() {
	configFile := getConfigLocation()
	f, err := os.Open(configFile)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed opening configFile [%s]: %s", configFile, err))
	}
	defer f.Close()
	jsonParser := json.NewDecoder(f)
	jsonErr := jsonParser.Decode(&config)
	if jsonErr != nil {
		log.Fatal(fmt.Sprintf("Failed decoding configFile [%s]: %s", configFile, jsonErr))
	}
	// Ensure required config items are set
	reqConfigItems := []string{config.SlackApiUrl, config.SlackApiToken}
	for _, item := range reqConfigItems {
		if len(item) < 1 {
			log.Fatal("Missing required config item(s); Required: SlackApiUrl, SlackApiToken")
		}
	}
}

// getConfigLocation ...
func getConfigLocation() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed getting users home directory: %s", err))
	}
	logDebug(fmt.Sprintf("discovered config file %s/etc/databot.json", usr.HomeDir))
	return fmt.Sprintf("%s/etc/databot.json", usr.HomeDir)
}

// logDebug ...
func logDebug(msg string) {
	if len(os.Getenv("databot_debug")) == 0 {
		return
	}
	log.Printf(fmt.Sprintf("DEBUG: %s", msg))
}
