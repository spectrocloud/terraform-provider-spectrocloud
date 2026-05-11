package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
Type - Unit Test
Description - Testing ToAlert function for email schema
*/
func TestToAlertEmail(t *testing.T) {
	rd := resourceAlert().TestResourceData()
	err := rd.Set("type", "email")
	if err != nil {
		return
	}
	err = rd.Set("is_active", true)
	if err != nil {
		return
	}
	err = rd.Set("alert_all_users", false)
	if err != nil {
		return
	}
	emails := []string{"testuser1@spectrocloud.com", "testuser2@spectrocloud.com"}
	err = rd.Set("identifiers", emails)
	if err != nil {
		return
	}
	alertChannelEmail := toAlert(rd)
	if alertChannelEmail.Type != "email" || alertChannelEmail.IsActive != true ||
		alertChannelEmail.AlertAllUsers != false || alertChannelEmail == nil {
		t.Fail()
		t.Logf("Alert email channel schema definition is failing")
	}
	if !reflect.DeepEqual(emails, alertChannelEmail.Identifiers) {
		t.Fail()
		t.Logf("Alert email identifiers are not matching with input")
	}
}

/*
Type - Unit Test
Description - Testing ToAlert function for http schema
*/
func TestToAlertHttp(t *testing.T) {
	rd := resourceAlert().TestResourceData()
	rd.Set("type", "http")
	rd.Set("is_active", true)
	rd.Set("alert_all_users", false)
	rd.Set("identifiers", []string{})
	var http []map[string]interface{}
	hookConfig := map[string]interface{}{
		"method": "POST",
		"url":    "https://www.openhook.com/spc/notify",
		"body":   "{ \"text\": \"{{message}}\" }",
		"headers": map[string]interface{}{
			"tag":    "Health",
			"source": "spectrocloud",
		},
	}
	http = append(http, hookConfig)
	rd.Set("http", http)
	alertChannelHttp := toAlert(rd)
	if alertChannelHttp.Type != "http" || alertChannelHttp.IsActive != true ||
		alertChannelHttp.AlertAllUsers != false || alertChannelHttp == nil {
		t.Fail()
		t.Logf("Alert http channel schema definition is failing")
	}
	if http[0]["method"] != alertChannelHttp.HTTP.Method || http[0]["url"] != alertChannelHttp.HTTP.URL ||
		http[0]["body"] != alertChannelHttp.HTTP.Body || len(http[0]["headers"].(map[string]interface{})) != len(alertChannelHttp.HTTP.Headers) {
		t.Fail()
		t.Logf("Alert http configurations are not matching with test http input")
	}
}

/*
Type - Unit Test
Description - Testing ToAlertChannels function for auto-detect with both email and http schema.
*/
func TestToAlertChannelsAutoDetect(t *testing.T) {
	rd := resourceAlert().TestResourceData()
	err := rd.Set("type", "") // Auto-detect based on configuration
	if err != nil {
		return
	}
	err = rd.Set("is_active", true)
	if err != nil {
		return
	}
	err = rd.Set("alert_all_users", false)
	if err != nil {
		return
	}
	emails := []string{"testuser1@spectrocloud.com", "testuser2@spectrocloud.com"}
	err = rd.Set("identifiers", emails)
	if err != nil {
		return
	}
	var http []map[string]interface{}
	hookConfig := map[string]interface{}{
		"method": "POST",
		"url":    "https://www.openhook.com/spc/notify",
		"body":   "{ \"text\": \"{{message}}\" }",
		"headers": map[string]interface{}{
			"tag":    "Health",
			"source": "spectrocloud",
		},
	}
	http = append(http, hookConfig)
	err = rd.Set("http", http)
	if err != nil {
		return
	}

	// Test toAlertChannels returns both email and http channels
	channels := toAlertChannels(rd)
	if len(channels) != 2 {
		t.Fail()
		t.Logf("Expected 2 channels (email and http), got %d", len(channels))
		return
	}

	// First channel should be email
	emailChannel := channels[0]
	if emailChannel.Type != "email" || emailChannel.IsActive != true {
		t.Fail()
		t.Logf("Email channel schema definition is failing")
	}
	if !reflect.DeepEqual(emails, emailChannel.Identifiers) {
		t.Fail()
		t.Logf("Alert email identifiers are not matching with input")
	}

	// Second channel should be http
	httpChannel := channels[1]
	if httpChannel.Type != "http" || httpChannel.IsActive != true {
		t.Fail()
		t.Logf("HTTP channel schema definition is failing")
	}
	if http[0]["method"] != httpChannel.HTTP.Method || http[0]["url"] != httpChannel.HTTP.URL ||
		http[0]["body"] != httpChannel.HTTP.Body || len(http[0]["headers"].(map[string]interface{})) != len(httpChannel.HTTP.Headers) {
		t.Fail()
		t.Logf("Alert http configurations are not matching with test http input")
	}
}

/*
Type - Unit Test
Description - Testing ToAlertChannels function with multiple HTTP configurations.
*/
func TestToAlertMultipleHttp(t *testing.T) {
	rd := resourceAlert().TestResourceData()
	err := rd.Set("type", "http")
	if err != nil {
		return
	}
	err = rd.Set("is_active", true)
	if err != nil {
		return
	}
	err = rd.Set("alert_all_users", false)
	if err != nil {
		return
	}

	var httpConfigs []map[string]interface{}
	// First HTTP webhook
	hookConfig1 := map[string]interface{}{
		"method": "POST",
		"url":    "https://www.webhook1.com/notify",
		"body":   "{ \"text\": \"{{message}}\" }",
		"headers": map[string]interface{}{
			"tag": "Health",
		},
	}
	// Second HTTP webhook
	hookConfig2 := map[string]interface{}{
		"method": "POST",
		"url":    "https://www.webhook2.com/alert",
		"body":   "{ \"alert\": \"{{message}}\" }",
		"headers": map[string]interface{}{
			"source": "spectrocloud",
		},
	}
	httpConfigs = append(httpConfigs, hookConfig1, hookConfig2)
	err = rd.Set("http", httpConfigs)
	if err != nil {
		return
	}

	// Test toAlertChannels returns multiple http channels
	channels := toAlertChannels(rd)
	if len(channels) != 2 {
		t.Fail()
		t.Logf("Expected 2 HTTP channels, got %d", len(channels))
		return
	}

	// Verify first HTTP channel
	if channels[0].Type != "http" || channels[0].HTTP.URL != "https://www.webhook1.com/notify" {
		t.Fail()
		t.Logf("First HTTP channel configuration is incorrect")
	}

	// Verify second HTTP channel
	if channels[1].Type != "http" || channels[1].HTTP.URL != "https://www.webhook2.com/alert" {
		t.Fail()
		t.Logf("Second HTTP channel configuration is incorrect")
	}
}

func prepareAlertTestData() *schema.ResourceData {
	rd := resourceAlert().TestResourceData()
	rd.SetId("test-alert-id")
	_ = rd.Set("type", "") // Auto-detect based on configuration
	_ = rd.Set("is_active", true)
	_ = rd.Set("alert_all_users", false)
	_ = rd.Set("project", "Default")
	_ = rd.Set("component", "ClusterHealth")
	emails := []string{"testuser1@spectrocloud.com", "testuser2@spectrocloud.com"}
	_ = rd.Set("identifiers", emails)
	var http []map[string]interface{}
	hookConfig := map[string]interface{}{
		"method": "POST",
		"url":    "https://www.openhook.com/spc/notify",
		"body":   "{ \"text\": \"{{message}}\" }",
		"headers": map[string]interface{}{
			"tag":    "Health",
			"source": "spectrocloud",
		},
	}
	http = append(http, hookConfig)
	_ = rd.Set("http", http)
	return rd
}

func TestResourceAlertCRUD(t *testing.T) {
	testResourceCRUD(t, prepareAlertTestData, unitTestMockAPIClient,
		resourceAlertCreate, resourceAlertRead, resourceAlertUpdate, resourceAlertDelete)
}
