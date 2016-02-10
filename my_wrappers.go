package main

import (
	"errors"
	"fmt"
	"net/http"
)

func createDeleteWrapper() func(http.ResponseWriter, *http.Request) {

	w := Wrapper{}

	w.SessionProvider = &SessionManager
	w.HadleFunc = handleDelete
	w.SuccessRenderer = new(AfterLaunchDeletedRedirectRenderer)
	return w.Wrap()
}

func handleDelete(c *Context) (interface{}, error) {
	launchId := (c.Params["launch_id"]).(int)
	launchInfo := DAO.GetLaunchInfo(int64(launchId))
	if launchInfo == nil {
		return nil, errors.New(fmt.Sprintf("Can't find launch with id %v", launchId))
	}

	err := DAO.DeleteLaunch(int64(launchId))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't delete launch &v", launchInfo))
	}
	return launchInfo, nil
}

type AfterLaunchDeletedRedirectRenderer struct {
}

func (r *AfterLaunchDeletedRedirectRenderer) Render(c *Context, data interface{}) {
	launchInfo := data.(TestLaunchEntity)
	http.Redirect(c.Resp, c.Req, "/branch?branch_name="+launchInfo.Branch, http.StatusMovedPermanently)
}

func createViewBranch() func(http.ResponseWriter, *http.Request) {
	w := Wrapper{}
	w.ExpectedParams = []Param{Param{Name: "branch_name", Mandatory: true, Type: PARAM_TYPE_STRING}}
	w.HadleFunc = handleViewBranch
	w.SuccessRenderer = &GoTemplateRenderer{TemplateName: "view_branch.html"}
	return w.Wrap()
}

func handleViewBranch(c *Context) (interface{}, error) {
	branchName := (c.Params["branch_name"]).(string)
	launches := DAO.GetAllLaunchesInBranch(branchName)
	return launches, nil
}
