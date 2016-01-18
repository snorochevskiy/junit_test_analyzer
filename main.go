// xmlparser project main.go
package main

import (
	"fmt"
	"os"
)

const PROMT = `
Usage:
program load-results <branch> <path-to-report-files>
program start-server <port>
`

type TestSuite struct {
	Name       string         `xml:"name,attr"`
	Properties *PropertiesTag `xml:"properties"`
	TestCases  []TestCase     `xml:"testcase"`

	TestsNumber  string `xml:"tests,attr"`
	TestsSkipped string `xml:"skipped,attr"`
	TestsFailed  string `xml:"failures,attr"`
	TestsErrors  string `xml:"errors,attr"`

	// Error of whole suite
	SystemErr string `xml:"system-err,omitempty"`
}

type PropertiesTag struct {
}

type TestCase struct {
	Name      string         `xml:"name,attr"`
	ClassName string         `xml:"classname,attr"`
	Skipped   *SkippedStatus `xml:"skipped"`
	Failure   *FailureStatus `xml:"failure"`
}

type SkippedStatus struct {
}

type FailureStatus struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

func (testcase *TestCase) IsSkipped() bool {
	return testcase.Skipped != nil
}

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
	if len(os.Args) > 2 {
		portStr := os.Args[2]
		//		port, parseErr := strconv.Atoi(portStr)
		//		if parseErr != nil || port < 1 || port > 65536 {
		//			fmt.Println("Invalid port")
		//			os.Exit(1)
		//		}
		startServer(portStr)
	} else {
		fmt.Println("No port specified")
	}

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

	ProcessAllResultsFiles(branch, dir)
}
