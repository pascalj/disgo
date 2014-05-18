package main

import (
	"flag"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/pascalj/disgo/handler"
	"github.com/pascalj/disgo/models"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	m          *martini.ClassicMartini
	cfg        models.Config
	cfgPath    string
	importPath string
	help       bool
	app        *handler.App
)

func init() {
	flag.StringVar(&cfgPath, "config", "disgo.gcfg", "path to the config file")
	flag.StringVar(&importPath, "import", "", "Disqus XML file to import")
	flag.Parse()

	var err error
	cfg, err = models.LoadConfig(cfgPath)
	checkErr(err, "Unable to load config file:")
}

func main() {
	// if importPath != "" {
	// 	file, err := os.Open(importPath)
	// 	checkErr(err, "Could not open Disqus XML file:")
	// 	reader := bufio.NewReader(file)
	// 	return
	// }
	app = handler.NewApp()
	app.SetRoutes()
	app.LoadConfig(cfgPath)
	app.ConnectDb()
	app.InitSession()
	checkErr(app.ParseTemplates(), "")
	http.Handle("/", app.Router)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func rateLimit(ren render.Render,
	req *http.Request,
	s sessions.Session,
	comment models.Comment,
	cfg models.Config,
	dbmap *gorp.DbMap) {
	if cfg.Rate_Limit.Enable {
		duration := time.Now().Unix() - cfg.Rate_Limit.Seconds
		ip, err := handler.RelevantIpBytes(req.RemoteAddr)
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

func checkErr(err error, description string) {
	if err != nil {
		log.Fatalln(description, err)
	}
}
