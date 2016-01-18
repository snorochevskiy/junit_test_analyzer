package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

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
