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
> jutra load-results <branch> <path-to-report-files> [--date=<from-fs | YYYY-MM-DD_HH:mm:ss>] [--label=<label>]
Loads results from junit XML reports in specified folder
	<branch>					- branch user in tests run (e.g. master, trunk)
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

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(PROMT)
		os.Exit(1)
	}

	initializeDriver()
	createDbIfNotExists()

	cmd := os.Args[1]

	if cmd == "load-results" {
		LoadTestResults()
	} else if cmd == "start-server" {
		StartServer()
	}

}

func StartServer() {
	if len(os.Args) < 2 {
		fmt.Println("No port specified")
		os.Exit(1)
	}

	portStr := os.Args[2]

	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "--daemon" {
			Reborn()
		}
	}

	startServer(portStr)
}

func LoadTestResults() {
	branch := os.Args[2]
	dir := os.Args[3]

	fileInfo, err := os.Stat(dir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if !fileInfo.IsDir() {
		fmt.Println(dir + " is not a directory")
		os.Exit(1)
	}

	processor := JUnitResultsFolderProcessor{Branch: branch, FullDirPath: dir}

	for i := 4; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "--date=") {
			timeStr := os.Args[i][strings.Index(os.Args[i], "=")+1:]
			if timeStr == "from-fs" {
				processor.TakeLaunchTimeFromDir = true
			} else {
				parsedTime, err := time.Parse("2006-01-02_15:04:05", timeStr)
				if err != nil {
					log.Fatalln(err)
				}
				processor.ExplicitlySetTime = parsedTime
			}

		} else if strings.HasPrefix(os.Args[i], "--label=") {
			labelStr := os.Args[i][strings.Index(os.Args[i], "=")+1:]
			processor.LaunchLabel = labelStr
		}
	}

	processor.ProcessAllResultsFiles()
}
