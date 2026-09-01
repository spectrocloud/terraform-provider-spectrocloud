package spectrocloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-openapi/runtime"
)

// PE-9338 — the Palette management API (hubble) blocks two classes of Kubernetes
// version changes with an HTTP 400 carrying a stable machine `code` in a
// `{code, message, data}` envelope:
//
//	K8sMultiMinorUpgradeNotSupported     — target skips one or more minor versions
//	K8sDowngradeAfterUpgradeNotSupported — target reverts a successful upgrade
//
// The provider is a CONSUMER of this contract, not the enforcer (SD D-7): the
// server remains the hard gate. Its job is to make the block LEGIBLE. Today the
// generated palette-sdk-go response reader (v1_spectro_clusters_patch_profiles_responses.go)
// defines only the 204 case; every other status funnels through
// `runtime.NewAPIError(..., response, response.Code())`. `runtime.APIError.Error()`
// marshals the `ClientResponse` interface — which exposes no serializable fields —
// so the server body is dropped and the user sees only `"<op> (status 400): {}"`.
// Because the message text never reaches `err.Error()`, the repo's usual
// `strings.Contains(err.Error(), ...)` idiom cannot work here; we must decode the
// unread response body ourselves and branch on the machine `code`.
const (
	codeK8sMultiMinorUpgradeNotSupported     = "K8sMultiMinorUpgradeNotSupported"
	codeK8sDowngradeAfterUpgradeNotSupported = "K8sDowngradeAfterUpgradeNotSupported"
)

// clusterVersionErrorEnvelope mirrors the hubble/herr error body. Consumers branch
// on `Code` (stable). The structured payload lives under `data`; the SDK may
// alternatively surface it under `details`, so we tolerate both (see mgmt vs edge
// framework split, api-contracts.md Contract 1 / "two-shape reality").
type clusterVersionErrorEnvelope struct {
	Code    string                   `json:"code"`
	Message string                   `json:"message"`
	Data    clusterVersionErrorData  `json:"data"`
	Details *clusterVersionErrorData `json:"details,omitempty"`
}

type clusterVersionErrorData struct {
	CurrentVersion string      `json:"currentVersion"`
	TargetVersion  string      `json:"targetVersion"`
	CurrentMinor   json.Number `json:"currentMinor"`
	TargetMinor    json.Number `json:"targetMinor"`
}

// payload returns the structured payload, preferring `data` and falling back to
// `details` (edge/stylus framework) when `data` is empty.
func (e clusterVersionErrorEnvelope) payload() clusterVersionErrorData {
	if e.Data != (clusterVersionErrorData{}) {
		return e.Data
	}
	if e.Details != nil {
		return *e.Details
	}
	return e.Data
}

// mapClusterProfileUpdateErr recognizes the PE-9338 version-block 400s coming back
// from a cluster profile update and returns a clear, actionable Terraform
// diagnostic that distinguishes the two machine codes. Any error it does not
// recognize (nil, non-APIError, non-400, unknown/empty code, unreadable body) is
// returned UNCHANGED — the mapping is forward-safe and never masks other failures.
func mapClusterProfileUpdateErr(err error) error {
	if err == nil {
		return nil
	}

	var apiErr *runtime.APIError
	if !errors.As(err, &apiErr) {
		return err
	}
	if apiErr.Code != 400 {
		return err
	}

	env, ok := extractClusterVersionErrorEnvelope(apiErr)
	if !ok {
		return err
	}

	switch env.Code {
	case codeK8sMultiMinorUpgradeNotSupported:
		return errors.New(formatMultiMinorMessage(env))
	case codeK8sDowngradeAfterUpgradeNotSupported:
		return errors.New(formatDowngradeMessage(env))
	default:
		// A 400 that isn't one of our two version blocks — leave it untouched.
		return err
	}
}

// extractClusterVersionErrorEnvelope reads the (still-unconsumed) response body off
// the opaque *runtime.APIError and decodes the {code, message, data} envelope. The
// generated reader returns the error at its `default:` branch WITHOUT reading the
// body, so the body is available here. If a future SDK regen changes that, decoding
// simply fails and the caller falls back to the raw error (fail-safe, OQ-5).
func extractClusterVersionErrorEnvelope(apiErr *runtime.APIError) (clusterVersionErrorEnvelope, bool) {
	var env clusterVersionErrorEnvelope

	resp, ok := apiErr.Response.(runtime.ClientResponse)
	if !ok {
		return env, false
	}
	body := resp.Body()
	if body == nil {
		return env, false
	}
	defer func() { _ = body.Close() }()

	raw, readErr := io.ReadAll(body)
	if readErr != nil || len(raw) == 0 {
		return env, false
	}
	if jsonErr := json.Unmarshal(raw, &env); jsonErr != nil {
		return env, false
	}
	if env.Code == "" {
		return env, false
	}
	return env, true
}

// formatMultiMinorMessage builds the multi-minor-upgrade diagnostic. It echoes the
// server-localized message (authoritative copy) and appends sequential-upgrade
// guidance plus the current/target versions when the server supplied them.
func formatMultiMinorMessage(env clusterVersionErrorEnvelope) string {
	serverMsg := strings.TrimSpace(env.Message)
	if serverMsg == "" {
		serverMsg = "Kubernetes upgrades across multiple minor versions are not supported. " +
			"Please update your cluster profile to sequentially upgrade the Kubernetes pack across each minor version."
	}
	var b strings.Builder
	b.WriteString("Kubernetes multi-minor upgrade blocked (")
	b.WriteString(codeK8sMultiMinorUpgradeNotSupported)
	b.WriteString("): ")
	b.WriteString(serverMsg)
	b.WriteString(versionSuffix(env))
	b.WriteString(" Upgrade the Kubernetes version one minor version at a time.")
	return b.String()
}

// formatDowngradeMessage builds the post-upgrade-downgrade diagnostic.
func formatDowngradeMessage(env clusterVersionErrorEnvelope) string {
	serverMsg := strings.TrimSpace(env.Message)
	if serverMsg == "" {
		serverMsg = "Kubernetes downgrades are not supported after a successful upgrade. " +
			"Please update your cluster profile to a Kubernetes version equal to or newer than the currently running version."
	}
	var b strings.Builder
	b.WriteString("Kubernetes downgrade blocked (")
	b.WriteString(codeK8sDowngradeAfterUpgradeNotSupported)
	b.WriteString("): ")
	b.WriteString(serverMsg)
	b.WriteString(versionSuffix(env))
	b.WriteString(" Set the Kubernetes version to one equal to or newer than the currently running version.")
	return b.String()
}

// versionSuffix renders " (current: X, target: Y)" when both versions are present,
// so the diagnostic is actionable without the operator digging through server logs.
func versionSuffix(env clusterVersionErrorEnvelope) string {
	p := env.payload()
	cur := strings.TrimSpace(p.CurrentVersion)
	tgt := strings.TrimSpace(p.TargetVersion)
	if cur == "" && tgt == "" {
		return ""
	}
	return fmt.Sprintf(" (current: %s, target: %s)", fallbackDash(cur), fallbackDash(tgt))
}

func fallbackDash(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}
