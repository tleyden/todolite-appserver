package libtodolite

import (
	"fmt"
	"log"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
)

type changes map[string]interface{}

type TodoliteChanges struct {
	Changes      []TodoliteChange `json:"results"`
	LastSequence interface{}      `json:"last_seq"`
}

type TodoliteChange struct {
	Sequence    interface{}        `json:"seq"`
	Id          string             `json:"id"`
	ChangedRevs []couch.ChangedRev `json:"changes"`
	Deleted     bool               `json:"deleted"`
	Type        string
	Title       string
	Parent      string // The parent list, or N/A
}

func NewTodoLiteChange(database couch.Database, change couch.Change) *TodoliteChange {

	todoliteChange := TodoliteChange{}
	todoliteChange.Sequence = change.Sequence
	todoliteChange.Id = change.Id
	todoliteChange.ChangedRevs = change.ChangedRevs
	todoliteChange.Deleted = change.Deleted

	// load the doc from sync gateway to figure out its type
	if !change.Deleted {
		todoItem := TodoItem{}
		err := database.Retrieve(change.Id, &todoItem)
		if err != nil {
			errMsg := fmt.Errorf("Didn't retrieve: %v Err: %v", change.Id, err)
			logg.LogError(errMsg)
			return &todoliteChange
		}
		todoliteChange.Type = todoItem.Type
		switch todoItem.Type {
		case "task":
			todoliteChange.Title = todoItem.Title
			listItem := TodoItem{}
			err := database.Retrieve(todoItem.ListId, &listItem)
			if err != nil {
				errMsg := fmt.Errorf("Didn't retrieve list: %v Err: %v", todoItem.ListId, err)
				logg.LogError(errMsg)
				todoliteChange.Parent = todoItem.ListId
			}
			todoliteChange.Parent = listItem.Title

		case "list":
			todoliteChange.Title = todoItem.Title
			todoliteChange.Parent = "N/A"
		case "profile":
			todoliteChange.Title = todoItem.Id
			todoliteChange.Parent = "N/A"
		}

		log.Printf("todoItem: %+v", todoItem)
	}

	return &todoliteChange

}
