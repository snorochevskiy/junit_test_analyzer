package main

import (
	"fmt"
	"log"
	"net/http"
)

func StartServer(port string) {

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", maxAgeHandler(3600, http.StripPrefix("/static/", fs)))

	rh := RoutedHandler{}
	rh.AddRoute("/", serveShowBranches)
	rh.AddRoute("/all-branches", serveShowBranches)
	rh.AddRoute("/filter-branches", serveFilterBranches)
	rh.AddRoute("/branch", serveLaunchesInBranchEx)
	rh.AddRoute("/launch", serverLaunchEx)
	rh.AddRoute("/packages", serverLaunchPackagesEx)
	rh.AddRoute("/package", servePackageEx)
	rh.AddRoute("/test", serverTestCaseEx)
	rh.AddRoute("/dynamics", serverTestDymanicsEx)
	rh.AddRoute("/diff", serveDiffLaunchesEx)
	rh.AddRoute("/delete-launch", serveDeleteLaunchEx)
	rh.AddRoute("/delete-this-and-previous-launches", serveDeleteThisAndPreviousLaunch)
	rh.AddRoute("/delete-branch", serveDeleteBranch)

	rh.AddRoute("/admin/list-users", serveListUsersEx)
	rh.AddRoute("/admin/edit-user", serveEditUserEx)
	rh.AddRoute("/admin/add-user", serveAddUser)
	rh.AddRoute("/admin/db-managment", serveManageDatabase)

	rh.AddRoute("/api/branches/status", serveApiBranchesStatus)

	rh.AddRoute("/login", handleLoginEx)
	rh.AddRoute("/logout", handleLogoutEx)

	http.HandleFunc("/", rh.ServeHTTP)

	log.Println("Listening...")
	http.ListenAndServe(":"+port, nil)
}

func maxAgeHandler(seconds int, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", seconds))
		h.ServeHTTP(w, r)
	})
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
