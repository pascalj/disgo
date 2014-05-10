package main

import (
	"database/sql"
	"flag"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/method"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/pascalj/disgo/handler"
	"github.com/pascalj/disgo/models"
	"github.com/pascalj/disgo/service"
	"github.com/russross/blackfriday"
	"github.com/ungerik/go-gravatar"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	m       *martini.ClassicMartini
	cfg     models.Config
	cfgPath string
	help    bool
)

func init() {
	flag.StringVar(&cfgPath, "config", "disgo.gcfg", "path to the config file")
	flag.Parse()

	var err error
	cfg, err = models.LoadConfig(cfgPath)
	checkErr(err, "Unable to load config file:")

	setupMartini()
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
	default:
		panic("No valid SQL driver specified. Options are: mysql, postgres.")
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
		ip, err := relevantIpBytes(req.RemoteAddr)
		errors := map[string]string{"overall": "Rate limit reached."}
		if err != nil {
			ren.JSON(429, errors)
			return
		}
		count, err := dbmap.SelectInt("select count(*) from comments where ClientIp=$1 and Created>$2", ip, duration)

		if err != nil || count >= cfg.Rate_Limit.Max_Comments {
			ren.JSON(429, errors)
			return
		}
	}
}

// Get relevant bytes from the IP address. This is used to rate limit v6 addresses as
// the last 64 will get shuffled.
func relevantIpBytes(remoteAddr string) (string, error) {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return "", err
	}

	parsedIp := net.ParseIP(ip)

	if parsedIp.To4() != nil {
		return ip, nil
	} else {
		// we got a v6 address, just grab the first 8 bytes
		for i := 8; i < len(parsedIp); i++ {
			parsedIp[i] = 0
		}
		return parsedIp.String(), nil
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
			"times": func(args ...int) []struct{} {
				return make([]struct{}, args[0])
			},
			"add": func(args ...int) int {
				return args[0] + args[1]
			},
			"base": func() string {
				return cfg.General.Prefix
			},
		},
	}
}

func getIndex(ren render.Render, req *http.Request, cfg models.Config) {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	base := []string{scheme, "://", req.Host, cfg.General.Prefix}
	ren.HTML(200, "index", strings.Join(base, ""))
}

func setupMartini() {
	m = martini.Classic()
	m.Map(cfg)
	m.Map(initDb(cfg))
	m.Use(sessions.Sessions("session", sessions.NewCookieStore([]byte(cfg.General.Secret))))
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     cfg.General.Origin,
		AllowCredentials: true,
	}))
	m.Use(method.Override())
	templates := cfg.General.Templates
	if templates == "" {
		templates = "templates"
	}
	m.Use(render.Renderer(render.Options{
		Funcs:     viewhelper(),
		Directory: templates,
	}))
	m.Use(handler.MapView)
	m.Map(service.MapNotifier(cfg))
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
