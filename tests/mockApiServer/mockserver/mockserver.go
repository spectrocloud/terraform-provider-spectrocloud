// Package mockserver runs the terraform-provider-spectrocloud mock Palette
// API in-process. It is imported by:
//
//   - spectrocloud/common_test.go — from TestMain, so unit tests get a fresh
//     server on every `go test` invocation without shelling out to a script
//     or building a separate binary.
//   - tests/mockApiServer/apiServerMock.go — thin main() wrapper that keeps
//     the standalone binary form available for manual/integration use.
//
// The server generates a self-signed TLS certificate at startup (no openssl
// dependency, no committed cert files needed), then serves the positive
// route set on :8088 and the negative route set on :8888 — matching the
// ports the tests have historically hard-coded in common_test.go.
package mockserver

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// APIKey is the fixed key the mock accepts. Any request whose ApiKey header
// does not match receives HTTP 403.
const APIKey = "12345"

// PositivePort and NegativePort are the fixed loopback ports. Tests dial
// 127.0.0.1:<port> directly, so these must not change without updating
// spectrocloud/common_test.go in lockstep.
const (
	PositivePort = 8088
	NegativePort = 8888
)

// Server groups the two HTTPS listeners so callers can shut them down as a
// unit. Stop is safe to call at most once and is a no-op after the first
// call (subsequent errors from Shutdown are ignored).
type Server struct {
	positive *http.Server
	negative *http.Server
}

// Stop shuts down both servers. It blocks until in-flight requests complete
// or the 5-second grace deadline elapses, whichever comes first.
func (s *Server) Stop() {
	if s == nil {
		return
	}
	// Best-effort shutdown; ignore errors — we're tearing down a test process.
	stop := func(srv *http.Server) {
		if srv == nil {
			return
		}
		_ = srv.Close()
	}
	stop(s.positive)
	stop(s.negative)
}

// Start boots both mock servers and returns once they're accepting
// connections. It aggregates the default positive+negative route sets from
// the routes package; call StartWith to inject a custom aggregation.
//
// If either listener fails to bind (e.g. the previous test run leaked the
// process), Start returns an error and cleans up whichever half succeeded.
func Start() (*Server, error) {
	return StartWith(DefaultPositiveRoutes(), DefaultNegativeRoutes())
}

// StartWith is Start with caller-supplied route slices. Exposed so tests
// that want to override behavior (e.g. verify a Handler-based dynamic
// response) can layer routes without editing the aggregation in init().
func StartWith(positive, negative []routes.Route) (*Server, error) {
	cert, err := generateSelfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("mockserver: generate cert: %w", err)
	}

	pos, err := listenAndServe(PositivePort, positive, cert)
	if err != nil {
		return nil, fmt.Errorf("mockserver: start positive server: %w", err)
	}

	neg, err := listenAndServe(NegativePort, negative, cert)
	if err != nil {
		// Clean up the half we already started.
		_ = pos.Close()
		return nil, fmt.Errorf("mockserver: start negative server: %w", err)
	}

	return &Server{positive: pos, negative: neg}, nil
}

// listenAndServe binds a TLS listener on 127.0.0.1:<port> and starts serving
// in a goroutine. Returning after the bind (rather than after Serve) means
// callers can rely on the port being ready for connections the moment Start
// returns — no health-check retry loop required.
func listenAndServe(port int, routeSet []routes.Route, cert tls.Certificate) (*http.Server, error) {
	router := mux.NewRouter()
	router.Use(apiKeyAuthMiddleware)
	registerRoutes(router, routeSet)

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}

	tlsLn := tls.NewListener(ln, srv.TLSConfig)
	go func() {
		if err := srv.Serve(tlsLn); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// The test binary has no useful place to surface this — log to stderr.
			fmt.Printf("mockserver on :%d exited: %v\n", port, err)
		}
	}()

	return srv, nil
}

// apiKeyAuthMiddleware mirrors the check the previous standalone binary
// performed: any request without the fixed ApiKey header receives 403.
// The Project-ID logging that used to live here has been dropped — it
// produced noise in test output and nothing depended on it.
func apiKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ApiKey") != APIKey {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// registerRoutes wires each Route into the router. When Route.Handler is
// set the route uses it directly; otherwise the fallback writes
// Response.StatusCode + JSON-encoded Response.Payload — same behavior the
// old apiServerMock.go had before the Handler field existed.
func registerRoutes(router *mux.Router, routeSet []routes.Route) {
	for _, route := range routeSet {
		route := route // capture range variable

		if route.Handler != nil {
			router.HandleFunc(route.Path, route.Handler).Methods(route.Method)
			continue
		}

		router.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(route.Response.StatusCode)
			if route.Response.Payload != nil {
				_ = json.NewEncoder(w).Encode(route.Response.Payload)
			}
		}).Methods(route.Method)
	}
}

// aggregate flattens a set of route-provider functions into a single slice.
func aggregate(routeFuncs ...func() []routes.Route) []routes.Route {
	var out []routes.Route
	for _, fn := range routeFuncs {
		out = append(out, fn()...)
	}
	return out
}

// DefaultPositiveRoutes returns the happy-path route set served on :8088.
// Keep this list in sync with new mock*.go files added under routes/ — the
// standalone apiServerMock.go also consumes it, so both paths stay
// consistent.
func DefaultPositiveRoutes() []routes.Route {
	return aggregate(
		routes.CommonProjectRoutes,
		routes.ProjectRoutes,
		routes.AppliancesRoutes,
		routes.TenantRoutes,
		routes.UserRoutes,
		routes.FilterRoutes,
		routes.RolesRoutes,
		routes.RegistriesRoutes,
		routes.PacksRoutes,
		routes.ClusterProfileRoutes,
		routes.CloudAccountsRoutes,
		routes.ClusterCommonRoutes,
		routes.AksClusterRoutes,
		routes.AzureClusterRoutes,
		routes.EksClusterRoutes,
		routes.ClusterRoutes,
		routes.CustomCloudClusterRoutes,
		routes.AwsClusterRoutes,
		routes.GcpClusterRoutes,
		routes.GkeClusterRoutes,
		routes.VsphereClusterRoutes,
		routes.MaasClusterRoutes,
		routes.EdgeNativeClusterRoutes,
		routes.CloudStackClusterRoutes,
		routes.VirtualClusterRoutes,
		routes.PlatformSettingRoutes,
		routes.UserRolesRoutes,
		routes.SSORoutes,
		routes.AppProfilesRoutes,
		routes.TeamRoutes,
		routes.ApplicationRoutes,
		routes.BackupRoutes,
		routes.IPPoolRoutes,
		routes.MacrosRoutes,
		routes.WorkspaceRoutes,
		routes.AlertRoutes,
		routes.ClusterGroupRoutes,
		routes.ClusterConfigTemplateRoutes,
		routes.ClusterConfigPolicyRoutes,
		routes.SSHKeyRoutes,
		routes.PCGDNSMapRoutes,
		routes.KubevirtVMRoutes,
		routes.RegistrationTokenRoutes,
		routes.PasswordPolicyRoutes,
		routes.DeveloperSettingRoutes,
		routes.ResourceLimitRoutes,
		routes.AuditTrailRoutes,
	)
}

// DefaultNegativeRoutes returns the error-path route set served on :8888.
func DefaultNegativeRoutes() []routes.Route {
	return aggregate(
		routes.CommonProjectRoutes,
		routes.ProjectNegativeRoutes,
		routes.AppliancesNegativeRoutes,
		routes.UserNegativeRoutes,
		routes.FilterNegativeRoutes,
		routes.RolesNegativeRoutes,
		routes.RegistriesNegativeRoutes,
		routes.PacksNegativeRoutes,
		routes.ClusterProfileNegativeRoutes,
		routes.CloudAccountsNegativeRoutes,
		routes.ClusterCommonNegativeRoutes,
		routes.ClusterNegativeRoutes,
		routes.MacrosNegativeRoutes,
		routes.TenantNegativeRoutes,
		routes.WorkspaceNegativeRoutes,
		routes.SSHKeyNegativeRoutes,
		routes.RegistrationTokenNegativeRoutes,
		routes.PasswordPolicyNegativeRoutes,
		routes.DeveloperSettingNegativeRoutes,
		routes.ResourceLimitNegativeRoutes,
		routes.AuditTrailNegativeRoutes,
	)
}

// generateSelfSignedCert produces a short-lived ECDSA cert usable for
// 127.0.0.1 TLS. Regenerating on every process start (rather than reading
// mock_server.crt / mock_server.key from disk) removes the openssl
// dependency and eliminates the "stale cert" failure mode the shell script
// used to hit when the cert expired between runs.
func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "spectrocloud-mock"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.IPv6loopback},
		DNSNames:     []string{"localhost"},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	return tls.X509KeyPair(certPEM, keyPEM)
}
