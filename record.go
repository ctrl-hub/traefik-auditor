package traefik_auditor

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

func newLogRecordResponse(response *responseWriter, ignoreHeaders []string) *logRecordResponse {

	t := time.Now()

	headers := make(map[string][]string)
	for name, values := range response.Header() {
		if !contains(ignoreHeaders, name) {
			for _, value := range values {
				headers[name] = append(headers[name], value)
			}
		}
	}

	return &logRecordResponse{
		Time:       t,
		Body:       string(response.body),
		Headers:    headers,
		StatusCode: response.status,
	}
}

func newLogRecordRequest(req *http.Request, ignoreHeaders []string) *logRecordRequest {

	t := time.Now()

	// read the body into a buffer and immediately assign it back to the body (io.ReadCloser)
	buf, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(buf))

	headers := make(map[string][]string)
	for name, values := range req.Header {
		if !contains(ignoreHeaders, name) {
			for _, value := range values {
				headers[name] = append(headers[name], value)
			}
		}
	}

	return &logRecordRequest{
		Time:          t,
		Headers:       headers,
		Body:          string(buf),
		Method:        req.Method,
		ContentLength: req.ContentLength,
		Path:          req.URL.Path,
		RawQuery:      req.URL.RawQuery,
		Query:         req.URL.Query(),
	}
}

type logRecordRequest struct {
	Time          time.Time           `json:"time"`
	Headers       map[string][]string `json:"headers"`
	Body          string              `json:"body"`
	Path          string              `json:"path"`
	Query         url.Values          `json:"query"`
	RawQuery      string              `json:"raw_query"`
	Method        string              `json:"method"`
	ContentLength int64               `json:"content_length"`
}

type logRecordResponse struct {
	Time       time.Time           `json:"time"`
	Body       string              `json:"body"`
	Headers    map[string][]string `json:"headers"`
	StatusCode int                 `json:"status"`
}

type logRecord struct {
	Request  *logRecordRequest  `json:"request"`
	Duration int64              `json:"duration"`
	Response *logRecordResponse `json:"response"`
}

func (record *logRecord) json() ([]byte, error) {
	return json.Marshal(record)
}
