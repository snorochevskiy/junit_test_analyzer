package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func StartServer(port string) {

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", maxAgeHandler(3600, http.StripPrefix("/static/", fs)))

	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)

	http.HandleFunc("/branch", serveLaunchesInBranch)
	http.HandleFunc("/launch", serverLaunch)
	http.HandleFunc("/packages", serverLaunchPackages)
	http.HandleFunc("/package", servePackage)
	http.HandleFunc("/test", serverTestCase)
	http.HandleFunc("/dynamics", serverTestDymanics)
	http.HandleFunc("/diff", serveDiffLaunches)
	http.HandleFunc("/delete-launch", serveDeleteLaunch)

	http.HandleFunc("/admin/list-users", serveListUsers)

	http.HandleFunc("/", serveRoot)

	log.Println("Listening...")
	http.ListenAndServe(":"+port, nil)
}

func maxAgeHandler(seconds int, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", seconds))
		h.ServeHTTP(w, r)
	})
}

func serveRoot(w http.ResponseWriter, r *http.Request) {

	branches := DAO.GetAllBranches()

	err := RenderInCommonTemplate(w, branches, "list_branches.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveLaunchesInBranch(w http.ResponseWriter, r *http.Request) {

	branchName := r.URL.Query().Get("branch_name")

	launches := DAO.GetAllLaunchesInBranch(branchName)

	err := RenderInCommonTemplate(w, launches, "view_branch.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func TestsWithStatusNum(tests []*TestCaseEntity, status string) int {
	counter := 0
	for i := 0; i < len(tests); i++ {
		if tests[i].Status == status {
			counter++
		}
	}
	return counter
}

func serverLaunch(w http.ResponseWriter, r *http.Request) {
	launchIdParam := r.URL.Query().Get("launch_id")
	launchId, parseErr := strconv.Atoi(launchIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

	err := RenderInCommonTemplate(w, dto, "view_launch.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func servePackage(w http.ResponseWriter, r *http.Request) {
	launchIdParam := r.URL.Query().Get("launch_id")
	launchId, parseErr := strconv.Atoi(launchIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	packageParam := r.URL.Query().Get("package")
	if packageParam == "" {
		http.Error(w, "package should be specified", http.StatusInternalServerError)
		return
	}

	testCases := DAO.GetAllTestsForPackage(int64(launchId), packageParam)

	var dto ViewPackageDTO
	dto.LaunchId = launchId
	dto.Package = packageParam
	dto.Tests = testCases

	err := RenderInCommonTemplate(w, dto, "view_package.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serverLaunchPackages(w http.ResponseWriter, r *http.Request) {
	launchIdParam := r.URL.Query().Get("launch_id")
	launchId, parseErr := strconv.Atoi(launchIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	packages, err := DAO.GetPackagesForLaunch(int64(launchId))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	var dto PackagesDTO
	dto.LaunchId = launchId
	dto.Packages = packages

	err = RenderInCommonTemplate(w, dto, "view_packages.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serverTestCase(w http.ResponseWriter, r *http.Request) {
	testCaseIdParam := r.URL.Query().Get("test_id")
	testCaseId, parseErr := strconv.Atoi(testCaseIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	testCase := DAO.GetTestCaseDetails(int64(testCaseId))

	err := RenderInCommonTemplate(w, testCase, "view_test_case.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serverTestDymanics(w http.ResponseWriter, r *http.Request) {
	testCaseIdParam := r.URL.Query().Get("test_id")
	testCaseId, parseErr := strconv.Atoi(testCaseIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tests := DAO.GetTestDynamics(int64(testCaseId))

	err := RenderInCommonTemplate(w, tests, "test_dynamics.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveDiffLaunches(w http.ResponseWriter, r *http.Request) {
	launchId1Param := r.URL.Query().Get("launch_id1")
	launchId1, launchId1ParseErr := strconv.Atoi(launchId1Param)
	if launchId1ParseErr != nil {
		log.Println(launchId1ParseErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	launchId2Param := r.URL.Query().Get("launch_id2")
	launchId2, launchId2ParseErr := strconv.Atoi(launchId2Param)
	if launchId2ParseErr != nil {
		log.Println(launchId2ParseErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

	err := RenderInCommonTemplate(w, dto, "view_launches_diff.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveDeleteLaunch(w http.ResponseWriter, r *http.Request) {
	session := SessionManager.GetSessionForRequest(r)
	if session == nil || session.User == nil {
		if renderErr := RenderInCommonTemplate(w, HttpErrDTO{Code: 403, Message: "No permissions"}, "error.html"); renderErr != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		return
	}

	launchIdParam := r.URL.Query().Get("launch_id")
	launchId, parseErr := strconv.Atoi(launchIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	launchInfo := DAO.GetLaunchInfo(int64(launchId))
	if launchInfo == nil {
		http.Error(w, "Unable to find launch "+launchIdParam, http.StatusInternalServerError)
		return
	}

	err := DAO.DeleteLaunch(int64(launchId))
	if err != nil {
		if renderErr := RenderInCommonTemplate(w, err.Error(), "error.html"); renderErr != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		return
	}

	http.Redirect(w, r, "/branch?branch_name="+launchInfo.Branch, http.StatusMovedPermanently)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {

	session := SessionManager.GetSessionForRequest(r)
	if session != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if r.Method != "POST" {
		if err := RenderInCommonTemplate(w, nil, "login.html"); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")
	userInfo := DAO.FindUser(login, password)

	errMsg := ""
	if login == "" {

	} else if userInfo == nil {
		errMsg = "Can't find user with login " + login
	} else if userInfo.Password != password {
		errMsg = "Wrong password"
	}

	if login == "" || errMsg != "" {

		if err := RenderInCommonTemplate(w, userInfo, "login.html"); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

	} else {
		SessionManager.InitSession(w, userInfo)
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Pragma", "no-cache")
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	SessionManager.ClearSession(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}
