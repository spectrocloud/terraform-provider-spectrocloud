package spectrocloud

import (
	"fmt"
	"os"
	"testing"
)

type Cred struct {
	hubbleHost string
	project    string
	apikey     string
	component  string
	AlertUid   string
}

var baseConfig Cred

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	// Setting up test credentials & base config from env variables
	baseConfig.hubbleHost = getEnvWithFallBack("TEST_HOST")
	baseConfig.project = getEnvWithFallBack("TEST_PROJECT")
	baseConfig.apikey = getEnvWithFallBack("TEST_API_KEY")
	baseConfig.component = "ClusterHealth"
	baseConfig.AlertUid = ""
	if IsIntegrationTestEnvSet(baseConfig) {
		fmt.Printf("\033[1;36m%s\033[0m", "> Credentials & Base config setup completed\n")
		fmt.Printf("\033[1;36m%s\033[0m", "-- Test Runnig with below crdentials & base config\n")
		fmt.Printf("* Test host - %s \n", baseConfig.hubbleHost)
		fmt.Printf("* Test project - %s \n", baseConfig.project)
		fmt.Printf("* Test key - %s \n", "***********************")
		fmt.Printf("\033[1;36m%s\033[0m", "-------------------------------\n")
	} else {
		fmt.Printf("\033[1;36m%s\033[0m", "> Since env variable not sipping integration test\n")
	}
	fmt.Printf("\033[1;36m%s\033[0m", "> Setup completed \n")
}
func IsIntegrationTestEnvSet(config Cred) (envSet bool) {
	if config.hubbleHost != "" && config.project != "" && config.apikey != "" {
		return true
	} else {
		return false
	}
}
func getEnvWithFallBack(key string) (response string) {
	value := os.Getenv(key)
	if len(value) == 0 {
		return ""
	}
	return value
}
func teardown() {
	fmt.Printf("\033[1;36m%s\033[0m", "> Teardown completed \n")
}
