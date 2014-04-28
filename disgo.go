package main

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/method"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pascalj/disgo/handler"
	"github.com/pascalj/disgo/models"
	"github.com/pascalj/disgo/service"
	"github.com/russross/blackfriday"
	"github.com/ungerik/go-gravatar"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	m   *martini.ClassicMartini
	cfg models.Config
)

func init() {
	m = martini.Classic()
	cfg, err := models.LoadConfig()
	checkErr(err, "Unable to load config file: ")
	m.Map(cfg)
	m.Map(initDb(cfg))
	m.Use(sessions.Sessions("session", sessions.NewCookieStore([]byte(cfg.General.Secret))))
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     cfg.General.Origin,
		AllowCredentials: true,
	}))
	m.Use(method.Override())
	m.Use(render.Renderer(render.Options{
		Funcs: viewhelper(),
	}))
	m.Use(handler.MapView)
	m.Map(service.MapNotifier(cfg))
}

func main() {
	r := martini.NewRouter()

	r.Group(`/comments`, func(r martini.Router) {
		r.Get(`/:id`, handler.GetComment)
		r.Post(``, binding.Bind(models.Comment{}), rateLimit, handler.CreateComment)
		r.Get(``, handler.GetComments)
		r.Post(`/approve/:id`, handler.RequireLogin, handler.ApproveComment)
		r.Delete(`/:id`, handler.RequireLogin, handler.DestroyComment)
	})
	r.Group(`/admin`, func(r martini.Router) {
		r.Get(``, handler.RequireLogin, handler.AdminIndex)
		r.Get(`/unapproved`, handler.RequireLogin, handler.UnapprovedComments)
	})
	r.Get(`/login`, handler.GetLogin)
	r.Post(`/login`, handler.PostLogin)
	r.Post(`/logout`, handler.PostLogout)
	r.Get(`/register`, handler.GetRegister)
	r.Post(`/user`, handler.PostUser)
	r.Get(`/`, getIndex)
	m.Action(r.Handle)
	m.Run()
}

func initDb(cfg models.Config) *gorp.DbMap {
	db, err := sql.Open(cfg.Database.Driver, cfg.Database.Access)
	checkErr(err, "Could not open database: ")
	var dbmap *gorp.DbMap
	switch cfg.Database.Driver {
	case "mysql":
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	case "postgres":
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	case "sqlite3":
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	default:
		panic("No valid SQL driver specified. Options are: mysql, postgres, sqlite3.")
	}
	dbmap.AddTableWithName(models.Comment{}, "comments").SetKeys(true, "Id")
	dbmap.AddTableWithName(models.User{}, "users").SetKeys(true, "Id")
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Could not create database tables: ")
	return dbmap
}

func rateLimit(ren render.Render,
	req *http.Request,
	s sessions.Session,
	comment models.Comment,
	cfg models.Config,
	dbmap *gorp.DbMap) {
	if cfg.Rate_Limit.Enable {
		duration := time.Now().Unix() - cfg.Rate_Limit.Seconds
		count, err := dbmap.SelectInt("select count(*) from comments where ClientIp=$1 and Created>$2", strings.Split(req.RemoteAddr, ":")[0], duration)

		if err != nil || count >= cfg.Rate_Limit.Max_Comments {
			errors := map[string]string{"overall": "Rate limit reached."}
			ren.JSON(429, errors)
			return
		}
	}
}

func viewhelper() []template.FuncMap {
	return []template.FuncMap{
		{
			"formatTime": func(args ...interface{}) string {
				t1 := time.Unix(args[0].(int64), 0)
				return t1.Format(time.Stamp)
			},
			"gravatar": func(args ...interface{}) string {
				return gravatar.Url(args[0].(string))
			},
			"awaitingApproval": func(args ...models.Comment) bool {
				return !args[0].Approved && cfg.General.Approval
			},
			"usesMarkdown": func() bool {
				return cfg.General.Markdown
			},
			"markdown": func(args ...string) template.HTML {
				output := blackfriday.MarkdownCommon([]byte(args[0]))
				return template.HTML(output)
			},
		},
	}
}

func getIndex(ren render.Render, req *http.Request) {
	base := []string{"http://", req.Host, req.URL.Path}
	ren.HTML(200, "index", strings.Join(base, ""))
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
