package main

import (
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sathiraumesh/wireguard_exporter/wgprometheus"
	"github.com/stretchr/testify/assert"
)

// mocked default flag values that we expect for exporter
var defaultPort = flag.Int("test-default-p", DEFALULT_PORT, "the port to listen on")
var defaultInterface = flag.String("test-default-i", "", "comma-separated list of interfaces")
var defaultListenAddr = flag.String("test-default-l", "", "the address to listen on")

// mocked custom flag values that we expect for exporter
var customPort = flag.Int("test-custom-i", 8080, "the port to listen on")
var customInterface = flag.String("test-custom-p", "wg0,wg1", "comma-separated list of interfaces")
var customListenAddr = flag.String("test-custom-l", "127.0.0.1", "the address to listen on")

func TestValidatesDefaultFlags(t *testing.T) {
	interfaces, port, _ := validateReturnFlags(*defaultInterface, *defaultPort, *defaultListenAddr)

	assert.Empty(t, interfaces, "default flags for interface should be empty")

	expectedPort := ":" + strconv.Itoa(*defaultPort)
	assert.Equalf(t, port, expectedPort, "invalid default port %s", port)
}

func TestValidateCustomFlags(t *testing.T) {
	interfaces, port, _ := validateReturnFlags(*customInterface, *customPort, *defaultListenAddr)

	assert.NotEmpty(t, interfaces, "custom flags for (-i) interface should not be empty")
	assert.Equal(t, len(interfaces), 2, "invalid interface count")

	expectedPort := ":" + strconv.Itoa(*customPort)
	assert.Equalf(t, port, expectedPort, "invalid custom port %s", port)
}

func TestValidateCustomListenAddr(t *testing.T) {
	interfaces, port, _ := validateReturnFlags(*defaultInterface, *customPort, *customListenAddr)

	assert.Empty(t, interfaces, "default flags for interface should be empty")

	expectedPort := *customListenAddr + ":" + strconv.Itoa(*customPort)
	assert.Equalf(t, port, expectedPort, "invalid listen address and port combination, expected %s, got %s", expectedPort, port)
}

func TestMetricsEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test that requires WireGuard interfaces")
	}

	registry := wgprometheus.GetRegistry()

	go wgprometheus.ScrapConnectionStats([]string{}, SCRAP_INTERVAL)

	handler := promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{},
	)

	testServer := httptest.NewServer(handler)
	defer testServer.Close()

	// we wait for some time for some time until the connection stats goroutine is run
	time.Sleep(2 * time.Second)

	resp, err := http.Get(testServer.URL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	responseText := string(body)

	assert.Contains(t, responseText, `public_key="HYf+yNzgj3uhARFlNy3Pawuk/yLC+WYoY2qwjjlSxxI="`)
}

func TestBasicAuthMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap it with basic auth
	authHandler := basicAuthMiddleware(testHandler, "testuser", "testpass")

	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid credentials",
			username:       "testuser",
			password:       "testpass",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Invalid username",
			username:       "wronguser",
			password:       "testpass",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "Invalid password",
			username:       "testuser",
			password:       "wrongpass",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "Empty credentials",
			username:       "",
			password:       "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/metrics", nil)
			if tt.username != "" || tt.password != "" {
				req.SetBasicAuth(tt.username, tt.password)
			}

			rr := httptest.NewRecorder()
			authHandler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

			if tt.expectedStatus == http.StatusUnauthorized {
				assert.Equal(t, `Basic realm="Wireguard Exporter"`, rr.Header().Get("WWW-Authenticate"))
			}
		})
	}
}

func TestBasicAuthMiddlewareWithoutAuth(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()

	// Test handler without authentication middleware
	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())
}

