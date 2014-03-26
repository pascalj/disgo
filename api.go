package main

import (
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/coopernurse/gorp"
	"github.com/martini-contrib/binding"
	"github.com/nu7hatch/gouuid"
	"net/http"
	"strings"
	"time"
)

func GetComments(res http.ResponseWriter, ren render.Render, view View, params martini.Params, dbmap *gorp.DbMap) {
	var comments []Comment
	dbmap.Select(&comments, "select * from comments order by created asc")
	randomId, _ := uuid.NewV4()
	cookie := []string{"disgo_id=", randomId.String()}
	res.Header().Add("Set-Cookie", strings.Join(cookie, ""))
	if comments != nil {
		view.RenderComments(comments, ren)
	} else {
		view.RenderComments([]Comment{}, ren)
	}
}

func GetComment(ren render.Render, view View, params martini.Params, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*Comment)
		view.RenderComment(*comment, ren)
	}
}

func UpdateComment(ren render.Render, params martini.Params, req *http.Request, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*Comment)
		comment.Email = req.FormValue("email")
		comment.Body = req.FormValue("body")
		comment.Url = req.FormValue("url")
		ren.JSON(200, comment)
	}
}

func CreateComment(ren render.Render, view View, comment Comment, req *http.Request, dbmap *gorp.DbMap) {
	comment.Created = time.Now().Unix()
	err := dbmap.Insert(&comment)
	if err != nil {
		ren.JSON(400, err.Error())
	} else {
		view.RenderComment(comment, ren)
	}
}

func DestroyComment(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
	obj, err := dbmap.Get(Comment{}, params["id"])
	if err != nil || obj == nil {
		ren.JSON(404, nil)
	} else {
		comment := obj.(*Comment)
		count, err := dbmap.Delete(comment)
		if count != 1 || err != nil {
			ren.JSON(500, err.Error())
		} else {
			ren.Redirect("/admin")
		}
	}
}

func (comment Comment) Validate(errors *binding.Errors, req *http.Request) {
	if len(comment.Name) == 0 {
		errors.Fields["name"] = "Please enter a name."
	}
	if len(comment.Body) == 0 {
		errors.Fields["body"] = "You must enter a comment text."
	}
	if len(comment.Email) == 0 {
		errors.Fields["email"] = "Please enter an email address."
	}
}

func MapView(c martini.Context, w http.ResponseWriter, r *http.Request) {
	accept := r.Header["Accept"]
	fmt.Printf(accept[0])
	if accept[0] != "" {
		accept = strings.Split(accept[0], ",")
	}
	switch accept[0] {
	case "text/html":
		c.MapTo(HtmlView{}, (*View)(nil))
		w.Header().Set("Content-Type", "text/html")
	default:
		c.MapTo(JsonView{}, (*View)(nil))
		w.Header().Set("Content-Type", "application/json")
	}
}
