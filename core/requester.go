package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Requester interface {
	RequestJson(url string) (interface{}, error)
}

type RealRequester struct {
}

func (r RealRequester) RequestJson(url string) (interface{}, error) {
	urlToQuery, err := encodeUrl(url)
	if err != nil {
		return nil, err
	}

	RP.Print(fmt.Sprintf("Querying url: %s\n", urlToQuery))

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
