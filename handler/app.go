package handler

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/pascalj/disgo/models"
	"github.com/pascalj/disgo/service"
	"html/template"
	"net/http"
)

type App struct {
	Router       *mux.Router
	Db           *sql.DB
	Config       models.Config
	SessionStore sessions.Store
	Templates    *template.Template
	Notifier     *service.Notifier
}

func NewApp(cfgPath string) (*App, error) {
	router := mux.NewRouter()
	app := &App{Router: router}
	err := app.setup(cfgPath)
	return app, err
}

// Setup the app. Loads the config, parses templates, connects to DB.
func (app *App) setup(cfgPath string) error {
	if err := app.LoadConfig(cfgPath); err != nil {
		return err
	}
	if err := app.ConnectDb(); err != nil {
		return err
	}
	if err := app.ParseTemplates(); err != nil {
		return err
	}
	app.SetRoutes()
	app.InitSession()
	app.Notifier = &service.Notifier{app.Config}
	return nil
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
	r.StrictSlash(true)
	r.Handle("/comments", app.handle(CreateComment)).Methods("POST")
	r.Handle("/comments", app.handle(GetComments)).Methods("GET", "HEAD")
	r.Handle("/comments/approve/{id}", app.handle(ApproveComment)).Methods("POST")
	// r.HandleFunc("/comments/{id}", DestroyComment).Methods("DELETE")

	r.Handle("/admin/", app.handle(AdminIndex)).Methods("GET", "HEAD")
	// r.HandleFunc("/admin/unapproved", UnapprovedComments).Methods("GET")
	// r.HandleFunc("/login", GetLogin).Methods("GET")
	// r.HandleFunc("/login", PostLogin).Methods("POST")
	// r.HandleFunc("/logout", PostLogout).Methods("POST")
	// r.HandleFunc("/register", GetRegister).Methods("GET")
	// r.HandleFunc("/user", PostUser).Methods("POST")
	r.Handle("/", app.handle(GetIndex)).Methods("GET", "HEAD")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("public/")))
}

func (app *App) ConnectDb() error {
	db, err := sql.Open(app.Config.Database.Driver, app.Config.Database.Access)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	app.Db = db
	// _, err = db.Exec(sqlCreate)
	// if err != nil {
	// 	return err
	// }
	return nil
}
