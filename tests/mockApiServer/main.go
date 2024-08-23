package mockApiServer

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
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

// Define userRoutes as a separate slice
var userRoutes = []Route{
	{
		Method: "GET",
		Path:   "/api/v1/users",
		Response: ResponseData{
			StatusCode: http.StatusOK,
			Payload: []map[string]interface{}{
				{"id": 1, "name": "John Doe"},
				{"id": 2, "name": "Jane Doe"},
			},
		},
	},
	{
		Method: "POST",
		Path:   "/api/v1/users",
		Response: ResponseData{
			StatusCode: http.StatusCreated,
			Payload: map[string]interface{}{
				"id":   3,
				"name": "New User",
			},
		},
	},
	{
		Method: "GET",
		Path:   "/api/v1/users/{userId}",
		Response: ResponseData{
			StatusCode: http.StatusOK,
			Payload: map[string]interface{}{
				"id":   1,
				"name": "John Doe",
			},
		},
	},
	{
		Method: "PUT",
		Path:   "/api/v1/users/{userId}",
		Response: ResponseData{
			StatusCode: http.StatusOK,
			Payload: map[string]interface{}{
				"id":   1,
				"name": "Updated User",
			},
		},
	},
	{
		Method: "DELETE",
		Path:   "/api/v1/users/{userId}",
		Response: ResponseData{
			StatusCode: http.StatusNoContent,
			Payload:    nil,
		},
	},
}

// Aggregate all routes into a single slice
var allRoutes = append(userRoutes)

// Middleware to check for the API key in the header
func apiKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ApiKey") != apiKey {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	router := mux.NewRouter()

	// Apply API key middleware to all routes
	router.Use(apiKeyAuthMiddleware)

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
	log.Fatal(http.ListenAndServe(":8080", router))
}
