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
	reqConfigItems := []string{
		config.SlackApiUrl,
		config.SlackApiToken,
		config.SlackDilbertChannel,
		config.JiraUrl,
		config.JiraUser,
		config.JiraPass}
	for _, item := range reqConfigItems {
		if len(item) < 1 {
			log.Fatal("Missing required config item(s)")
		}
	}
}

func getHomeEtc() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed getting users home directory: %s", err))
	}
	logDebug(fmt.Sprintf("discovered HOME etc %s/etc", usr.HomeDir))
	return fmt.Sprintf("%s/etc", usr.HomeDir)
}

// getConfigLocation ...
func getConfigLocation() string {
	homeDir := getHomeEtc()
	return fmt.Sprintf("%s/databot.json", homeDir)
}

// logDebug ...
func logDebug(msg string) {
	if len(os.Getenv("databot_debug")) == 0 {
		return
	}
	log.Printf(fmt.Sprintf("DEBUG: %s", msg))
}

func createDirIfMissing(fileLoc string, perm uint32) error {
	if !pathExists(fileLoc) {
		if err := os.MkdirAll(fileLoc, os.FileMode(perm)); err != nil {
			return fmt.Errorf("Failed creating %s; [%s]", fileLoc, err)
		}
	}
	return nil
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func createFile(path string) error {
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return err
	}
	return nil
}
