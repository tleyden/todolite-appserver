package todolite

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/couchbaselabs/logg"
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

func (t TodoLiteApp) FollowChangesFeed(startingSince string) {

	handleChange := func(reader io.Reader) interface{} {
		logg.LogTo("TODOLITE", "handleChange() callback called")
		changes, err := decodeChanges(reader)
		if err != nil {
			logg.LogTo("TODOLITE", "error decoding changes: %v", err)
			return nil // stop changes feed
		}

		logg.LogTo("TODOLITE", "changes: %v", changes)

		t.processChanges(changes)

		since := changes.LastSequence
		logg.LogTo("TODOLITE", "returning since: %v", since)

		return since

	}

	options := changes{}
	if startingSince != "" {
		logg.LogTo("TODOLITE", "startingSince not empty: %v", startingSince)
		options["since"] = startingSince
	} else {
		// find the sequence of most recent change
		lastSequence, err := t.Database.LastSequence()
		if err != nil {
			logg.LogPanic("Error getting LastSequence: %v", err)
			return
		}
		options["since"] = lastSequence
	}

	options["feed"] = "longpoll"
	logg.LogTo("TODOLITE", "Following changes feed: %+v", options)
	t.Database.Changes(handleChange, options)

}

func (t TodoLiteApp) processChanges(changes couch.Changes) {

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

func decodeChanges(reader io.Reader) (couch.Changes, error) {

	changes := couch.Changes{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&changes)
	if err != nil {
		logg.LogTo("TODOLITE", "Err decoding changes: %v", err)
	}
	return changes, err

}
