package routes

import "net/http"

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetUsersHandler"))
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetUserHandler"))
}

func PostUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PostUsersHandler"))
}

func DeleteUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("DeleteUsersHandler"))
}
