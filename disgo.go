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
	"github.com/russross/blackfriday"
	"github.com/ungerik/go-gravatar"
	"html/template"
	"log"
	"net"
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
				if cfg.General.Prefix != "" {
					return cfg.General.Prefix
				} else {
					return "/"
				}
			},
		},
	}
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
