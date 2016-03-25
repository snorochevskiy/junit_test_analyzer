package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

func serveListUsersEx(context *HttpContext) {

	users := DAO.GetAllUsers()

	err := RenderInCommonTemplateEx(context, users, "list_users.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveEditUserEx(context *HttpContext) {

	session := context.Session
	if !session.IsLoggedIn() {
		errDto := HttpErrDTO{Code: 403, Message: "No permissions"}
		if renderErr := RenderInCommonTemplateEx(context, errDto, "error.html"); renderErr != nil {
			http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
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

		updateErr := DAO.UpdateUser(user)
		if updateErr != nil {
			http.Error(context.Resp, "Unable to update user. Reason: "+updateErr.Error(), http.StatusInternalServerError)
			return
		}
	}

	userIdParam := context.Req.URL.Query().Get("user_id")
	userId, parseErr := strconv.Atoi(userIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user := DAO.GetUserById(int64(userId))

	err := RenderInCommonTemplateEx(context, user, "edit_user.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveAddUser(context *HttpContext) {
	session := context.Session
	if !session.IsLoggedIn() {
		errDto := HttpErrDTO{Code: 403, Message: "No permissions"}
		if renderErr := RenderInCommonTemplateEx(context, errDto, "error.html"); renderErr != nil {
			http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if context.Req.Method == "POST" {
		user := extractUserFromFormData(context.Req)

		if user.Password != context.Req.FormValue("confirmPassword") {
			http.Error(context.Resp, "password mistmatch", http.StatusBadRequest)
			return
		}

		err := DAO.InsertUser(user)
		if err != nil {
			http.Error(context.Resp, "Unable to create user. Reason: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err := RenderInCommonTemplateEx(context, nil, "add_user.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveManageDatabase(context *HttpContext) {

	rendingObject := DbManagmentRO{}

	action := context.Req.URL.Query().Get("action")
	switch action {
	case "vacuum":
		rendingObject.ActionErr = DB_UTIL.vacuum()
	}

	dbFileName := calculateFullDbFilePath()
	fileInfo, fInfoErr := os.Stat(dbFileName)
	if fInfoErr != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	rendingObject.DbInfo.DbFileName = dbFileName
	rendingObject.DbInfo.DbFileSize = fileInfo.Size()

	err := RenderInCommonTemplateEx(context, rendingObject, "database_managment.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
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
