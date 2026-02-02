// Copyright (c) Spectro Cloud
// SPDX-License-Identifier: MPL-2.0

// Package vcr provides HTTP recording and replay functionality for testing.
// This implements a VCR (Video Cassette Recorder) pattern that:
// - Records real HTTP interactions during "record" mode
// - Replays recorded interactions during "replay" mode
// - Eliminates the need for external mock servers
package vcr

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Mode represents the VCR operating mode
type Mode int

const (
	// ModeDisabled - VCR is disabled, pass through to real transport
	ModeDisabled Mode = iota
	// ModeRecording - Record real HTTP interactions to cassette
	ModeRecording
	// ModeReplaying - Replay interactions from cassette
	ModeReplaying
)

// Interaction represents a single HTTP request/response pair
type Interaction struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// Request represents a recorded HTTP request
type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// Response represents a recorded HTTP response
type Response struct {
	StatusCode int               `json:"status_code"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body"`
}

// Cassette stores all interactions for a test scenario
type Cassette struct {
	Name         string        `json:"name"`
	Interactions []Interaction `json:"interactions"`

	mu    sync.Mutex
	index int
}

// Recorder is the main VCR recorder/player
type Recorder struct {
	cassette      *Cassette
	mode          Mode
	realTransport http.RoundTripper
	mu            sync.Mutex

	// FilterHeaders contains headers to filter out (e.g., Authorization)
	FilterHeaders []string

	// MatcherFunc allows custom request matching logic
	MatcherFunc func(r *http.Request, recorded *Request) bool
}

// NewRecorder creates a new VCR recorder
func NewRecorder(cassetteName string, mode Mode) (*Recorder, error) {
	r := &Recorder{
		mode:          mode,
		realTransport: http.DefaultTransport,
		FilterHeaders: []string{"Authorization", "ApiKey", "X-Api-Key"},
	}

	cassetteDir := getCassetteDir()
	cassettePath := filepath.Join(cassetteDir, cassetteName+".json")

	if mode == ModeReplaying {
		// Load existing cassette
		cassette, err := LoadCassette(cassettePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load cassette %s: %w", cassetteName, err)
		}
		r.cassette = cassette
	} else if mode == ModeRecording {
		// Create new cassette
		r.cassette = &Cassette{
			Name:         cassetteName,
			Interactions: []Interaction{},
		}
	}

	return r, nil
}

// SetRealTransport sets the underlying transport for recording mode
func (r *Recorder) SetRealTransport(transport http.RoundTripper) {
	r.realTransport = transport
}

// RoundTrip implements http.RoundTripper
func (r *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.mode == ModeDisabled {
		return r.realTransport.RoundTrip(req)
	}

	if r.mode == ModeReplaying {
		return r.replay(req)
	}

	return r.record(req)
}

// record makes a real request and records the interaction
func (r *Recorder) record(req *http.Request) (*http.Response, error) {
	// Read request body
	var reqBody []byte
	if req.Body != nil {
		var err error
		reqBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	// Make real request
	resp, err := r.realTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	// Record interaction
	interaction := Interaction{
		Request: Request{
			Method:  req.Method,
			URL:     req.URL.String(),
			Headers: r.filterHeaders(req.Header),
			Body:    string(reqBody),
		},
		Response: Response{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Headers:    headerToMap(resp.Header),
			Body:       string(respBody),
		},
	}

	r.mu.Lock()
	r.cassette.Interactions = append(r.cassette.Interactions, interaction)
	r.mu.Unlock()

	return resp, nil
}

// replay finds a matching interaction and returns the recorded response
func (r *Recorder) replay(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Read request body for matching
	var reqBody []byte
	if req.Body != nil {
		var err error
		reqBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	// Find matching interaction
	for _, interaction := range r.cassette.Interactions {
		if r.matches(req, &interaction.Request, string(reqBody)) {
			return r.buildResponse(&interaction.Response, req), nil
		}
	}

	return nil, fmt.Errorf("no matching interaction found for %s %s", req.Method, req.URL.String())
}

// matches checks if a request matches a recorded request
func (r *Recorder) matches(req *http.Request, recorded *Request, reqBody string) bool {
	// Use custom matcher if provided
	if r.MatcherFunc != nil {
		return r.MatcherFunc(req, recorded)
	}

	// Default matching: method + URL path (without query params for flexibility)
	if req.Method != recorded.Method {
		return false
	}

	// Compare URL paths
	reqPath := req.URL.Path
	recordedURL := recorded.URL
	if idx := strings.Index(recordedURL, "?"); idx != -1 {
		recordedURL = recordedURL[:idx]
	}
	if !strings.HasSuffix(recordedURL, reqPath) && !strings.Contains(recordedURL, reqPath) {
		return false
	}

	// For POST/PUT/PATCH, also match body hash for uniqueness
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		if hashBody(reqBody) != hashBody(recorded.Body) {
			return false
		}
	}

	return true
}

// buildResponse creates an http.Response from a recorded Response
func (r *Recorder) buildResponse(recorded *Response, req *http.Request) *http.Response {
	header := make(http.Header)
	for k, v := range recorded.Headers {
		header.Set(k, v)
	}

	return &http.Response{
		StatusCode:    recorded.StatusCode,
		Status:        recorded.Status,
		Header:        header,
		Body:          io.NopCloser(bytes.NewReader([]byte(recorded.Body))),
		ContentLength: int64(len(recorded.Body)),
		Request:       req,
	}
}

// filterHeaders removes sensitive headers from the recording
func (r *Recorder) filterHeaders(h http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range h {
		skip := false
		for _, filter := range r.FilterHeaders {
			if strings.EqualFold(key, filter) {
				skip = true
				break
			}
		}
		if !skip && len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// Stop saves the cassette if in recording mode
func (r *Recorder) Stop() error {
	if r.mode == ModeRecording && r.cassette != nil {
		return r.cassette.Save()
	}
	return nil
}

// Save writes the cassette to disk
func (c *Cassette) Save() error {
	cassetteDir := getCassetteDir()
	if err := os.MkdirAll(cassetteDir, 0755); err != nil {
		return fmt.Errorf("failed to create cassette directory: %w", err)
	}

	cassettePath := filepath.Join(cassetteDir, c.Name+".json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cassette: %w", err)
	}

	if err := os.WriteFile(cassettePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cassette: %w", err)
	}

	return nil
}

// LoadCassette loads a cassette from disk
func LoadCassette(path string) (*Cassette, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cassette Cassette
	if err := json.Unmarshal(data, &cassette); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cassette: %w", err)
	}

	return &cassette, nil
}

// getCassetteDir returns the directory for storing cassettes
func getCassetteDir() string {
	// Check for custom cassette directory
	if dir := os.Getenv("VCR_CASSETTE_DIR"); dir != "" {
		return dir
	}

	// Try multiple locations for cassettes
	// 1. Current working directory
	cwd, _ := os.Getwd()

	// 2. testdata/cassettes relative to cwd
	dir := filepath.Join(cwd, "testdata", "cassettes")
	if _, err := os.Stat(dir); err == nil {
		return dir
	}

	// 3. spectrocloud/testdata/cassettes (from project root)
	dir = filepath.Join(cwd, "spectrocloud", "testdata", "cassettes")
	if _, err := os.Stat(dir); err == nil {
		return dir
	}

	// 4. ../testdata/cassettes (from spectrocloud dir)
	dir = filepath.Join(cwd, "..", "testdata", "cassettes")
	if _, err := os.Stat(dir); err == nil {
		return dir
	}

	// Default to testdata/cassettes in current directory
	return filepath.Join("testdata", "cassettes")
}

// headerToMap converts http.Header to map[string]string
func headerToMap(h http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range h {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// hashBody creates a simple hash of request body for matching
func hashBody(body string) string {
	if body == "" {
		return ""
	}

	// Normalize JSON for consistent hashing
	var normalized interface{}
	if err := json.Unmarshal([]byte(body), &normalized); err == nil {
		if sortedBody, err := json.Marshal(sortKeys(normalized)); err == nil {
			body = string(sortedBody)
		}
	}

	hash := md5.Sum([]byte(body))
	return hex.EncodeToString(hash[:])
}

// sortKeys recursively sorts map keys for consistent JSON comparison
func sortKeys(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		sorted := make(map[string]interface{})
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sorted[k] = sortKeys(v[k])
		}
		return sorted
	case []interface{}:
		for i, item := range v {
			v[i] = sortKeys(item)
		}
		return v
	default:
		return v
	}
}

// GetMode returns the VCR mode based on environment variables
func GetMode() Mode {
	if os.Getenv("VCR_RECORD") == "true" || os.Getenv("VCR_RECORD") == "1" {
		return ModeRecording
	}
	if os.Getenv("VCR_DISABLED") == "true" || os.Getenv("VCR_DISABLED") == "1" {
		return ModeDisabled
	}
	return ModeReplaying
}
