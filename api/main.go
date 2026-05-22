package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

var redisClient = redis.NewClient(&redis.Options{
	Addr: "redis:6379",
})

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"method", "path", "status"},
	)
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

type Job struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		// Generate requestID - this will track requests from API gateway -> worker servcice -> DB -> ML services
		requestID := uuid.New().String()
		// Set response header BEFORE handler runs
		w.Header().Set("X-Request-ID", requestID)

		// The real handler runs here: /work, /healthz, /metrics, etc.
		next.ServeHTTP(recorder, r)

		// Now the handler is done, so recorder.statusCode is known.
		duration := time.Since(start)
		status := strconv.Itoa(recorder.statusCode)

		requestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Inc()

		requestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Observe(duration.Seconds())

		log.Printf(
			"method=%s path=%s status=%s duration=%s",
			r.Method,
			r.URL.Path,
			status,
			duration,
		)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func workHandler(w http.ResponseWriter, r *http.Request) {
	delay := time.Duration(rand.Intn(500)) * time.Millisecond
	time.Sleep(delay)

	if rand.Intn(10) == 0 {
		http.Error(w, "simulated error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "work completed in %s\n", delay)
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	job := Job{
		ID:   uuid.New().String(),
		Type: "process_work",
	}

	payload, err := json.Marshal(job)
	if err != nil {
		http.Error(w, "failed to marshal job", http.StatusInternalServerError)
		return
	}

	err = redisClient.RPush(ctx, "jobs", payload).Err()
	if err != nil {
		http.Error(w, "failed to enqueue job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(map[string]string{
		"job_id": job.ID,
		"status": "queued",
	})
}

func main() {
	prometheus.MustRegister(
		requestsTotal,
		requestDuration,
		errorsTotal,
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/work", workHandler)
	mux.HandleFunc("/jobs", jobsHandler)
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":8080",
		Handler: loggingMiddleware(mux),
	}

	log.Println("server running on :8080")
	log.Fatal(server.ListenAndServe())
}
