package main

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBasicAuthMiddleware(t *testing.T) {
	// Create a test handler that just returns 200
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name               string
		username           string
		password           string
		authHeader         string
		expectedStatusCode int
		expectAuthHeader   bool
	}{
		{
			name:               "No auth configured, no auth header",
			username:           "",
			password:           "",
			authHeader:         "",
			expectedStatusCode: http.StatusOK,
			expectAuthHeader:   false,
		},
		{
			name:               "No auth configured, with auth header",
			username:           "",
			password:           "",
			authHeader:         "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass")),
			expectedStatusCode: http.StatusOK,
			expectAuthHeader:   false,
		},
		{
			name:               "Auth configured, correct credentials",
			username:           "testuser",
			password:           "testpass",
			authHeader:         "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:testpass")),
			expectedStatusCode: http.StatusOK,
			expectAuthHeader:   false,
		},
		{
			name:               "Auth configured, wrong password",
			username:           "testuser",
			password:           "testpass",
			authHeader:         "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:wrongpass")),
			expectedStatusCode: http.StatusUnauthorized,
			expectAuthHeader:   true,
		},
		{
			name:               "Auth configured, wrong username",
			username:           "testuser",
			password:           "testpass",
			authHeader:         "Basic " + base64.StdEncoding.EncodeToString([]byte("wronguser:testpass")),
			expectedStatusCode: http.StatusUnauthorized,
			expectAuthHeader:   true,
		},
		{
			name:               "Auth configured, no auth header",
			username:           "testuser",
			password:           "testpass",
			authHeader:         "",
			expectedStatusCode: http.StatusUnauthorized,
			expectAuthHeader:   true,
		},
		{
			name:               "Auth configured, malformed auth header",
			username:           "testuser",
			password:           "testpass",
			authHeader:         "Basic invalid",
			expectedStatusCode: http.StatusUnauthorized,
			expectAuthHeader:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the basic auth credentials
			basicAuthUsername = tt.username
			basicAuthPassword = tt.password

			// Create the middleware-wrapped handler
			handler := basicAuthMiddleware(testHandler)

			// Create a test request
			req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler(rr, req)

			// Check the status code
			if rr.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, rr.Code)
			}

			// Check if WWW-Authenticate header is set when expected
			wwwAuth := rr.Header().Get("WWW-Authenticate")
			if tt.expectAuthHeader && wwwAuth == "" {
				t.Error("Expected WWW-Authenticate header, but it was not set")
			}
			if !tt.expectAuthHeader && wwwAuth != "" {
				t.Errorf("Did not expect WWW-Authenticate header, but got: %s", wwwAuth)
			}
		})
	}
}

func TestWebhookHandlerWithBasicAuth(t *testing.T) {
	// Save original values and restore after test
	origUsername := basicAuthUsername
	origPassword := basicAuthPassword
	origRedisClient := redisClient
	defer func() {
		basicAuthUsername = origUsername
		basicAuthPassword = origPassword
		redisClient = origRedisClient
	}()

	// Set test credentials
	basicAuthUsername = "webhookuser"
	basicAuthPassword = "webhookpass"
	redisClient = nil // Disable Redis for this test

	// Create the middleware-wrapped handler
	handler := basicAuthMiddleware(webhookHandler)

	tests := []struct {
		name               string
		method             string
		authHeader         string
		body               string
		expectedStatusCode int
	}{
		{
			name:               "Valid request with correct auth",
			method:             http.MethodPost,
			authHeader:         "Basic " + base64.StdEncoding.EncodeToString([]byte("webhookuser:webhookpass")),
			body:               `{"type": "transaction.created", "data": {}}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Valid request with wrong auth",
			method:             http.MethodPost,
			authHeader:         "Basic " + base64.StdEncoding.EncodeToString([]byte("webhookuser:wrongpass")),
			body:               `{"type": "transaction.created", "data": {}}`,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "Valid request with no auth",
			method:             http.MethodPost,
			authHeader:         "",
			body:               `{"type": "transaction.created", "data": {}}`,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "GET request with correct auth (should fail method check)",
			method:             http.MethodGet,
			authHeader:         "Basic " + base64.StdEncoding.EncodeToString([]byte("webhookuser:webhookpass")),
			body:               "",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/webhook", bytes.NewBufferString(tt.body))
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()
			handler(rr, req)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, rr.Code)
			}
		})
	}
}

func TestWebhookHandlerWithoutBasicAuth(t *testing.T) {
	// Save original values and restore after test
	origUsername := basicAuthUsername
	origPassword := basicAuthPassword
	origRedisClient := redisClient
	defer func() {
		basicAuthUsername = origUsername
		basicAuthPassword = origPassword
		redisClient = origRedisClient
	}()

	// Disable basic auth
	basicAuthUsername = ""
	basicAuthPassword = ""
	redisClient = nil // Disable Redis for this test

	// Create the middleware-wrapped handler
	handler := basicAuthMiddleware(webhookHandler)

	tests := []struct {
		name               string
		authHeader         string
		body               string
		expectedStatusCode int
	}{
		{
			name:               "Request without auth header should succeed",
			authHeader:         "",
			body:               `{"type": "transaction.created", "data": {}}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Request with auth header should also succeed",
			authHeader:         "Basic " + base64.StdEncoding.EncodeToString([]byte("anyuser:anypass")),
			body:               `{"type": "transaction.created", "data": {}}`,
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(tt.body))
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler(rr, req)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, rr.Code)
			}
		})
	}
}

func TestMainBasicAuthEnvVarLoading(t *testing.T) {
	tests := []struct {
		name             string
		envUsername      string
		envPassword      string
		expectedUsername string
		expectedPassword string
		description      string
	}{
		{
			name:             "Both credentials set",
			envUsername:      "testuser",
			envPassword:      "testpass",
			expectedUsername: "testuser",
			expectedPassword: "testpass",
			description:      "Should enable basic auth",
		},
		{
			name:             "Neither credential set",
			envUsername:      "",
			envPassword:      "",
			expectedUsername: "",
			expectedPassword: "",
			description:      "Should disable basic auth",
		},
		{
			name:             "Only username set",
			envUsername:      "testuser",
			envPassword:      "",
			expectedUsername: "",
			expectedPassword: "",
			description:      "Should disable basic auth (partial config)",
		},
		{
			name:             "Only password set",
			envUsername:      "",
			envPassword:      "testpass",
			expectedUsername: "",
			expectedPassword: "",
			description:      "Should disable basic auth (partial config)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origUsername := os.Getenv("WEBHOOK_USERNAME")
			origPassword := os.Getenv("WEBHOOK_PASSWORD")
			defer func() {
				os.Setenv("WEBHOOK_USERNAME", origUsername)
				os.Setenv("WEBHOOK_PASSWORD", origPassword)
			}()

			// Set test env vars
			if tt.envUsername != "" {
				os.Setenv("WEBHOOK_USERNAME", tt.envUsername)
			} else {
				os.Unsetenv("WEBHOOK_USERNAME")
			}
			if tt.envPassword != "" {
				os.Setenv("WEBHOOK_PASSWORD", tt.envPassword)
			} else {
				os.Unsetenv("WEBHOOK_PASSWORD")
			}

			// Simulate the env var loading logic from main()
			username := os.Getenv("WEBHOOK_USERNAME")
			password := os.Getenv("WEBHOOK_PASSWORD")

			// Apply the same logic as in main()
			if username != "" && password != "" {
				// Both set - keep them
			} else if username != "" || password != "" {
				// Partially configured - clear both
				username = ""
				password = ""
			}
			// else both empty - keep them empty

			if username != tt.expectedUsername {
				t.Errorf("Expected username '%s', got '%s'", tt.expectedUsername, username)
			}
			if password != tt.expectedPassword {
				t.Errorf("Expected password '%s', got '%s'", tt.expectedPassword, password)
			}
		})
	}
}
