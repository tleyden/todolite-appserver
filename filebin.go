package todolite

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/couchbaselabs/logg"
)

func init() {
	logg.LogKeys["CLI"] = true
	logg.LogKeys["TODOLITE"] = true
	logg.LogKeys["TEST"] = true
}

// given a url, download the url contents and upload it to FileBin.ca,
// and return the FileBin URL where it's stored.
func copyUrlToFileBin(sourceUrl string) (string, error) {

	fileBinUrl := "http://filebin.ca/upload.php"

	// read bytes from attachment url
	res, err := http.Get(sourceUrl)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	//responseBody, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//	return nil, err
	//}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "file.png")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, res.Body)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", fileBinUrl, body)
	if err != nil {
		return "", err
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	uploadResponse, err := client.Do(request)

	defer uploadResponse.Body.Close()

	uploadResponseBody, err := ioutil.ReadAll(uploadResponse.Body)
	if err != nil {
		return "", err
	}

	logg.LogTo("TODOLITE", "uploadResponseBody: %s", uploadResponseBody)

	// save to file

	// post file to filebin (see http://matt.aimonetti.net/posts/2013/07/01/golang-multipart-file-upload-example/)

	// read filebin from response

	// return it

	return sourceUrl, nil

}
