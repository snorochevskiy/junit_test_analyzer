package main

import (
	"log"
	"os"
)

func main() {
	initializeDriver()
	createDbIfNotExists()

	if CLI.IsServerMode() {
		serverConfig := CLI.ParseServerConfiguration()
		if serverConfig.DaemonMode {
			Reborn()
			SetLogToFile()
		}
		StartServer(serverConfig.Port)
	} else if CLI.IsImporterMode() {
		importConfig := CLI.ParseImportConfiguration()
		ProcessAllResultsFiles(importConfig)
	} else {
		CLI.Promt()
	}

}

func SetLogToFile() {
	f, err := os.OpenFile("jutra.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
}
