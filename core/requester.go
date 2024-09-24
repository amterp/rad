package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

type Requester struct {
	jsonPathsByMockedUrlRegex map[string]string
}

func NewRequester() *Requester {
	return &Requester{
		jsonPathsByMockedUrlRegex: make(map[string]string),
	}
}

func (r *Requester) AddMockedResponse(urlRegex string, jsonPath string) {
	r.jsonPathsByMockedUrlRegex[urlRegex] = jsonPath
}

func (r *Requester) RequestJson(url string) (interface{}, error) {
	mockJson, ok := r.resolveMockedJson(url)
	if ok {
		return mockJson, nil
	}

	urlToQuery, err := encodeUrl(url)
	if err != nil {
		return nil, err
	}

	RP.RadInfo(fmt.Sprintf("Querying url: %s\n", urlToQuery))

	resp, err := http.Get(urlToQuery)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading HTTP body (%v): %w", body, err)
	}

	isValidJson := json.Valid(body)
	if !isValidJson {
		return nil, fmt.Errorf("received invalid JSON in response (truncated max 50 chars): [%s]", body[:50])
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}
	return data, nil
}

// todo test this more, might need additional query param encoding
func encodeUrl(rawUrl string) (string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("error parsing URL %v: %w", rawUrl, err)
	}
	return parsedUrl.String(), nil
}

func (r *Requester) resolveMockedJson(url string) (interface{}, bool) {
	for urlRegex, jsonPath := range r.jsonPathsByMockedUrlRegex {
		re, err := regexp.Compile(urlRegex)
		if err != nil {
			RP.ErrorExit(fmt.Sprintf("Failed to compile mock response regex %q: %v\n", urlRegex, err))
		}

		if re.MatchString(url) {
			RP.RadInfo(fmt.Sprintf("Mocking response for url (matched %q): %s\n", urlRegex, url))
			data := r.loadMockedResponse(jsonPath)
			return data, true
		} else {
			RP.RadDebug(fmt.Sprintf("No match for url %q against regex %q", url, urlRegex))
		}
	}
	return nil, false
}

func (r *Requester) loadMockedResponse(path string) interface{} {
	file, err := os.Open(path)
	if err != nil {
		RP.ErrorExit(fmt.Sprintf("Error opening file %s: %v\n", path, err))
	}
	defer file.Close()

	var data interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		RP.ErrorExit(fmt.Sprintf("Error decoding JSON from file %s: %v\n", path, err))
	}

	return data
}
