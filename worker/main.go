package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

type Job struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Attempts int    `json:"attempts"`
}

var (
	jobsProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_processed_total",
			Help: "Total number of jobs processed by the worker",
		},
		[]string{"type", "status"},
	)

	jobDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "job_duration_seconds",
			Help:    "Job processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type", "status"},
	)
)

const maxAttempts = 3

func main() {
	ctx := context.Background()

	prometheus.MustRegister(jobsProcessedTotal, jobDuration)

	go startMetricsServer()

	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	log.Println("worker started, waiting for jobs...")

	for {
		result, err := redisClient.BLPop(ctx, 0*time.Second, "jobs").Result()
		if err != nil {
			log.Printf("failed to pop job: %v", err)
			continue
		}

		rawJob := result[1]

		var job Job
		if err := json.Unmarshal([]byte(rawJob), &job); err != nil {
			log.Printf("failed to unmarshal job: %v", err)
			continue
		}

		processJob(ctx, redisClient, job)
	}
}

func startMetricsServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	log.Println("worker metrics server running on :8081")

	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("worker metrics server failed: %v", err)
	}
}

func processJob(ctx context.Context, redisClient *redis.Client, job Job) {
	start := time.Now()

	log.Printf(
		"processing job id=%s type=%s attempts=%d",
		job.ID,
		job.Type,
		job.Attempts,
	)

	delay := time.Duration(rand.Intn(2000)) * time.Millisecond
	time.Sleep(delay)

	if rand.Intn(5) == 0 {
		job.Attempts++

		if job.Attempts >= maxAttempts {
			status := "dead_lettered"

			jobsProcessedTotal.WithLabelValues(job.Type, status).Inc()
			jobDuration.WithLabelValues(job.Type, status).Observe(time.Since(start).Seconds())

			payload, err := json.Marshal(job)
			if err != nil {
				log.Printf("failed to marshal dead-letter job id=%s err=%v", job.ID, err)
				return
			}

			if err := redisClient.RPush(ctx, "dead_jobs", payload).Err(); err != nil {
				log.Printf("failed to dead-letter job id=%s err=%v", job.ID, err)
				return
			}

			log.Printf(
				"dead-lettered job id=%s type=%s attempts=%d duration=%s",
				job.ID,
				job.Type,
				job.Attempts,
				delay,
			)

			return
		}

		status := "retried"

		jobsProcessedTotal.WithLabelValues(job.Type, status).Inc()
		jobDuration.WithLabelValues(job.Type, status).Observe(time.Since(start).Seconds())

		payload, err := json.Marshal(job)
		if err != nil {
			log.Printf("failed to marshal retry job id=%s err=%v", job.ID, err)
			return
		}

		if err := redisClient.RPush(ctx, "jobs", payload).Err(); err != nil {
			log.Printf("failed to requeue job id=%s err=%v", job.ID, err)
			return
		}

		log.Printf(
			"requeued job id=%s type=%s attempts=%d duration=%s",
			job.ID,
			job.Type,
			job.Attempts,
			delay,
		)

		return
	}

	status := "success"

	jobsProcessedTotal.WithLabelValues(job.Type, status).Inc()
	jobDuration.WithLabelValues(job.Type, status).Observe(time.Since(start).Seconds())

	log.Printf(
		"completed job id=%s type=%s attempts=%d duration=%s",
		job.ID,
		job.Type,
		job.Attempts,
		delay,
	)
}
