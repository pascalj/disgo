package handler

import (
	"github.com/gorilla/mux"
	"net/http"
)

type App struct {
	Router *mux.Router
}

func NewApp() *App {
	router := mux.NewRouter()
	return &App{router}
}

func (app *App) SetRoutes() {
	r := app.Router
	r.HandleFunc("/comments", CreateComment).Methods("POST")
	r.HandleFunc("/comments", GetComments).Methods("GET")
	r.HandleFunc("/comments/{id}", GetComment).Methods("GET")
	r.HandleFunc("/comments/approve/:id", ApproveComment).Methods("POST")
	r.HandleFunc("/comments/{id}", DestroyComment).Methods("DELETE")

	// r.HandleFunc("/admin", AdminIndex).Methods("GET")
	// r.HandleFunc("/admin/unapproved", UnapprovedComments).Methods("GET")
	// r.HandleFunc("/login", GetLogin).Methods("GET")
	// r.HandleFunc("/login", PostLogin).Methods("POST")
	// r.HandleFunc("/logout", PostLogout).Methods("POST")
	// r.HandleFunc("/register", GetRegister).Methods("GET")
	// r.HandleFunc("/user", PostUser).Methods("POST")
	// r.HandleFunc("/", getIndex).Methods("GET")
}
