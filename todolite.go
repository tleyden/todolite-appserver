package todolite

import (
	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
	"io"
)

type TodoLiteApp struct {
	DatabaseURL string
	Database    couch.Database
}

func NewTodoLiteApp(DatabaseURL string) *TodoLiteApp {
	return &TodoLiteApp{
		DatabaseURL: DatabaseURL,
	}
}

type Changes map[string]interface{}

func (t *TodoLiteApp) InitApp() error {
	db, err := couch.Connect(t.DatabaseURL)
	if err != nil {
		logg.LogPanic("Error connecting to db: %v", err)
		// logg.LogError(err)
		return err
	}
	t.Database = db
	return nil
}

func (t TodoLiteApp) FollowChangesFeed() {

	handleChange := func(reader io.Reader) int64 {
		logg.LogTo("TODOLITE", "handleChange() callback called")
		return 0
	}

	options := Changes{"since": 0}
	options["feed"] = "longpoll"
	t.Database.Changes(handleChange, options)

}
