package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTP duration histogram
var httpRequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP requests in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	},
	[]string{"path"},
)

var httpRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "HTTP requests processed",
	},
	[]string{"path", "status"},
)

type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	// first WriteHeader call determines status
	if rw.wroteHeader {
		return
	}

	rw.statusCode = code
	rw.wroteHeader = true

	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// if no status has been written yet get 200
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}

	return rw.ResponseWriter.Write(b)
}

// Prometheus metrics
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		statusCodeStr := strconv.Itoa(rw.statusCode)

		// avoid creating a separate Prometheus times for every task ID
		path := normalizePath(r.URL.Path)

		httpRequestDuration.WithLabelValues(path).Observe(duration)
		httpRequestsTotal.WithLabelValues(path, statusCodeStr).Inc()
	})
}

func normalizePath(path string) string {
	if path == "/tasks" || strings.HasPrefix(path, "/tasks/") {
		return "/tasks"
	}

	return path
}

func main() {
	port := getEnv("PORT", "8080")

	// Run internal health check and exit
	isHealthcheck := flag.Bool(
		"healthcheck",
		false,
		"Run internal health check query",
	)
	flag.Parse()

	if *isHealthcheck {
		runHealthcheck(port)
		return
	}

	store := NewMemoryStore()

	mux := http.NewServeMux()

	// must respond 200
	mux.HandleFunc("GET /healthz", HealthHandler)

	// Metrics — application metrics+Prometheus
	oldMetricsFunc := MetricsHandler(store)

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		oldMetricsFunc(w, r)

		// do not return compressed output
		r.Header.Del("Accept-Encoding")

		rec := httptest.NewRecorder()
		promhttp.Handler().ServeHTTP(rec, r)

		_, _ = w.Write([]byte("\n"))
		_, _ = w.Write(rec.Body.Bytes())
	})

	// Task CRUD.
	mux.HandleFunc("GET /tasks", ListTasksHandler(store))
	mux.HandleFunc("POST /tasks", CreateTaskHandler(store))
	mux.HandleFunc("GET /tasks/{id}", GetTaskHandler(store))
	mux.HandleFunc("PUT /tasks/{id}", UpdateTaskHandler(store))
	mux.HandleFunc("DELETE /tasks/{id}", DeleteTaskHandler(store))

	addr := ":" + port

	log.Printf("task-api starting on %s", addr)

	if err := http.ListenAndServe(
		addr,
		prometheusMiddleware(mux),
	); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func runHealthcheck(port string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://localhost:%s/healthz", port)

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Health check failed: HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("Health check OK")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
