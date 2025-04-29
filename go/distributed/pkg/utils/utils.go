package utils

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// Task represents a job to be processed
type Task struct {
	ID       string
	Type     string
	Payload  []byte
	Deadline time.Time
}

// TaskResult represents the result of a processed task
type TaskResult struct {
	TaskID         string
	Success        bool
	Result         []byte
	ErrorMessage   string
	CompletionTime time.Time
}

// GenerateRandomID creates a random ID for tasks
func GenerateRandomID(prefix string) string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s-%d", prefix, rand.Intn(1000000))
}

// SimulateWork simulates processing time for a task
func SimulateWork(taskType string) ([]byte, error) {
	// Simulate different processing times based on task type
	var processingTime time.Duration

	switch taskType {
	case "fast":
		processingTime = time.Duration(rand.Intn(500)) * time.Millisecond
	case "medium":
		processingTime = time.Duration(rand.Intn(1000)+500) * time.Millisecond
	case "slow":
		processingTime = time.Duration(rand.Intn(2000)+1000) * time.Millisecond
	default:
		processingTime = time.Duration(rand.Intn(1000)) * time.Millisecond
	}

	log.Printf("Processing task of type '%s' for %v", taskType, processingTime)
	time.Sleep(processingTime)

	// 10% chance of failure for realistic simulation
	if rand.Float32() < 0.1 {
		return nil, fmt.Errorf("random task failure occurred")
	}

	// Generate some dummy result data
	resultData := []byte(fmt.Sprintf("Result data for task type: %s", taskType))
	return resultData, nil
}

// GetLocalIP returns a string representation of the server address with port
func GetServerAddress(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
