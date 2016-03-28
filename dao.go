package main

import (
	"database/sql"
	"log"
	"strings"
)

type DaoService struct {
}

var DAO DaoService = DaoService{}

const TEST_CASE_STATUS_FAILED = "FAILED"
const TEST_CASE_STATUS_SKIPPED = "SKIPPED"
const TEST_CASE_STATUS_PASSED = "PASSED"

func (*DaoService) PersistLaunch(launchInfo ParsedLaunchInfo) error {
	connection, err := OpenDbConnection()
	if err != nil {
		return nil
	}
	defer closeDb(connection)

	transaction, err := connection.Begin()
	if err != nil {
		return err
	}

	res, err := transaction.Exec("INSERT INTO test_launches(branch, creation_date, label, test_num, failed_num, skipped_num, passed_num) values(?, ?, ?, ?, ?, ?, ?)",
		launchInfo.Branch, launchInfo.LaunchTime, launchInfo.Label, launchInfo.OveralNum, launchInfo.FailedNum, launchInfo.SkippedNum, launchInfo.PassedNum)
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

	for _, test := range launchInfo.Tests {
		res, err := testStmt.Exec(test.Name, test.Package, test.ClassName, test.Md5Hash, test.Status, launchId)
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

	if commitErr := transaction.Commit(); commitErr != nil {
		log.Printf("Unable to commit new launch INSERT. Reason: %v\n", commitErr)
	}

	return nil
}

func (*DaoService) GetAllBranchesNames() []string {
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

func (*DaoService) getFilteredBranches(connection *sql.DB, filter *BranchesFilter) ([]*BranchInfoEntity, error) {

	sqlText := "SELECT DISTINCT branch FROM test_launches"
	params := make([]interface{}, 0, 5)

	if filter != nil && filter.HasSomethingToFilter() {
		sqlText += " WHERE"
		if filter.LabelTemplate != "" {
			sqlText += " label LIKE ?"
			params = append(params, strings.Replace(filter.LabelTemplate, "*", "%", -1))
		}
	}

	//	log.Printf("SQL: %v\n", sqlText)
	//	for i, v := range params {
	//		log.Printf("param %v = %v\n", i, v)
	//	}

	rows, err := connection.Query(sqlText, params...)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	branches := make([]*BranchInfoEntity, 0, 10)
	for rows.Next() {
		bi := new(BranchInfoEntity)
		scanErr := rows.Scan(&bi.BranchName)
		if scanErr != nil {
			log.Println(scanErr)
			continue
		}
		branches = append(branches, bi)
	}
	return branches, nil
}

func (dao *DaoService) GetAllBranchesInfo(filter *BranchesFilter) ([]*BranchInfoEntity, error) {
	connection, err := OpenDbConnection()
	if err != nil {
		return nil, err
	}
	defer closeDb(connection)

	branches, err := dao.getFilteredBranches(connection, filter)

	//	rows, err := connection.Query("SELECT DISTINCT branch FROM test_launches")
	//	if err != nil {
	//		return nil, err
	//	}
	//	defer closeRows(rows)

	//	branches := make([]*BranchInfoEntity, 0, 10)
	//	for rows.Next() {
	//		bi := new(BranchInfoEntity)
	//		scanErr := rows.Scan(&bi.BranchName)
	//		if scanErr != nil {
	//			log.Println(scanErr)
	//			continue
	//		}
	//		branches = append(branches, bi)
	//	}

	for i := 0; i < len(branches); i++ {
		rows, err := connection.Query("SELECT launch_id, creation_date, failed_num FROM test_launches WHERE branch = ? ORDER BY creation_date DESC LIMIT 1", branches[i].BranchName)
		if err != nil {
			log.Println(err)
			continue
		}
		if rows.Next() {
			scanErr := ScanStruct(rows, branches[i])
			if scanErr != nil {
				log.Println(scanErr)
			}
		}
		rows.Close()

		if !branches[i].LastLaunchFailedNum.Valid {
			failRows, err := connection.Query("SELECT test_case_id FROM test_cases JOIN test_case_failures ON test_case_id = parent_test_case_id WHERE parent_launch_id = ?", branches[i].LastLaunchId)
			if err != nil {
				log.Println(err)
				continue
			}
			if failRows.Next() {
				branches[i].LastLauchFailed = true
			}
			failRows.Close()
		} else {
			branches[i].LastLauchFailed = branches[i].LastLaunchFailedNum.Int64 > 0
		}
	}

	return branches, nil
}

func (dao *DaoService) GetAllLaunchesInBranch(branch string) []*TestLaunchEntity {
	rows, err := ExecuteSelect("SELECT launch_id, branch, label, creation_date FROM test_launches WHERE branch = ? ORDER BY creation_date", branch)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer closeRows(rows)

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
	defer closeRows(rows)

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
	defer closeRows(rows)

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

func (*DaoService) GetTestDynamics(testId int64) []*TestFullInfoEntity {
	rows, err := ExecuteSelect(
		"SELECT test_case_id, branch, name, package, class_name, status, parent_launch_id, creation_date, test_case_failure_id "+
			"FROM test_cases LEFT JOIN test_case_failures ON test_case_id = parent_test_case_id JOIN test_launches ON parent_launch_id = launch_id "+
			"WHERE md5_hash IN ( SELECT md5_hash FROM test_cases WHERE test_case_id=? ) "+
			"ORDER BY creation_date DESC",
		testId)
	if err != nil {
		log.Println(err)
		return nil
	}

	results := make([]*TestFullInfoEntity, 0, 10)
	for rows.Next() {
		testInfo := new(TestFullInfoEntity)
		ScanStruct(rows, testInfo)
		results = append(results, testInfo)
	}

	return results
}

func (*DaoService) GetAddedTestsInDiff(launchId1 int64, launchId2 int64) []*TestCaseEntity {
	newTestsRows, newTestRowsErr := ExecuteSelect(
		"SELECT test_case_id, name, package, class_name, status, parent_launch_id FROM test_cases WHERE parent_launch_id = ? AND md5_hash IN ( "+
			"SELECT md5_hash FROM test_cases WHERE parent_launch_id = ? EXCEPT "+
			"SELECT md5_hash FROM test_cases WHERE parent_launch_id = ?"+
			" ) ORDER BY status", launchId2, launchId2, launchId1)
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
		log.Println(err)
		return err
	}
	return nil
}

func (dao *DaoService) DeleteGivenLaunchWithAllPrevious(launchId int64) error {

	launchInfo := dao.GetLaunchInfo(launchId)
	if launchInfo == nil {
		return nil
	}

	_, err := ExecuteDelete("DELETE FROM test_launches WHERE branch = ? AND creation_date <= ?", launchInfo.Branch, launchInfo.CreateDate)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (*DaoService) DeleteBranch(branchName string) error {
	_, err := ExecuteDelete("DELETE FROM test_launches WHERE branch = ?", branchName)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (*DaoService) DeleteOrphans() error {
	_, err := ExecuteDelete(SQL_REMOVED_ORPHAN_TESTS)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = ExecuteDelete(SQL_REMOVED_ORPHAN_FAILURES)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (*DaoService) FindUser(login string, password string) *UserEntity {
	rows, err := ExecuteSelect(
		"SELECT user_id, login, password, is_active, first_name, last_name FROM users WHERE login = ? AND password = ?", login, password)
	if err != nil {
		log.Printf("Error selecting user with login = %v. Reason: %v\n", login, err)
		return nil
	}
	defer closeRows(rows)

	if !rows.Next() {
		return nil
	}

	userEntity := new(UserEntity)
	ScanStruct(rows, userEntity)

	return userEntity
}

func (*DaoService) GetUserById(userId int64) *UserEntity {
	rows, err := ExecuteSelect(
		"SELECT user_id, login, password, is_active, first_name, last_name FROM users WHERE user_id = ?", userId)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer closeRows(rows)

	if !rows.Next() {
		return nil
	}

	userEntity := new(UserEntity)
	ScanStruct(rows, userEntity)

	return userEntity
}

func (*DaoService) GetAllUsers() []*UserEntity {
	rows, err := ExecuteSelect(
		"SELECT user_id, login, password, is_active, first_name, last_name FROM users")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	users := make([]*UserEntity, 0, 10)
	for rows.Next() {
		user := new(UserEntity)
		ScanStruct(rows, user)
		users = append(users, user)
	}
	return users
}

func (*DaoService) UpdateUser(user *UserEntity) error {
	_, err := ExecuteInsert("UPDATE users SET is_active = ?, first_name = ?, last_name = ? WHERE user_id = ?",
		ConvertBool(user.IsActive), user.FirstName, user.LastName, user.UserId)

	return err
}

func (*DaoService) InsertUser(user *UserEntity) error {
	_, err := ExecuteInsert("INSERT INTO users (login, password, is_active, first_name, last_name) VALUES(?, ?, ?, ?, ?)",
		user.Login, user.Password, ConvertBool(user.IsActive), user.FirstName, user.LastName)

	return err
}

func (*DaoService) CreateUser(user *UserEntity) error {
	_, err := ExecuteInsert("INSERT INTO users(login, password, is_active, first_name, last_name) VALUES(?, ?, ?, ?, ?)",
		user.Login, user.Password, ConvertBool(user.IsActive), user.FirstName, user.LastName, user.UserId)

	return err
}
