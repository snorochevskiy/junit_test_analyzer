package main

import (
	"net/http"
)

func serveListUsers(w http.ResponseWriter, r *http.Request) {

	users := DAO.GetAllUsers()

	err := RenderInCommonTemplate(w, users, "list_users.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
