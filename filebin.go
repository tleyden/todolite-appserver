package todolite

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

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

	requestGenerator := func() (*http.Request, error) {

		fileBinUrl := "http://filebin.ca/upload.php"

		// read bytes from attachment url
		res, err := http.Get(sourceUrl)
		if err != nil {
			return nil, err
		}

		defer res.Body.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("key", "7dH67qRO07yxZc7k3BcgBLheeMIpXw3p") // api key
		part, err := writer.CreateFormFile("file", "file.png")
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(part, res.Body)
		if err != nil {
			return nil, err
		}

		err = writer.Close()
		if err != nil {
			return nil, err
		}

		request, err := http.NewRequest("POST", fileBinUrl, body)
		if err != nil {
			return nil, err
		}

		request.Header.Add("Content-Type", writer.FormDataContentType())
		return request, nil
	}

	uploadToFileBinExtractUrl := func(request *http.Request) (string, error) {

		client := &http.Client{}
		uploadResponse, err := client.Do(request)
		if err != nil {
			return "", err
		}

		defer uploadResponse.Body.Close()

		scanner := bufio.NewScanner(uploadResponse.Body)
		for scanner.Scan() {
			line := scanner.Text()
			prefix := "url:"
			if strings.HasPrefix(line, prefix) {
				suffix := line[len(prefix):]
				return suffix, nil
			}
		}

		return "", fmt.Errorf("Did not find url in filebin response")

	}

	request, err := requestGenerator()
	if err != nil {
		return "", err
	}

	uploadedFileUrl, err := uploadToFileBinExtractUrl(request)
	if err != nil {
		return "", err
	}

	logg.LogTo("TODOLITE", "uploadedFileUrl: %s", uploadedFileUrl)

	return uploadedFileUrl, nil

}
