package todolite

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

		sinceStr := fmt.Sprintf("%v", changes.LastSequence)

		t.saveLastProcessedSeq(sinceStr)

		logg.LogTo("TODOLITE", "returning since: %v", changes.LastSequence)

		return changes.LastSequence

	}

	options := changes{}
	options["since"] = t.determineStartingSince(startingSince)

	options["feed"] = "longpoll"
	logg.LogTo("TODOLITE", "Following changes feed: %+v", options)
	t.Database.Changes(handleChange, options)

}

// the startingSince param will override any last processed sequence
// we have stored.  if empty, use the stored last processed sequence.
// if _that's_ empty too, then fast forward to end of _changes feed
func (t TodoLiteApp) determineStartingSince(startingSince string) interface{} {

	if startingSince != "" {
		// if we have been passed a starting since, use it
		logg.LogTo("TODOLITE", "Using startingSince param: %v", startingSince)
		return startingSince
	} else {
		// otherwise try to get the stored last processed sequence
		lastProcessedSeq, err := t.lastProcessedSeq()
		if err == nil {
			logg.LogTo("TODOLITE", "Using saved last seq: %v", lastProcessedSeq)
			return lastProcessedSeq
		} else {
			logg.LogTo("TODOLITE", "Error getting stored last seq: %v", err)

			// if that's empty, find the sequence of most recent change
			lastSequence, err := t.Database.LastSequence()
			if err != nil {
				logg.LogPanic("Error getting LastSequence: %v", err)
			}
			logg.LogTo("TODOLITE", "Using end of changes feed: %v", lastSequence)
			return lastSequence
		}

	}

}

func (t TodoLiteApp) lastProcessedSeq() (string, error) {

	infile, err := os.Open("lastprocessed.db")
	if err != nil {
		logg.LogTo("TODOLITE", "could not open lastprocessed.db file for reading")
		return "", err
	}
	defer infile.Close()
	reader := bufio.NewReader(infile)
	line, err := reader.ReadString('\n')
	if err != nil {
		logg.LogTo("TODOLITE", "could not read from lastprocessed.db file")
		return "", err
	}
	return strings.TrimSpace(line), nil

}

func (t TodoLiteApp) saveLastProcessedSeq(seq string) (err error) {
	outfile, err := os.Create("lastprocessed.db")
	if err != nil {
		logg.LogTo("TODOLITE", "could not open lastprocessed.db file")
		return nil
	}
	defer outfile.Close()

	writer := bufio.NewWriter(outfile)
	defer func() {
		if err == nil {
			err = writer.Flush()
		}
	}()
	seqWithNewline := fmt.Sprintf("%s\n", seq)
	if _, err = writer.WriteString(seqWithNewline); err != nil {
		return err
	}
	return nil
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
