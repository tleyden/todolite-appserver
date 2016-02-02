package libtodolite

import (
	"fmt"
	"net/http"

	"github.com/gocraft/web"
	"github.com/tleyden/go-couch"
)

type Context struct {
	Database *couch.Database
}

func (c *Context) ConnectToSyncGw(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	// if the sync gateway db connection is nil, then connect
	dbUrl := "http://localhost:4985/todolite"

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

	// connect to sync gateway admin port

	// get the changes feed

	// pass to a template to render

	fmt.Fprint(rw, "Changes")
}
