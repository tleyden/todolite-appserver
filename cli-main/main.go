package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/couchbaselabs/logg"
	"github.com/tleyden/todolite-appserver"
)

// This follows the changes feed of the TodoLite sync gateway database and:
// When a new image is uploaded, it runs it through OCR and saves the decoded text into the JSON

var (
	urlDescription        = "Sync gateway url, with db name and no trailing slash"
	url                   = kingpin.Arg("url", urlDescription).Required().String()
	openOcrUrlDescription = "OpenOCR API root url, eg http://api.openocr.net"
	openOcrUrl            = kingpin.Arg("openOcrUrl", openOcrUrlDescription).Required().String()
	sinceDescription      = "Since parameter to changes feed"
	since                 = kingpin.Arg("since", sinceDescription).String()
)

func init() {
	logg.LogKeys["CLI"] = true
	logg.LogKeys["TODOLITE"] = true
}

func main() {
	kingpin.Parse()
	if *url == "" {
		kingpin.UsageErrorf("URL is empty")
		return
	}
	if *openOcrUrl == "" {
		kingpin.UsageErrorf("OpenOcr URL is empty")
		return
	}

	logg.LogTo("CLI", "url: %v openOcrUrl: %v", *url, *openOcrUrl)
	todoliteApp := todolite.NewTodoLiteApp(*url, *openOcrUrl)
	err := todoliteApp.InitApp()
	if err != nil {
		logg.LogPanic("Error initializing todo lite app: %v", err)
	}
	go todoliteApp.FollowChangesFeed(*since)
	select {}

}
