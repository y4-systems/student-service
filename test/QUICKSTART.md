# Quick Start Guide - Rate Limiter Tests

## 🚀 Instant Test Commands

### 1. Run Unit Tests (Fastest)
```bash
cd /workspaces/student-service
go test -v -short ./test
```

**Expected:** All tests pass in < 1 second

### 2. Run with Coverage
```bash
go test -v -short -cover ./test
```

**Expected:** Shows test coverage percentage

### 3. Run Benchmarks
```bash
go test -bench=. -benchmem ./test
```

**Expected:** Performance metrics (~77 ns/op for GetLimiter, ~118 ns/op for Allow)

### 4. Interactive Test Runner
```bash
./test/run_tests.sh
```

**Expected:** Menu with all test options

---

## 🧪 Testing with Running Service

### Start the Service
```bash
# Terminal 1
cd /workspaces/student-service
go run main.go
```

### Run Demo (Go)
```bash
# Terminal 2
cd /workspaces/student-service
go run test/demo/rate_limiter_demo.go
```

### Run Demo (Bash)
```bash
# Terminal 2
./test/test_rate_limiter.sh
```

---

## 📊 Expected Results

### Unit Tests Output
```
=== RUN   Test_NewIPRateLimiter
--- PASS: Test_NewIPRateLimiter (0.00s)
=== RUN   Test_GetLimiter
--- PASS: Test_GetLimiter (0.00s)
=== RUN   Test_RateLimitEnforcement
--- PASS: Test_RateLimitEnforcement (0.00s)
=== RUN   Test_MultipleIPAddresses
--- PASS: Test_MultipleIPAddresses (0.00s)
=== RUN   Test_ConcurrentAccess
--- PASS: Test_ConcurrentAccess (0.00s)
=== RUN   Test_GetIPAddress_XForwardedFor
--- PASS: Test_GetIPAddress_XForwardedFor (0.00s)
=== RUN   Test_GetIPAddress_XRealIP
--- PASS: Test_GetIPAddress_XRealIP (0.00s)
=== RUN   Test_GetIPAddress_RemoteAddr
--- PASS: Test_GetIPAddress_RemoteAddr (0.00s)
=== RUN   Test_GetIPAddress_Priority
--- PASS: Test_GetIPAddress_Priority (0.00s)
PASS
ok      github.com/y4-systems/student-service/test      0.006s
```




## 🎯 What Gets Tested

✅ **Rate limiter initialization**  
✅ **IP-based limiting** (different IPs get separate limits)  
✅ **Rate enforcement** (blocks rapid requests)  
✅ **IP extraction** (from headers: X-Forwarded-For, X-Real-IP, RemoteAddr)  
✅ **Thread safety** (concurrent access)  
✅ **Rate recovery** (limits reset after time)  
✅ **Performance** (< 120 ns per operation)  

---