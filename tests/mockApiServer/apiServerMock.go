package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// API key for authentication
const apiKey = "12345"

// Aggregate all routes into slices for different servers
var allRoutesPositive []routes.Route
var allRoutesNegative []routes.Route

// Middleware to check for the API key and log the Project-ID if present
func apiKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ApiKey") != apiKey {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		// Log the Project-ID if it is present
		if projectID := r.Header.Get("Project-ID"); projectID != "" {
			log.Printf("Project-ID: %s", projectID)
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Create routers for different ports
	router8088 := mux.NewRouter()
	router8888 := mux.NewRouter()

	// Set up routes for port 8088
	setupRoutes(router8088, allRoutesPositive)

	// Set up routes for port 8888
	setupRoutes(router8888, allRoutesNegative)

	// Start servers on different ports
	go func() {
		log.Println("Starting server on :8088...")
		server := &http.Server{
			Addr:         ":8088",
			Handler:      router8088,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		if err := server.ListenAndServeTLS("mock_server.crt", "mock_server.key"); err != nil {
			log.Fatalf("Server failed to start on port 8088: %v", err)
		}
	}()

	log.Println("Starting server on :8888...")

	server := &http.Server{
		Addr:         ":8888",
		Handler:      router8888,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if err := server.ListenAndServeTLS("mock_server.crt", "mock_server.key"); err != nil {
		log.Fatalf("Server failed to start on port 8888: %v", err)
	}
}

// setupRoutes configures the given router with the provided routes
func setupRoutes(router *mux.Router, routes []routes.Route) {
	// Apply API key middleware to all routes
	router.Use(apiKeyAuthMiddleware)

	// Register all routes
	for _, route := range routes {
		route := route // capture the range variable

		router.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(route.Response.StatusCode)
			if route.Response.Payload != nil {
				err := json.NewEncoder(w).Encode(route.Response.Payload)
				if err != nil {
					return
				}
			}
		}).Methods(route.Method)
	}
}

func aggregateRoutes(routeFuncs ...func() []routes.Route) []routes.Route {
	var aggregatedRoutes []routes.Route
	for _, routeFunc := range routeFuncs {
		aggregatedRoutes = append(aggregatedRoutes, routeFunc()...)
	}
	return aggregatedRoutes
}

func init() {
	// Initialize routes for port 8088
	allRoutesPositive = aggregateRoutes(
		routes.CommonProjectRoutes,
		routes.ProjectRoutes,
		routes.AppliancesRoutes,
		routes.UserRoutes,
		routes.FilterRoutes,
		routes.RolesRoutes,
		routes.RegistriesRoutes,
		routes.PacksRoutes,
		routes.ClusterProfileRoutes,
		routes.CloudAccountsRoutes,
		routes.ClusterCommonRoutes,
		routes.ClusterRoutes,
		routes.AppProfilesRoutes,
		routes.TeamRoutes,
		routes.ApplicationRoutes,
		routes.BackupRoutes,
		routes.IPPoolRoutes,
		routes.MacrosRoutes,
		routes.TenantRoutes,
		routes.WorkspaceRoutes,
		routes.AlertRoutes,
		routes.ClusterGroupRoutes,
		routes.ClusterConfigTemplateRoutes,
		routes.ClusterConfigPolicyRoutes,
		routes.DeveloperSettingRoutes,
	)
	// Initialize routes for port 8888
	allRoutesNegative = aggregateRoutes(
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
		routes.MacrosNegativeRoutes,
		routes.TenantNegativeRoutes,
		routes.WorkspaceNegativeRoutes,
	)
}
