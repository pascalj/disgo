package service

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/pascalj/disgo/models"
	"io"
	"time"
)

type disqus struct {
	Threads []thread `xml:"thread"`
	Posts   []post   `xml:"post"`
}

type thread struct {
	Id       string `xml:"id"`
	ThreadId string `xml:"http://disqus.com/disqus-internals id,attr"`
	Link     string `xml:"link"`
	Title    string `xml:"title"`
}

type post struct {
	Message   string    `xml:"message"`
	CreatedAt string    `xml:"createdAt"`
	IsDeleted string    `xml:"isDeleted"`
	IsSpam    string    `xml:"isSpam"`
	Author    author    `xml:"author"`
	IpAddress string    `xml:"ipAddress"`
	Thread    threadRef `xml:"thread"`
}

type author struct {
	Email string `xml:"email"`
	Name  string `xml:"name"`
}

type threadRef struct {
	Id string `xml:"http://disqus.com/disqus-internals id,attr"`
}

func Import(dbmap *gorp.DbMap, xmlReader io.Reader) error {
	parsed := &disqus{}
	decoder := xml.NewDecoder(xmlReader)
	if err := decoder.Decode(parsed); err != nil {
		return err
	}
	comments := make([]models.Comment, 0)
	for _, post := range parsed.Posts {
		thread, err := parsed.findThread(post.Thread.Id)
		if err != nil {
			fmt.Errorf("Could not find thread reference", post.Thread.Id)
		}
		if post.IsDeleted == "true" {
			continue
		}
		createdAt, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			createdAt = time.Now()
		}
		comment := models.NewComment(post.Author.Email, post.Author.Name, "", post.Message, thread.Link, post.IpAddress, "")
		comment.Created = createdAt.Unix()
		comment.Approved = post.IsSpam == "false"
		comments = append(comments, comment)
	}

	total := len(parsed.Posts)
	fmt.Println("Read", total, "from the file.")

	for i, comment := range comments {
		if err := dbmap.Insert(&comment); err != nil {
			total--
			fmt.Errorf("Could not import comment", i)
		}
	}

	fmt.Println("Wrote", total, "comments to the database.")

	return nil
}

func (dis *disqus) findThread(id string) (thread, error) {
	for _, thread := range dis.Threads {
		if thread.Id == id {
			return thread, nil
		}
	}
	return thread{}, errors.New("Could not find thread")
}
