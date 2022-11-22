package test

import (
	"github.com/spectrocloud/hapi/models"
	userC "github.com/spectrocloud/hapi/user/client/v1"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"testing"
)

func TestAlertResourceEmailSchema(t *testing.T) {
	channelEmail := &models.V1Channel{
		IsActive:      true,
		Type:          "email",
		AlertAllUsers: true,
		Identifiers:   []string{"test@spectrocloud.com", "test2@spectrocloud.com"},
	}
	conn := client.New(baseConfig.hubbleHost, baseConfig.email, baseConfig.pwd, baseConfig.project, baseConfig.apikey, false, 3)
	projectId, _ := conn.GetProjectUID(baseConfig.project)
	params := userC.NewV1ProjectsUIDAlertCreateParams().WithUID(projectId).WithBody(channelEmail).WithComponent(baseConfig.component)
	if params == nil {
		t.Fail()
		t.Logf("Failed with email alert schema for allert creation")
	}

}

func TestAlertResourceHTTPSchema(t *testing.T) {
	header := map[string]string{
		"type": "CH-Notification",
		"tag":  "Spectro",
	}
	channelHttp := &models.V1Channel{
		IsActive:      true,
		Type:          "http",
		AlertAllUsers: true,
		Identifiers:   []string{},
		HTTP: &models.V1ChannelHTTP{
			Body:    "{ \"text\": \"{{message}}\" }",
			Method:  "POST",
			URL:     "https://openhook.com/put/edit2",
			Headers: header,
		},
	}
	conn := client.New(baseConfig.hubbleHost, baseConfig.email, baseConfig.pwd, baseConfig.project, baseConfig.apikey, false, 3)
	projectId, _ := conn.GetProjectUID(baseConfig.project)
	params := userC.NewV1ProjectsUIDAlertCreateParams().WithUID(projectId).WithBody(channelHttp).WithComponent(baseConfig.component)
	if params == nil {
		t.Fail()
		t.Logf("Failed with Http alert schema for allert creation")
	}

}

func TestAlertResourceEmailHTTPSchema(t *testing.T) {
	header := map[string]string{
		"type": "CH-Notification",
		"tag":  "Spectro",
	}
	channelEmailhttp := &models.V1Channel{
		IsActive:      true,
		Type:          "http",
		AlertAllUsers: true,
		Identifiers:   []string{"test@spectrocloud.com", "test2@spectrocloud.com"},
		HTTP: &models.V1ChannelHTTP{
			Body:    "{ \"text\": \"{{message}}\" }",
			Method:  "POST",
			URL:     "https://openhook.com/put/edit2",
			Headers: header,
		},
	}
	conn := client.New(baseConfig.hubbleHost, baseConfig.email, baseConfig.pwd, baseConfig.project, baseConfig.apikey, false, 3)
	projectId, _ := conn.GetProjectUID(baseConfig.project)
	params := userC.NewV1ProjectsUIDAlertCreateParams().WithUID(projectId).WithBody(channelEmailhttp).WithComponent(baseConfig.component)
	if params == nil {
		t.Fail()
		t.Logf("Failed with email alert schema for allert creation")
	}

}

func TestAlertCRUDEmail(t *testing.T) {
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

func TestAlertCRUDHttp(t *testing.T) {
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
