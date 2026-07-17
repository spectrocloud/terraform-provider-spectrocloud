package spectrocloud

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/mockserver"
	"github.com/stretchr/testify/assert"
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
	host          = "127.0.0.1:8088"
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

// mockAPIServer is the in-process mock started by TestMain. It replaces the
// previous shell-script-driven MockBuild process — no openssl, no nohup,
// no /tmp binary to clean up.
var mockAPIServer *mockserver.Server

func TestMain(m *testing.M) {
	// Batch 22 — zero out the retry.StateChangeConf.Delay across every
	// wait loop in the provider (see wait_delay.go). Without this the
	// initial 30-second Delay would block Create-path tests for the
	// full test-binary timeout, hiding coverage of the post-wait
	// branches. Setting the override before m.Run means every test
	// goroutine that reads it sees the zeroed value.
	zero := time.Duration(0)
	waitDelayOverride = &zero

	srv, err := mockserver.Start()
	if err != nil {
		fmt.Printf("Error starting mock API server: %v\n", err)
		os.Exit(1)
	}
	mockAPIServer = srv
	fmt.Printf("\033[1;36m%s\033[0m", "> Started Mock Api Server at https://127.0.0.1:8088 (positive) & https://127.0.0.1:8888 (negative)\n")

	ctx := context.Background()
	unitTestMockAPIClient, _ = unitTestProviderConfigure(ctx)
	unitTestMockAPINegativeClient, _ = unitTestNegativeCaseProviderConfigure(ctx)
	fmt.Printf("\033[1;36m%s\033[0m", "> Setup completed \n")

	code := m.Run()

	mockAPIServer.Stop()
	fmt.Printf("\033[1;36m%s\033[0m", "> Stopped Mock Api Server \n")
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

func assertFirstDiagMessage(t *testing.T, diags diag.Diagnostics, msg string) {
	if assert.NotEmpty(t, diags, "Expected diags to contain at least one element") {
		assert.Contains(t, diags[0].Summary, msg, "The first diagnostic message does not contain the expected error message")
	}
}

// assertAnyDiagContains passes if any diag's Summary OR Detail contains msg.
// Prefer this over assertFirstDiagMessage when the diag order isn't stable —
// e.g. a Create that fans out and reports two independent failures.
func assertAnyDiagContains(t *testing.T, diags diag.Diagnostics, msg string) {
	if !assert.NotEmpty(t, diags, "Expected diags to contain at least one element") {
		return
	}
	for _, d := range diags {
		if strings.Contains(d.Summary, msg) || strings.Contains(d.Detail, msg) {
			return
		}
	}
	// Build a readable failure that surfaces every diag we saw, not just diags[0].
	var got []string
	for _, d := range diags {
		got = append(got, fmt.Sprintf("{summary=%q detail=%q}", d.Summary, d.Detail))
	}
	t.Errorf("no diag matched %q; got %s", msg, strings.Join(got, ", "))
}

// resourceCRUDFunc is the signature of resource Create/Read/Update/Delete functions.
type resourceCRUDFunc func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics

// testResourceCRUD runs a full CRUD cycle (Create -> Read -> Update -> Read -> Delete) and asserts no diagnostics.
func testResourceCRUD(t *testing.T, prepareData func() *schema.ResourceData, meta interface{}, create, read, update, delete resourceCRUDFunc) {
	ctx := context.Background()
	d := prepareData()

	diags := create(ctx, d, meta)
	assert.Empty(t, diags, "Create should not return diagnostics")
	assert.NotEmpty(t, d.Id(), "Create should set resource ID")

	diags = read(ctx, d, meta)
	assert.Empty(t, diags, "Read should not return diagnostics")

	diags = update(ctx, d, meta)
	assert.Empty(t, diags, "Update should not return diagnostics")

	diags = read(ctx, d, meta)
	assert.Empty(t, diags, "Read after Update should not return diagnostics")

	diags = delete(ctx, d, meta)
	assert.Empty(t, diags, "Delete should not return diagnostics")
}

// testResourceCreateReadDelete runs Create -> Read -> Delete without the Update
// leg. Use it for resources where Update is a no-op, ForceNew-only, or where
// the mock's Update endpoint doesn't yet round-trip cleanly. Keeping this
// separate from testResourceCRUD means new coverage doesn't have to fake an
// Update path just to satisfy the harness.
func testResourceCreateReadDelete(t *testing.T, prepareData func() *schema.ResourceData, meta interface{},
	create, read, delete resourceCRUDFunc) {
	ctx := context.Background()
	d := prepareData()

	diags := create(ctx, d, meta)
	assert.Empty(t, diags, "Create should not return diagnostics")
	assert.NotEmpty(t, d.Id(), "Create should set resource ID")

	diags = read(ctx, d, meta)
	assert.Empty(t, diags, "Read should not return diagnostics")

	diags = delete(ctx, d, meta)
	assert.Empty(t, diags, "Delete should not return diagnostics")
}

// testResourceCRUDReturningDiags runs the full Create/Read/Update/Read/Delete
// cycle and returns the diags from every step so the caller can make
// finer-grained assertions — e.g. "Update returned a Warning but no Error".
// testResourceCRUD is a strict wrapper around this; pick this one when you
// need to inspect diags instead of just failing on non-empty.
func testResourceCRUDReturningDiags(t *testing.T, prepareData func() *schema.ResourceData, meta interface{},
	create, read, update, delete resourceCRUDFunc) (createDiags, readDiags, updateDiags, readAfterUpdateDiags, deleteDiags diag.Diagnostics) {
	ctx := context.Background()
	d := prepareData()

	createDiags = create(ctx, d, meta)
	readDiags = read(ctx, d, meta)
	updateDiags = update(ctx, d, meta)
	readAfterUpdateDiags = read(ctx, d, meta)
	deleteDiags = delete(ctx, d, meta)
	return
}

// testResourceCRUDNegative runs one CRUD op with negative client and asserts diags contain msgSubstr.
func testResourceCRUDNegative(t *testing.T, op string, prepare func() *schema.ResourceData, meta interface{},
	create, read, update, delete resourceCRUDFunc, setID bool, msgSubstr string) {
	ctx := context.Background()
	d := prepare()
	if setID {
		d.SetId("12763471256725")
	}
	var diags diag.Diagnostics
	switch op {
	case "Create":
		diags = create(ctx, d, meta)
	case "Read":
		diags = read(ctx, d, meta)
	case "Update":
		diags = update(ctx, d, meta)
	case "Delete":
		diags = delete(ctx, d, meta)
	default:
		t.Fatalf("unknown op %s", op)
	}
	if len(diags) == 0 {
		t.Errorf("expected diagnostics containing %q", msgSubstr)
		return
	}
	if !strings.Contains(diags[0].Summary, msgSubstr) {
		t.Errorf("diag summary %q does not contain %q", diags[0].Summary, msgSubstr)
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
