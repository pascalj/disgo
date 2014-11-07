package handler

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pascalj/disgo/models"
	"github.com/pascalj/disgo/service"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

// App stores the context.
type App struct {
	Router       *mux.Router
	Db           *sqlx.DB
	Config       models.Config
	SessionStore sessions.Store
	Templates    map[string]*template.Template
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
	cfg, err := models.LoadConfig(cfgPath)
	if err != nil {
		return err
	}
	app.Config = cfg
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

// Initialize the session.
func (app *App) InitSession() {
	app.SessionStore = sessions.NewCookieStore([]byte(app.Config.General.Secret))
}

// Parse and store all templates.
func (app *App) ParseTemplates() error {
	app.Templates = app.buildTemplates()
	return nil
}

// Set all routes for the app.
func (app *App) SetRoutes() {
	r := app.Router
	r.StrictSlash(true)
	r.Handle("/comments", app.handle(CreateComment).addMiddleware(cors).addMiddleware(rateLimit)).Methods("POST")
	r.Handle("/comments", app.handle(GetComments).addMiddleware(cors)).Methods("GET", "HEAD")
	r.Handle("/comments/{id}/approve", app.handle(ApproveComment).addMiddleware(requireLogin)).Methods("POST")
	r.Handle("/comments/{id}/delete", app.handle(DestroyComment).addMiddleware(requireLogin)).Methods("POST")
	r.Handle("/comments/{id}", app.handle(DestroyComment).addMiddleware(requireLogin)).Methods("DELETE")

	r.Handle("/admin/", app.handle(AdminIndex).addMiddleware(requireLogin)).Methods("GET", "HEAD")
	r.Handle("/login", app.handle(GetLogin)).Methods("GET", "HEAD")
	r.Handle("/session", app.handle(PostSession)).Methods("POST")
	r.Handle("/logout", app.handle(PostLogout).addMiddleware(requireLogin)).Methods("POST")
	r.Handle("/register", app.handle(GetRegister)).Methods("GET", "HEAD")
	r.Handle("/user", app.handle(PostUser)).Methods("POST")

	r.Handle("/", app.handle(GetIndex)).Methods("GET", "HEAD")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("public/")))
}

// Connect the DB and store the db pool reference.
func (app *App) ConnectDb() error {
	var db *sqlx.DB
	var err error
	switch app.Config.Database.Driver {
	case "sqlite3":
		db, err = sqlx.Connect(app.Config.Database.Driver, app.Config.Database.Path)
		break
	case "postgresql":
	case "mysql":
		db, err = sqlx.Connect(app.Config.Database.Driver,
			fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				app.Config.Database.Host,
				app.Config.Database.Port,
				app.Config.Database.Username,
				app.Config.Database.Password,
				app.Config.Database.Password))
		break
	default:
		err = errors.New("Missing or wrong database configuration.")
	}

	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	app.Db = db
	sqlFile, err := ioutil.ReadFile("data/sql/schema." + app.Config.Database.Driver + ".sql")
	if err != nil {
		return err
	}
	statements := strings.Split(string(sqlFile), ";")
	for _, statement := range statements {
		if strings.Trim(statement, " \t\n\r") == "" {
			continue
		}
		if _, err := db.Exec(string(statement)); err != nil {
			return err
		}
	}
	return nil
}
