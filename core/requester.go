package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/samber/lo"
)

type Requester struct {
	jsonPathsByMockedUrlRegex map[string]string
}

type RequestDef struct {
	Url     string
	Body    string
	Headers map[string]string
}

type ResponseDef struct {
	Body       string
	StatusCode int
}

func (r *ResponseDef) ToRslMap(i *MainInterpreter, t Token) RslMapOld {
	rslMap := NewOldRslMap()
	out, _ := TryConvertJsonToNativeTypes(i, t, r.Body)
	rslMap.SetStr("body", out)
	rslMap.SetStr("status_code", int64(r.StatusCode))
	// todo we should add more e.g. reason, message, response headers
	return *rslMap
}

func NewRequester() *Requester {
	return &Requester{
		jsonPathsByMockedUrlRegex: make(map[string]string),
	}
}

func (r *Requester) AddMockedResponse(urlRegex string, jsonPath string) {
	r.jsonPathsByMockedUrlRegex[urlRegex] = jsonPath
}

func (r *Requester) Get(url string, headers map[string]string) (*ResponseDef, error) {
	req := newGetRequest(url, headers)
	return r.request(req, func(encodedUrl string) (*http.Response, error) {
		return http.Get(encodedUrl)
	})
}

func (r *Requester) PutOrPost(method string, url string, body string, headers map[string]string) (*ResponseDef, error) {
	req := newPutOrPostRequest(url, body, headers)
	return r.request(req, func(encodedUrl string) (*http.Response, error) {
		request, err := http.NewRequest(method, encodedUrl, strings.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("error creating %s request: %w", method, err)
		}

		for key, value := range headers {
			request.Header.Add(key, value)
		}

		return http.DefaultClient.Do(request)
	})
}

func (r *Requester) RequestJson(url string) (interface{}, error) {
	resp, err := r.Get(url, nil)
	if err != nil {
		return nil, err
	}

	body := resp.Body
	bodyBytes := []byte(body)
	isValidJson := json.Valid(bodyBytes)
	if !isValidJson {
		return nil, fmt.Errorf("received invalid JSON in response (truncated max 50 chars): [%s]", body[:50])
	}

	var data interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}
	return data, nil
}

// todo test this more, might need additional query param encoding
func encodeUrl(rawUrl string) (string, error) {
	rawUrl = strings.ReplaceAll(rawUrl, "%", "%25")
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("error parsing URL %v: %w", rawUrl, err)
	}
	parsedUrl.RawQuery = parsedUrl.Query().Encode()
	return parsedUrl.String(), nil
}

func newGetRequest(url string, headers map[string]string) RequestDef {
	return RequestDef{
		Url:     url,
		Headers: headers,
	}
}

func newPutOrPostRequest(encodedUrl string, body string, headers map[string]string) RequestDef {
	if !lo.Contains(lo.Keys(headers), "Content-Type") {
		headers["Content-Type"] = "application/json"
	}
	return RequestDef{
		Url:     encodedUrl,
		Body:    body,
		Headers: headers,
	}
}

func (r *Requester) request(def RequestDef, reqFunc func(encodedUrl string) (*http.Response, error)) (*ResponseDef, error) {
	url := def.Url
	mockJson, ok := r.resolveMockedResponse(def.Url)
	if ok {
		return &ResponseDef{
			Body:       mockJson,
			StatusCode: 200,
		}, nil
	}

	urlToQuery, err := encodeUrl(url)
	if err != nil {
		return nil, err
	}

	RP.RadInfo(fmt.Sprintf("Querying url: %s\n", urlToQuery))

	resp, err := reqFunc(urlToQuery)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading HTTP body (%v): %w", body, err)
	}

	return &ResponseDef{
		Body:       string(body),
		StatusCode: resp.StatusCode,
	}, nil
}

func (r *Requester) resolveMockedResponse(url string) (string, bool) {
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
			RP.RadDebugf(fmt.Sprintf("No match for url %q against regex %q", url, urlRegex))
		}
	}
	return "", false
}

func (r *Requester) loadMockedResponse(path string) string {
	file, err := os.Open(path)
	if err != nil {
		RP.ErrorExit(fmt.Sprintf("Error opening file %s: %v\n", path, err))
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		RP.ErrorExit(fmt.Sprintf("Error reading file %s: %v\n", path, err))
	}
	return string(data)
}
