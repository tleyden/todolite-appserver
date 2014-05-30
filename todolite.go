package todolite

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/couchbaselabs/logg"
	sgrepl "github.com/couchbaselabs/sg-replicate"
	"github.com/tleyden/go-couch"
	openocr "github.com/tleyden/open-ocr-client"
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

func (t TodoLiteApp) FollowChangesFeed(since int) {

	handleChange := func(reader io.Reader) int64 {
		logg.LogTo("TODOLITE", "handleChange() callback called")
		changes := decodeChanges(reader)
		logg.LogTo("TODOLITE", "changes: %v", changes)

		t.processChanges(changes)

		return int64(changes.LastSequence)
	}

	options := changes{"since": since}
	options["feed"] = "longpoll"
	t.Database.Changes(handleChange, options)

}

func (t TodoLiteApp) processChanges(changes sgrepl.Changes) {

	for _, change := range changes.Results {
		logg.LogTo("TODOLITE", "change: %v", change)

		if change.Deleted {
			logg.LogTo("TODOLITE", "change was deleted, skipping")
			continue
		}

		todoItem := TodoItem{}
		err := t.Database.Retrieve(change.Id, &todoItem)
		if err != nil {
			errMsg := fmt.Errorf("Didn't retrieve: %v - %v", change.Id, err)
			logg.LogError(errMsg)
			continue
		}
		logg.LogTo("TODOLITE", "todo item: %+v", todoItem)

		if todoItem.OcrDecoded != "" {
			logg.LogTo("TODOLITE", "%v already ocr decoded, skipping", change.Id)
			continue
		}

		attachmentUrl := todoItem.AttachmentUrl(t.Database.DBURL())
		if attachmentUrl == "" {
			logg.LogTo("TODOLITE", "todo item has no attachment, skipping")
			continue
		}
		logg.LogTo("TODOLITE", "OCR Decoding: %v", attachmentUrl)

		ocrDecoded, err := t.ocrDecode(attachmentUrl)
		if err != nil {
			errMsg := fmt.Errorf("OCR failed: %+v - %v", todoItem, err)
			logg.LogError(errMsg)
			ocrDecoded = "failed"
		}
		err = t.updateTodoItemWithOcr(todoItem, ocrDecoded)
		if err != nil {
			errMsg := fmt.Errorf("Update failed: %+v - %v", todoItem, err)
			logg.LogError(errMsg)
			continue
		}

	}

}

func (t TodoLiteApp) ocrDecode(attachmentUrl string) (ocrDecoded string, err error) {
	openOcrUrl := "http://api.openocr.net"
	openOcrClient := openocr.NewHttpClient(openOcrUrl)
	ocrDecoded, err = openOcrClient.DecodeImageUrl(attachmentUrl, openocr.ENGINE_TESSERACT)
	if err != nil {
		return "", err
	}
	return ocrDecoded, nil
}

func (t TodoLiteApp) updateTodoItemWithOcr(i TodoItem, ocrDecoded string) error {
	i.OcrDecoded = ocrDecoded
	revid, err := t.Database.Edit(i)
	logg.LogTo("TODOLITE", "new revid: %v", revid)
	return err

}

func decodeChanges(reader io.Reader) (decodedChanges sgrepl.Changes) {

	changes := sgrepl.Changes{}
	decoder := json.NewDecoder(reader)
	decoder.Decode(&changes)
	return changes

}
