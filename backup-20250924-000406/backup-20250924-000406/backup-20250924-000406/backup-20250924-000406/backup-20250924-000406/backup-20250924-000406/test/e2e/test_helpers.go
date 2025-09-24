package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"
)

// makeSimpleRequest is a helper function to make HTTP requests for E2E tests
func makeSimpleRequest(method, url string, body interface{}, server *httptest.Server) (*http.Response, []byte, error) {
	var reqBody bytes.Buffer
	if body != nil {
		json.NewEncoder(&reqBody).Encode(body)
	}

	req, err := http.NewRequest(method, server.URL+url, &reqBody)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody := make([]byte, resp.ContentLength)
	if resp.ContentLength > 0 {
		resp.Body.Read(respBody)
	}

	return resp, respBody, nil
}
