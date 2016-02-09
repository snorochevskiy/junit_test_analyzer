package main

import (
	"errors"
	"fmt"
	"net/http"
)

type AfterLaunchDeletedRedirectRenderer struct {
}

func (r *AfterLaunchDeletedRedirectRenderer) Render(c *Context, data interface{}) {
	launchInfo := data.(TestLaunchEntity)
	http.Redirect(c.Resp, c.Req, "/branch?branch_name="+launchInfo.Branch, http.StatusMovedPermanently)
}

func createDeleteWrapper() {

	w := Wrapper{}

	w.SessionProvider = &SessionManager
	w.HadleFunc = handleDelete
	w.SuccessRenderer = new(AfterLaunchDeletedRedirectRenderer)
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
