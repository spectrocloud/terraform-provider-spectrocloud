package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// MockHardenedImagesClusterCount is the number of synthetic clusters the mock tenant reports.
// It is deliberately larger than the 50-cluster API page size so that the provider's paging loop
// is exercised across more than one request.
const MockHardenedImagesClusterCount = 60

// hardenedImagesPageSize mirrors the maximum page size the real status API accepts.
const hardenedImagesPageSize = 50

// hardenedImagesState holds the exclusion flags the mock tenant currently has set, so that a PUT
// to /v1/spectroclusters/imagePullSecret/clusters is visible to the following GET of
// /v1/spectroclusters/imagePullSecret/status. Without that round trip the resource's
// create-then-read cycle could not be verified.
var hardenedImagesState = struct {
	sync.Mutex
	excluded map[string]bool
}{excluded: map[string]bool{}}

// ResetHardenedImagesState clears the recorded exclusions. Tests call it so that one test's
// mutations do not leak into the next.
func ResetHardenedImagesState() {
	hardenedImagesState.Lock()
	defer hardenedImagesState.Unlock()
	hardenedImagesState.excluded = map[string]bool{}
}

// MockHardenedImagesClusterUID returns the synthetic UID of the nth cluster, counting from 1.
func MockHardenedImagesClusterUID(n int) string {
	return fmt.Sprintf("test-hardened-cluster-%d", n)
}

// hardenedImagesClusterStatus builds one row of the tenant rollout report. Every third cluster is
// reported as Failed so the failure reason and message fields have non-empty coverage.
func hardenedImagesClusterStatus(n int, excluded bool) *models.V1ImagePullSecretTenantPropagationClusterStatus {
	status := &models.V1ImagePullSecretTenantPropagationClusterStatus{
		Cluster: &models.V1ObjectReference{
			UID:  MockHardenedImagesClusterUID(n),
			Name: fmt.Sprintf("hardened-cluster-%d", n),
		},
		Project: &models.V1ObjectReference{
			UID:  "test-project-uid",
			Name: "Default",
		},
		Exclude: excluded,
		State:   models.V1ImagePullSecretPropagationStateCompleted,
	}
	if n%3 == 0 {
		status.State = models.V1ImagePullSecretPropagationStateFailed
		status.Reason = "ConnectivityIssue"
		status.Message = "cluster agent is unreachable"
	}
	return status
}

// hardenedImagesStatusHandler serves GET /v1/spectroclusters/imagePullSecret/status. It honours
// the limit and offset query parameters so the provider's paging loop runs for real.
func hardenedImagesStatusHandler(w http.ResponseWriter, r *http.Request) {
	limit := hardenedImagesPageSize
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= hardenedImagesPageSize {
			limit = parsed
		}
	}
	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			offset = parsed
		}
	}

	hardenedImagesState.Lock()
	excluded := make(map[string]bool, len(hardenedImagesState.excluded))
	for uid, v := range hardenedImagesState.excluded {
		excluded[uid] = v
	}
	hardenedImagesState.Unlock()

	items := make([]*models.V1ImagePullSecretTenantPropagationClusterStatus, 0, limit)
	var failed int64
	for n := 1; n <= MockHardenedImagesClusterCount; n++ {
		uid := MockHardenedImagesClusterUID(n)
		status := hardenedImagesClusterStatus(n, excluded[uid])
		if status.State == models.V1ImagePullSecretPropagationStateFailed {
			failed++
		}
		if n <= offset || len(items) >= limit {
			continue
		}
		items = append(items, status)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&models.V1ImagePullSecretTenantPropagationStatus{
		State:    models.V1ImagePullSecretPropagationStateInProgress,
		Clusters: &models.V1ImagePullSecretTenantPropagationClusterCounts{Failed: failed},
		Items:    items,
		Listmeta: &models.V1ListMetaData{
			Count:  int64(MockHardenedImagesClusterCount),
			Limit:  int64(limit),
			Offset: int64(offset),
		},
	})
}

// hardenedImagesClustersUpdateHandler serves PUT /v1/spectroclusters/imagePullSecret/clusters,
// recording the exclusion flags so the next status read reflects them.
func hardenedImagesClustersUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var body models.V1ImagePullSecretPropagationClustersUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(getError(strconv.Itoa(http.StatusBadRequest), "malformed image pull secret cluster update"))
		return
	}
	if body.Exclude == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(getError(strconv.Itoa(http.StatusBadRequest), "exclude is required"))
		return
	}

	hardenedImagesState.Lock()
	for _, uid := range body.ClusterUids {
		if *body.Exclude {
			hardenedImagesState.excluded[uid] = true
		} else {
			delete(hardenedImagesState.excluded, uid)
		}
	}
	hardenedImagesState.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// HardenedImagesRoutes wires the two endpoints that resource_hardened_images.go and
// data_source_hardened_images.go touch:
//   - /v1/spectroclusters/imagePullSecret/status (GET)
//   - /v1/spectroclusters/imagePullSecret/clusters (PUT)
func HardenedImagesRoutes() []Route {
	return []Route{
		{
			Method:  "GET",
			Path:    "/v1/spectroclusters/imagePullSecret/status",
			Handler: hardenedImagesStatusHandler,
		},
		{
			Method:  "PUT",
			Path:    "/v1/spectroclusters/imagePullSecret/clusters",
			Handler: hardenedImagesClustersUpdateHandler,
		},
	}
}

// HardenedImagesNegativeRoutes returns error responses on both paths so the resource's error
// branches can be exercised.
func HardenedImagesNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/imagePullSecret/status",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid hardened images status request"),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/spectroclusters/imagePullSecret/clusters",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid hardened images cluster update"),
			},
		},
	}
}
