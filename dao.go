package main

import (
	"log"
	"time"
)

type DaoService struct {
}

var DAO DaoService = DaoService{}

const TEST_CASE_STATUS_FAILED = "FAILED"
const TEST_CASE_STATUS_SKIPPED = "SKIPPED"
const TEST_CASE_STATUS_PASSED = "PASSED"

func (*DaoService) PersistLaunch(branch string, testCases []*TestCase, launchTime time.Time, launchLabel string) error {
	connection, err := OpenDbConnection()
	if err != nil {
		return nil
	}
	defer connection.Close()

	transaction, err := connection.Begin()
	if err != nil {
		return err
	}

	res, err := transaction.Exec("INSERT INTO test_launches(branch, creation_date, label) values(?, ?, ?)", branch, launchTime, launchLabel)
	if err != nil {
		transaction.Rollback()
		return err
	}
	launchId, err := res.LastInsertId()
	if err != nil {
		transaction.Rollback()
		return err
	}

	testStmt, err := transaction.Prepare("INSERT INTO test_cases(name, package, class_name, md5_hash, status, parent_launch_id) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		transaction.Rollback()
		return err
	}

	failureStmt, err := transaction.Prepare("INSERT INTO test_case_failures(failure_type, failure_message, failure_text, parent_test_case_id) values(?, ?, ?, ?)")
	if err != nil {
		transaction.Rollback()
		return err
	}

	for _, test := range testCases {
		res, err := testStmt.Exec(test.Name, test.Package, test.ClassName, test.Md5Hash, test.TestCaseStatus, launchId)
		if err != nil {
			transaction.Rollback()
			return err
		}
		if test.Failure != nil {
			testId, err := res.LastInsertId()
			if err != nil {
				transaction.Rollback()
				return err
			}
			_, failureAddErr := failureStmt.Exec(test.Failure.Type, test.Failure.Message, test.Failure.Text, testId)
			if failureAddErr != nil {
				transaction.Rollback()
				return failureAddErr
			}
		}
	}
	transaction.Commit()
	return nil
}

func (*DaoService) GetAllBranches() []string {
	rows, err := ExecuteSelect("SELECT DISTINCT branch FROM test_launches ORDER BY branch")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

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

func (dao *DaoService) GetAllLaunchesInBranch(branch string) []*TestLaunchEntity {
	rows, err := ExecuteSelect("SELECT launch_id, branch, label, creation_date FROM test_launches WHERE branch = ? ORDER BY creation_date", branch)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	testLaunches := make([]*TestLaunchEntity, 0, 10)
	for rows.Next() {
		testLaunch := new(TestLaunchEntity)
		ScanStruct(rows, testLaunch)

		testLaunch.FailedTestsNum = dao.GetNumberOfFailedTestInLaunch(testLaunch.Id)
		testLaunches = append(testLaunches, testLaunch)
	}
	return testLaunches
}

func (*DaoService) GetLaunchInfo(launchId int64) *TestLaunchEntity {

	rows, err := ExecuteSelect("SELECT launch_id, branch, label, creation_date FROM test_launches WHERE launch_id = ?", launchId)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	if !rows.Next() {
		return nil
	}

	testLaunch := new(TestLaunchEntity)
	scanErr := ScanStruct(rows, testLaunch)

	if scanErr != nil {
		log.Println(scanErr)
		return nil
	}

	return testLaunch
}

func (*DaoService) GetAllTestsForLaunch(launchId int64) []*TestCaseEntity {
	rows, err := ExecuteSelect("SELECT test_case_id, name, package, class_name, status, parent_launch_id FROM test_cases WHERE parent_launch_id = ? ORDER BY status", launchId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	testCases := make([]*TestCaseEntity, 0, 10)
	for rows.Next() {
		testCase := new(TestCaseEntity)
		ScanStruct(rows, testCase)
		testCases = append(testCases, testCase)
	}
	return testCases
}

func (*DaoService) GetAllTestsForPackage(launchId int64, packageName string) []*TestCaseEntity {
	rows, err := ExecuteSelect("SELECT test_case_id, name, package, class_name, status, parent_launch_id FROM test_cases WHERE parent_launch_id = ? AND package = ? ORDER BY status", launchId, packageName)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	testCases := make([]*TestCaseEntity, 0, 10)
	for rows.Next() {
		testCase := new(TestCaseEntity)
		ScanStruct(rows, testCase)
		testCases = append(testCases, testCase)
	}
	return testCases
}

func (*DaoService) GetPackagesForLaunch(launchId int64) ([]*PackageEntity, error) {
	connection, err := OpenDbConnection()
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	stmt, err := connection.Prepare("SELECT count(*) FROM test_cases WHERE parent_launch_id = ? AND package = ? AND status = ?")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	rows, err := connection.Query("SELECT package, count(*) as tests_num FROM test_cases WHERE parent_launch_id = ? GROUP BY package ORDER BY package", launchId)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	packages := make([]*PackageEntity, 0, 10)
	for rows.Next() {
		packageEntity := new(PackageEntity)
		ScanStruct(rows, packageEntity)

		failedTestsNumRow := stmt.QueryRow(launchId, packageEntity.Package, TEST_CASE_STATUS_FAILED)
		var failedTestIntPackage int
		err := failedTestsNumRow.Scan(&failedTestIntPackage)
		if err != nil {
			log.Panicln(err)
			return nil, err
		}
		packageEntity.FailedTestsNum = failedTestIntPackage

		passedTestsNumRow := stmt.QueryRow(launchId, packageEntity.Package, TEST_CASE_STATUS_PASSED)
		var passedTestIntPackage int
		err = passedTestsNumRow.Scan(&passedTestIntPackage)
		if err != nil {
			log.Panicln(err)
			return nil, err
		}
		packageEntity.PassedTestsNum = passedTestIntPackage

		skippedTestsNumRow := stmt.QueryRow(launchId, packageEntity.Package, TEST_CASE_STATUS_SKIPPED)
		var skippedTestIntPackage int
		err = skippedTestsNumRow.Scan(&skippedTestIntPackage)
		if err != nil {
			log.Panicln(err)
			return nil, err
		}
		packageEntity.SkippedTestsNum = skippedTestIntPackage

		packages = append(packages, packageEntity)
	}
	return packages, nil
}

func (*DaoService) GetTestCaseDetails(testCaseId int64) *TestCaseEntity {
	rows, err := ExecuteSelect("SELECT test_case_id, name, package, class_name, status, parent_launch_id FROM test_cases WHERE test_case_id = ?", testCaseId)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

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
	row := SelectOneRow("SELECT count(*) FROM test_cases WHERE parent_launch_id = (?) AND status IN ('FAILED')", launchId)
	num := new(int)
	row.Scan(num)
	return *num
}

func (*DaoService) GetAddedTestsInDiff(launchId1 int64, launchId2 int64) []*TestCaseEntity {
	newTestsRows, newTestRowsErr := ExecuteSelect(
		"SELECT test_case_id, name, package, class_name, status, parent_launch_id FROM test_cases WHERE parent_launch_id = ? AND md5_hash IN ( "+
			"SELECT md5_hash FROM test_cases WHERE parent_launch_id = ? EXCEPT "+
			"SELECT md5_hash FROM test_cases WHERE parent_launch_id = ?"+
			" ) ORDER BY status", launchId1, launchId2, launchId1)
	if newTestRowsErr != nil {
		log.Println(newTestRowsErr)
		return nil
	}
	defer newTestsRows.Close()

	testCases := make([]*TestCaseEntity, 0, 10)
	for newTestsRows.Next() {
		testCase := new(TestCaseEntity)
		ScanStruct(newTestsRows, testCase)
		testCases = append(testCases, testCase)
	}
	return testCases
}

func (*DaoService) GetTestsFromStatus1ToStatus2(launchId1 int64, launchId2 int64, status1 string, status2 string) []*TestCaseEntity {
	rows, err := ExecuteSelect(
		"SELECT test_case_id, name, package, class_name, status, parent_launch_id FROM test_cases WHERE status = '"+status2+"' AND parent_launch_id = ? AND md5_hash IN ( "+
			"SELECT md5_hash FROM test_cases WHERE parent_launch_id = ? AND status = '"+status1+"'"+
			" )", launchId2, launchId1)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	tests := make([]*TestCaseEntity, 0, 10)
	for rows.Next() {
		testCase := new(TestCaseEntity)
		ScanStruct(rows, testCase)
		tests = append(tests, testCase)
	}
	return tests
}

func (*DaoService) DeleteLaunch(launchId int64) error {
	_, err := ExecuteDelete("DELETE FROM test_launches WHERE launch_id = ?", launchId)
	if err != nil {
		// TODO process an error ?
		return err
	}
	return nil
}
