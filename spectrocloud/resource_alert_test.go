package spectrocloud

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

/*
Type - Unit Test
Description - Testing ToAlert function for email schema
*/
func TestToAlertEmail(t *testing.T) {
	rd := resourceAlert().TestResourceData()
	rd.Set("type", "email")
	rd.Set("is_active", true)
	rd.Set("alert_all_users", false)
	emails := []string{"testuser1@spectrocloud.com", "testuser2@spectrocloud.com"}
	rd.Set("identifiers", emails)
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
Description - Testing ToAlert function for http schema with email schema.
*/
func TestToAlertHttpEmail(t *testing.T) {
	rd := resourceAlert().TestResourceData()
	rd.Set("type", "http")
	rd.Set("is_active", true)
	rd.Set("alert_all_users", false)
	emails := []string{"testuser1@spectrocloud.com", "testuser2@spectrocloud.com"}
	rd.Set("identifiers", emails)
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
	alertChannelHttpEmail := toAlert(rd)
	if alertChannelHttpEmail.Type != "http" || alertChannelHttpEmail.IsActive != true ||
		alertChannelHttpEmail.AlertAllUsers != false || alertChannelHttpEmail == nil {
		t.Fail()
		t.Logf("Alert http channel schema definition is failing")
	}
	if http[0]["method"] != alertChannelHttpEmail.HTTP.Method || http[0]["url"] != alertChannelHttpEmail.HTTP.URL ||
		http[0]["body"] != alertChannelHttpEmail.HTTP.Body || len(http[0]["headers"].(map[string]interface{})) != len(alertChannelHttpEmail.HTTP.Headers) {
		t.Fail()
		t.Logf("Alert http configurations are not matching with test http input")
	}
	if !reflect.DeepEqual(emails, alertChannelHttpEmail.Identifiers) {
		t.Fail()
		t.Logf("Alert email identifiers are not matching with input")
	}
}

/*
Type - Integration Test
Description - Testing all CRUD function for email alerts.
*/
func TestAlertCRUDEmail(t *testing.T) {
	if !IsIntegrationTestEnvSet(baseConfig) {
		t.Skip("Skipping integration test env variable not set")
	}
	conn := client.New(baseConfig.hubbleHost, baseConfig.email, baseConfig.pwd, baseConfig.project, baseConfig.apikey, false, 3)
	var err error
	channelEmail := &models.V1Channel{
		IsActive:      true,
		Type:          "email",
		AlertAllUsers: true,
		Identifiers:   []string{"test@spectrocloud.com", "test2@spectrocloud.com"},
	}
	projectId, err := conn.GetProjectUID(baseConfig.project)
	if err != nil {
		t.Fail()
		t.Logf("\n Unable to read project UID for name - %s", baseConfig.project)
	}
	baseConfig.AlertUid, err = conn.CreateAlert(channelEmail, projectId, baseConfig.component)
	if err != nil {
		t.Fail()
		t.Log("\n Email Alert Creation failed")
	}
	payload, err := conn.ReadAlert(projectId, baseConfig.component, baseConfig.AlertUid)
	if err != nil {
		t.Fail()
		t.Logf("\n Email Alert Read Failed for UID - %s", baseConfig.AlertUid)
	}
	if payload.UID != baseConfig.AlertUid || payload.AlertAllUsers != channelEmail.AlertAllUsers {
		t.Fail()
		t.Logf("\n Email Alert Read Response is not matching with payload - %s", baseConfig.AlertUid)
	}
	channelEmail.IsActive = false
	_, err = conn.UpdateAlert(channelEmail, projectId, baseConfig.component, baseConfig.AlertUid)
	if err != nil {
		t.Fail()
		t.Logf("\n Unable to update email alert for UID - %s", baseConfig.AlertUid)
	}
	payload, err = conn.ReadAlert(projectId, baseConfig.component, baseConfig.AlertUid)
	if err != nil {
		t.Fail()
		t.Logf("\n Unable to read email alert for UID - %s", baseConfig.AlertUid)
	}
	if payload.IsActive != false {
		t.Fail()
		t.Logf("\n Email alert update failed - %s", baseConfig.AlertUid)
	}
	err = conn.DeleteAlerts(projectId, baseConfig.component, baseConfig.AlertUid)
	payload, _ = conn.ReadAlert(projectId, baseConfig.component, baseConfig.AlertUid)
	if err == nil && payload == nil {
		println("> Test TestCRUDAlertEmail Completed Successfully ")
	} else {
		t.Fail()
		t.Logf("\n Email Alert Delete Failed - %s", baseConfig.AlertUid)
	}
}

/*
Type - Integration Test
Description - Testing all CRUD function for http(webhook) alerts.
*/
func TestAlertCRUDHttp(t *testing.T) {
	if !IsIntegrationTestEnvSet(baseConfig) {
		t.Skip("Skipping integration test env variable not set")
	}
	conn := client.New(baseConfig.hubbleHost, baseConfig.email, baseConfig.pwd, baseConfig.project, baseConfig.apikey, false, 3)
	var err error
	header := map[string]string{
		"type": "CH-Notification",
		"tag":  "Spectro",
	}
	channelHttp := &models.V1Channel{
		IsActive:      true,
		Type:          "email",
		AlertAllUsers: true,
		Identifiers:   []string{},
		HTTP: &models.V1ChannelHTTP{
			Body:    "{ \"text\": \"{{message}}\" }",
			Method:  "POST",
			URL:     "https://openhook.com/put/edit2",
			Headers: header,
		},
	}
	projectId, err := conn.GetProjectUID(baseConfig.project)
	if err != nil {
		t.Fail()
		t.Logf("\n Unable to read project UID for name - %s", baseConfig.project)
	}
	baseConfig.AlertUid, err = conn.CreateAlert(channelHttp, projectId, baseConfig.component)
	if err != nil {
		t.Fail()
		t.Log("\n HTTP Alert Creation failed")
	}
	payload, err := conn.ReadAlert(projectId, baseConfig.component, baseConfig.AlertUid)
	if err != nil {
		t.Fail()
		t.Logf("\n HTTP Alert Read Failed for UID - %s", baseConfig.AlertUid)
	}
	if payload.UID != baseConfig.AlertUid || payload.AlertAllUsers != channelHttp.AlertAllUsers {
		t.Fail()
		t.Logf("\n HTTP Alert Read Response is not matching with payload - %s", baseConfig.AlertUid)
	}
	channelHttp.IsActive = false
	_, err = conn.UpdateAlert(channelHttp, projectId, baseConfig.component, baseConfig.AlertUid)
	if err != nil {
		t.Fail()
		t.Logf("\n Unable to update email alert for UID - %s", baseConfig.AlertUid)
	}
	payload, err = conn.ReadAlert(projectId, baseConfig.component, baseConfig.AlertUid)
	if err != nil {
		t.Fail()
		t.Logf("\n Unable to read email alert for UID - %s", baseConfig.AlertUid)
	}
	if payload.IsActive != false {
		t.Fail()
		t.Logf("\n HTTP alert update failed - %s", baseConfig.AlertUid)
	}
	err = conn.DeleteAlerts(projectId, baseConfig.component, baseConfig.AlertUid)
	payload, _ = conn.ReadAlert(projectId, baseConfig.component, baseConfig.AlertUid)
	if err == nil && payload == nil {
		println("> Test TestCRUDAlertHttp Completed Successfully ")
	} else {
		t.Fail()
		t.Logf("\n HTTP Alert Delete Failed - %s", baseConfig.AlertUid)
	}
}

func prepareAlertTestData() *schema.ResourceData {
	rd := resourceAlert().TestResourceData()
	rd.Set("type", "email")
	rd.Set("is_active", true)
	rd.Set("alert_all_users", false)
	rd.Set("project", "Default")
	emails := []string{"testuser1@spectrocloud.com", "testuser2@spectrocloud.com"}
	rd.Set("identifiers", emails)
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
	return rd
}

func TestGetProjectID(t *testing.T) {
	assert := assert.New(t)
	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil
		},
	}
	pjtUid, err := getProjectID(rd, m)
	if err != nil {
		assert.Error(errors.New("unable to read project uid"))
	}
	assert.Equal("test-project-uid", pjtUid)
}

func TestGetProjectIDError(t *testing.T) {
	assert := assert.New(t)
	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "", errors.New("unable to read project uid")
		},
	}
	pjtUid, err := getProjectID(rd, m)
	if err == nil {
		assert.Error(errors.New("unexpected Error"))
	}
	assert.Equal(err.Error(), "unable to read project uid")
	assert.Equal("", pjtUid)
}

func TestResourceAlertCreate(t *testing.T) {
	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil
		},
		CreateAlertFn: func(body *models.V1Channel, projectUID, component string) (string, error) {
			return "test-alert-ui", nil
		},
	}
	ctx := context.Background()
	diags := resourceAlertCreate(ctx, rd, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceAlertCreateProjectUIDError(t *testing.T) {
	assert := assert.New(t)
	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "", errors.New("unable to read project uid")

		},
		CreateAlertFn: func(body *models.V1Channel, projectUID, component string) (string, error) {
			return "test-alert-uid", nil
		},
	}
	ctx := context.Background()
	diags := resourceAlertCreate(ctx, rd, m)
	assert.Equal(diags[0].Summary, "unable to read project uid")
}

func TestResourceAlertCreateAlertUIDError(t *testing.T) {
	assert := assert.New(t)
	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil
		},
		CreateAlertFn: func(body *models.V1Channel, projectUID, component string) (string, error) {
			return "", errors.New("alert creation failed")
		},
	}
	ctx := context.Background()
	diags := resourceAlertCreate(ctx, rd, m)
	assert.Equal(diags[0].Summary, "alert creation failed")
}

func TestResourceAlertUpdate(t *testing.T) {

	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil
		},
		UpdateAlertFn: func(body *models.V1Channel, projectUID, component, alertUID string) (string, error) {
			return "success", nil
		},
	}
	ctx := context.Background()
	diags := resourceAlertUpdate(ctx, rd, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceAlertUpdateError(t *testing.T) {
	assert := assert.New(t)
	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil
		},
		UpdateAlertFn: func(body *models.V1Channel, projectUID, component, alertUID string) (string, error) {
			return "", errors.New("alert update failed")
		},
	}
	ctx := context.Background()
	diags := resourceAlertUpdate(ctx, rd, m)
	assert.Equal(diags[0].Summary, "alert update failed")
}

func TestResourceAlertDelete(t *testing.T) {

	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil
		},
		DeleteAlertsFn: func(projectUID, component, alertUID string) error {
			return nil
		},
	}
	ctx := context.Background()
	diags := resourceAlertDelete(ctx, rd, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceAlertDeleteProjectUIDError(t *testing.T) {
	assert := assert.New(t)
	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "", errors.New("unable to read project uid")
		},
		DeleteAlertsFn: func(projectUID, component, alertUID string) error {
			return nil
		},
	}
	ctx := context.Background()
	diags := resourceAlertDelete(ctx, rd, m)
	assert.Equal(diags[0].Summary, "unable to read project uid")
}

func TestResourceAlertDeleteError(t *testing.T) {
	assert := assert.New(t)
	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil

		},
		DeleteAlertsFn: func(projectUID, component, alertUID string) error {
			return errors.New("unable to delete alert")
		},
	}
	ctx := context.Background()
	diags := resourceAlertDelete(ctx, rd, m)
	assert.Equal(diags[0].Summary, "unable to delete alert")
}

func TestResourceAlertReadAlertNil(t *testing.T) {
	rd := prepareAlertTestData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil

		},
		ReadAlertFn: func(projectUID, component, alertUID string) (*models.V1Channel, error) {
			return nil, nil
		},
	}
	ctx := context.Background()
	diags := resourceAlertRead(ctx, rd, m)

	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceAlertReadAlertEmail(t *testing.T) {
	rd := resourceAlert().TestResourceData()
	rd.Set("type", "email")
	rd.Set("is_active", true)
	rd.Set("alert_all_users", false)
	rd.Set("project", "Default")
	emails := []string{"testuser1@spectrocloud.com", "testuser2@spectrocloud.com"}
	rd.Set("identifiers", emails)
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil

		},
		ReadAlertFn: func(projectUID, component, alertUID string) (*models.V1Channel, error) {
			rd.Set("UID", "alert-test-uid")
			return toAlert(rd), nil
		},
	}
	ctx := context.Background()
	diags := resourceAlertRead(ctx, rd, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceAlertReadAlertHttp(t *testing.T) {
	rd := resourceAlert().TestResourceData()
	rd.Set("type", "http")
	rd.Set("is_active", true)
	rd.Set("alert_all_users", false)
	rd.Set("project", "Default")
	emails := []string{"testuser1@spectrocloud.com", "testuser2@spectrocloud.com"}
	rd.Set("identifiers", emails)
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
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil

		},
		ReadAlertFn: func(projectUID, component, alertUID string) (*models.V1Channel, error) {
			rd.Set("UID", "alert-test-uid")
			return toAlert(rd), nil
		},
	}
	ctx := context.Background()
	diags := resourceAlertRead(ctx, rd, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceAlertReadNegative(t *testing.T) {
	rd := resourceAlert().TestResourceData()
	m := &client.V1Client{
		GetProjectUIDFn: func(projectName string) (string, error) {
			return "test-project-uid", nil

		},
		ReadAlertFn: func(projectUID, component, alertUID string) (*models.V1Channel, error) {
			rd.Set("UID", "alert-test-uid")
			return toAlert(rd), nil
		},
	}
	ctx := context.Background()
	diags := resourceAlertRead(ctx, rd, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}
