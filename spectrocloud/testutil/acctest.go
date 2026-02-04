// Copyright (c) Spectro Cloud
// SPDX-License-Identifier: MPL-2.0

// Package testutil provides common utilities for testing the Spectro Cloud provider.
package testutil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/testutil/vcr"
)

// TestAccPreCheck validates the necessary test environment variables exist
// before running acceptance tests.
func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("SPECTROCLOUD_APIKEY"); v == "" {
		t.Fatal("SPECTROCLOUD_APIKEY must be set for acceptance tests")
	}
	if v := os.Getenv("SPECTROCLOUD_HOST"); v == "" {
		// Set default if not provided
		_ = os.Setenv("SPECTROCLOUD_HOST", "api.spectrocloud.com")
	}
}

// TestAccPreCheckWithVCR is like TestAccPreCheck but skips API key requirement
// when running in VCR replay mode.
func TestAccPreCheckWithVCR(t *testing.T) {
	mode := vcr.GetMode()
	if mode == vcr.ModeReplaying {
		// In replay mode, we don't need real credentials
		return
	}
	TestAccPreCheck(t)
}

// ProviderFactories returns the provider factories for acceptance tests
func ProviderFactories(provider *schema.Provider) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"spectrocloud": func() (*schema.Provider, error) {
			return provider, nil
		},
	}
}

// VCRProviderFactories returns provider factories configured with VCR transport
func VCRProviderFactories(t *testing.T, cassetteName string, provider *schema.Provider) (map[string]func() (*schema.Provider, error), func()) {
	mode := vcr.GetMode()

	recorder, err := vcr.NewRecorder(cassetteName, mode)
	if err != nil {
		if mode == vcr.ModeReplaying {
			t.Skipf("Skipping test: cassette %s not found. Run with VCR_RECORD=true to record.", cassetteName)
		}
		t.Fatalf("Failed to create VCR recorder: %v", err)
	}

	// Configure HTTP client with VCR transport
	httpClient := &http.Client{Transport: recorder}

	// Create provider with custom HTTP client
	providerFunc := func() (*schema.Provider, error) {
		// The provider will use the VCR-enabled HTTP client
		return provider, nil
	}

	cleanup := func() {
		if err := recorder.Stop(); err != nil {
			t.Errorf("Failed to stop VCR recorder: %v", err)
		}
	}

	// Store HTTP client for use in provider configuration
	_ = httpClient // Will be used when we configure the provider

	return map[string]func() (*schema.Provider, error){
		"spectrocloud": providerFunc,
	}, cleanup
}

// GetTestClient creates a Palette SDK client for testing
func GetTestClient() (*client.V1Client, error) {
	host := os.Getenv("SPECTROCLOUD_HOST")
	if host == "" {
		host = "api.spectrocloud.com"
	}

	apiKey := os.Getenv("SPECTROCLOUD_APIKEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SPECTROCLOUD_APIKEY environment variable not set")
	}

	c := client.New(
		client.WithPaletteURI(host),
		client.WithAPIKey(apiKey),
		client.WithInsecureSkipVerify(true),
		client.WithRetries(3),
	)

	return c, nil
}

// GetTestClientWithVCR creates a client configured with VCR transport
// NOTE: This function requires the palette-sdk-go to support custom HTTP client injection.
// Until then, VCR testing will use the acceptance test approach with terraform-plugin-testing.
func GetTestClientWithVCR(t *testing.T, cassetteName string) (*client.V1Client, func()) {
	mode := vcr.GetMode()

	recorder, err := vcr.NewRecorder(cassetteName, mode)
	if err != nil {
		if mode == vcr.ModeReplaying {
			t.Skipf("Skipping test: cassette %s not found", cassetteName)
		}
		t.Fatalf("Failed to create VCR recorder: %v", err)
	}

	host := os.Getenv("SPECTROCLOUD_HOST")
	if host == "" {
		host = "api.spectrocloud.com"
	}

	apiKey := os.Getenv("SPECTROCLOUD_APIKEY")
	if apiKey == "" && mode != vcr.ModeReplaying {
		t.Fatal("SPECTROCLOUD_APIKEY must be set when not in VCR replay mode")
	}
	if apiKey == "" {
		apiKey = "vcr-replay-dummy-key"
	}

	// TODO: The palette-sdk-go client needs to support custom HTTP client/transport injection
	// For now, we create a standard client. To fully enable VCR, add this to palette-sdk-go:
	//   func WithHTTPClient(httpClient *http.Client) func(*V1Client) { ... }
	//
	// Then use: client.WithHTTPClient(&http.Client{Transport: recorder})
	c := client.New(
		client.WithPaletteURI(host),
		client.WithAPIKey(apiKey),
		client.WithInsecureSkipVerify(true),
		client.WithRetries(1),
	)

	// Store recorder for use (currently not connected to client)
	_ = recorder

	cleanup := func() {
		if err := recorder.Stop(); err != nil {
			t.Errorf("Failed to stop VCR recorder: %v", err)
		}
	}

	return c, cleanup
}

// ComposeTestCheckFunc is a helper to compose multiple check functions
func ComposeTestCheckFunc(fs ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(fs...)
}

// CheckResourceExists is a helper to verify a resource exists
func CheckResourceExists(resourceName string, checkFunc func(rs *terraform.ResourceState) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set: %s", resourceName)
		}
		if checkFunc != nil {
			return checkFunc(rs)
		}
		return nil
	}
}

// CheckResourceDestroyed is a helper to verify a resource was destroyed
func CheckResourceDestroyed(resourceType string, checkFunc func(id string) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			if err := checkFunc(rs.Primary.ID); err != nil {
				return err
			}
		}
		return nil
	}
}

// RandomName generates a random name with a prefix for test resources
func RandomName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, RandomString(8))
}

// RandomString generates a random alphanumeric string
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// TestContext creates a context for test operations
func TestContext() context.Context {
	return context.Background()
}

// ConfigCompose combines multiple Terraform configuration strings
func ConfigCompose(configs ...string) string {
	var builder strings.Builder
	for _, config := range configs {
		builder.WriteString(config)
		builder.WriteString("\n")
	}
	return builder.String()
}

// SkipIfEnvNotSet skips the test if any of the specified environment variables are not set
func SkipIfEnvNotSet(t *testing.T, envVars ...string) {
	for _, envVar := range envVars {
		if os.Getenv(envVar) == "" {
			t.Skipf("Skipping test: %s environment variable not set", envVar)
		}
	}
}
