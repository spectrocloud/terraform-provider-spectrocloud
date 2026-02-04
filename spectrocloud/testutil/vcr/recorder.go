// Copyright (c) Spectro Cloud
// SPDX-License-Identifier: MPL-2.0

package vcr

import "os"

// Mode is the VCR operation mode (record or replay).
type Mode int

const (
	// ModeReplaying replays from an existing cassette.
	ModeReplaying Mode = iota
	// ModeRecording records interactions to a cassette.
	ModeRecording
)

// GetMode returns the VCR mode from the environment.
// Set VCR_RECORD=true to record; otherwise replay.
func GetMode() Mode {
	if os.Getenv("VCR_RECORD") == "true" {
		return ModeRecording
	}
	return ModeReplaying
}

// Recorder is a no-op recorder for tests that only need the mode/pattern.
// Full record/replay is done via LoadCassette and an httptest.Server in tests.
type Recorder struct{}

// NewRecorder creates a recorder for the given cassette name and mode.
// Returns a no-op recorder; actual replay is done with LoadCassette + httptest.Server.
func NewRecorder(_ string, _ Mode) (*Recorder, error) {
	return &Recorder{}, nil
}

// Stop stops the recorder (no-op).
func (r *Recorder) Stop() error {
	return nil
}
