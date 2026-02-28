#!/bin/bash

# Quick Test Runner
# Provides easy access to all rate limiter tests

echo "=================================================="
echo "   Rate Limiter Test Runner"
echo "=================================================="
echo ""
echo "Choose a test to run:"
echo ""
echo "  1. Unit Tests (fast)"
echo "  2. Unit Tests with Coverage"
echo "  3. All Tests (including time-dependent)"
echo "  4. Performance Benchmarks"
echo "  5. Interactive Demo (requires running service)"
echo "  6. Bash Script Test (requires running service)"
echo "  7. Run All Tests"
echo ""
echo "  0. Exit"
echo ""
read -p "Enter choice [0-7]: " choice

case $choice in
    1)
        echo ""
        echo "Running fast unit tests..."
        go test -v -short ./test
        ;;
    2)
        echo ""
        echo "Running unit tests with coverage..."
        go test -v -short -cover ./test
        ;;
    3)
        echo ""
        echo "Running all tests (this may take time)..."
        go test -v ./test
        ;;
    4)
        echo ""
        echo "Running performance benchmarks..."
        go test -bench=. -benchmem ./test
        ;;
    5)
        echo ""
        echo "Starting interactive demo..."
        echo "Make sure the service is running on port 5001!"
        echo ""
        read -p "Press Enter to continue or Ctrl+C to cancel..."
        go run test/demo/rate_limiter_demo.go
        ;;
    6)
        echo ""
        echo "Running bash script test..."
        echo "Make sure the service is running on port 5001!"
        echo ""
        read -p "Press Enter to continue or Ctrl+C to cancel..."
        ./test/test_rate_limiter.sh
        ;;
    7)
        echo ""
        echo "Running all tests..."
        echo ""
        echo "1. Unit tests..."
        go test -v -short ./test
        echo ""
        echo "2. Benchmarks..."
        go test -bench=. ./test
        echo ""
        echo "All tests complete!"
        ;;
    0)
        echo "Exiting..."
        exit 0
        ;;
    *)
        echo "Invalid choice!"
        exit 1
        ;;
esac

echo ""
echo "=================================================="
echo "   Test Complete"
echo "=================================================="
