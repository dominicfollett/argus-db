package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	var logBuffer bytes.Buffer
	out := io.Writer(&logBuffer)

	// test context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a test environment variable getter
	getenv := func(key string) string {
		switch key {
		case "HOST":
			return "localhost"
		case "PORT":
			return "8124"
		default:
			return ""
		}
	}

	// Start the server in a separate goroutine
	go func() {
		if err := run(ctx, getenv, out); err != nil {
			t.Errorf("Error starting server: %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	// Define the number of concurrent threads and requests per thread
	numThreads := 10
	numRequestsPerThread := 10

	// Create a wait group to synchronize the threads
	var wg sync.WaitGroup
	wg.Add(numThreads)

	// Create a slice to store the responses
	responses := make([]string, numThreads*numRequestsPerThread)

	// Mutex to protect 'responses' slice
	var mu sync.Mutex

	// Create a custom transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     30 * time.Second,
	}

	// Create an HTTP client with the custom transport
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	for i := 0; i < numThreads; i++ {
		go func(threadID int) {
			defer wg.Done()

			for j := 0; j < numRequestsPerThread; j++ {
				// Create a test request payload
				payload := limitArgs{
					Key:      fmt.Sprint("test_key_", threadID*numRequestsPerThread+j),
					Capacity: 10,
					Interval: 60,
					Unit:     "s",
				}

				jsonPayload, _ := json.Marshal(payload)

				req, _ := http.NewRequest(http.MethodPost, "http://localhost:8124/api/v1/limit", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("Error making request: %v", err)
					return
				}

				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Error reading response body: %v", err)
					return
				}
				result := string(respBody)
				resp.Body.Close()

				mu.Lock()
				responses[threadID*numRequestsPerThread+j] = result
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Check the responses
	for _, resp := range responses {
		if resp != "OK" {
			t.Errorf("Unexpected response: %s", resp)
		}
	}

	// Cancel the context to gracefully shutdown the server
	cancel()

	// Wait for the server to shutdown
	time.Sleep(2 * time.Second)

	// TODO: Check the log output for any errors or race conditions?
	// write the log output to a file?
}
