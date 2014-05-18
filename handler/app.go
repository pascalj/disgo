package handler

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pascalj/disgo/models"
	"github.com/russross/blackfriday"
	"github.com/ungerik/go-gravatar"
	"html/template"
	"net/http"
	"time"
)

type App struct {
	Router       *mux.Router
	Db           *sql.DB
	Config       models.Config
	SessionStore sessions.Store
	Templates    *template.Template
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

func (app *App) InitSession() {
	app.SessionStore = sessions.NewCookieStore([]byte(app.Config.General.Secret))
}

func (app *App) ParseTemplates() error {
	var err error
	templates := template.New("")
	templates.Funcs(app.viewhelpers())
	templates, err = templates.ParseGlob("templates" + "/*.tmpl")
	if err != nil {
		return err
	}
	templates, err = templates.ParseGlob("templates" + "/admin/*.tmpl")
	if err != nil {
		return err
	}
	app.Templates = templates
	return nil
}

func (app *App) SetRoutes() {
	r := app.Router
	// r.Handle("/comments", app.handle(CreateComment)).Methods("POST")
	r.Handle("/comments", app.handle(GetComments)).Methods("GET")
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

func (app *App) viewhelpers() template.FuncMap {
	return template.FuncMap{
		"formatTime": func(args ...interface{}) string {
			t1 := time.Unix(args[0].(int64), 0)
			return t1.Format(time.Stamp)
		},
		"gravatar": func(args ...interface{}) string {
			return gravatar.Url(args[0].(string))
		},
		"awaitingApproval": func(args ...models.Comment) bool {
			return !args[0].Approved && app.Config.General.Approval

		},
		"usesMarkdown": func() bool {
			return app.Config.General.Markdown
		},
		"markdown": func(args ...string) template.HTML {
			output := blackfriday.MarkdownCommon([]byte(args[0]))
			return template.HTML(output)
		},
		"times": func(args ...int) []struct{} {
			return make([]struct{}, args[0])
		},
		"add": func(args ...int) int {
			return args[0] + args[1]
		},
		"base": func() string {
			if app.Config.General.Prefix != "" {
				return app.Config.General.Prefix
			} else {
				return "/"
			}
		},
	}
}
