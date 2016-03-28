package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
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
	Name          string         `xml:"name,attr"`
	FullClassName string         `xml:"classname,attr"`
	Skipped       *SkippedStatus `xml:"skipped"`
	Failure       *FailureStatus `xml:"failure"`

	Package   string
	ClassName string
	Md5Hash   string
	Status    string
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

type ParsedLaunchInfo struct {
	Branch     string
	LaunchTime time.Time
	Label      string
	Tests      []*TestCase

	FailedNum  int
	SkippedNum int
	PassedNum  int
	OveralNum  int
}

func ProcessAllResultsFiles(importConfig *ImportConfiguration) error {

	launchInfo := ParsedLaunchInfo{}
	launchInfo.LaunchTime = determineLaunchTime(importConfig)
	launchInfo.Branch = importConfig.Branch
	launchInfo.Label = importConfig.LaunchLabel

	reportFiles, reportFilesErr := ioutil.ReadDir(importConfig.FullDirPath)
	if reportFilesErr != nil {
		log.Panic(reportFilesErr)
		return reportFilesErr
	}

	allTests := make([]*TestCase, 0, 100)

	for _, reportFile := range reportFiles {
		if !reportFile.IsDir() && strings.HasSuffix(reportFile.Name(), ".xml") {
			fullReportFilePath := path.Join(importConfig.FullDirPath, reportFile.Name())
			suite, suitePathErr := parseTestSuite(fullReportFilePath)
			if suitePathErr != nil {
				log.Println(suitePathErr)
				continue
			}

			for i := 0; i < len(suite.TestCases); i++ {

				test := suite.TestCases[i]
				prepareTestCase(&test)
				allTests = append(allTests, &test)

				if test.Status == TEST_CASE_STATUS_FAILED {
					launchInfo.FailedNum++
				}
				if test.Status == TEST_CASE_STATUS_SKIPPED {
					launchInfo.SkippedNum++
				}
				if test.Status == TEST_CASE_STATUS_PASSED {
					launchInfo.PassedNum++
				}
				launchInfo.OveralNum++
			}
		}
	}

	launchInfo.Tests = allTests

	return DAO.PersistLaunch(launchInfo)

}

func determineLaunchTime(importConfig *ImportConfiguration) time.Time {
	var launchTime = time.Now()
	if !importConfig.ExplicitlySetTime.IsZero() {
		launchTime = importConfig.ExplicitlySetTime
	} else if importConfig.TakeLaunchTimeFromDir {
		dirInfo, err := os.Stat(importConfig.FullDirPath)
		if err != nil {
			log.Fatalln(err)
		}
		launchTime = dirInfo.ModTime()
	}
	return launchTime
}

func prepareTestCase(tc *TestCase) {
	md5Hash := md5.Sum([]byte(tc.FullClassName + "#" + tc.Name))
	tc.Md5Hash = hex.EncodeToString(md5Hash[:])

	tc.Package = tc.FullClassName[0:strings.LastIndex(tc.FullClassName, ".")]
	tc.ClassName = tc.FullClassName[strings.LastIndex(tc.FullClassName, ".")+1:]

	if tc.Failure != nil {
		tc.Status = TEST_CASE_STATUS_FAILED
	} else if tc.Skipped != nil {
		tc.Status = TEST_CASE_STATUS_SKIPPED
	} else {
		tc.Status = TEST_CASE_STATUS_PASSED
	}
}

func parseTestSuite(fullFilePath string) (*TestSuite, error) {
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
