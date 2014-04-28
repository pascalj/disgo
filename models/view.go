package models

import (
	"github.com/martini-contrib/render"
)

// View is an interface that's used to render comments to the proper format.
type View interface {
	RenderComment(Comment, map[string]interface{}, render.Render)
	RenderComments([]Comment, map[string]interface{}, render.Render)
}

// HtmlView represents HTML output.
type HtmlView struct{}

// JsonView represents JSON output.
type JsonView struct{}

// RenderComment will render a comment to HTML.
func (view HtmlView) RenderComment(comment Comment, context map[string]interface{}, ren render.Render) {
	ren.HTML(200, "comment", &comment)
}

// RenderComment will render a comment to JSON.
func (view JsonView) RenderComment(comment Comment, context map[string]interface{}, ren render.Render) {
	ren.JSON(200, &comment)
}

// RenderComment will render a comments to HTML.
func (view HtmlView) RenderComments(comments []Comment, context map[string]interface{}, ren render.Render) {
	context["comments"] = comments
	ren.HTML(200, "comments", context)
}

// RenderComment will render a comments to JSON.
func (view JsonView) RenderComments(comments []Comment, context map[string]interface{}, ren render.Render) {
	ren.JSON(200, comments)
}
