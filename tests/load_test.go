package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	baseURL := "http://localhost:8080"
	concurrency := 10
	requests := 100

	// Test health endpoint load
	t.Run("HealthEndpointLoad", func(t *testing.T) {
		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		start := time.Now()
		for i := 0; i < requests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := http.Get(baseURL + "/health")
				if err == nil && resp.StatusCode == 200 {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
				if resp != nil {
					resp.Body.Close()
				}
			}()

			if (i+1)%concurrency == 0 {
				time.Sleep(10 * time.Millisecond)
			}
		}

		wg.Wait()
		duration := time.Since(start)

		t.Logf("Load test completed: %d/%d successful requests in %v", successCount, requests, duration)
		assert.Greater(t, successCount, requests*8/10) // 80% success rate
	})

	// Test registration endpoint load
	t.Run("RegistrationLoad", func(t *testing.T) {
		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				reqBody := map[string]interface{}{
					"email":    fmt.Sprintf("load-test-%d@example.com", id),
					"password": "password123",
					"role":     "customer",
				}

				jsonBody, _ := json.Marshal(reqBody)
				resp, err := http.Post(baseURL+"/api/auth/register", "application/json", bytes.NewBuffer(jsonBody))

				if err == nil && (resp.StatusCode == 201 || resp.StatusCode == 400) {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
				if resp != nil {
					resp.Body.Close()
				}
			}(i)
		}

		wg.Wait()
		t.Logf("Registration load test: %d/50 requests completed", successCount)
	})
}
