package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pascalj/disgo/models"
	"github.com/russross/blackfriday"
	"github.com/ungerik/go-gravatar"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	// Create table instuctions.
	// BUG(pascalj): Use migrations instead.
	sqlCreate = []string{`
    CREATE TABLE IF NOT EXISTS comments (
      Id bigint(20) NOT NULL AUTO_INCREMENT,
      Created bigint(20) DEFAULT NULL,
      Email varchar(255) DEFAULT NULL,
      Name varchar(255) DEFAULT NULL,
      Body varchar(255) DEFAULT NULL,
      Url varchar(255) DEFAULT NULL,
      ClientIp varchar(255) DEFAULT NULL,
      Approved tinyint(1) DEFAULT NULL,
      PRIMARY KEY (Id)
    ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8`,

		`CREATE TABLE IF NOT EXISTS users (
      Id bigint(20) NOT NULL AUTO_INCREMENT,
      Created bigint(20) DEFAULT NULL,
      Email varchar(255) DEFAULT NULL,
      Password varchar(255) DEFAULT NULL,
      PRIMARY KEY (Id)
    ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`}
	SessionName = "disgo"
)

// Viewhelpers for rendering.
func (app *App) viewhelpers() template.FuncMap {
	return template.FuncMap{
		"formatTime": func(args ...interface{}) string {
			t1 := time.Unix(args[0].(int64), 0)
			return t1.Format(time.Stamp)
		},
		"gravatar": func(args ...interface{}) string {
			return gravatar.Url(args[0].(string))
		},
		"awaitingApproval": func(args ...models.Comment) bool {
			return !args[0].Approved && app.Config.General.Approval

		},
		"usesMarkdown": func() bool {
			return app.Config.General.Markdown
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
		"content": func() string {
			return "No template selected."
		},
		"base": func() string {
			if app.Config.General.Prefix != "" {
				return app.Config.General.Prefix
			} else {
				return "/"
			}
		},
	}
}

// rateLimit checks if a client is still allowed to post comments.
// func rateLimit(ren render.Render,
// 	req *http.Request,
// 	s sessions.Session,
// 	comment models.Comment,
// 	cfg models.Config,
// 	dbmap *gorp.DbMap) {
// 	if cfg.Rate_Limit.Enable {
// 		duration := time.Now().Unix() - cfg.Rate_Limit.Seconds
// 		ip, err := relevantIpBytes(req.RemoteAddr)
// 		errors := map[string]string{"overall": "Rate limit reached."}
// 		if err != nil {
// 			ren.JSON(429, errors)
// 			return
// 		}
// 		count, err := dbmap.SelectInt("select count(*) from comments where ClientIp=$1 and Created>$2", ip, duration)

// 		if err != nil || count >= cfg.Rate_Limit.Max_Comments {
// 			ren.JSON(429, errors)
// 			return
// 		}
// 	}
// }

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

func checkErr(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
	}
}

// Handler wrapper that injects the App into the handler functions.
func (app *App) handle(handler disgoHandler) *appHandler {
	return &appHandler{handler, app}
}

type disgoHandler func(http.ResponseWriter, *http.Request, *App)
type appHandler struct {
	handler disgoHandler
	app     *App
}

func (h *appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, origin := range h.app.Config.General.Origin {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	h.handler(w, r, h.app)
}

// Render a single comment.
func renderComment(w http.ResponseWriter, tmpl string, comment models.Comment, app *App) {
	render(w, tmpl, map[string]interface{}{"comment": comment}, app)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
}

// Render errors. Always writes JSON encoded
func renderErrors(w http.ResponseWriter, errors map[string]string, code int) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(&errors); err != nil {
		http.Error(w, fmt.Sprintf("Cannot encode response data"), 500)
	}
}

func paginatedComments(db *sql.DB, page int) *models.PaginatedComments {
	comments, pages := models.AllCommentsPaginated(db, page)
	return &models.PaginatedComments{pages, page, 10, comments}
}

func render(w http.ResponseWriter, tmpl string, ctx map[string]interface{}, app *App) {
	funcs := template.FuncMap{
		"content": func() template.HTML {
			buf := new(bytes.Buffer)
			app.Templates.ExecuteTemplate(buf, tmpl, ctx)
			return template.HTML(buf.String())
		},
	}
	app.Templates.Funcs(funcs)
	app.Templates.ExecuteTemplate(w, "layout", ctx)
	// if err := app.Templates[tmpl].Execute(os.Stdout, ctx); err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
}

func (app *App) buildTemplates() *template.Template {
	dir := app.Config.General.Templates
	if dir == "" {
		dir = "templates/"
	}
	t := template.New(dir)
	template.Must(t.Parse("Disgo"))

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		r, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := filepath.Ext(r)
		if ext == ".tmpl" {

			buf, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}

			name := (r[0 : len(r)-len(ext)])
			tmpl := t.New(filepath.ToSlash(name))
			tmpl.Funcs(app.viewhelpers())

			template.Must(tmpl.Funcs(app.viewhelpers()).Parse(string(buf)))
		}

		return nil
	})

	return t
}
