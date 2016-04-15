package main

import (
	"log"
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

func serveShowBranches(context *HttpContext) {

	filter := extractBranchesFilter(context.Req)

	branches, err := DAO.GetAllBranchesInfo(filter)

	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	sort.Sort(sort.Reverse(SortableSlice(branches)))
	sort.Reverse(SortableSlice(branches))

	rendRrr := RenderInCommonTemplateEx(context, branches, "list_branches.html")
	if rendRrr != nil {
		http.Error(context.Resp, rendRrr.Error(), http.StatusInternalServerError)
	}
}

func serveFilterBranches(context *HttpContext) {

	rendRrr := RenderInCommonTemplateEx(context, nil, "filter_branches.html")
	if rendRrr != nil {
		http.Error(context.Resp, rendRrr.Error(), http.StatusInternalServerError)
	}
}

func serveLaunchesInBranchEx(context *HttpContext) {

	branchName := context.Req.URL.Query().Get("branch_name")

	launches := DAO.GetAllLaunchesInBranch(branchName)

	err := RenderInCommonTemplateEx(context, launches, "view_branch.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serverLaunchEx(context *HttpContext) {
	launchIdParam := context.Req.URL.Query().Get("launch_id")
	launchId, parseErr := strconv.Atoi(launchIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	testCases := DAO.GetAllTestsForLaunch(int64(launchId))
	launchInfo := DAO.GetLaunchInfo(int64(launchId))

	var dto ViewLaunchDTO
	dto.LaunchId = launchId
	dto.Branch = launchInfo.Branch
	dto.Label = launchInfo.Label
	dto.Tests = testCases

	dto.FailedTestsNum = TestsWithStatusNum(dto.Tests, TEST_CASE_STATUS_FAILED)
	dto.PassedTestsNum = TestsWithStatusNum(dto.Tests, TEST_CASE_STATUS_PASSED)
	dto.SkippedTestsNum = TestsWithStatusNum(dto.Tests, TEST_CASE_STATUS_SKIPPED)

	err := RenderInCommonTemplateEx(context, dto, "view_launch.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func servePackageEx(context *HttpContext) {
	launchIdParam := context.Req.URL.Query().Get("launch_id")
	launchId, parseErr := strconv.Atoi(launchIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

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

	err := RenderInCommonTemplateEx(context, dto, "view_package.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serverLaunchPackagesEx(context *HttpContext) {
	launchIdParam := context.Req.URL.Query().Get("launch_id")
	launchId, parseErr := strconv.Atoi(launchIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	packages, err := DAO.GetPackagesForLaunch(int64(launchId))
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	var dto PackagesDTO
	dto.LaunchId = launchId
	dto.Packages = packages

	err = RenderInCommonTemplateEx(context, dto, "view_packages.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serverTestCaseEx(context *HttpContext) {
	testCaseIdParam := context.Req.URL.Query().Get("test_id")
	testCaseId, parseErr := strconv.Atoi(testCaseIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	testCase := DAO.GetTestCaseDetails(int64(testCaseId))

	err := RenderInCommonTemplateEx(context, testCase, "view_test_case.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serverTestDymanicsEx(context *HttpContext) {
	testCaseIdParam := context.Req.URL.Query().Get("test_id")
	testCaseId, parseErr := strconv.Atoi(testCaseIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tests := DAO.GetTestDynamics(int64(testCaseId))

	err := RenderInCommonTemplateEx(context, tests, "test_dynamics.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveDiffLaunchesEx(context *HttpContext) {
	launchId1Param := context.Req.URL.Query().Get("launch_id1")
	launchId1, launchId1ParseErr := strconv.Atoi(launchId1Param)
	if launchId1ParseErr != nil {
		log.Println(launchId1ParseErr)
		http.Error(context.Resp, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	launchId2Param := context.Req.URL.Query().Get("launch_id2")
	launchId2, launchId2ParseErr := strconv.Atoi(launchId2Param)
	if launchId2ParseErr != nil {
		log.Println(launchId2ParseErr)
		http.Error(context.Resp, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

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

	err := RenderInCommonTemplateEx(context, dto, "view_launches_diff.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveDeleteLaunchEx(context *HttpContext) {
	session := context.Session
	if !session.IsLoggedIn() {
		errDto := HttpErrDTO{Code: 403, Message: "No permissions"}
		if renderErr := RenderInCommonTemplateEx(context, errDto, "error.html"); renderErr != nil {
			http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	launchIdParam := context.Req.URL.Query().Get("launch_id")
	launchId, parseErr := strconv.Atoi(launchIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(context.Resp, "Invalid launch id", http.StatusBadRequest)
		return
	}

	launchInfo := DAO.GetLaunchInfo(int64(launchId))
	if launchInfo == nil {
		http.Error(context.Resp, "Unable to find launch "+launchIdParam, http.StatusBadRequest)
		return
	}

	err := DAO.DeleteLaunch(int64(launchId))
	if err != nil {
		daoErr := HttpErrDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		if renderErr := RenderInCommonTemplateEx(context, daoErr, "error.html"); renderErr != nil {
			http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// TODO : Find why orphans tests occure after launche is deleted
	DAO.DeleteOrphans()

	http.Redirect(context.Resp, context.Req, "/branch?branch_name="+launchInfo.Branch, http.StatusMovedPermanently)
}

func serveDeleteThisAndPreviousLaunch(context *HttpContext) {
	session := context.Session
	if !session.IsLoggedIn() {
		errDto := HttpErrDTO{Code: 403, Message: "No permissions"}
		if renderErr := RenderInCommonTemplateEx(context, errDto, "error.html"); renderErr != nil {
			http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	launchIdParam := context.Req.URL.Query().Get("launch_id")
	launchId, parseErr := strconv.Atoi(launchIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(context.Resp, "Invalid launch id", http.StatusBadRequest)
		return
	}

	launchInfo := DAO.GetLaunchInfo(int64(launchId))
	if launchInfo == nil {
		http.Error(context.Resp, "Unable to find launch "+launchIdParam, http.StatusBadRequest)
		return
	}

	err := DAO.DeleteGivenLaunchWithAllPrevious(int64(launchId))
	if err != nil {
		daoErr := HttpErrDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		if renderErr := RenderInCommonTemplateEx(context, daoErr, "error.html"); renderErr != nil {
			http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// TODO : Find why orphans tests occure after launche is deleted
	DAO.DeleteOrphans()

	http.Redirect(context.Resp, context.Req, "/branch?branch_name="+launchInfo.Branch, http.StatusMovedPermanently)
	//http.Redirect(context.Resp, context.Req, "/", http.StatusMovedPermanently)
}

func serveDeleteBranch(context *HttpContext) {
	session := context.Session
	if !session.IsLoggedIn() {
		errDto := HttpErrDTO{Code: 403, Message: "No permissions"}
		if renderErr := RenderInCommonTemplateEx(context, errDto, "error.html"); renderErr != nil {
			http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	branchName := context.Req.URL.Query().Get("branch")
	if branchName == "" {
		http.Error(context.Resp, "Invalid branch name", http.StatusBadRequest)
		return
	}

	err := DAO.DeleteBranch(branchName)
	if err != nil {
		daoErr := HttpErrDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		if renderErr := RenderInCommonTemplateEx(context, daoErr, "error.html"); renderErr != nil {
			http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// TODO : Find why orphans tests occure after launche is deleted
	DAO.DeleteOrphans()

	http.Redirect(context.Resp, context.Req, "/all-branches", http.StatusMovedPermanently)
}

func handleLoginEx(context *HttpContext) {

	session := context.Session
	if session.IsLoggedIn() {
		http.Redirect(context.Resp, context.Req, "/", http.StatusFound)
		return
	}

	if context.Req.Method != "POST" {
		if err := RenderInCommonTemplateEx(context, nil, "login.html"); err != nil {
			http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
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
		if err := RenderInCommonTemplateEx(context, errMsg, "login.html"); err != nil {
			http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

	} else {
		SESSION_MANAGER.InitSession(context.Resp, userInfo)
		context.Resp.Header().Set("Cache-Control", "no-cache")
		context.Resp.Header().Set("Pragma", "no-cache")
		http.Redirect(context.Resp, context.Req, "/", http.StatusFound)
	}
}

func handleLogoutEx(context *HttpContext) {
	SESSION_MANAGER.ClearSession(context.Req, context.Resp)
	http.Redirect(context.Resp, context.Req, "/login", http.StatusFound)
}
