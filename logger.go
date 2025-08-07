// Package traefik_auditor is a traefik plugin that intercepts the request / response
// lifecycle and forwards the encapsulated log to a remote server
package traefik_auditor

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	errRemoteFailed            = errors.New("remote request failed")
	errNoRemoteserversSupplied = errors.New("no remote servers supplied")
)

// Config the plugin configuration.
type Config struct {
	RemoteServer  string   `json:"remoteServer" yaml:"remoteServer" toml:"remoteServer"`
	IgnoreHeaders []string `json:"ignoreHeaders" yaml:"ignoreHeaders" toml:"ignoreHeaders"`
	Timeout       string   `json:"timeout" yaml:"timeout" toml:"timeout"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		RemoteServer:  "",
		IgnoreHeaders: []string{},
		Timeout:       "5s",
	}
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.RemoteServer == "" {
		return nil, errNoRemoteserversSupplied
	}
	for i, header := range config.IgnoreHeaders {
		config.IgnoreHeaders[i] = strings.TrimSpace(header)
	}

	return &Logger{
		remoteServer:  config.RemoteServer,
		ignoreHeaders: config.IgnoreHeaders,
		timeout:       config.Timeout,
		next:          next,
		name:          name,
	}, nil
}

type Logger struct {
	next          http.Handler
	remoteServer  string
	ignoreHeaders []string
	timeout       string
	name          string
}

func (logger *Logger) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	record := new(logRecord)
	record.Request = newLogRecordRequest(req, logger.ignoreHeaders)

	rw := &responseWriter{ResponseWriter: w}
	logger.next.ServeHTTP(rw, req)
	record.Response = newLogRecordResponse(rw, logger.ignoreHeaders)

	record.Duration = record.Response.Time.Sub(record.Request.Time).Milliseconds()

	// send the record to the remote server
	go func() {
		err := logger.sendRecord(record)
		if err != nil {
			log.Println(err)
		}
	}()
}

func (logger *Logger) sendRecord(record *logRecord) error {
	outboundBody, err := record.json()
	if err != nil {
		return err
	}
	outboundReq, err := http.NewRequest(http.MethodPost, logger.remoteServer, bytes.NewBuffer(outboundBody))
	if err != nil {
		return err
	}
	outboundReq.Header.Add("Content-Type", "application/json")

	to, err := time.ParseDuration(logger.timeout)

	client := &http.Client{
		Timeout: to,
	}
	outboundRes, err := client.Do(outboundReq)
	if err != nil {
		return err
	}
	outboundRes.Body.Close()
	if outboundRes.StatusCode != http.StatusOK {
		return errRemoteFailed
	}
	return nil
}

type responseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) Write(p []byte) (int, error) {
	rw.body = append(rw.body, p...)
	return rw.ResponseWriter.Write(p)
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.ToLower(a) == strings.ToLower(e) {
			return true
		}
	}
	return false
}
