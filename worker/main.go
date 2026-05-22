package worker

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

type Job struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func main() {
	ctx := context.Background()

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

		log.Printf("processing job id=%s type=%s", job.ID, job.Type)

		delay := time.Duration(rand.Intn(2000)) * time.Millisecond
		time.Sleep(delay)

		log.Printf("completed job id=%s duration=%s", job.ID, delay)
	}
}
