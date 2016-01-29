package main

import ()

func main() {
	initializeDriver()

	if CLI.IsServerMode() {
		serverConfig := CLI.ParseServerConfiguration()
		if serverConfig.DaemonMode {
			Reborn()
		}
		StartServer(serverConfig.Port)
	} else if CLI.IsImporterMode() {
		importConfig := CLI.ParseImportConfiguration()
		ProcessAllResultsFiles(importConfig)
	} else {
		CLI.Promt()
	}

}
