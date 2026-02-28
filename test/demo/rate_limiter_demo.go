package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	// Configuration
	baseURL := "http://localhost:5001"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	loginEndpoint := baseURL + "/auth/login"
	numRequests := 10

	fmt.Println("==================================================")
	fmt.Println("   Rate Limiter Demo")
	fmt.Println("==================================================")
	fmt.Printf("\n🎯 Target: %s\n", loginEndpoint)
	fmt.Printf("📊 Number of requests: %d\n", numRequests)
	fmt.Printf("⏱️  Rate limit: 5 requests per minute per IP\n\n")

	// Test credentials
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	fmt.Println("Starting rapid login attempts...")
	fmt.Println()

	successCount := 0
	rateLimitedCount := 0
	otherCount := 0

	for i := 1; i <= numRequests; i++ {
		fmt.Printf("Request %d: ", i)

		statusCode, body := makeLoginRequest(loginEndpoint, loginReq)

		switch statusCode {
		case 429:
			fmt.Println("❌ RATE LIMITED (429 Too Many Requests)")
			var errResp ErrorResponse
			if err := json.Unmarshal([]byte(body), &errResp); err == nil {
				fmt.Printf("   Message: %s\n", errResp.Error)
			}
			rateLimitedCount++
		case 401:
			fmt.Println("✅ REQUEST PROCESSED (401 Unauthorized - expected)")
			successCount++
		case 400:
			fmt.Println("✅ REQUEST PROCESSED (400 Bad Request)")
			successCount++
		case 0:
			fmt.Println("❌ CONNECTION FAILED - Service may not be running")
			otherCount++
		default:
			fmt.Printf("⚠️  UNEXPECTED (%d)\n", statusCode)
			otherCount++
		}

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("   Test Results")
	fmt.Println("==================================================")
	fmt.Printf("✅ Processed requests: %d\n", successCount)
	fmt.Printf("❌ Rate limited: %d\n", rateLimitedCount)
	fmt.Printf("⚠️  Other responses: %d\n", otherCount)
	fmt.Println()

	// Verify rate limiting is working
	if rateLimitedCount > 0 {
		fmt.Println("✅ SUCCESS: Rate limiting is working!")
		fmt.Println("   Expected behavior: First request allowed, subsequent rapid requests blocked.")
	} else if otherCount == numRequests {
		fmt.Println("❌ ERROR: Could not connect to service!")
		fmt.Printf("   Make sure the service is running at %s\n", baseURL)
		os.Exit(1)
	} else {
		fmt.Println("⚠️  WARNING: No rate limiting detected!")
		fmt.Println("   This might indicate:")
		fmt.Println("   1. Rate limit configuration is too permissive")
		fmt.Println("   2. Rate limiter is not properly initialized")
	}

	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("Testing rate limit recovery...")
	fmt.Println("==================================================")
	fmt.Println()
	fmt.Println("⏳ Waiting 13 seconds for rate limit to reset...")

	// Countdown
	for i := 13; i >= 1; i-- {
		fmt.Printf("%d... ", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Println()
	fmt.Println()

	fmt.Print("Request after wait: ")
	statusCode, _ := makeLoginRequest(loginEndpoint, loginReq)

	switch statusCode {
	case 429:
		fmt.Println("❌ STILL RATE LIMITED (unexpected)")
	case 401, 400:
		fmt.Println("✅ REQUEST PROCESSED (rate limit recovered)")
	case 0:
		fmt.Println("❌ CONNECTION FAILED")
	default:
		fmt.Printf("⚠️  UNEXPECTED (%d)\n", statusCode)
	}

	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("   Test Complete")
	fmt.Println("==================================================")
}

// makeLoginRequest sends a login request and returns status code and response body
func makeLoginRequest(url string, loginReq LoginRequest) (int, string) {
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return 0, ""
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, ""
	}

	return resp.StatusCode, string(body)
}
