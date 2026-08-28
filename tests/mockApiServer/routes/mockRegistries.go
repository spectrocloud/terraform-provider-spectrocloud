package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func getPackRegistryPayload() *models.V1PackRegistry {
	return &models.V1PackRegistry{
		APIVersion: "",
		Kind:       "",
		Metadata: &models.V1ObjectMeta{
			Annotations:           nil,
			CreationTimestamp:     models.V1Time{},
			DeletionTimestamp:     models.V1Time{},
			Labels:                nil,
			LastModifiedTimestamp: models.V1Time{},
			Name:                  "test-registry-name",
			UID:                   "test-registry-uid",
		},
		Spec: &models.V1PackRegistrySpec{
			Auth:     &models.V1RegistryAuth{Type: "basic"},
			Endpoint: strPtr("https://pack.example.com"),
			Name:     "test-registry-name",
			Scope:    "project",
		},
		Status: nil,
	}
}

func getHelmRegistryPayload() *models.V1HelmRegistry {
	return &models.V1HelmRegistry{
		APIVersion: "",
		Kind:       "",
		Metadata: &models.V1ObjectMeta{
			Annotations:           nil,
			CreationTimestamp:     models.V1Time{},
			DeletionTimestamp:     models.V1Time{},
			Labels:                nil,
			LastModifiedTimestamp: models.V1Time{},
			Name:                  "Public",
			UID:                   "test-registry-uid",
		},
		Spec: &models.V1HelmRegistrySpec{
			Auth: &models.V1RegistryAuth{
				Password: "test=pwd",
				TLS:      nil,
				Token:    "as",
				Type:     "token",
				Username: "sf",
			},
			Endpoint:    strPtr("test.com"),
			IsPrivate:   false,
			Name:        "Public",
			RegistryUID: generateRandomStringUID(),
			Scope:       "project",
		},
		Status: &models.V1HelmRegistryStatus{
			HelmSyncStatus: &models.V1RegistrySyncStatus{
				LastRunTime:    models.V1Time{},
				LastSyncedTime: models.V1Time{},
				Message:        "",
				Status:         "Active",
			},
		},
	}
}

// helmRegistryFixtureFor dispatches GET /v1/registries/helm/{uid} by UID,
// mirroring ecrRegistryFixtureFor/basicRegistryFixtureFor below. PLT-2356:
// added the "helm-uid-tls" case so resourceRegistryHelmRead's TLS mapping
// (registry.Spec.Auth.TLS != nil branch) can be exercised directly. The
// default branch preserves the original static payload so every pre-existing
// helm registry test keeps passing unmodified.
func helmRegistryFixtureFor(uid string) (*models.V1HelmRegistry, int) {
	switch uid {
	case "helm-uid-tls":
		r := getHelmRegistryPayload()
		r.Metadata.UID = uid
		r.Spec.Auth.TLS = &models.V1TLSConfiguration{
			Ca:                 "test-ca-pem",
			Certificate:        "test-cert-pem",
			Enabled:            true,
			InsecureSkipVerify: true,
			Key:                "test-key-pem",
		}
		return r, http.StatusOK
	default:
		return getHelmRegistryPayload(), http.StatusOK
	}
}

func helmRegistryGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	payload, status := helmRegistryFixtureFor(uid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// ecrRegistryFixtureFor dispatches GET /v1/registries/oci/{uid}/ecr by UID so
// resourceRegistryEcrRead's credential-type switch and the TLS nil/non-nil
// branch can each be exercised directly. The default branch preserves the
// original static STS payload so every pre-existing test keeps passing
// unmodified.
func ecrRegistryFixtureFor(uid string) (*models.V1EcrRegistry, int) {
	base := func() *models.V1EcrRegistry {
		return &models.V1EcrRegistry{
			Metadata: &models.V1ObjectMeta{
				Name: "testSecretRegistry",
				UID:  "testSecretRegistry-id",
			},
			Spec: &models.V1EcrRegistrySpec{
				BaseContentPath: "test-path",
				DefaultRegion:   "test-region",
				Endpoint:        strPtr("test.point"),
				IsPrivate:       boolPtr(false),
				ProviderType:    strPtr("test-type"),
				RegistryUID:     "test-reg-uid",
				Scope:           "project",
			},
			Status: &models.V1OciRegistryStatus{
				SyncStatus: &models.V1RegistrySyncStatus{
					IsSyncSupported: false,
				},
			},
		}
	}

	switch uid {
	case "ecr-uid-secret-plain":
		r := base()
		r.Spec.Credentials = &models.V1AwsCloudAccount{
			CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
			AccessKey:      "plain-access-key",
			SecretKey:      "plain-secret-key",
		}
		// TLS intentionally left nil to cover the `registry.Spec.TLS != nil` false branch.
		return r, http.StatusOK
	case "ecr-uid-secret-masked":
		r := base()
		r.Spec.Credentials = &models.V1AwsCloudAccount{
			CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
			AccessKey:      "masked-access-key",
			SecretKey:      "abc***def",
		}
		return r, http.StatusOK
	case "ecr-uid-unknown-cred":
		r := base()
		unknown := models.V1AwsCloudAccountCredentialType("weird-type")
		r.Spec.Credentials = &models.V1AwsCloudAccount{
			CredentialType: &unknown,
		}
		return r, http.StatusOK
	default:
		r := base()
		r.Spec.Credentials = &models.V1AwsCloudAccount{
			AccessKey:      "test-key",
			CredentialType: models.V1AwsCloudAccountCredentialTypeSts.Pointer(),
			Partition:      strPtr("test-part"),
			PolicyARNs:     []string{"test-arns"},
			SecretKey:      "test-secret-key",
			Sts: &models.V1AwsStsCredentials{
				Arn:        "test-arn",
				ExternalID: "test-external-id",
			},
		}
		r.Spec.TLS = &models.V1TLSConfiguration{
			Ca:                 "test-ca",
			Certificate:        "test-cert",
			Enabled:            false,
			InsecureSkipVerify: false,
			Key:                "test-key",
		}
		return r, http.StatusOK
	}
}

func ecrRegistryGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	payload, status := ecrRegistryFixtureFor(uid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// basicRegistryFixtureFor dispatches GET /v1/registries/oci/{uid}/basic by
// UID, mirroring ecrRegistryFixtureFor. The default branch keeps the
// original static zarf/basic-auth payload (now with a populated Status so
// the Read function's unconditional `registry.Status.SyncStatus` access no
// longer nil-derefs).
func basicRegistryFixtureFor(uid string) (*models.V1BasicOciRegistry, int) {
	base := func() *models.V1BasicOciRegistry {
		return &models.V1BasicOciRegistry{
			Metadata: &models.V1ObjectMeta{
				Name: "test-zarf-registry",
				UID:  "test-zarf-oci-reg-basic-uid",
			},
			Spec: &models.V1BasicOciRegistrySpec{
				Endpoint:        strPtr("https://registry.example.com"),
				BasePath:        "",
				BaseContentPath: "/",
				ProviderType:    strPtr("zarf"),
				IsSyncSupported: true,
				Auth: &models.V1RegistryAuth{
					Username: "test-username",
					Password: "test-password",
					Type:     "basic",
					TLS: &models.V1TLSConfiguration{
						Certificate:        "",
						Enabled:            false,
						InsecureSkipVerify: false,
					},
				},
				Scope: "tenant",
			},
			Status: &models.V1OciRegistryStatus{
				SyncStatus: &models.V1RegistrySyncStatus{
					IsSyncSupported: true,
				},
			},
		}
	}

	switch uid {
	case "basic-uid-noauth":
		r := base()
		r.Spec.Auth = &models.V1RegistryAuth{
			Type: "noAuth",
			TLS:  &models.V1TLSConfiguration{},
		}
		return r, http.StatusOK
	default:
		return base(), http.StatusOK
	}
}

func basicRegistryGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	payload, status := basicRegistryFixtureFor(uid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// ociRegistrySyncStatusFixtureFor dispatches the OCI ecr/basic sync-status
// endpoints by UID, giving tests direct control over
// resourceOciRegistrySyncRefreshFunc's status-mapping switch and the
// Message-vs-Status-only formatting branches in
// resourceRegistryEcrCreate/Update. The default branch preserves the
// original static "Success" fixture.
func ociRegistrySyncStatusFixtureFor(uid string) *models.V1RegistrySyncStatus {
	switch uid {
	case "oci-sync-completed-no-message":
		return &models.V1RegistrySyncStatus{IsSyncSupported: true, Status: "Completed"}
	case "oci-sync-not-supported":
		return &models.V1RegistrySyncStatus{IsSyncSupported: false}
	case "oci-sync-empty-status":
		return &models.V1RegistrySyncStatus{IsSyncSupported: true}
	case "oci-sync-failed-with-message":
		return &models.V1RegistrySyncStatus{IsSyncSupported: true, Status: "Failed", Message: "boom"}
	case "oci-sync-failed-no-message":
		return &models.V1RegistrySyncStatus{IsSyncSupported: true, Status: "Error"}
	case "oci-sync-inprogress":
		return &models.V1RegistrySyncStatus{IsSyncSupported: true, Status: "Running"}
	case "oci-sync-unknown":
		return &models.V1RegistrySyncStatus{IsSyncSupported: true, Status: "Weird"}
	default:
		return &models.V1RegistrySyncStatus{
			IsSyncSupported: true,
			Status:          "Success",
			Message:         "Registry synchronized successfully",
		}
	}
}

func ociBasicRegistrySyncStatusHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ociRegistrySyncStatusFixtureFor(uid))
}

// ociEcrRegistrySyncStatusErrorUID preserves TestWaitForOciRegistrySync_APIError's
// expectation that this specific UID 404s on the ECR sync-status endpoint —
// there was no route at all for this path before this handler was added.
const ociEcrRegistrySyncStatusErrorUID = "test-oci-registry-uid"

func ociEcrRegistrySyncStatusHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	if uid == ociEcrRegistrySyncStatusErrorUID {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ociRegistrySyncStatusFixtureFor(uid))
}

func RegistriesRoutes() []Route {
	return []Route{
		{
			Method: "PUT",
			Path:   "/v1/registries/oci/{uid}/ecr",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/registries/oci/{uid}/ecr",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method:  "GET",
			Path:    "/v1/registries/oci/{uid}/ecr",
			Handler: ecrRegistryGetHandler,
		},
		{
			Method: "POST",
			Path:   "/v1/registries/oci/ecr",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-sts-oci-reg-ecr-uid"},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/registries/oci/ecr/validate",
			Response: ResponseData{
				// The generated V1EcrRegistriesValidate client only accepts
				// 204 (see v1_ecr_registries_validate_responses.go); 200 was
				// silently untested until validateRegistryCred got direct
				// coverage.
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/registries/oci/basic/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/registries/oci/basic",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-zarf-oci-reg-basic-uid"},
			},
		},
		{
			Method:  "GET",
			Path:    "/v1/registries/oci/{uid}/basic",
			Handler: basicRegistryGetHandler,
		},
		{
			Method: "PUT",
			Path:   "/v1/registries/oci/{uid}/basic",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/registries/oci/{uid}/basic",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/oci/summary",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1OciRegistries{
					Items: []*models.V1OciRegistry{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-registry-oci",
								UID:  generateRandomStringUID(),
							},
							Spec:   nil,
							Status: nil,
						},
					},
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/registries/helm",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": generateRandomStringUID()},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/registries/helm/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/registries/helm/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/pack",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1PackRegistries{
					Items: []*models.V1PackRegistry{getPackRegistryPayload()},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/helm",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1HelmRegistries{
					Items: []*models.V1HelmRegistry{getHelmRegistryPayload()},
				},
			},
		},
		{
			Method:  "GET",
			Path:    "/v1/registries/helm/{uid}",
			Handler: helmRegistryGetHandler,
		},
		{
			Method: "GET",
			Path:   "/v1/registries/helm/{uid}/sync/status",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1RegistrySyncStatus{
					IsSyncSupported: true,
					Status:          "Success",
					Message:         "Registry synchronized successfully",
				},
			},
		},
		{
			Method:  "GET",
			Path:    "/v1/registries/oci/{uid}/basic/sync/status",
			Handler: ociBasicRegistrySyncStatusHandler,
		},
		{
			// Previously unmocked entirely (any UID 404s) — see
			// TestWaitForOciRegistrySync_APIError. Now UID-dispatched so
			// other tests can also drive the ecr sync-status success/failure
			// scenarios; ociEcrRegistrySyncStatusErrorUID keeps that
			// original 404 behavior for the one UID it relies on.
			Method:  "GET",
			Path:    "/v1/registries/oci/{uid}/ecr/sync/status",
			Handler: ociEcrRegistrySyncStatusHandler,
		},
		{
			Method: "GET",
			Path:   "/v1/registries/metadata",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1RegistriesMetadata{
					Items: []*models.V1RegistryMetadata{
						{
							IsDefault: false,
							IsPrivate: false,
							Kind:      "",
							Name:      "test-registry-name",
							Scope:     "project",
							UID:       "test-registry-uid",
						},
					},
				},
			},
		},
	}
}

func RegistriesNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/registries/helm/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getHelmRegistryPayload(),
			},
		},
	}
}
