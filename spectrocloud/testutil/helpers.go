// Copyright (c) Spectro Cloud
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// RandomName returns a name with the given prefix and a random suffix (timestamp-based).
// Used by acceptance tests to avoid resource name collisions.
func RandomName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()%100000)
}

// TestAccPreCheck skips the test if acceptance test environment is not configured.
// Checks for SPECTROCLOUD_APIKEY so that tests requiring real API are skipped when not set.
func TestAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("SPECTROCLOUD_APIKEY") == "" {
		t.Skip("Skipping acceptance test: SPECTROCLOUD_APIKEY not set")
	}
}
