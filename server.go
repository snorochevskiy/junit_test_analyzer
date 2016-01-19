package main

import (
	"log"
	"net/http"
	"strconv"
)

func startServer(port string) {
	http.HandleFunc("/", serveRoot)
	http.HandleFunc("/branch", serveLaunchesInBranch)
	http.HandleFunc("/launch", serverLaunch)
	http.HandleFunc("/test", serverTestCase)

	log.Println("Listening...")
	http.ListenAndServe(":"+port, nil)
}

func serveRoot(w http.ResponseWriter, r *http.Request) {

	branches := DAO.GetAllBranches()

	err := RenderInCommonTemplate(w, branches, "list_branches.template")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveLaunchesInBranch(w http.ResponseWriter, r *http.Request) {

	launches := DAO.GetAllLaunches()

	err := RenderInCommonTemplate(w, launches, "test_launches.template")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
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

	err := RenderInCommonTemplate(w, testCases, "view_launch.template")
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

	err := RenderInCommonTemplate(w, testCase, "view_test_case.template")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
