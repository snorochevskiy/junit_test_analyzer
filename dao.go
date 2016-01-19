package main

import (
	"log"
)

type DaoService struct {
}

var DAO DaoService = DaoService{}

const TEST_CASE_STATUS_FAILED = "FAILURE"
const TEST_CASE_STATUS_SKIPPED = "SKIPPED"
const TEST_CASE_STATUS_PASSED = "PASSED"

func (*DaoService) GetAllBranches() []string {
	rows, err := ExecuteSelect("SELECT DISTINCT branch FROM test_launches ORDER BY branch")
	if err != nil {
		log.Fatal(err)
	}

	branchNames := make([]string, 0, 10)
	for rows.Next() {
		var branchName string
		scanErr := rows.Scan(&branchName)
		if scanErr != nil {
			log.Println(scanErr)
			continue
		}
		branchNames = append(branchNames, branchName)
	}
	return branchNames
}

func (*DaoService) CreateTestsLaunch(branchName string) int64 {
	res, err := ExecuteInsert("INSERT INTO test_launches(branch) values(?)", branchName)
	if err != nil {
		log.Println(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
	}
	return id
}

func (dao *DaoService) AddTestCase(launchId int64, testCase *TestCase) {

	var testCaseStatus string
	if testCase.Failure != nil {
		testCaseStatus = TEST_CASE_STATUS_FAILED
	} else if testCase.Skipped != nil {
		testCaseStatus = TEST_CASE_STATUS_SKIPPED
	} else {
		testCaseStatus = TEST_CASE_STATUS_PASSED
	}

	res, err := ExecuteInsert(
		"INSERT INTO test_cases(name, class_name, status, parent_launch_id) values(?, ?, ?, ?)",
		testCase.Name, testCase.ClassName, testCaseStatus, launchId)
	if err != nil {
		log.Println(err)
		return
	}

	insertedTestCaseId, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return
	}

	if testCase.Failure != nil {
		dao.AddTestCaseFailure(insertedTestCaseId, testCase.Failure)
	}

}

func (*DaoService) AddTestCaseFailure(testCaseId int64, failure *FailureStatus) {
	_, err := ExecuteInsert(
		"INSERT INTO test_case_failures(failure_type, failure_message, failure_text, parent_test_case_id) values(?, ?, ?, ?)",
		failure.Type, failure.Message, failure.Text, testCaseId)
	if err != nil {
		log.Println(err)
		return
	}
}

func (dao *DaoService) GetAllLaunches() []*TestLaunchEntity {
	rows, err := ExecuteSelect("SELECT launch_id, branch, creation_date FROM test_launches ORDER BY launch_id")
	if err != nil {
		log.Fatal(err)
	}

	testLaunches := make([]*TestLaunchEntity, 0, 10)
	for rows.Next() {
		testLaunch := new(TestLaunchEntity)
		ScanStruct(rows, testLaunch)
		log.Println(*testLaunch)
		testLaunch.FailedTestsNum = dao.GetNumberOfFailedTestInLaunch(testLaunch.Id)
		testLaunches = append(testLaunches, testLaunch)
	}
	return testLaunches
}

func (*DaoService) GetAllTestsForLaunch(launchId int64) []*TestCaseEntity {
	rows, err := ExecuteSelect("SELECT test_case_id, name, class_name, status, parent_launch_id FROM test_cases WHERE parent_launch_id = ?", launchId)
	if err != nil {
		log.Fatal(err)
	}

	testCases := make([]*TestCaseEntity, 0, 10)
	for rows.Next() {
		testCase := new(TestCaseEntity)
		ScanStruct(rows, testCase)
		testCases = append(testCases, testCase)
	}
	return testCases
}

func (*DaoService) GetTestCaseDetails(testCaseId int64) *TestCaseEntity {
	rows, err := ExecuteSelect("SELECT test_case_id, name, class_name, status, parent_launch_id FROM test_cases WHERE test_case_id = ?", testCaseId)
	if err != nil {
		log.Println(err)
		return nil
	}

	if !rows.Next() {
		return nil
	}

	testCase := new(TestCaseEntity)
	scanErr := ScanStruct(rows, testCase)
	if scanErr != nil {
		log.Println(scanErr)
	}

	if testCase.Status == TEST_CASE_STATUS_FAILED {
		failedInfoRows, failedInfoErr := ExecuteSelect("SELECT test_case_failure_id, failure_message, failure_type, failure_text FROM test_case_failures WHERE parent_test_case_id = ?", testCaseId)
		if failedInfoErr != nil {
			log.Println(failedInfoErr)
		} else if failedInfoRows.Next() {
			testFailure := new(FailureEntity)
			scanErr := ScanStruct(failedInfoRows, testFailure)
			if scanErr != nil {
				log.Println(scanErr)
			}

			testCase.FailureInfo = testFailure
		} else {
			log.Printf("No failed info for %v", testCaseId)
		}

	}

	return testCase
}

func (*DaoService) GetNumberOfFailedTestInLaunch(launchId int64) int {
	row := SelectOneRow("SELECT count(*) FROM test_cases WHERE parent_launch_id = (?) AND status IN ('FAILURE')", launchId)
	num := new(int)
	row.Scan(num)
	return *num
}

func (*DaoService) GetDiffBetweenLaunches(launchId1 int64, launchId2 int64) {

}
