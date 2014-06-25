package handler

import (
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
	"strings"
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
      Body text,
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

// DisgoHandler is a wrapper for a HandlerFunc that adds the app struct.
type disgoHandler func(http.ResponseWriter, *http.Request, *App)

// A middleware can write something to the ResponseWriter but has to return true
// to stop the middleware chain.
type middleware func(http.ResponseWriter, *http.Request, *App) bool
type appHandler struct {
	handler    disgoHandler
	app        *App
	middleware []middleware
}

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
		"linebreak": func(args ...string) template.HTML {
			return template.HTML(strings.Replace(template.HTMLEscapeString(args[0]), "\n", "<br>", -1))
		},
		"times": func(args ...int) []struct{} {
			return make([]struct{}, args[0])
		},
		"add": func(args ...int) int {
			return args[0] + args[1]
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
	return &appHandler{handler, app, make([]middleware, 0)}
}

// Implement the Handler
func (h *appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, handler := range h.middleware {
		halt := handler(w, r, h.app)
		if halt {
			return
		}
	}

	h.handler(w, r, h.app)
}

// Add a middleware to a handler.
func (h *appHandler) addMiddleware(mw middleware) *appHandler {
	h.middleware = append(h.middleware, mw)
	return h
}

// Require a user to be logged in. This middleware will redirect to the login page
// of the userId is not set in the session and do nothing if it is set.
func requireLogin(rw http.ResponseWriter, req *http.Request, app *App) bool {
	ses, _ := app.SessionStore.Get(req, SessionName)
	var err error
	var id int64
	if val := ses.Values["userId"]; val != nil {
		id = val.(int64)
	}

	if err == nil {
		_, err = models.UserById(app.Db, id)
	}

	if err != nil {
		http.Redirect(rw, req, app.Config.General.Prefix+"/login", http.StatusSeeOther)
		return true
	}
	return false
}

// rateLimit checks if a client is still allowed to post comments.
func rateLimit(rw http.ResponseWriter, req *http.Request, app *App) bool {
	if app.Config.Rate_Limit.Enable {
		duration := time.Now().Unix() - app.Config.Rate_Limit.Seconds
		ip, err := relevantIpBytes(req.RemoteAddr)
		errors := map[string]string{"overall": "Rate limit reached."}

		if err != nil {
			renderErrors(rw, errors, 429)
			return true
		}

		var count int64
		row := app.Db.QueryRow("select count(*) from comments where ClientIp=? and Created>?", ip, duration)
		err = row.Scan(&count)

		if err != nil || count >= app.Config.Rate_Limit.Max_Comments {
			renderErrors(rw, errors, 429)
			return true
		}
	}
	return false
}

// Middleware to send CORS headers.
func cors(rw http.ResponseWriter, req *http.Request, app *App) bool {
	origin := req.Header.Get("Origin")
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range app.Config.General.Origin {
		if allowedOrigin == origin {
			rw.Header().Set("Access-Control-Allow-Origin", origin)
			rw.Header().Set("Access-Control-Allow-Credentials", "true")
			if req.Method == "OPTIONS" {
				return true
			}
		}
	}

	return false
}

// Render a single comment.
func renderComment(w http.ResponseWriter, tmpl string, comment models.Comment, app *App) {
	render(w, tmpl, map[string]interface{}{"comment": comment}, app)
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

// PaginatedComments loads the paginated comments on a page. Default page is 0.
func paginatedComments(db *sql.DB, page int) *models.PaginatedComments {
	comments, pages := models.AllCommentsPaginated(db, page)
	return &models.PaginatedComments{pages, page, 10, comments}
}

// Render a template.
func render(w http.ResponseWriter, tmpl string, ctx map[string]interface{}, app *App) {
	err := app.Templates[tmpl].Execute(w, ctx)
	if err != nil {
		renderErrors(w, map[string]string{"overall": "Internal server error."}, http.StatusInternalServerError)
	}
}

// Load and build all templates and store them in the app struct.
// Templates without the layout will be store in templates['partial'+templateName].
func (app *App) buildTemplates() map[string]*template.Template {
	dir := app.Config.General.Templates
	templates := make(map[string]*template.Template, 0)
	must := template.Must

	if dir == "" {
		dir = "templates/"
	}

	buf, err := ioutil.ReadFile(dir + "layout.tmpl")
	if err != nil {
		panic(err)
	}
	layout := must(template.New("main").Funcs(app.viewhelpers()).Parse(string(buf)))

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
			if name == "layout" {
				return nil
			}
			newLayout := must(layout.Clone())
			must(newLayout.New("body").Parse(string(buf)))
			templates[name] = newLayout
			// store the template without layout as a partial
			templates["partial/"+name] = must(template.New(name).Funcs(app.viewhelpers()).Parse(string(buf)))
		}
		return nil
	})

	return templates
}
