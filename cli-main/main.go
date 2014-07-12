package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/couchbaselabs/logg"
	"github.com/tleyden/todolite-appserver"
)

// This follows the changes feed of the TodoLite sync gateway database and:
// When a new image is uploaded, it runs it through OCR and saves the decoded text into the JSON

var (
	urlDescription   = "Sync gateway url, with db name and no trailing slash"
	url              = kingpin.Arg("url", urlDescription).String()
	sinceDescription = "Since parameter to changes feed"
	since            = kingpin.Arg("since", sinceDescription).String()
)

func init() {
	logg.LogKeys["CLI"] = true
	logg.LogKeys["TODOLITE"] = true
}

func main() {
	kingpin.Parse()
	logg.LogTo("CLI", "url: %v", *url)
	todoliteApp := todolite.NewTodoLiteApp(*url)
	err := todoliteApp.InitApp()
	if err != nil {
		logg.LogPanic("Error initializing todo lite app: %v", err)
	}
	go todoliteApp.FollowChangesFeed(*since)
	select {}

}
