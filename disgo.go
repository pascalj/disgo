package main

import (
	"database/sql"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/coopernurse/gorp"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ungerik/go-gravatar"
	"html/template"
	"log"
	"time"
)

var m *martini.Martini

func main() {
	m = martini.New()
	m.Map(initDb())
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins: []string{"http*"},
	}))
	m.Use(martini.Logger())
	m.Use(MapView)
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
	m.Use(martini.Static("public"))
	r := martini.NewRouter()
	r.Get(`/comments/:id`, GetComment)
	r.Post(`/comments`, binding.Bind(Comment{}), CreateComment)
	r.Get(`/comments`, GetComments)
	r.Get(`/comments/new`, Form)
	r.Delete(`/comments/:id`, DestroyComment)
	m.Action(r.Handle)
	m.Run()
}

func Form(ren render.Render) {
	ren.HTML(200, "form", nil)
}

func initDb() *gorp.DbMap {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("sqlite3", "/tmp/test_db.bin")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// add a table, setting the table name to 'posts' and
	// specifying that the Id property is an auto incrementing PK
	dbmap.AddTableWithName(Comment{}, "comments").SetKeys(true, "Id")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
