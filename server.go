package main

import (
	"fmt"
	"jutra/router"
	"log"
	"net/http"
)

type JutraPanicHandler struct {
}

type JutraHttpErrorRenderObject struct {
	Message string
}

func (ro JutraHttpErrorRenderObject) String() string {
	return ro.Message
}

func (h *JutraPanicHandler) HttpErrorForPanic(panicObject interface{}) (httpError int, errorMessage interface {
	String() string
}) {
	switch panicObject.(type) {
	case DaoPanicErr:
		log.Println(panicObject.(DaoPanicErr).Message)
		return 500, JutraHttpErrorRenderObject{Message: panicObject.(DaoPanicErr).Message}

	case ParsePanicErr:
		log.Println(panicObject.(ParsePanicErr).Message)
		return http.StatusBadRequest, JutraHttpErrorRenderObject{Message: panicObject.(ParsePanicErr).Message}

	case string:
		log.Println(panicObject.(string))
		return http.StatusBadRequest, JutraHttpErrorRenderObject{Message: panicObject.(string)}

	default:
		return 500, JutraHttpErrorRenderObject{Message: "Internal Server Error"}
	}
}

func StartServer(port string) {

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", maxAgeHandler(3600, http.StripPrefix("/static/", fs)))

	rh := router.RoutedHandler{}
	rh.PanicHandler = new(JutraPanicHandler)

	rh.AddRoute("/", serveMainPage)
	rh.AddRoute("/project/:projectId", serveProject)
	//rh.AddRoute("/all-branches", serveShowBranches)
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
	rh.AddRoute("/delete-branch/:branchId", serveDeleteBranch)

	rh.AddRoute("/admin/list-users", serveListUsersEx)
	rh.AddRoute("/admin/edit-user", serveEditUserEx)
	rh.AddRoute("/admin/add-user", serveAddUser)
	rh.AddRoute("/admin/db-managment", serveManageDatabase)

	rh.AddRoute("/api/v1/project/list", serveApiListProjects)
	rh.AddRoute("/api/v1/project/:projectId/status", serveApiBranchesStatus)

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
