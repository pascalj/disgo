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
	"github.com/russross/blackfriday"
	"github.com/ungerik/go-gravatar"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	m   *martini.Martini
	cfg Config
)

func init() {
	m = martini.New()
	cfg = LoadConfig()
	m.Map(cfg)
	m.Map(initDb(cfg))
	m.Use(sessions.Sessions("session", sessions.NewCookieStore([]byte("secret"))))
	m.Use(martini.Static("public"))
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     cfg.General.Origin,
		AllowCredentials: true,
	}))
	m.Use(method.Override())
	m.Use(render.Renderer(render.Options{
		Funcs: viewhelper(),
	}))
	m.Use(MapView)
}

func main() {
	r := martini.NewRouter()
	r.Get(`/`, getIndex)
	r.Get(`/comments/:id`, GetComment)
	r.Post(`/comments`, binding.Bind(Comment{}), rateLimit, CreateComment)
	r.Get(`/comments`, GetComments)
	r.Post(`/comments/approve/:id`, ApproveComment)
	r.Delete(`/comments/:id`, DestroyComment)
	r.Get(`/admin`, RequireLogin, AdminIndex)
	r.Get(`/admin/unapproved`, RequireLogin, UnapprovedComments)
	r.Get(`/login`, GetLogin)
	r.Post(`/login`, PostLogin)
	r.Post(`/logout`, PostLogout)
	r.Get(`/register`, GetRegister)
	r.Post(`/user`, PostUser)
	m.Action(r.Handle)
	m.Run()
}

func initDb(cfg Config) *gorp.DbMap {
	db, err := sql.Open(cfg.Database.Driver, cfg.Database.Access)
	checkErr(err, "sql.Open failed")
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
	dbmap.AddTableWithName(Comment{}, "comments").SetKeys(true, "Id")
	dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
	return dbmap
}

func rateLimit(ren render.Render, req *http.Request,
	s sessions.Session, comment Comment, cfg Config, dbmap *gorp.DbMap) {
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
			"awaitingApproval": func(args ...Comment) bool {
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
