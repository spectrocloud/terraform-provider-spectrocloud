package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"mockApiServer/routes"
	"net/http"
)

// ResponseData defines the structure of mock responses
type ResponseData struct {
	StatusCode int
	Payload    interface{}
}

// Route defines a mock route with method, path, and response
type Route struct {
	Method   string
	Path     string
	Response ResponseData
}

// API key for authentication
const apiKey = "12345"

// Aggregate all routes into a single slice
var allRoutes []routes.Route

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
	router := mux.NewRouter()

	// Apply API key middleware to all routes
	router.Use(apiKeyAuthMiddleware)
	setAllRoutes()
	// Register all routes
	for _, route := range allRoutes {
		route := route // capture the range variable

		router.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(route.Response.StatusCode)
			if route.Response.Payload != nil {
				json.NewEncoder(w).Encode(route.Response.Payload)
			}
		}).Methods(route.Method)
	}

	// Start the server
	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServeTLS(":8080", "mock_server.crt", "mock_server.key", router))
}

func setAllRoutes() {
	allRoutes = append(allRoutes, routes.ProjectRoutes()...)
}
