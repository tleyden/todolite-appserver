package libtodolite

import (
	"encoding/json"
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

	maxChars := 50
	funcMap := template.FuncMap{
		"Truncate": func(s string) string {
			if len(s) >= maxChars {
				return fmt.Sprintf("%s ...", s[:maxChars])
			}
			return s
		},
	}

	t := template.New("Changes template")
	t.Funcs(funcMap)
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

func (c *Context) WebhookReceiver(rw web.ResponseWriter, req *web.Request) {
	fmt.Println("/webhook_receiver POST request received")

	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)

	todoItem := TodoItem{}
	err := decoder.Decode(&todoItem)
	if err != nil {
		log.Printf("Error decoding POST body into a TodoItem, err: %v", err)
		http.Error(rw, err.Error(), 500)
		return
	}

	log.Printf("TodoItem: %+v", todoItem)
	if todoItem.Type != Task {
		log.Printf("Ignoring %+v since it's not a task", todoItem)
		fmt.Fprintf(rw, "Ignoring item")
		return
	}

	listItem, err := c.findList(todoItem)
	if err != nil {
		log.Printf("Error looking up list for TodoItem: %v, err: %v", todoItem, err)
		http.Error(rw, err.Error(), 500)
		return
	}
	log.Printf("list: %+v", listItem)

	err = c.sendPushNotification(todoItem, listItem)
	if err != nil {
		log.Printf("Error sending notifications for list: %v, err: %v", listItem, err)
		http.Error(rw, err.Error(), 500)
		return
	}

	fmt.Fprintf(rw, "Finished successfully")

}

func (c *Context) findList(i TodoItem) (TodoList, error) {

	l := TodoList{}
	err := c.Database.Retrieve(i.ListId, &l)
	if err != nil {
		return l, err
	}
	return l, nil

}

func (c *Context) sendPushNotification(i TodoItem, l TodoList) error {

	// TODO: since the todo item doesn't have a field that records
	// who added it, there's no way to not send the notification
	// to the adder.

	for _, member := range l.Members {
		log.Printf("Send notification to %v", member)

	}

	return nil
}

func ConfigMiddleware(databaseURL string) func(*Context, web.ResponseWriter, *web.Request, web.NextMiddlewareFunc) {
	return func(c *Context, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
		c.DatabaseURL = databaseURL
		next(rw, req)
	}
}
