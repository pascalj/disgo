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

// App stores the context.
type App struct {
	Router       *mux.Router
	Db           *sql.DB
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

// Load the config.
func (app *App) LoadConfig(path string) error {
	cfg, err := models.LoadConfig(path)
	if err != nil {
		return err
	}
	app.Config = cfg
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
	db, err := sql.Open(app.Config.Database.Driver, app.Config.Database.Access)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	app.Db = db
	for _, statement := range sqlCreate {
		_, err = db.Exec(statement)
		if err != nil {
			return err
		}
	}
	return nil
}
