package handler

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/pascalj/disgo/models"
	"net/http"
)

func AdminIndex(ren render.Render, dbmap *gorp.DbMap) {
	var comments []models.Comment
	dbmap.Select(&comments, "select * from comments order by created desc limit 10")
	ren.HTML(200, "admin/index", comments, render.HTMLOptions{
		Layout: "admin/layout",
	})
}

func UnapprovedComments(ren render.Render, dbmap *gorp.DbMap) {
	count, err := dbmap.SelectInt("select count(*) from comments where approved<>1")
	if err == nil && count > 0 {
		var comments []models.Comment
		dbmap.Select(&comments, "select * from comments where approved<>1 order by created desc")
		ren.HTML(200, "admin/unapproved", comments, render.HTMLOptions{
			Layout: "admin/layout",
		})
	} else {
		ren.Redirect("/admin")
	}

}

func GetRegister(ren render.Render) {
	ren.HTML(200, "admin/register", nil, render.HTMLOptions{
		Layout: "admin/layout",
	})
}

func PostUser(ren render.Render, req *http.Request, dbmap *gorp.DbMap) {
	if models.UserCount(dbmap) == 0 {
		email, password := req.FormValue("email"), req.FormValue("password")
		user := models.NewUser(email, password)
		err := dbmap.Insert(&user)
		if err != nil {
			ren.Redirect("/register")
		} else {
			ren.Redirect("/login")
		}
	}
}

func RequireLogin(rw http.ResponseWriter, req *http.Request,
	s sessions.Session, dbmap *gorp.DbMap, c martini.Context) {
	obj, err := dbmap.Get(models.User{}, s.Get("userId"))

	if err != nil || obj == nil {
		http.Redirect(rw, req, "/login", http.StatusFound)
		return
	}

	user := obj.(*models.User)
	c.Map(user)
}

func GetLogin(ren render.Render, dbmap *gorp.DbMap) {
	if models.UserCount(dbmap) > 0 {
		ren.HTML(200, "admin/login", nil, render.HTMLOptions{
			Layout: "layout",
		})
	} else {
		ren.Redirect("/register")
	}

}

func PostLogin(ren render.Render, req *http.Request, session sessions.Session, dbmap *gorp.DbMap) {
	var user models.User

	email, password := req.FormValue("email"), req.FormValue("password")
	err := dbmap.SelectOne(&user, "select * from users where email = :email", map[string]interface{}{"email": email})
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		ren.Redirect("/login")
		return
	}

	session.Set("userId", user.Id)
	ren.Redirect("/admin/")
}

func PostLogout(ren render.Render, session sessions.Session) {
	session.Clear()
	ren.Redirect("/login")
}