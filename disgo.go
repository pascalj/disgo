package main

import (
	"database/sql"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/coopernurse/gorp"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/method"
	"github.com/martini-contrib/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ungerik/go-gravatar"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var m *martini.Martini

func main() {
	m = martini.New()
	m.Map(initDb())
	m.Use(martini.Static("public"))

	store := sessions.NewCookieStore([]byte("secret"))
	m.Use(sessions.Sessions("session", store))
	m.Use(martini.Logger())
	m.Use(MapView)
	m.Use(method.Override())
	m.Use(render.Renderer(render.Options{
		Funcs: []template.FuncMap{
			{
				"formatTime": func(args ...interface{}) string {
					t1 := time.Unix(args[0].(int64), 0)
					return t1.Format(time.Stamp)
				},
				"gravatar": func(args ...interface{}) string {
					return gravatar.Url(args[0].(string))
				},
			},
		},
	}))
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"http*"},
		AllowCredentials: true,
	}))
	r := martini.NewRouter()
	r.Get(`/`, GetIndex)
	r.Get(`/comments/:id`, GetComment)
	r.Post(`/comments`, binding.Bind(Comment{}), CreateComment)
	r.Get(`/comments`, GetComments)
	r.Post(`/comments/approve/:id`, ApproveComment)
	r.Delete(`/comments/:id`, DestroyComment)
	r.Get(`/admin`, RequireLogin, AdminIndex)
	r.Get(`/login`, GetLogin)
	r.Post(`/login`, PostLogin)
	r.Post(`/logout`, PostLogout)
	r.Get(`/register`, GetRegister)
	r.Post(`/user`, PostUser)
	m.Action(r.Handle)
	m.Run()
}

func initDb() *gorp.DbMap {
	db, err := sql.Open("sqlite3", "/tmp/test_db.bin")
	checkErr(err, "sql.Open failed")
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbmap.AddTableWithName(Comment{}, "comments").SetKeys(true, "Id")
	dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
	return dbmap
}

func GetIndex(ren render.Render, req *http.Request) {
	base := []string{"http://", req.Host, req.URL.Path}
	ren.HTML(200, "index", strings.Join(base, ""))
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
