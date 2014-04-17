package service

import (
	"bytes"
	"github.com/pascalj/disgo/models"
	"net/smtp"
	"strconv"
	"text/template"
)

type Notifier struct {
	Config models.Config
}

func (notifier *Notifier) CommentCreated(comment *models.Comment) {
	if notifier.Config.Email.Notify {
		notifier.sendMail(notifier.Config.Email.From, notifier.Config.Email.To, newCommentTemplate(comment))
	}
}

func newCommentTemplate(comment *models.Comment) []byte {
	template, _ := template.ParseFiles("templates/mail/newcomment.tmpl")
	var wr bytes.Buffer
	err := template.Execute(&wr, comment)
	if err != nil {
		panic("error sending mail")
	}
	return wr.Bytes()
}

func (notifier *Notifier) sendMail(from string, to []string, text []byte) error {
	auth := smtp.PlainAuth("",
		notifier.Config.Email.Username,
		notifier.Config.Email.Password,
		notifier.Config.Email.Host,
	)
	return smtp.SendMail(notifier.Config.Email.Host+":"+strconv.Itoa(notifier.Config.Email.Port),
		auth,
		notifier.Config.Email.From,
		to,
		text)
}

func MapNotifier(cfg models.Config) *Notifier {
	notifier := new(Notifier)
	notifier.Config = cfg
	return notifier
}
