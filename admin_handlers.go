package main

import (
	"jutra/router"
	"log"
	"net/http"
	"os"
	"strconv"
)

func serveListUsersEx(context *router.HttpContext) {

	users := DAO.GetAllUsers()

	RenderInCommonTemplateEx(context, users, "list_users.html")
}

func serveEditUserEx(context *router.HttpContext) {

	session := context.Session
	if !session.IsLoggedIn() {
		errDto := HttpErrDTO{Code: 403, Message: "No permissions"}
		RenderInCommonTemplateEx(context, errDto, "error.html")
		return
	}

	if context.Req.Method == "POST" && context.Req.FormValue("userId") != "" && context.Req.FormValue("login") != "" {

		user := extractUserFromFormData(context.Req)

		if user.UserId == 0 {
			http.Error(context.Resp, "User error", http.StatusBadRequest)
			return
		}

		if user.Password != context.Req.FormValue("confirmPassword") {
			http.Error(context.Resp, "password mistmatch", http.StatusBadRequest)
			return
		}

		DAO.UpdateUser(user)
	}

	userIdParam := context.Req.URL.Query().Get("user_id")
	userId, parseErr := strconv.Atoi(userIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user := DAO.GetUserById(int64(userId))

	RenderInCommonTemplateEx(context, user, "edit_user.html")
}

func serveAddUser(context *router.HttpContext) {
	session := context.Session
	if !session.IsLoggedIn() {
		errDto := HttpErrDTO{Code: 403, Message: "No permissions"}
		RenderInCommonTemplateEx(context, errDto, "error.html")
		return
	}

	if context.Req.Method == "POST" {
		user := extractUserFromFormData(context.Req)

		if user.Password != context.Req.FormValue("confirmPassword") {
			http.Error(context.Resp, "password mistmatch", http.StatusBadRequest)
			return
		}

		DAO.InsertUser(user)
	}

	RenderInCommonTemplateEx(context, nil, "add_user.html")
}

func serveManageDatabase(context *router.HttpContext) {

	rendingObject := DbManagmentRO{}

	action := context.Req.URL.Query().Get("action")
	switch action {
	case "vacuum":
		rendingObject.ActionErr = DB_UTIL.vacuum()

	case "clean":
		DAO.DeleteOrphans()
	}

	dbFileName := calculateFullDbFilePath()
	fileInfo, fInfoErr := os.Stat(dbFileName)
	if fInfoErr != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	rendingObject.DbInfo.DbFileName = dbFileName
	rendingObject.DbInfo.DbFileSize = fileInfo.Size()

	RenderInCommonTemplateEx(context, rendingObject, "database_managment.html")
}

func extractUserFromFormData(r *http.Request) *UserEntity {
	user := UserEntity{
		Login:     r.FormValue("login"),
		Password:  r.FormValue("password"),
		IsActive:  r.FormValue("isActive") == "on",
		FirstName: r.FormValue("firstName"),
		LastName:  r.FormValue("lastName"),
	}

	userIdStr := r.FormValue("userId")

	if userIdStr != "" {
		if userId, err := strconv.Atoi(userIdStr); err == nil {
			user.UserId = int64(userId)
		}
	}

	return &user
}
