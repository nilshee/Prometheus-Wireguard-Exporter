package main

import (
	"crypto/subtle"
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sathiraumesh/wireguard_exporter/wgprometheus"
)

const SCRAP_INTERVAL = time.Second * 5
const DEFALULT_PORT = 9011

var port = flag.Int("p", DEFALULT_PORT, "the port to listen on")
var listenAddr = flag.String("l", "", "the address to listen on (default: all interfaces)")
var interfaces = flag.String("i", "", "comma-separated list of interfaces")
var authUser = flag.String("auth-user", "", "basic auth username (optional)")
var authPass = flag.String("auth-pass", "", "basic auth password (optional)")

func main() {

	flag.Parse()

	interfaces, port, _ := validateReturnFlags(*interfaces, *port, *listenAddr)

	registry := wgprometheus.GetRegistry()

	go wgprometheus.ScrapConnectionStats(interfaces, SCRAP_INTERVAL)

	metricsHandler := promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{},
	)

	// Wrap with basic auth if credentials are provided
	if *authUser != "" && *authPass != "" {
		log.Printf("Basic authentication enabled for user: %s", *authUser)
		http.Handle("/metrics", basicAuthMiddleware(metricsHandler, *authUser, *authPass))
	} else {
		log.Println("Basic authentication disabled")
		http.Handle("/metrics", metricsHandler)
	}

	log.Printf("Starting Wireguard exporter on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// basicAuthMiddleware wraps an http.Handler with basic authentication
func basicAuthMiddleware(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		// Use constant-time comparison to prevent timing attacks
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Wireguard Exporter"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized\n"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateReturnFlags(interfaceArg string, portArg int, listenAddrArg string) (interfaces []string, port string, configPath string) {

	if strings.TrimSpace(listenAddrArg) != "" {
		port = listenAddrArg + ":" + strconv.Itoa(portArg)
	} else {
		port = ":" + strconv.Itoa(portArg)
	}

	if strings.TrimSpace(interfaceArg) != "" {
		interfaces = strings.Split(interfaceArg, ",")
	}

	return
}
