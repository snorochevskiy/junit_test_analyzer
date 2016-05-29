package main

import (
	"database/sql"
	"log"
	"runtime"
)

type DaoPanicErr struct {
	Message string
}

func (e DaoPanicErr) String() string {
	return e.Message
}

func DaoChechAndPanic(err error) {
	if err == nil {
		return
	}
	pc, _, _, _ := runtime.Caller(1)
	msg := runtime.FuncForPC(pc).Name() + ": " + err.Error()
	panic(DaoPanicErr{Message: msg})
}

type DaoService struct {
}

var DAO DaoService = DaoService{}

const TEST_CASE_STATUS_FAILED = "FAILED"
const TEST_CASE_STATUS_SKIPPED = "SKIPPED"
const TEST_CASE_STATUS_PASSED = "PASSED"

func (*DaoService) GetAllProjects() []*ProjectEntity {
	rows, err := ExecuteSelect("SELECT project_id, project_name, description FROM test_projects ORDER BY project_name")
	DaoChechAndPanic(err)
	defer rows.Close()

	projects := make([]*ProjectEntity, 0, 10)
	for rows.Next() {
		project := new(ProjectEntity)
		scanErr := ScanStruct(rows, project)
		if scanErr != nil {
			log.Println(scanErr)
			continue
		}
		projects = append(projects, project)
	}
	return projects
}

func (*DaoService) GetBranchesInProject(projectId int64) []*ProjectBranchEntity {
	rows, err := ExecuteSelect("SELECT branch_id, branch_name FROM project_branches WHERE parent_project_id = ?", projectId)
	DaoChechAndPanic(err)
	defer rows.Close()

	branches := make([]*ProjectBranchEntity, 0, 10)
	for rows.Next() {
		branch := new(ProjectBranchEntity)
		DaoChechAndPanic(ScanStruct(rows, branch))
		branch.ParentProjectId = projectId
		branches = append(branches, branch)
	}
	return branches
}

func (*DaoService) GetParentProjectForBranch(branchId int64) (int64, error) {
	rows, err := ExecuteSelect("SELECT parent_project_id FROM project_branches WHERE branch_id = ?", branchId)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, nil
	}

	var projectId int64 = 0
	if err := rows.Scan(&projectId); err != nil {
		return 0, err
	}
	return projectId, nil
}

func (*DaoService) GetProjectIdByProjectName(porjectName string) int64 {
	rows, err := ExecuteSelect("SELECT project_id FROM test_projects WHERE project_name = ?", porjectName)
	DaoChechAndPanic(err)
	defer rows.Close()

	if !rows.Next() {
		return 0
	}

	var projectId int64 = 0
	DaoChechAndPanic(rows.Scan(&projectId))

	return projectId
}

func (dao *DaoService) getFilteredBranchesIds(connection *sql.DB, projectId int64, filter *BranchesFilter) []*ProjectBranchEntity {

	branches := dao.GetBranchesInProject(projectId)
	return branches

	//  -- ONE MORE STEP TO FILTER IF REQUIRED
	//	sqlText := "SELECT DISTINCT branch FROM test_launches"
	//	params := make([]interface{}, 0, 5)

	//	if filter != nil && filter.HasSomethingToFilter() {
	//		sqlText += " WHERE"
	//		if filter.LabelTemplate != "" {
	//			sqlText += " label LIKE ?"
	//			params = append(params, strings.Replace(filter.LabelTemplate, "*", "%", -1))
	//		}
	//	}

	//	rows, err := connection.Query(sqlText, params...)
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
	//	return branches, nil
}

func (dao *DaoService) GetAllBranchesInfo(projectId int64, filter *BranchesFilter) []*BranchDetailedInfoEntity {
	connection, err := OpenDbConnection()
	DaoChechAndPanic(err)
	defer closeDb(connection)

	projectBranches := dao.getFilteredBranchesIds(connection, projectId, filter)

	branches := make([]*BranchDetailedInfoEntity, 0, 10)
	for i := 0; i < len(projectBranches); i++ {
		rows, err := connection.Query("SELECT launch_id, creation_date, failed_num FROM test_launches WHERE parent_branch_id = ? ORDER BY creation_date DESC LIMIT 1", projectBranches[i].Id)
		if err != nil {
			log.Println(err)
			continue
		}
		if rows.Next() {
			branchInfo := new(BranchDetailedInfoEntity)
			scanErr := ScanStruct(rows, branchInfo)
			if scanErr != nil {
				log.Println(scanErr)
			}
			branchInfo.Id = projectBranches[i].Id
			branchInfo.BranchName = projectBranches[i].Name
			branches = append(branches, branchInfo)
		}
		rows.Close()

	}

	return branches
}

func (dao *DaoService) GetAllLaunchesInBranch(branchId int64) []*TestLaunchEntity {
	rows, err := ExecuteSelect("SELECT launch_id, label, creation_date FROM test_launches WHERE parent_branch_id = ? ORDER BY creation_date", branchId)
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

	rows, err := ExecuteSelect("SELECT launch_id, parent_branch_id, label, creation_date FROM test_launches WHERE launch_id = ?", launchId)
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
	DaoChechAndPanic(err)
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
	DaoChechAndPanic(err)
	defer rows.Close()

	testCases := make([]*TestCaseEntity, 0, 10)
	for rows.Next() {
		testCase := new(TestCaseEntity)
		ScanStruct(rows, testCase)
		testCases = append(testCases, testCase)
	}
	return testCases
}

func (*DaoService) GetPackagesForLaunch(launchId int64) []*PackageEntity {
	connection, err := OpenDbConnection()
	DaoChechAndPanic(err)
	defer connection.Close()

	stmt, err := connection.Prepare("SELECT count(*) FROM test_cases WHERE parent_launch_id = ? AND package = ? AND status = ?")
	DaoChechAndPanic(err)

	rows, err := connection.Query("SELECT package, count(*) as tests_num FROM test_cases WHERE parent_launch_id = ? GROUP BY package ORDER BY package", launchId)
	DaoChechAndPanic(err)
	defer rows.Close()

	packages := make([]*PackageEntity, 0, 10)
	for rows.Next() {
		packageEntity := new(PackageEntity)
		ScanStruct(rows, packageEntity)

		failedTestsNumRow := stmt.QueryRow(launchId, packageEntity.Package, TEST_CASE_STATUS_FAILED)
		var failedTestIntPackage int
		err := failedTestsNumRow.Scan(&failedTestIntPackage)
		DaoChechAndPanic(err)
		packageEntity.FailedTestsNum = failedTestIntPackage

		passedTestsNumRow := stmt.QueryRow(launchId, packageEntity.Package, TEST_CASE_STATUS_PASSED)
		var passedTestIntPackage int
		err = passedTestsNumRow.Scan(&passedTestIntPackage)
		DaoChechAndPanic(err)
		packageEntity.PassedTestsNum = passedTestIntPackage

		skippedTestsNumRow := stmt.QueryRow(launchId, packageEntity.Package, TEST_CASE_STATUS_SKIPPED)
		var skippedTestIntPackage int
		err = skippedTestsNumRow.Scan(&skippedTestIntPackage)
		DaoChechAndPanic(err)
		packageEntity.SkippedTestsNum = skippedTestIntPackage

		packages = append(packages, packageEntity)
	}
	return packages
}

func (*DaoService) GetTestCaseDetails(testCaseId int64) *TestCaseEntity {
	rows, err := ExecuteSelect("SELECT test_case_id, name, package, class_name, status, parent_launch_id FROM test_cases WHERE test_case_id = ?", testCaseId)
	if err != nil {
		panic(DaoPanicErr{Message: err.Error()})
	}
	defer rows.Close()

	if !rows.Next() {
		return nil
	}

	testCase := new(TestCaseEntity)
	scanErr := ScanStruct(rows, testCase)
	if scanErr != nil {
		panic(DaoPanicErr{Message: scanErr.Error()})
	}

	if testCase.Status == TEST_CASE_STATUS_FAILED {
		failedInfoRows, failedInfoErr := ExecuteSelect("SELECT test_case_failure_id, failure_message, failure_type, failure_text FROM test_case_failures WHERE parent_test_case_id = ?", testCaseId)
		if failedInfoErr != nil {
			panic(DaoPanicErr{Message: failedInfoErr.Error()})
		} else if failedInfoRows.Next() {
			testFailure := new(FailureEntity)
			scanErr := ScanStruct(failedInfoRows, testFailure)
			if scanErr != nil {
				panic(DaoPanicErr{Message: scanErr.Error()})
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
	if err := row.Scan(num); err != nil {
		panic(DaoPanicErr{Message: err.Error()})
	}
	return *num
}

func (*DaoService) GetTestDynamics(testId int64) []*TestFullInfoEntity {
	rows, err := ExecuteSelect(
		"SELECT test_case_id, parent_branch_id, name, package, class_name, status, parent_launch_id, creation_date, test_case_failure_id "+
			"FROM test_cases LEFT JOIN test_case_failures ON test_case_id = parent_test_case_id JOIN test_launches ON parent_launch_id = launch_id "+
			"WHERE md5_hash IN ( SELECT md5_hash FROM test_cases WHERE test_case_id=? ) "+
			"ORDER BY creation_date DESC",
		testId)
	if err != nil {
		panic(DaoPanicErr{Message: err.Error()})
	}

	results := make([]*TestFullInfoEntity, 0, 10)
	for rows.Next() {
		testInfo := new(TestFullInfoEntity)
		if err := ScanStruct(rows, testInfo); err != nil {
			panic(DaoPanicErr{Message: err.Error()})
		}
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

func (dao *DaoService) DeleteGivenLaunchWithAllPrevious(launchId int64) {

	launchInfo := dao.GetLaunchInfo(launchId)
	if launchInfo == nil {
		return
	}

	_, err := ExecuteDelete("DELETE FROM test_launches WHERE parent_branch_id = ? AND creation_date <= ?", launchInfo.BranchId, launchInfo.CreateDate)
	DaoChechAndPanic(err)
}

func (dao *DaoService) DeleteAllLaunchesInBranch(branchId int64) {
	DaoChechAndPanic(ExecuteDeleteNoResult("DELETE FROM test_launches WHERE parent_branch_id = ?", branchId))

	dao.DeleteBranch(branchId)
}

func (dao *DaoService) DeleteBranchIfEmpty(branchId int64) {
	rows, err := ExecuteSelect("SELECT Count(*) FROM test_launches WHERE parent_branch_id = ?", branchId)
	DaoChechAndPanic(err)

	rows.Next()
	var numberOfLaunches int
	rows.Scan(&numberOfLaunches)
	rows.Close()

	if numberOfLaunches > 0 {
		return
	}

	dao.DeleteBranch(branchId)
}

func (dao *DaoService) DeleteBranch(branchId int64) {
	DaoChechAndPanic(ExecuteDeleteNoResult("DELETE FROM project_branches WHERE branch_id = ?", branchId))
}

func (*DaoService) DeleteOrphans() {
	DaoChechAndPanic(ExecuteDeleteNoResult(SQL_REMOVED_ORPHAN_TESTS))

	DaoChechAndPanic(ExecuteDeleteNoResult(SQL_REMOVED_ORPHAN_FAILURES))
}

func (*DaoService) FindUser(login string, password string) *UserEntity {
	rows, err := ExecuteSelect(
		"SELECT user_id, login, password, is_active, first_name, last_name FROM users WHERE login = ? AND password = ?", login, password)
	DaoChechAndPanic(err)
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
	DaoChechAndPanic(err)
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
	DaoChechAndPanic(err)
	defer rows.Close()

	users := make([]*UserEntity, 0, 10)
	for rows.Next() {
		user := new(UserEntity)
		ScanStruct(rows, user)
		users = append(users, user)
	}
	return users
}

func (*DaoService) UpdateUser(user *UserEntity) {
	_, err := ExecuteInsert("UPDATE users SET is_active = ?, first_name = ?, last_name = ? WHERE user_id = ?",
		ConvertBool(user.IsActive), user.FirstName, user.LastName, user.UserId)

	DaoChechAndPanic(err)
}

func (*DaoService) InsertUser(user *UserEntity) {
	_, err := ExecuteInsert("INSERT INTO users (login, password, is_active, first_name, last_name) VALUES(?, ?, ?, ?, ?)",
		user.Login, user.Password, ConvertBool(user.IsActive), user.FirstName, user.LastName)

	DaoChechAndPanic(err)
}

func (*DaoService) CreateUser(user *UserEntity) {
	_, err := ExecuteInsert("INSERT INTO users(login, password, is_active, first_name, last_name) VALUES(?, ?, ?, ?, ?)",
		user.Login, user.Password, ConvertBool(user.IsActive), user.FirstName, user.LastName, user.UserId)

	DaoChechAndPanic(err)
}
