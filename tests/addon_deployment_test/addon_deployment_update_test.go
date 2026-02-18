package addon_deployment

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func TestUpdateAddonDeploymentIsNotAttached(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch &&
			strings.HasPrefix(r.URL.Path, "/v1/spectroclusters/") &&
			strings.HasSuffix(r.URL.Path, "/profiles") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}

	h := client.New(
		client.WithPaletteURI(u.Host),
		client.WithSchemes([]string{u.Scheme}),
	)

	// Create mock cluster
	cluster := &models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{
			UID:         "test-cluster",
			Annotations: map[string]string{"scope": "project"},
		},
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID:  "test-profile-uid",
					Name: "test-profile-name",
				},
			},
		},
	}

	// Create mock body
	body := &models.V1SpectroClusterProfiles{
		Profiles: []*models.V1SpectroClusterProfileEntity{
			{UID: "test-profile"},
		},
	}

	// Create mock newProfile
	newProfile := &models.V1ClusterProfile{
		Metadata: &models.V1ObjectMeta{
			UID: "new-test-profile-uid",
		},
	}

	// Call UpdateAddonDeployment
	err = h.UpdateAddonDeployment(cluster, body, newProfile)

	// Assert there was no error
	assert.NoError(t, err)
}

func TestUpdateAddonDeploymentIsAttached(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch &&
			strings.HasPrefix(r.URL.Path, "/v1/spectroclusters/") &&
			strings.HasSuffix(r.URL.Path, "/profiles") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}

	h := client.New(
		client.WithPaletteURI(u.Host),
		client.WithSchemes([]string{u.Scheme}),
	)

	// Create mock cluster
	cluster := &models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{
			UID:         "test-cluster",
			Annotations: map[string]string{"scope": "tenant"},
		},
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID:  "test-profile-uid",
					Name: "test-profile-name",
				},
			},
		},
	}

	// Create mock body
	body := &models.V1SpectroClusterProfiles{
		Profiles: []*models.V1SpectroClusterProfileEntity{
			{UID: "test-profile"},
		},
	}

	// Create mock newProfile
	newProfile := &models.V1ClusterProfile{
		Metadata: &models.V1ObjectMeta{
			Name: "test-profile-name",
		},
	}

	// Call UpdateAddonDeployment
	err = h.UpdateAddonDeployment(cluster, body, newProfile)

	// Assert there was no error
	assert.NoError(t, err)
}
