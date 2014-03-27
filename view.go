package main

import (
	"github.com/codegangsta/martini-contrib/render"
)

type View interface {
	RenderComment(Comment, map[string]interface{}, render.Render)
	RenderComments([]Comment, map[string]interface{}, render.Render)
}

type HtmlView struct{}
type JsonView struct{}

func (view HtmlView) RenderComment(comment Comment, context map[string]interface{}, ren render.Render) {
	ren.HTML(200, "comment", &comment)
}

func (view JsonView) RenderComment(comment Comment, context map[string]interface{}, ren render.Render) {
	ren.JSON(200, &comment)
}

func (view HtmlView) RenderComments(comments []Comment, context map[string]interface{}, ren render.Render) {
	context["comments"] = comments
	ren.HTML(200, "comments", context)
}

func (view JsonView) RenderComments(comments []Comment, context map[string]interface{}, ren render.Render) {
	ren.JSON(200, comments)
}
