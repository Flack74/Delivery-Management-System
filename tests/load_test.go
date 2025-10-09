package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
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

	t.Run("HealthEndpointLoad", func(t *testing.T) {
		testHealthEndpoint(t, baseURL, requests, concurrency)
	})

	t.Run("RegistrationLoad", func(t *testing.T) {
		testRegistrationEndpoint(t, baseURL)
	})
}

func testHealthEndpoint(t *testing.T, baseURL string, requests, concurrency int) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 50,
			IdleConnTimeout:     30 * time.Second,
		},
	}
	
	var wg sync.WaitGroup
	var successCount int32

	start := time.Now()
	semaphore := make(chan struct{}, concurrency)
	
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			resp, err := client.Get(baseURL + "/health")
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == 200 {
					atomic.AddInt32(&successCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	finalCount := int(atomic.LoadInt32(&successCount))
	t.Logf("Load test completed: %d/%d successful requests in %v", finalCount, requests, duration)
	assert.Greater(t, finalCount, requests*8/10)
}

func testRegistrationEndpoint(t *testing.T, baseURL string) {
	var wg sync.WaitGroup
	var successCount int32

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			reqBody := map[string]interface{}{
				"email":    fmt.Sprintf("load-test-%d@example.com", id),
				"password": "password123",
				"role":     "customer",
			}

			jsonBody, err := json.Marshal(reqBody)
			if err != nil {
				return
			}
			
			resp, err := http.Post(baseURL+"/api/auth/register", "application/json", bytes.NewBuffer(jsonBody))
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == 201 || resp.StatusCode == 400 {
					atomic.AddInt32(&successCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	finalCount := int(atomic.LoadInt32(&successCount))
	t.Logf("Registration load test: %d/50 requests completed", finalCount)
}