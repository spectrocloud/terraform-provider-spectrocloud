package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Route struct {
	Path     string            `json:"path"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers,omitempty"`
	Auth     *AuthConfig       `json:"auth,omitempty"`
	Response ResponseConfig    `json:"response"`
}

type AuthConfig struct {
	Type     string `json:"type"`
	Header   string `json:"header,omitempty"`
	Key      string `json:"key,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type ResponseConfig struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

func main() {
	// Load routes from JSON file
	routes := loadRoutes("./routes.json")

	// Setup handlers based on routes
	for _, route := range routes {
		http.HandleFunc(route.Path, createHandler(route))
	}

	// Start server
	log.Println("Starting server on :8080")
	err := http.ListenAndServeTLS(":8080", "./server.crt", "./server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadRoutes(file string) []Route {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Error reading routes file: %v", err)
	}

	var routes []Route
	if err := json.Unmarshal(data, &routes); err != nil {
		log.Fatalf("Error unmarshaling routes: %v", err)
	}
	return routes
}

func createHandler(route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != route.Method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check authentication if required
		if route.Auth != nil && !checkAuth(route.Auth, r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Set headers
		for key, value := range route.Headers {
			w.Header().Set(key, value)
		}

		// Handle dynamic project_uid replacement
		responseBody := string(route.Response.Body)
		if strings.Contains(responseBody, "{{project_uid}}") {
			projectUID := r.Header.Get("project_uid")
			responseBody = strings.ReplaceAll(responseBody, "{{project_uid}}", projectUID)
		}

		w.WriteHeader(route.Response.Status)
		_, err := w.Write([]byte(responseBody))
		if err != nil {
			return
		}
	}
}

func checkAuth(auth *AuthConfig, r *http.Request) bool {
	switch auth.Type {
	case "apikey":
		return r.Header.Get(auth.Header) == auth.Key
	case "basic":
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Basic ") {
			payload, err := base64.StdEncoding.DecodeString(authHeader[len("Basic "):])
			if err == nil {
				pair := strings.SplitN(string(payload), ":", 2)
				if len(pair) == 2 && pair[0] == auth.Username && pair[1] == auth.Password {
					return true
				}
			}
		}
	}
	return false
}
