package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"mockApiServer/routes"
	"net/http"
)

//// ResponseData defines the structure of mock responses
//type ResponseData struct {
//	StatusCode int
//	Payload    interface{}
//}
//
//// Route defines a mock route with method, path, and response
//type Route struct {
//	Method   string
//	Path     string
//	Response ResponseData
//}

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
	router8080 := mux.NewRouter()
	router8888 := mux.NewRouter()

	// Set up routes for port 8080
	setupRoutes(router8080, allRoutesPositive)

	// Set up routes for port 8888
	setupRoutes(router8888, allRoutesNegative)

	// Start servers on different ports
	go func() {
		log.Println("Starting server on :8080...")
		if err := http.ListenAndServeTLS(":8080", "mock_server.crt", "mock_server.key", router8080); err != nil {
			log.Fatalf("Server failed to start on port 8080: %v", err)
		}
	}()

	log.Println("Starting server on :8888...")
	//if err := http.ListenAndServe(":8888", router8888); err != nil {
	//	log.Fatalf("Server failed to start on port 8888: %v", err)
	//}
	if err := http.ListenAndServeTLS(":8888", "mock_server.crt", "mock_server.key", router8888); err != nil {
		log.Fatalf("Server failed to start on port 8080: %v", err)
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

func init() {
	// Initialize routes for port 8080
	allRoutesPositive = append(allRoutesPositive, routes.ProjectRoutes()...)
	allRoutesPositive = append(allRoutesPositive, routes.CommonProjectRoutes()...)

	// Initialize routes for port 8888
	allRoutesNegative = append(allRoutesNegative, routes.ProjectNegativeRoutes()...)
	allRoutesNegative = append(allRoutesNegative, routes.CommonProjectRoutes()...)
}
