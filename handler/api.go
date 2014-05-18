package handler

import (
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/pascalj/disgo/models"
	"net/http"
	"strings"
	"time"
)

// GetComments will display all comments for a given URL-parameter. If configured, it only
// displays approved comments.
func GetComments(
	ren render.Render,
	view models.View,
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

// GetComment show one comment by id.
func GetComment(ren render.Render, view models.View, params martini.Params, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(models.Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*models.Comment)
		view.RenderComment(*comment, nil, ren)
	}
}

// ApproveComment allows admins to approve a comment by id.
func ApproveComment(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
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

// CreateComment validates and creates a new comment. It also saves the client's IP-adress
// to reduce spam.
func CreateComment(w http.ResponseWriter, req *http.Request, app *App) {
	comment := models.NewComment("email", "name", "title", "body", "url", "ip", "id")
	comment.Created = time.Now().Unix()
	comment.ClientIp = strings.Split(req.RemoteAddr, ":")[0]
	// err := dbmap.Insert(&comment)
	// if err != nil {
	// 	ren.JSON(400, err.Error())
	// } else {
	// 	session.Set("email", comment.Email)
	// 	session.Set("name", comment.Name)
	// 	go notifier.CommentCreated(&comment)
	// 	view.RenderComment(comment, nil, ren)
	// }
}

// DestroyComment deletes a comment from the database by id.
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

// MapView maps the View type for martini depending on the accept header. It is used
// to generate the appropriate output for html and json.
func MapView(ctx martini.Context, res http.ResponseWriter, req *http.Request) {
	accept := req.Header["Accept"]
	if accept[0] != "" {
		accept = strings.Split(accept[0], ",")
	}
	switch accept[0] {
	case "text/html":
		ctx.MapTo(models.HtmlView{}, (*models.View)(nil))
		res.Header().Set("Content-Type", "text/html")
	default:
		ctx.MapTo(models.JsonView{}, (*models.View)(nil))
		res.Header().Set("Content-Type", "application/json")
	}
}
