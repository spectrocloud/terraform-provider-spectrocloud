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
Description - Testing ToAlert function for http schema with email schema.
*/
func TestToAlertHttpEmail(t *testing.T) {
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

//func TestGetProjectIDError(t *testing.T) {
//	assert := assert.New(t)
//	rd := prepareAlertTestData()
//	m := &client.V1Client{}
//	pjtUid, err := getProjectID(rd, m)
//	if err == nil {
//		assert.Error(errors.New("unexpected Error"))
//	}
//	assert.Equal(err.Error(), "unable to read project uid")
//	assert.Equal("", pjtUid)
//}

//func TestResourceAlertCreate(t *testing.T) {
//	rd := prepareAlertTestData()
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertCreate(ctx, rd, m)
//	if len(diags) > 0 {
//		t.Errorf("Unexpected diagnostics: %#v", diags)
//	}
//}

//func TestResourceAlertCreateProjectUIDError(t *testing.T) {
//	assert := assert.New(t)
//	rd := prepareAlertTestData()
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertCreate(ctx, rd, m)
//	assert.Equal(diags[0].Summary, "unable to read project uid")
//}

//func TestResourceAlertCreateAlertUIDError(t *testing.T) {
//	assert := assert.New(t)
//	rd := prepareAlertTestData()
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertCreate(ctx, rd, m)
//	assert.Equal(diags[0].Summary, "alert creation failed")
//}

//func TestResourceAlertUpdate(t *testing.T) {
//
//	rd := prepareAlertTestData()
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertUpdate(ctx, rd, m)
//	if len(diags) > 0 {
//		t.Errorf("Unexpected diagnostics: %#v", diags)
//	}
//}

//func TestResourceAlertUpdateError(t *testing.T) {
//	assert := assert.New(t)
//	rd := prepareAlertTestData()
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertUpdate(ctx, rd, m)
//	assert.Equal(diags[0].Summary, "alert update failed")
//}

//func TestResourceAlertDelete(t *testing.T) {
//
//	rd := prepareAlertTestData()
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertDelete(ctx, rd, m)
//	if len(diags) > 0 {
//		t.Errorf("Unexpected diagnostics: %#v", diags)
//	}
//}

//func TestResourceAlertDeleteProjectUIDError(t *testing.T) {
//	assert := assert.New(t)
//	rd := prepareAlertTestData()
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertDelete(ctx, rd, m)
//	assert.Equal(diags[0].Summary, "unable to read project uid")
//}

//func TestResourceAlertDeleteError(t *testing.T) {
//	assert := assert.New(t)
//	rd := prepareAlertTestData()
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertDelete(ctx, rd, m)
//	assert.Equal(diags[0].Summary, "unable to delete alert")
//}

//func TestResourceAlertReadAlertNil(t *testing.T) {
//	rd := prepareAlertTestData()
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertRead(ctx, rd, m)
//
//	if len(diags) > 0 {
//		t.Errorf("Unexpected diagnostics: %#v", diags)
//	}
//}

//func TestResourceAlertReadAlertEmail(t *testing.T) {
//	rd := resourceAlert().TestResourceData()
//	rd.Set("type", "email")
//	rd.Set("is_active", true)
//	rd.Set("alert_all_users", false)
//	rd.Set("project", "Default")
//	emails := []string{"testuser1@spectrocloud.com", "testuser2@spectrocloud.com"}
//	rd.Set("identifiers", emails)
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertRead(ctx, rd, m)
//	if len(diags) > 0 {
//		t.Errorf("Unexpected diagnostics: %#v", diags)
//	}
//}

//func TestResourceAlertReadAlertHttp(t *testing.T) {
//	rd := resourceAlert().TestResourceData()
//	rd.Set("type", "http")
//	rd.Set("is_active", true)
//	rd.Set("alert_all_users", false)
//	rd.Set("project", "Default")
//	emails := []string{"testuser1@spectrocloud.com", "testuser2@spectrocloud.com"}
//	rd.Set("identifiers", emails)
//	var http []map[string]interface{}
//	hookConfig := map[string]interface{}{
//		"method": "POST",
//		"url":    "https://www.openhook.com/spc/notify",
//		"body":   "{ \"text\": \"{{message}}\" }",
//		"headers": map[string]interface{}{
//			"tag":    "Health",
//			"source": "spectrocloud",
//		},
//	}
//	http = append(http, hookConfig)
//	rd.Set("http", http)
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertRead(ctx, rd, m)
//	if len(diags) > 0 {
//		t.Errorf("Unexpected diagnostics: %#v", diags)
//	}
//}

//func TestResourceAlertReadNegative(t *testing.T) {
//	rd := resourceAlert().TestResourceData()
//	m := &client.V1Client{}
//	ctx := context.Background()
//	diags := resourceAlertRead(ctx, rd, m)
//	if len(diags) > 0 {
//		t.Errorf("Unexpected diagnostics: %#v", diags)
//	}
//}
