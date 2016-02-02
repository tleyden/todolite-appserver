package libtodolite

import (
	"fmt"
	"log"
	"net/http"

	"html/template"

	"github.com/gocraft/web"
	"github.com/tleyden/go-couch"
)

type Context struct {
	Database    *couch.Database
	DatabaseURL string
}

func (c *Context) ConnectToSyncGw(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {

	if c.Database == nil {
		log.Printf("Connecting to Sync Gateway: %v", c.DatabaseURL)
		db, err := couch.Connect(c.DatabaseURL)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			return
		}
		c.Database = &db
	} else {
		log.Printf("Already connected to Sync Gateway: %v", c.DatabaseURL)
	}

	next(rw, req)
}

func (c *Context) Root(rw web.ResponseWriter, req *web.Request) {
	fmt.Fprint(rw, "Welcome to the TodoLite webserver.  Read code to see avail endpoints")
}

func (c *Context) ChangesFeed(rw web.ResponseWriter, req *web.Request) {

	rw.Header().Set("Content-Type", "text/html")

	// get the changes feed
	changesOptions := map[string]interface{}{}
	changes, err := c.Database.GetChanges(changesOptions)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}

	todoChanges := c.todoliteChanges(changes)

	t := template.New("Changes template")
	t, err = t.Parse(changesTemplate)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}

	err = t.Execute(rw, todoChanges)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}

}

func (c *Context) todoliteChanges(changes couch.Changes) TodoliteChanges {

	todoliteChanges := TodoliteChanges{}
	todoliteChanges.LastSequence = changes.LastSequence

	for _, change := range changes.Results {
		todoliteChange := NewTodoLiteChange(*c.Database, change)
		todoliteChanges.Changes = append(todoliteChanges.Changes, *todoliteChange)
	}

	return todoliteChanges

}

func ConfigMiddleware(databaseURL string) func(*Context, web.ResponseWriter, *web.Request, web.NextMiddlewareFunc) {
	return func(c *Context, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
		c.DatabaseURL = databaseURL
		next(rw, req)
	}
}
