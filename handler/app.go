package handler

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/pascalj/disgo/models"
	"net/http"
)

type App struct {
	Router *mux.Router
	Db     *sql.DB
	Config models.Config
}

const (
	sqlCreate = `
		CREATE TABLE IF NOT EXISTS comments (
		  Id bigint(20) NOT NULL AUTO_INCREMENT,
		  Created bigint(20) DEFAULT NULL,
		  Email varchar(255) DEFAULT NULL,
		  Name varchar(255) DEFAULT NULL,
		  Body varchar(255) DEFAULT NULL,
		  Url varchar(255) DEFAULT NULL,
		  ClientIp varchar(255) DEFAULT NULL,
		  Approved tinyint(1) DEFAULT NULL,
		  PRIMARY KEY (Id)
		) ENGINE=InnoDB AUTO_INCREMENT=368 DEFAULT CHARSET=utf8;

		CREATE TABLE IF NOT EXISTS users (
		  Id bigint(20) NOT NULL AUTO_INCREMENT,
		  Created bigint(20) DEFAULT NULL,
		  Email varchar(255) DEFAULT NULL,
		  Password varchar(255) DEFAULT NULL,
		  PRIMARY KEY (Id)
		) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;`
)

func NewApp() *App {
	router := mux.NewRouter()
	return &App{Router: router}
}

func (app *App) LoadConfig(path string) error {
	cfg, err := models.LoadConfig(path)
	if err != nil {
		return err
	}
	app.Config = cfg
	return nil
}

func (app *App) SetRoutes() {
	r := app.Router
	// r.Handle("/comments", app.handle(CreateComment)).Methods("POST")
	// r.HandleFunc("/comments", GetComments).Methods("GET")
	// r.HandleFunc("/comments/{id}", GetComment).Methods("GET")
	// r.HandleFunc("/comments/approve/:id", ApproveComment).Methods("POST")
	// r.HandleFunc("/comments/{id}", DestroyComment).Methods("DELETE")

	// r.HandleFunc("/admin", AdminIndex).Methods("GET")
	// r.HandleFunc("/admin/unapproved", UnapprovedComments).Methods("GET")
	// r.HandleFunc("/login", GetLogin).Methods("GET")
	// r.HandleFunc("/login", PostLogin).Methods("POST")
	// r.HandleFunc("/logout", PostLogout).Methods("POST")
	// r.HandleFunc("/register", GetRegister).Methods("GET")
	// r.HandleFunc("/user", PostUser).Methods("POST")
	r.Handle("/", app.handle(GetIndex)).Methods("GET")
}

func (app *App) ConnectDb() error {
	db, err := sql.Open(app.Config.Database.Driver, app.Config.Database.Access)
	if err != nil {
		return err
	}
	app.Db = db
	_, err = db.Exec(sqlCreate)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) handle(handler disgoHandler) *appHandler {
	return &appHandler{handler, app}
}

type disgoHandler func(http.ResponseWriter, *http.Request, *App)
type appHandler struct {
	handler disgoHandler
	app     *App
}

func (h *appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler(w, r, h.app)
}
