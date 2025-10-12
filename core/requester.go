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

	ts "github.com/tree-sitter/go-tree-sitter"
)

var (
	statusOk     = 200
	emptyHeaders = make(map[string][]string)
)

type Requester struct {
	jsonPathsByMockedUrlRegex map[string]string
}

type RequestDef struct {
	Method  string
	Url     string
	Headers map[string][]string
	Body    *string
}

func NewRequestDef(method, url string, headers map[string][]string, body *string) RequestDef {
	return RequestDef{
		Method:  method,
		Url:     url,
		Headers: headers,
		Body:    body,
	}
}

func (r RequestDef) BodyReader() io.Reader {
	if r.Body == nil {
		return strings.NewReader("") // todo same as nil?
	}
	return strings.NewReader(*r.Body)
}

// todo we should add more e.g. reason, message
type ResponseDef struct {
	Success         bool // true if 2xx response, else false
	StatusCode      *int
	Headers         *map[string][]string
	Body            *string
	Error           *string // signifies error making request
	DurationSeconds float64
}

func NewResponseDef(
	statusCode *int,
	headers *map[string][]string,
	body *string,
	error *string,
	durationSeconds float64,
) ResponseDef {
	success := false
	if statusCode != nil && *statusCode >= 200 && *statusCode < 300 {
		success = true
	}

	return ResponseDef{
		Success:         success,
		StatusCode:      statusCode,
		Headers:         headers,
		Body:            body,
		Error:           error,
		DurationSeconds: durationSeconds,
	}
}

func (r ResponseDef) ToRadMap(i *Interpreter, callNode *ts.Node) *RadMap {
	radMap := NewRadMap()

	radMap.SetPrimitiveBool("success", r.Success)
	if r.StatusCode != nil {
		radMap.SetPrimitiveInt("status_code", *r.StatusCode)
	}
	if r.Headers != nil {
		// todo should this *always* be present, but potentially empty?
		headers := NewRadMap()
		for key, values := range *r.Headers {
			headers.Set(newRadValue(i, callNode, key), newRadValue(i, callNode, values))
		}
	}
	if r.Body != nil {
		out, _ := TryConvertJsonToNativeTypes(i, callNode, *r.Body)
		radMap.Set(newRadValue(i, callNode, "body"), out)
	}
	if r.Error != nil {
		radMap.SetPrimitiveStr("error", *r.Error)
	}
	radMap.SetPrimitiveFloat("duration_seconds", r.DurationSeconds)

	return radMap
}

func NewRequester() *Requester {
	return &Requester{
		jsonPathsByMockedUrlRegex: make(map[string]string),
	}
}

func (r *Requester) AddMockedResponse(urlRegex string, jsonPath string) {
	r.jsonPathsByMockedUrlRegex[urlRegex] = jsonPath
}

func (r *Requester) Request(def RequestDef) ResponseDef {

	req, err := http.NewRequest(def.Method, def.Url, def.BodyReader())
	for key, values := range def.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if err != nil {
		msg := fmt.Sprintf("Failed to create HTTP request: %v", err)
		return NewResponseDef(nil, nil, nil, &msg, 0)
	}

	return r.request(req)
}

func (r *Requester) RequestJson(url string) (interface{}, error) {
	reqDef := NewRequestDef("GET", url, emptyHeaders, nil)
	response := r.Request(reqDef)

	if !response.Success {
		if response.Error != nil {
			return nil, fmt.Errorf("request failed: %s", *response.Error)
		} else if response.StatusCode != nil {
			return nil, fmt.Errorf("request failed: non-successful status code %d", *response.StatusCode)
		} else {
			return nil, fmt.Errorf("request failed: unknown reason") // this probably signifies a bug in Rad
		}
	}

	body := *response.Body
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

func (r *Requester) request(req *http.Request) ResponseDef {
	mockJson, ok := r.resolveMockedResponse(req.URL.String())
	if ok {
		return NewResponseDef(&statusOk, &emptyHeaders, &mockJson, nil, 0)
	}

	urlToQuery, err := encodeUrl(req.URL.String())
	if err != nil {
		msg := fmt.Sprintf("Failed to encode url: %v", err)
		return NewResponseDef(nil, nil, &mockJson, &msg, 0)
	}

	RP.RadStderrf("Querying url: %s\n", urlToQuery)
	start := RClock.Now()
	resp, err := http.DefaultClient.Do(req)
	durationSeconds := RClock.Now().Sub(start).Seconds()

	if err != nil {
		msg := fmt.Sprintf("Failed to make HTTP request: %v", err)
		return NewResponseDef(nil, nil, nil, &msg, durationSeconds)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("Failed to read response body: %v", err)
		return NewResponseDef(nil, nil, nil, &msg, durationSeconds)
	}
	bodyStr := string(body)
	headers := map[string][]string(resp.Header)

	return NewResponseDef(&resp.StatusCode, &headers, &bodyStr, nil, durationSeconds)
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

func (r *Requester) resolveMockedResponse(url string) (string, bool) {
	for urlRegex, jsonPath := range r.jsonPathsByMockedUrlRegex {
		re, err := regexp.Compile(urlRegex)
		if err != nil {
			RP.ErrorExit(fmt.Sprintf("Failed to compile mock response regex %q: %v\n", urlRegex, err))
		}

		if re.MatchString(url) {
			RP.RadStderrf(fmt.Sprintf("Mocking response for url (matched %q): %s\n", urlRegex, url))
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
