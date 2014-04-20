package handler

import (
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/pascalj/disgo/models"
	"github.com/pascalj/disgo/service"
	"net/http"
	"strings"
	"time"
)

func GetComments(
	res http.ResponseWriter,
	ren render.Render,
	view models.View,
	params martini.Params,
	dbmap *gorp.DbMap,
	session sessions.Session,
	req *http.Request,
	cfg models.Config) {
	var comments []models.Comment
	qry := req.URL.Query()
	if cfg.General.Approval {
		dbmap.Select(&comments, "select * from comments where (approved = 1 OR email = :email) and url = :url",
			map[string]interface{}{"email": session.Get("email"), "url": qry["url"][0]})
	} else {
		dbmap.Select(&comments, "select * from comments where url=:url", map[string]interface{}{"url": qry["url"][0]})
	}
	ctx := map[string]interface{}{
		"email": session.Get("email"),
		"name":  session.Get("name"),
	}
	if comments != nil {
		view.RenderComments(comments, ctx, ren)
	} else {
		view.RenderComments([]models.Comment{}, ctx, ren)
	}
}

func GetComment(ren render.Render, view models.View, params martini.Params, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(models.Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*models.Comment)
		view.RenderComment(*comment, nil, ren)
	}
}

func UpdateComment(ren render.Render, params martini.Params, comment models.Comment, req *http.Request, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(models.Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*models.Comment)
		comment.Email = req.FormValue("email")
		comment.Body = req.FormValue("body")
		comment.Url = req.FormValue("url")
		ren.JSON(200, comment)
	}
}

func ApproveComment(ren render.Render, params martini.Params, req *http.Request, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(models.Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.Error(404)
	} else {
		comment := obj.(*models.Comment)
		comment.Approved = true
		dbmap.Update(comment)
		ren.Redirect("/admin")
	}
}

func CreateComment(ren render.Render,
	view models.View,
	comment models.Comment,
	req *http.Request,
	dbmap *gorp.DbMap,
	session sessions.Session,
	notifier *service.Notifier) {
	comment.Created = time.Now().Unix()
	comment.ClientIp = strings.Split(req.RemoteAddr, ":")[0]
	err := dbmap.Insert(&comment)
	if err != nil {
		ren.JSON(400, err.Error())
	} else {
		session.Set("email", comment.Email)
		session.Set("name", comment.Name)
		go notifier.CommentCreated(&comment)
		view.RenderComment(comment, nil, ren)
	}
}

func DestroyComment(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(models.Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*models.Comment)
		count, err := dbmap.Delete(comment)
		if count != 1 || err != nil {
			ren.JSON(500, err.Error())
		} else {
			ren.Redirect("/admin")
		}
	}
}

func MapView(c martini.Context, w http.ResponseWriter, r *http.Request) {
	accept := r.Header["Accept"]
	if accept[0] != "" {
		accept = strings.Split(accept[0], ",")
	}
	switch accept[0] {
	case "text/html":
		c.MapTo(models.HtmlView{}, (*models.View)(nil))
		w.Header().Set("Content-Type", "text/html")
	default:
		c.MapTo(models.JsonView{}, (*models.View)(nil))
		w.Header().Set("Content-Type", "application/json")
	}
}