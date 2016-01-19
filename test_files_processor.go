package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

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

func ProcessAllResultsFiles(branch string, fullDirPath string) {
	reportFiles, reportFilesErr := ioutil.ReadDir(fullDirPath)
	if reportFilesErr != nil {
		log.Panic(reportFilesErr)
		return
	}

	launchId := DAO.CreateTestsLaunch(branch)

	for _, reportFile := range reportFiles {
		if !reportFile.IsDir() && strings.HasSuffix(reportFile.Name(), ".xml") {
			fullReportFilePath := path.Join(fullDirPath, reportFile.Name())
			suite, suitePathErr := ParseTestSuite(fullReportFilePath)
			if suitePathErr != nil {
				log.Println(suitePathErr)
				continue
			}
			PersistSuite(launchId, suite)

		}
	}
}

func ParseTestSuite(fullFilePath string) (*TestSuite, error) {
	xmlFile, err := os.Open(fullFilePath)
	if err != nil {
		log.Println("Error opening file:", err)
		return nil, err
	}
	defer xmlFile.Close()

	bytes, _ := ioutil.ReadAll(xmlFile)

	var suite TestSuite
	err = xml.Unmarshal(bytes, &suite)
	if err != nil {
		return nil, err
	}

	return &suite, nil
}

func PersistSuite(launchId int64, suite *TestSuite) {
	for _, testCase := range suite.TestCases {
		DAO.AddTestCase(launchId, &testCase)
	}
}
