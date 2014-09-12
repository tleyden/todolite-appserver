package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/couchbaselabs/logg"
	"github.com/tleyden/todolite-appserver"
)

// This follows the changes feed of the TodoLite sync gateway database and:
// When a new image is uploaded, it runs it through OCR and saves the decoded text into the JSON

var (
	urlDescription        = "Sync gateway url, with db name and no trailing slash"
	sgUrl                 = kingpin.Arg("sgUrl", urlDescription).Required().String()
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
	if *sgUrl == "" {
		kingpin.UsageErrorf("SG URL is empty")
		return
	}
	if *openOcrUrl == "" {
		kingpin.UsageErrorf("OpenOcr URL is empty")
		return
	}

	logg.LogTo("CLI", "sgRrl: %v openOcrUrl: %v", *sgUrl, *openOcrUrl)
	todoliteApp := todolite.NewTodoLiteApp(*sgUrl, *openOcrUrl)
	err := todoliteApp.InitApp()
	if err != nil {
		logg.LogPanic("Error initializing todo lite app: %v", err)
	}
	go todoliteApp.FollowChangesFeed(*since)

	// start a reverse proxy
	target := &url.URL{Scheme: "http", Host: "localhost:4985", Path: "/"}

	// proxy := httputil.NewSingleHostReverseProxy(target)
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		logg.LogTo("CLI", "path: %s", req.URL.Path)

		if !isRequestAllowed(req) {
			logg.LogTo("CLI", "forbideen url, redirect to google")
			req.URL.Host = "google.com"
		}

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}
	proxy := &httputil.ReverseProxy{Director: director}

	err = http.ListenAndServe(":8081", proxy)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func createReverseProxyDirector() httputil.Director {

}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func isRequestAllowed(req *http.Request) bool {

	if !strings.HasPrefix(req.URL.Path, "/todolite/_user") {
		return false
	}

	if req.Method != "POST" {
		return false
	}

	return true

}
