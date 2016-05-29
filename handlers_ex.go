package main

import (
	"jutra/router"
	sm "jutra/session"
	"net/http"
	"sort"
	"strconv"
)

type BranchesFilter struct {
	LabelTemplate string
}

func (f *BranchesFilter) HasSomethingToFilter() bool {
	return f.LabelTemplate != ""
}

func extractBranchesFilter(r *http.Request) *BranchesFilter {
	filter := new(BranchesFilter)

	lblTemplate := r.URL.Query().Get("label")
	if lblTemplate != "" {
		filter.LabelTemplate = lblTemplate
	}

	return filter
}

func serveMainPage(context *router.HttpContext) {
	projects := DAO.GetAllProjects()

	var ro MainPageRO
	ro.Projects = projects

	RenderInCommonTemplateEx(context, ro, "main_page.html")

}

func serveProject(context *router.HttpContext) {

	projectIdStr := context.PathParams["projectId"]
	var projectId int64

	if projectIdStr == "" {
		// Get ID of default project
		projectId = DAO.GetProjectIdByProjectName("")
	} else {
		projectId = ParseInt64(projectIdStr, "Invalid project ID")
	}

	filter := extractBranchesFilter(context.Req)

	branches := DAO.GetAllBranchesInfo(projectId, filter)

	sort.Sort(sort.Reverse(SortableSlice(branches)))
	sort.Reverse(SortableSlice(branches))

	RenderInCommonTemplateEx(context, branches, "list_branches.html")

}

func serveFilterBranches(context *router.HttpContext) {

	RenderInCommonTemplateEx(context, nil, "filter_branches.html")

}

func serveLaunchesInBranchEx(context *router.HttpContext) {

	branchIdStr := context.Req.URL.Query().Get("branchId")
	branchId := ParseInt64(branchIdStr, "Wrong branch ID")

	launches := DAO.GetAllLaunchesInBranch(branchId)

	RenderInCommonTemplateEx(context, launches, "view_branch.html")
}

func serverLaunchEx(context *router.HttpContext) {
	launchIdParam := context.Req.URL.Query().Get("launch_id")
	launchId := ParseInt64(launchIdParam, "Invalid test run ID")

	testCases := DAO.GetAllTestsForLaunch(launchId)
	launchInfo := DAO.GetLaunchInfo(launchId)

	var dto ViewLaunchDTO
	dto.LaunchId = launchId
	dto.BranchId = launchInfo.BranchId
	dto.Label = launchInfo.Label
	dto.Tests = testCases

	dto.FailedTestsNum = TestsWithStatusNum(dto.Tests, TEST_CASE_STATUS_FAILED)
	dto.PassedTestsNum = TestsWithStatusNum(dto.Tests, TEST_CASE_STATUS_PASSED)
	dto.SkippedTestsNum = TestsWithStatusNum(dto.Tests, TEST_CASE_STATUS_SKIPPED)

	RenderInCommonTemplateEx(context, dto, "view_launch.html")
}

func servePackageEx(context *router.HttpContext) {
	launchIdParam := context.Req.URL.Query().Get("launch_id")
	launchId := ParseInt64(launchIdParam, "Invalid test run ID")

	packageParam := context.Req.URL.Query().Get("package")
	if packageParam == "" {
		http.Error(context.Resp, "package should be specified", http.StatusInternalServerError)
		return
	}

	testCases := DAO.GetAllTestsForPackage(int64(launchId), packageParam)

	var dto ViewPackageDTO
	dto.LaunchId = launchId
	dto.Package = packageParam
	dto.Tests = testCases

	RenderInCommonTemplateEx(context, dto, "view_package.html")
}

func serverLaunchPackagesEx(context *router.HttpContext) {
	launchIdParam := context.Req.URL.Query().Get("launch_id")
	launchId := ParseInt64(launchIdParam, "Invalid test run ID")

	packages := DAO.GetPackagesForLaunch(int64(launchId))

	var dto PackagesDTO
	dto.LaunchId = launchId
	dto.Packages = packages

	RenderInCommonTemplateEx(context, dto, "view_packages.html")
}

func serverTestCaseEx(context *router.HttpContext) {
	testCaseIdParam := context.Req.URL.Query().Get("test_id")
	testCaseId := ParseInt64(testCaseIdParam, "Invalid test ID")

	testCase := DAO.GetTestCaseDetails(int64(testCaseId))

	RenderInCommonTemplateEx(context, testCase, "view_test_case.html")
}

func serverTestDymanicsEx(context *router.HttpContext) {
	testCaseIdParam := context.Req.URL.Query().Get("test_id")
	testCaseId := ParseInt64(testCaseIdParam, "Invalid test ID")

	tests := DAO.GetTestDynamics(testCaseId)

	RenderInCommonTemplateEx(context, tests, "test_dynamics.html")
}

func serveDiffLaunchesEx(context *router.HttpContext) {
	launchId1Param := context.Req.URL.Query().Get("launch_id1")
	launchId1 := ParseInt64(launchId1Param, "Invalid left test run ID")

	launchId2Param := context.Req.URL.Query().Get("launch_id2")
	launchId2 := ParseInt64(launchId2Param, "Invalid right test run ID")

	var dto LaunchesDiffDTO
	dto.LaunchId1 = launchId1
	dto.LaunchId2 = launchId2
	dto.AddedTests = DAO.GetAddedTestsInDiff(int64(launchId1), int64(launchId2))
	dto.RemovedTests = DAO.GetAddedTestsInDiff(int64(launchId2), int64(launchId1))
	dto.PassedToFailedTests = DAO.GetTestsFromStatus1ToStatus2(int64(launchId1), int64(launchId2), TEST_CASE_STATUS_PASSED, TEST_CASE_STATUS_FAILED)
	dto.PassedToSkippedTests = DAO.GetTestsFromStatus1ToStatus2(int64(launchId1), int64(launchId2), TEST_CASE_STATUS_PASSED, TEST_CASE_STATUS_SKIPPED)
	dto.FailedToPassedTests = DAO.GetTestsFromStatus1ToStatus2(int64(launchId1), int64(launchId2), TEST_CASE_STATUS_FAILED, TEST_CASE_STATUS_PASSED)
	dto.FailedToSkippedTests = DAO.GetTestsFromStatus1ToStatus2(int64(launchId1), int64(launchId2), TEST_CASE_STATUS_FAILED, TEST_CASE_STATUS_SKIPPED)
	dto.SkippedToFailedTests = DAO.GetTestsFromStatus1ToStatus2(int64(launchId1), int64(launchId2), TEST_CASE_STATUS_SKIPPED, TEST_CASE_STATUS_FAILED)
	dto.SkippedToPassedTests = DAO.GetTestsFromStatus1ToStatus2(int64(launchId1), int64(launchId2), TEST_CASE_STATUS_SKIPPED, TEST_CASE_STATUS_PASSED)

	RenderInCommonTemplateEx(context, dto, "view_launches_diff.html")
}

func serveDeleteLaunchEx(context *router.HttpContext) {
	session := context.Session
	if !session.IsLoggedIn() {
		errDto := HttpErrDTO{Code: 403, Message: "No permissions"}
		RenderInCommonTemplateEx(context, errDto, "error.html")
		return
	}

	launchIdParam := context.Req.URL.Query().Get("launch_id")
	launchId := ParseInt64(launchIdParam, "Invalid test run ID")

	launchInfo := DAO.GetLaunchInfo(int64(launchId))
	if launchInfo == nil {
		http.Error(context.Resp, "Unable to find launch "+launchIdParam, http.StatusBadRequest)
		return
	}

	err := DAO.DeleteLaunch(int64(launchId))
	if err != nil {
		daoErr := HttpErrDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		RenderInCommonTemplateEx(context, daoErr, "error.html")
		return
	}

	// TODO : Find why orphans tests occure after launche is deleted
	DAO.DeleteOrphans()

	http.Redirect(context.Resp, context.Req, "/branch?branchId="+strconv.FormatInt(launchInfo.BranchId, 10), http.StatusMovedPermanently)
}

func serveDeleteThisAndPreviousLaunch(context *router.HttpContext) {
	session := context.Session
	if !session.IsLoggedIn() {
		errDto := HttpErrDTO{Code: 403, Message: "No permissions"}
		RenderInCommonTemplateEx(context, errDto, "error.html")
		return
	}

	launchIdParam := context.Req.URL.Query().Get("launch_id")
	launchId := ParseInt64(launchIdParam, "Invalid launch id")

	launchInfo := DAO.GetLaunchInfo(launchId)
	if launchInfo == nil {
		http.Error(context.Resp, "Can't find run", http.StatusBadRequest)
		return
	}

	DAO.DeleteGivenLaunchWithAllPrevious(launchId)

	// TODO : Find why orphans tests occure after launche is deleted
	DAO.DeleteOrphans()

	http.Redirect(context.Resp, context.Req, "/branch?branchId="+strconv.FormatInt(launchInfo.BranchId, 10), http.StatusMovedPermanently)
	//http.Redirect(context.Resp, context.Req, "/", http.StatusMovedPermanently)
}

func serveDeleteBranch(context *router.HttpContext) {
	session := context.Session
	if !session.IsLoggedIn() {
		errDto := HttpErrDTO{Code: 403, Message: "No permissions"}
		RenderInCommonTemplateEx(context, errDto, "error.html")
		return
	}

	branchIdStr := context.PathParams["branchId"]
	branchId := ParseInt64(branchIdStr, "Invalid branch ID")

	projectId, err := DAO.GetParentProjectForBranch(branchId)
	if err != nil {
		http.Error(context.Resp, "Can't find project for given branch ID", http.StatusBadRequest)
		return
	}

	DAO.DeleteAllLaunchesInBranch(branchId)

	// TODO : Find why orphans tests occure after launche is deleted
	DAO.DeleteOrphans()

	http.Redirect(context.Resp, context.Req, "/project/"+strconv.FormatInt(projectId, 10), http.StatusMovedPermanently)
}

func handleLoginEx(context *router.HttpContext) {

	session := context.Session
	if session.IsLoggedIn() {
		http.Redirect(context.Resp, context.Req, "/", http.StatusFound)
		return
	}

	if context.Req.Method != "POST" {
		RenderInCommonTemplateEx(context, nil, "login.html")
		return
	}

	login := context.Req.FormValue("login")
	password := context.Req.FormValue("password")
	userInfo := DAO.FindUser(login, password)

	errMsg := ""
	if login == "" {

	} else if userInfo == nil {
		errMsg = "Can't find user with login " + login
	} else if userInfo.Password != password {
		errMsg = "Wrong password"
	}

	if login == "" || errMsg != "" {
		RenderInCommonTemplateEx(context, errMsg, "login.html")
		return

	}
	sm.InitSession(context.Resp, userInfo)
	context.Resp.Header().Set("Cache-Control", "no-cache")
	context.Resp.Header().Set("Pragma", "no-cache")
	http.Redirect(context.Resp, context.Req, "/", http.StatusFound)
}

func handleLogoutEx(context *router.HttpContext) {
	sm.ClearSession(context.Req, context.Resp)
	http.Redirect(context.Resp, context.Req, "/login", http.StatusFound)
}
