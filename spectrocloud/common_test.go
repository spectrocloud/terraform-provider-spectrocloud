package spectrocloud

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

//type Cred struct {
//	hubbleHost string
//	project    string
//	apikey     string
//	component  string
//	AlertUid   string
//}

const (
	negativeHost  = "127.0.0.1:8888"
	host          = "127.0.0.1:8080"
	trace         = false
	retryAttempts = 10
	apiKey        = "12345"
	projectName   = "unittest"
	projectUID    = "testprojectuid"
)

type CodedError struct {
	Code    string
	Message string
}

func (e CodedError) Error() string {
	return e.Message
}

// var baseConfig Cred
var unitTestMockAPIClient interface{}
var unitTestMockAPINegativeClient interface{}

var basePath = ""
var startMockApiServerScript = ""
var stopMockApiServerScript = ""

func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	_ = os.Setenv("TF_SRC", filepath.Dir(cwd))
	basePath = os.Getenv("TF_SRC")
	startMockApiServerScript = basePath + "/tests/mockApiServer/start_mock_api_server.sh"
	stopMockApiServerScript = basePath + "/tests/mockApiServer/stop_mock_api_server.sh"
	fmt.Printf("\033[1;36m%s\033[0m", "> [Debug] Basepath -"+basePath+" \n")
	err := setup()
	if err != nil {
		fmt.Printf("Error during setup: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	teardown()
	os.Exit(code)
}

func unitTestProviderConfigure(ctx context.Context) (interface{}, diag.Diagnostics) {
	host := host
	apiKey := apiKey
	retryAttempts := retryAttempts

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := client.New(
		client.WithPaletteURI(host),
		client.WithAPIKey(apiKey),
		client.WithRetries(retryAttempts),
		client.WithInsecureSkipVerify(true),
		client.WithRetries(1))

	//// comment to trace flag
	//client.WithTransportDebug()(c)

	uid := projectUID
	ProviderInitProjectUid = uid
	client.WithScopeProject(uid)(c)
	return c, diags
}

func unitTestNegativeCaseProviderConfigure(ctx context.Context) (interface{}, diag.Diagnostics) {
	apiKey := apiKey
	retryAttempts := retryAttempts

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := client.New(
		client.WithPaletteURI(negativeHost),
		client.WithAPIKey(apiKey),
		client.WithRetries(retryAttempts),
		client.WithInsecureSkipVerify(true),
		client.WithRetries(1))

	//// comment to trace flag
	//client.WithTransportDebug()(c)

	uid := projectUID
	ProviderInitProjectUid = uid
	client.WithScopeProject(uid)(c)
	return c, diags
}

func checkMockServerHealth() error {
	maxRetries := 5
	delay := 2 * time.Second

	// Skip TLS verification (use with caution; not recommended for production)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{Transport: tr}

	for i := 0; i < maxRetries; i++ {
		// Create a new HTTP request
		req, err := http.NewRequest("GET", "https://127.0.0.1:8080/v1/health", nil)
		if err != nil {
			return err
		}

		// Add the API key as a header
		req.Header.Set("ApiKey", "12345")

		// Send the request
		resp, err := c.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			// Server is up and running
			err := resp.Body.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if resp != nil {
			err := resp.Body.Close()
			if err != nil {
				return err
			}
		}

		// Wait before retrying
		time.Sleep(delay)
	}

	return errors.New("server is not responding after multiple attempts")
}

func setup() error {
	fmt.Printf("\033[1;36m%s\033[0m", "> Starting Mock API Server \n")
	var ctx context.Context

	cmd := exec.Command("sh", startMockApiServerScript)
	output, err := cmd.CombinedOutput()
	err = checkMockServerHealth()
	if err != nil {
		fmt.Printf("Failed to run start api server script: %s\nError: %s", output, err)
		return err
	}

	fmt.Printf("\033[1;36m%s\033[0m", "> Started Mock Api Server at https://127.0.0.1:8080 & https://127.0.0.1:8888 \n")
	unitTestMockAPIClient, _ = unitTestProviderConfigure(ctx)
	unitTestMockAPINegativeClient, _ = unitTestNegativeCaseProviderConfigure(ctx)
	fmt.Printf("\033[1;36m%s\033[0m", "> Setup completed \n")
	return nil
}

func teardown() {
	cmd := exec.Command("bash", stopMockApiServerScript)
	_, _ = cmd.CombinedOutput()
	fmt.Printf("\033[1;36m%s\033[0m", "> Stopped Mock Api Server \n")
	fmt.Printf("\033[1;36m%s\033[0m", "> Teardown completed \n")
	err := deleteBuild()
	if err != nil {
		fmt.Printf("Test Clean up is incomplete: %v\n", err)
	}
}

func deleteBuild() error {
	err := os.Remove(basePath + "/tests/mockApiServer/MockBuild")
	if err != nil {
		return err
	}
	return nil
}

func assertFirstDiagMessage(t *testing.T, diags diag.Diagnostics, msg string) {
	if assert.NotEmpty(t, diags, "Expected diags to contain at least one element") {
		assert.Contains(t, diags[0].Summary, msg, "The first diagnostic message does not contain the expected error message")
	}
}

func TestHandleReadError_NotFound(t *testing.T) {
	resource := resourceProject().TestResourceData()

	resource.SetId("something")

	err := error(CodedError{
		Code:    "ResourceNotFound",
		Message: "ResourceNotFound: not found",
	})

	_ = handleReadError(resource, err, nil)

	assert.Equal(t, "something", resource.Id())
}

func TestHandleReadError_OtherError(t *testing.T) {
	resource := resourceProject().TestResourceData()

	err := fmt.Errorf("unexpected error")

	diags := handleReadError(resource, err, nil)

	assert.Len(t, diags, 1)
	assert.Equal(t, diag.Error, diags[0].Severity)
	assert.Contains(t, diags[0].Summary, "unexpected error")
}
