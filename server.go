package main

import (
	"log"
	"net/http"
)

func startServer(port string) {
	http.HandleFunc("/", serveRoot)

	log.Println("Listening...")
	http.ListenAndServe(":"+port, nil)
}

func serveRoot(w http.ResponseWriter, r *http.Request) {

	launches := DAO.GetAllTestLaunches()

	template, err := createCommonTemplate("test_launches.template")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := template.ExecuteTemplate(w, "layout", launches); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
