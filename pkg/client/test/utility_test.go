package test

import (
	"fmt"
	"os"
	"testing"
)

type Cred struct {
	hubbleHost string
	email      string
	project    string
	apikey     string
	pwd        string
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
	baseConfig.hubbleHost = os.Getenv("TEST_HOST")
	baseConfig.email = os.Getenv("TEST_EMAIL")
	baseConfig.pwd = os.Getenv("TEST_PWD")
	baseConfig.project = os.Getenv("TEST_PROJECT")
	baseConfig.apikey = os.Getenv("TEST_API_KEY")
	baseConfig.component = "ClusterHealth"
	baseConfig.AlertUid = ""
	fmt.Printf("\033[1;36m%s\033[0m", "> Credentials & Base config setup completed\n")
	fmt.Printf("\033[1;36m%s\033[0m", "-- Test Runnig with below crdentials & base config\n")
	fmt.Printf("* Test host - %s \n", baseConfig.hubbleHost)
	fmt.Printf("* Test email - %s \n", baseConfig.email)
	fmt.Printf("* Test pwd - %s \n", baseConfig.pwd)
	fmt.Printf("* Test project - %s \n", baseConfig.project)
	fmt.Printf("\033[1;36m%s\033[0m", "-------------------------------\n")
}

func teardown() {
	fmt.Printf("\033[1;36m%s\033[0m", "> Teardown completed \n")
}
