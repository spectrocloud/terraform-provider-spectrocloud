package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/client"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

//type Cred struct {
//	hubbleHost string
//	project    string
//	apikey     string
//	component  string
//	AlertUid   string
//}

const (
	host          = "127.0.0.1:8080"
	trace         = false
	retryAttempts = 10
	apiKey        = "12345"
	projectName   = "unittest"
	projectUID    = "testprojectuid"
)

// var baseConfig Cred
var unitTestMockAPIClient interface{}

var basePath = ""
var startMockApiServerScript = ""
var stopMockApiServerScript = ""

func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	_ = os.Setenv("TF_SRC", filepath.Dir(cwd))
	basePath = os.Getenv("TF_SRC")
	startMockApiServerScript = basePath + "/tests/mockApiServer/start_mock_api_server.sh"
	stopMockApiServerScript = basePath + "/tests/mockApiServer/stop_mock_api_server.sh"

	setup()
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

func setup() {
	fmt.Printf("\033[1;36m%s\033[0m", "> Starting Mock API Server \n")
	var ctx context.Context

	cmd := exec.Command("sh", startMockApiServerScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to run start api server script: %s\nError: %s", output, err)
	}
	fmt.Printf("\033[1;36m%s\033[0m", "> Started Mock Api Server at https://127.0.0.1:8080 \n")
	unitTestMockAPIClient, _ = unitTestProviderConfigure(ctx)

	fmt.Printf("\033[1;36m%s\033[0m", "> Setup completed \n")
}

func teardown() {
	cmd := exec.Command("bash", stopMockApiServerScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to run stop api server script: %s\nError: %s", output, err)
	}
	fmt.Printf("\033[1;36m%s\033[0m", "> Stopped Mock Api Server \n")
	fmt.Printf("\033[1;36m%s\033[0m", "> Teardown completed \n")
}
