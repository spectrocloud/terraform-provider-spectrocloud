// Copyright (c) Spectro Cloud
// SPDX-License-Identifier: MPL-2.0

package vcr

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Cassette represents a recorded set of HTTP interactions for replay in tests.
type Cassette struct {
	Name         string        `json:"name"`
	Interactions []Interaction `json:"interactions"`
}

// Interaction is a single request/response pair.
type Interaction struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// Request holds the recorded HTTP request.
type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// Response holds the recorded HTTP response.
type Response struct {
	StatusCode int               `json:"status_code"`
	Status     string            `json:"status,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body"`
}

// LoadCassette loads a cassette from the given path. Tries the path as-is,
// then relative to current working directory (e.g. testdata/cassettes/...),
// so tests can run from repo root or from spectrocloud/.
func LoadCassette(path string) (*Cassette, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Try relative to cwd (e.g. when running from spectrocloud/)
		alt := filepath.Join("testdata", "cassettes", filepath.Base(path))
		data, err = os.ReadFile(alt)
		if err != nil {
			return nil, err
		}
	}
	var c Cassette
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
