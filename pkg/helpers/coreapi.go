package helpers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

var coreAPIURL = ""

// InitCoreAPI sets the URL and make sure CoreAPI is running. If it's not the site can't start.
func InitCoreAPI(url string) bool {
	coreAPIURL = url

	req, err := http.NewRequest("GET", coreAPIURL+"/general/ping", nil) // TODO switch this to /ping once CoreAPI is updated
	if err != nil {
		LogToSentry(err)
		return false
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogToSentry(err)
		return false
	}
	if resp.StatusCode != http.StatusOK {
		LogToSentry(fmt.Errorf("CoreAPI Init: Wanted 200 response, got %d instead", resp.StatusCode))
		return false
	}

	return true
}

func CoreAPIFile(file []byte, formData map[string]string, endpoint string) (data []byte, success bool, err error) {
	formBody := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(formBody)
	formFile, err := bodyWriter.CreateFormFile("pkmn", "temp.pkm")
	if err != nil {
		return nil, false, err
	}
	_, err = formFile.Write(file)
	if err != nil {
		return nil, false, err
	}

	if formData != nil {
		for formKey, formVal := range formData {
			bodyWriter.WriteField(formKey, formVal)
		}
	}
	err = bodyWriter.Close()
	if err != nil {
		return nil, false, err
	}
	req, err := http.NewRequest("POST", coreAPIURL+endpoint, formBody)
	if err != nil {
		return nil, false, err
	}
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("got unexpected http response code %d. Check CoreAPI logs", resp.StatusCode)
	}

	success = true
	return data, success, err
}
