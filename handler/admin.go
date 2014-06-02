package handler

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/pascalj/disgo/models"
	"net/http"
	"strconv"
	"strings"
)

// AdminIndex shows the overview of the admin interface with latest comments.
func AdminIndex(w http.ResponseWriter, req *http.Request, app *App) {
	qry := req.URL.Query()
	page, err := strconv.Atoi(qry.Get("page"))

	if err != nil {
		page = 0
	}
	comments := paginatedComments(app.Db, page)
	render(w, "admin/index", map[string]interface{}{"comments": comments}, app)
}

// UnapprovedComments will only list unapproved comments, else it behaves like AdminIndex.
func UnapprovedComments(w http.ResponseWriter, req *http.Request, app *App) {
	count, err := models.UnapprovedCommentsCount(app.Db)
	if err == nil && count > 0 {
		comments := models.UnapprovedComments(app.Db)
		ctx := map[string]interface{}{"comments": comments}
		render(w, "unapproved", ctx, app)
	} else {
		http.Redirect(w, req, app.Config.General.Prefix+"/admin", http.StatusTemporaryRedirect)
	}
}

// GetRegister shows the register form.
func GetRegister(w http.ResponseWriter, req *http.Request, app *App) {
	render(w, "admin/register", nil, app)
}

// PostUser will create a new user when no other users are in the database.
// If users are present, it will redirect to the login.
func PostUser(w http.ResponseWriter, req *http.Request, app *App) {
	if models.UserCount(app.Db) == 0 {
		email, password := req.FormValue("email"), req.FormValue("password")
		user := models.NewUser(email, password)
		err := user.Save(app.Db)
		if err != nil {
			http.Redirect(w, req, app.Config.General.Prefix+"/register", http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, req, app.Config.General.Prefix+"/login", http.StatusTemporaryRedirect)
		}
	}
}

// RegireLogin is a middleware that ensures that only an admin call the following
// handler(s).
// func RequireLogin(rw http.ResponseWriter, req *http.Request,
// 	s sessions.Session, dbmap *gorp.DbMap, c martini.Context, cfg models.Config) {
// 	obj, err := dbmap.Get(models.User{}, s.Get("userId"))

// 	if err != nil || obj == nil {
// 		http.Redirect(rw, req, cfg.General.Prefix+"/login", http.StatusFound)
// 		return
// 	}

// 	user := obj.(*models.User)
// 	c.Map(user)
// }

// GetLogin shows the login form for the backend.
func GetLogin(w http.ResponseWriter, req *http.Request, app *App) {
	if models.UserCount(app.Db) > 0 {
		render(w, "admin/login", nil, app)
	} else {
		http.Redirect(w, req, app.Config.General.Prefix+"/register", http.StatusTemporaryRedirect)
	}

}

// PostLogin takes the email and password parameter and logs the user in if they are correct.
func PostSession(w http.ResponseWriter, req *http.Request, app *App) {
	var user models.User

	email, password := req.FormValue("email"), req.FormValue("password")
	user, err := models.UserByEmail(app.Db, email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		http.Redirect(w, req, app.Config.General.Prefix+"/login", http.StatusSeeOther)
		return
	}

	session, _ := app.SessionStore.Get(req, SessionName)
	session.Values["userId"] = user.Id
	session.Save(req, w)

	http.Redirect(w, req, app.Config.General.Prefix+"/admin/", http.StatusTemporaryRedirect)
}

func GetIndex(w http.ResponseWriter, req *http.Request, app *App) {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	base := []string{scheme, "://", req.Host, app.Config.General.Prefix}
	render(w, "index", map[string]interface{}{"base": strings.Join(base, ""), "hideNav": true}, app)
}

// PostLogout logs the user out and redirects to the login page.
func PostLogout(w http.ResponseWriter, req *http.Request, app *App) {
	session, _ := app.SessionStore.Get(req, SessionName)
	session.Values["userId"] = nil
	session.Save(req, w)
	http.Redirect(w, req, app.Config.General.Prefix+"/login", http.StatusTemporaryRedirect)
}
