package main

import (
	"log"
	"net/http"
	"strconv"
)

func serveListUsers(w http.ResponseWriter, r *http.Request) {

	users := DAO.GetAllUsers()

	err := RenderInCommonTemplate(w, users, "list_users.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serveEditUser(w http.ResponseWriter, r *http.Request) {

	log.Println(r.Method)
	log.Println(r.FormValue("userId"))
	log.Println(r.FormValue("password"))
	log.Println(r.FormValue("isActive"))
	if r.Method == "POST" && r.FormValue("userId") != "" {
		log.Println(r.FormValue("login"))
	}

	userIdParam := r.URL.Query().Get("user_id")
	userId, parseErr := strconv.Atoi(userIdParam)
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user := DAO.GetUserById(int64(userId))

	err := RenderInCommonTemplate(w, user, "edit_user.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
