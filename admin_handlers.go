package main

import (
	"log"
	"net/http"
	"strconv"
)

func serveListUsersEx(context *HttpContext) {

	users := DAO.GetAllUsers()

	ro := RenderObject{
		User: context.Session.GetUserRenderInfo(),
		Data: users,
	}
	err := RenderInCommonTemplate(context.Resp, ro, "list_users.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveEditUserEx(context *HttpContext) {

	if context.Req.Method == "POST" && context.Req.FormValue("userId") != "" && context.Req.FormValue("login") != "" {

		if context.Req.FormValue("password") != "" && context.Req.FormValue("password") != context.Req.FormValue("confirmPassword") {
			http.Error(context.Resp, "password mistmatch", http.StatusBadRequest)
			return
		}
		userId, err := strconv.Atoi(context.Req.FormValue("userId"))
		if err != nil {
			http.Error(context.Resp, "User error", http.StatusBadRequest)
			return
		}

		user := UserEntity{
			UserId:    int64(userId),
			Login:     context.Req.FormValue("login"),
			Password:  context.Req.FormValue("password"),
			IsActive:  context.Req.FormValue("isActive") == "on",
			FirstName: context.Req.FormValue("firstName"),
			LastName:  context.Req.FormValue("lastName"),
		}
		updateErr := DAO.UpdateUser(&user)
		if updateErr != nil {
			http.Error(context.Resp, "Unable to update user", http.StatusInternalServerError)
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

	ro := RenderObject{
		User: context.Session.GetUserRenderInfo(),
		Data: user,
	}
	err := RenderInCommonTemplate(context.Resp, ro, "edit_user.html")
	if err != nil {
		http.Error(context.Resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
