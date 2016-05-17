package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const PROMT = `
To import JUnit reports to database:
> jutra load-results <path-to-report-files> [--project=<PROJECT>] [--branch=<BRANCH>] [--date=<from-fs | YYYY-MM-DD_HH:mm:ss>] [--label=<label>]
Loads results from junit XML reports in specified folder
	<PROJECT>				- project that is tested
	<BRANCH>					- branch used in tests run (e.g. master, trunk)
	<path-to-report-files>	- folder with XML junit reports
	--date					- specify launch time
		* from-fs				-tells to take launch time as folder last mod time
		* YYYY-MM-DD_HH:mm:ss	- date in format YYYY-MM-DD_HH:mm:ss
	--label					- Some label you want to sign launch with

*Default launch time is NOW()

To start jutra web application server:
> jutra start-server <port> [--daemon]
Starts web application on specified port
`

type Cli struct{}

var CLI Cli

type ServerConfiguration struct {
	Port       string
	DaemonMode bool
}

type ImportConfiguration struct {
	Project               string
	Branch                string
	FullDirPath           string
	LaunchLabel           string
	TakeLaunchTimeFromDir bool
	ExplicitlySetTime     time.Time
}

func (*Cli) IsServerMode() bool {
	return len(os.Args) > 2 && os.Args[1] == "start-server"
}

func (*Cli) IsImporterMode() bool {
	return len(os.Args) >= 3 && os.Args[1] == "load-results"
}

func (*Cli) Promt() {
	fmt.Println(PROMT)
}

func (*Cli) ParseServerConfiguration() *ServerConfiguration {
	if len(os.Args) < 2 {
		fmt.Println("No port specified")
		os.Exit(1)
	}

	serverConfiguration := new(ServerConfiguration)
	serverConfiguration.Port = os.Args[2]

	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "--daemon" {
			serverConfiguration.DaemonMode = true
		}
	}
	return serverConfiguration
}

func (*Cli) ParseImportConfiguration() *ImportConfiguration {
	if len(os.Args) < 3 {
		fmt.Println(PROMT)
		os.Exit(1)
	}

	importConfig := new(ImportConfiguration)
	importConfig.FullDirPath = os.Args[2]

	fileInfo, err := os.Stat(importConfig.FullDirPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if !fileInfo.IsDir() {
		fmt.Println(importConfig.FullDirPath + " is not a directory")
		os.Exit(1)
	}

	for i := 3; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "--date=") {
			timeStr := os.Args[i][strings.Index(os.Args[i], "=")+1:]
			if timeStr == "from-fs" {
				importConfig.TakeLaunchTimeFromDir = true
			} else {
				parsedTime, err := time.Parse("2006-01-02_15:04:05", timeStr)
				if err != nil {
					log.Fatalln(err)
				}
				importConfig.ExplicitlySetTime = parsedTime
			}

		} else if strings.HasPrefix(os.Args[i], "--label=") {
			labelStr := os.Args[i][strings.Index(os.Args[i], "=")+1:]
			importConfig.LaunchLabel = labelStr
		} else if strings.HasPrefix(os.Args[i], "--project=") {
			projectStr := os.Args[i][strings.Index(os.Args[i], "=")+1:]
			importConfig.Project = projectStr
		} else if strings.HasPrefix(os.Args[i], "--branch=") {
			branchStr := os.Args[i][strings.Index(os.Args[i], "=")+1:]
			importConfig.Branch = branchStr
		}
	}

	return importConfig
}
