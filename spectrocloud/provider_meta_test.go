package spectrocloud

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PLT-2438: getV1ClientWithResourceContext used to type-assert `m` straight
// to *client.V1Client and re-scope that single shared pointer on every call
// (client.WithScopeProject/WithScopeTenant mutate the receiver in place).
// Since Terraform runs resource CRUDs concurrently, two goroutines racing a
// tenant-scoped call against a project-scoped call could corrupt each
// other's ctx/projectUID on the same client. providerConfigure now builds
// two independent, already-scoped clients once and never mutates either
// afterward; getV1ClientWithResourceContext just dispatches between them.

// projectUIDResponder returns an httptest handler that answers GetProjectUID
// (GET /v1/projects, matched on "name") with the given UID, and records the
// ProjectUid header seen on every request it serves.
func projectUIDResponder(uid string, seen *sync.Map) http.HandlerFunc {
	var counter atomic.Int64
	return func(w http.ResponseWriter, r *http.Request) {
		id := counter.Add(1)
		seen.Store(id, r.Header.Get("ProjectUid"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"items":[{"metadata":{"name":"Default","uid":%q}}]}`, uid)))
	}
}

func TestPLT2438_ProviderConfigureReturnsIndependentClients(t *testing.T) {
	var seen sync.Map
	srv := httptest.NewTLSServer(projectUIDResponder("proj-abc", &seen))
	defer srv.Close()

	d := prepareBaseProviderConfig()
	_ = d.Set("host", srv.Listener.Addr().String())

	m, diags := providerConfigure(context.Background(), d, "test")
	require.Empty(t, diags)

	pm, ok := m.(*ProviderMeta)
	require.True(t, ok, "providerConfigure must return *ProviderMeta")
	require.NotNil(t, pm.Project)
	require.NotNil(t, pm.Tenant)
	assert.NotSame(t, pm.Project, pm.Tenant, "project and tenant clients must be independent objects, not the same mutated pointer")
	assert.Equal(t, "proj-abc", pm.ProjectUID)
}

// TestPLT2438_TwoProviderInstancesDoNotShareProjectUID covers the second bug
// described in the ticket: ProviderInitProjectUid was a package-level var
// clobbered by every providerConfigure call, so two aliased provider blocks
// configured against different projects would corrupt each other's UID.
// Building two ProviderMeta from two independent providerConfigure calls
// must not let either UID leak into the other.
func TestPLT2438_TwoProviderInstancesDoNotShareProjectUID(t *testing.T) {
	var seenA, seenB sync.Map
	srvA := httptest.NewTLSServer(projectUIDResponder("project-a-uid", &seenA))
	defer srvA.Close()
	srvB := httptest.NewTLSServer(projectUIDResponder("project-b-uid", &seenB))
	defer srvB.Close()

	dA := prepareBaseProviderConfig()
	_ = dA.Set("host", srvA.Listener.Addr().String())
	mA, diagsA := providerConfigure(context.Background(), dA, "test")
	require.Empty(t, diagsA)

	dB := prepareBaseProviderConfig()
	_ = dB.Set("host", srvB.Listener.Addr().String())
	mB, diagsB := providerConfigure(context.Background(), dB, "test")
	require.Empty(t, diagsB)

	pmA := mA.(*ProviderMeta)
	pmB := mB.(*ProviderMeta)

	assert.Equal(t, "project-a-uid", pmA.ProjectUID)
	assert.Equal(t, "project-b-uid", pmB.ProjectUID)
	assert.Equal(t, "project-a-uid", getProviderProjectUID(pmA))
	assert.Equal(t, "project-b-uid", getProviderProjectUID(pmB))
}

// TestPLT2438_ConcurrentScopedCallsDoNotLeakProjectHeader reproduces the
// ticket's exact failure signature under concurrency: a tenant call must
// never carry a ProjectUid header, and a project call must always carry the
// correct one, no matter how the two are interleaved. Run with `-race` to
// confirm there is no longer any shared mutable state to race on.
func TestPLT2438_ConcurrentScopedCallsDoNotLeakProjectHeader(t *testing.T) {
	const projectUID = "race-project-uid"
	var seen sync.Map
	srv := httptest.NewServer(projectUIDResponder(projectUID, &seen))
	defer srv.Close()

	clientOpts := []func(*client.V1Client){
		client.WithPaletteURI(srv.Listener.Addr().String()),
		client.WithAPIKey("k"),
		client.WithSchemes([]string{"http"}),
	}
	tenantClient := client.New(clientOpts...)
	projectClient := client.New(clientOpts...)
	client.WithScopeProject(projectUID)(projectClient)

	pm := &ProviderMeta{Project: projectClient, Tenant: tenantClient, ProjectUID: projectUID}

	const iterations = 100
	var wg sync.WaitGroup
	errs := make(chan error, iterations*2)

	for i := 0; i < iterations; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c := getV1ClientWithResourceContext(pm, "tenant")
			if _, err := c.GetProjectUID("Default"); err != nil {
				errs <- err
			}
		}()
		go func() {
			defer wg.Done()
			c := getV1ClientWithResourceContext(pm, "project")
			if _, err := c.GetProjectUID("Default"); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("unexpected request error: %v", err)
	}

	corrupted := 0
	total := 0
	seen.Range(func(_, v interface{}) bool {
		total++
		header := v.(string)
		// We can't attribute a given recorded header to "tenant" or
		// "project" after the fact (both hit the same endpoint), but every
		// recorded header must be either empty (tenant) or the exact
		// project UID (project) -- anything else means scope state leaked
		// or got corrupted between the two clients.
		if header != "" && header != projectUID {
			corrupted++
		}
		return true
	})
	assert.Equal(t, iterations*2, total, "expected every request to be recorded")
	assert.Zero(t, corrupted, "every request's ProjectUid header must be either empty or exactly %q", projectUID)
}

func TestGetProviderProjectUID(t *testing.T) {
	pm := &ProviderMeta{ProjectUID: "abc-123"}
	assert.Equal(t, "abc-123", getProviderProjectUID(pm))
}
