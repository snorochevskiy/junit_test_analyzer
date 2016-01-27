package main

import (
	"log"
	"net/http"
	"strconv"
)

type MessageDTO struct {
	MessageText string
}

func startServer(port string) {

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/branch", serveLaunchesInBranch)
	http.HandleFunc("/launch", serverLaunch)
	http.HandleFunc("/packages", serverLaunchPackages)
	http.HandleFunc("/package", servePackage)
	http.HandleFunc("/test", serverTestCase)
	http.HandleFunc("/diff", serveDiffLaunches)
	http.HandleFunc("/delete-launch", serveDeleteLaunch)
	http.HandleFunc("/", serveRoot)

	log.Println("Listening...")
	http.ListenAndServe(":"+port, nil)
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

type ViewLaunchDTO struct {
	LaunchId int
	Label    string
	Branch   string
	Tests    []*TestCaseEntity
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

	err := RenderInCommonTemplate(w, dto, "view_launch.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

type ViewPackageDTO struct {
	LaunchId int
	Package  string
	Tests    []*TestCaseEntity
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

type PackagesDTO struct {
	LaunchId int
	Packages []*PackageEntity
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

type LaunchesDiffDTO struct {
	LaunchId1            int
	LaunchId2            int
	NewTests             []*TestCaseEntity
	PassedToFailedTests  []*TestCaseEntity
	PassedToSkippedTests []*TestCaseEntity
	FailedToPassedTests  []*TestCaseEntity
	FailedToSkippedTests []*TestCaseEntity
	SkippedToFailedTests []*TestCaseEntity
	SkippedToPassedTests []*TestCaseEntity
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
	dto.NewTests = DAO.GetNewTestsInDiff(int64(launchId1), int64(launchId2))
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
	launchIdParam := r.URL.Query().Get("launch_id")
	launchId, parseErr := strconv.Atoi(launchIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	//TODO Redirect on this branch page

	launchInfo := DAO.GetLaunchInfo(int64(launchId))
	if launchInfo == nil {
		http.Error(w, "Unable to find launch "+launchIdParam, http.StatusInternalServerError)
		return
	}

	err := DAO.DeleteLaunch(int64(launchId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/branch?branch_name="+launchInfo.Branch, http.StatusMovedPermanently)
}
