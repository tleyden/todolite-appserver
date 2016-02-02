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
	Database *couch.Database
}

func (c *Context) ConnectToSyncGw(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	// if the sync gateway db connection is nil, then connect
	dbUrl := "http://localhost:4985/todolite12rc2b-cc"

	if c.Database == nil {
		db, err := couch.Connect(dbUrl)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			return
		}
		c.Database = &db
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
	log.Printf("changes: %v", changes)

	// convert raw changes to a slice of todo lite changes
	// object type (user/list/task) | id | is_delete | name | container
	// user                         | 1  | false       foo    n/a
	// list                         | 2  | false       hey    self
	// task                         | 3  | false       lol    hey

	todoChanges := c.todoliteChanges(changes)

	// pass to a template to render
	log.Printf("todolite changes: %+v", todoChanges)

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
		log.Printf("todolite change: %+v", todoliteChange)
		todoliteChanges.Changes = append(todoliteChanges.Changes, *todoliteChange)
	}

	return todoliteChanges

}
