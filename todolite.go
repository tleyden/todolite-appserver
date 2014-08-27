package todolite

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/couchbaselabs/logg"
	sgrepl "github.com/couchbaselabs/sg-replicate"
	"github.com/tleyden/go-couch"
	openocr "github.com/tleyden/open-ocr-client"
)

type TodoLiteApp struct {
	DatabaseURL string
	OpenOCRURL  string
	Database    couch.Database
}

func NewTodoLiteApp(DatabaseURL, openOCRURL string) *TodoLiteApp {
	return &TodoLiteApp{
		DatabaseURL: DatabaseURL,
		OpenOCRURL:  openOCRURL,
	}
}

func (t *TodoLiteApp) InitApp() error {
	db, err := couch.Connect(t.DatabaseURL)
	if err != nil {
		logg.LogPanic("Error connecting to db: %v", err)
		return err
	}
	t.Database = db
	return nil
}

func (t TodoLiteApp) FollowChangesFeed(since interface{}) {

	handleChange := func(reader io.Reader) interface{} {
		logg.LogTo("TODOLITE", "handleChange() callback called")
		changes, err := decodeChanges(reader)
		if err == nil {
			logg.LogTo("TODOLITE", "changes: %v", changes)

			t.processChanges(changes)

			since = changes.LastSequence

		} else {
			logg.LogTo("TODOLITE", "error decoding changes: %v", err)

		}

		logg.LogTo("TODOLITE", "returning since: %v", since)
		return since

	}

	options := changes{"since": since}
	options["feed"] = "longpoll"
	t.Database.Changes(handleChange, options)

}

// TODO: remove dependency on sgrepl.Changes and use go-couch.Changes instead
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

		if todoItem.OcrDecoded != "" && todoItem.OcrDecoded != "failed" {
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

func (t TodoLiteApp) ocrDecode(attachmentUrl string) (string, error) {

	openOcrClient := openocr.NewHttpClient(t.OpenOCRURL)

	res, err := http.Get(attachmentUrl)
	if err != nil {
		errMsg := fmt.Errorf("Unable to open reader for %s: %s", attachmentUrl, err)
		logg.LogError(errMsg)
		return "", errMsg
	}
	defer res.Body.Close()

	ocrRequest := openocr.OcrRequest{
		EngineType: openocr.ENGINE_TESSERACT,
	}

	ocrDecoded, err := openOcrClient.DecodeImageReader(res.Body, ocrRequest)

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

func decodeChanges(reader io.Reader) (sgrepl.Changes, error) {

	changes := sgrepl.Changes{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&changes)
	if err != nil {
		logg.LogTo("TODOLITE", "Err decoding changes: %v", err)
	}
	return changes, err

}
