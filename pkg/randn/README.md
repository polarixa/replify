# randn

**randn** is a Go library providing convenient utilities for generating random numbers, UUIDs, unique identifiers, and random data. It offers both simple random generation for general use and cryptographically secure generation for security-sensitive applications.

## Overview

The `randn` package simplifies random data generation in Go by providing:

- **UUID Generation**: Standard and custom-delimited UUIDs using `/dev/urandom`
- **Unique IDs**: Alphanumeric IDs, crypto-secure IDs, and time-based IDs
- **Random Numbers**: Integers, floats (32/64-bit), and ranged random values
- **Random Bytes**: Generate random byte arrays for any purpose

**Problem Solved:** Go's standard `math/rand` package requires manual seeding and verbose setup. The `randn` package provides pre-configured, ready-to-use functions with sensible defaults, plus additional utilities like UUID and ID generation that aren't in the standard library.

## Use Cases

### When to Use

- ✅ **UUID generation** - create unique identifiers for database records
- ✅ **Session tokens** - generate secure session IDs
- ✅ **API keys** - create cryptographically secure keys
- ✅ **Random IDs** - generate readable alphanumeric identifiers
- ✅ **Testing data** - create random test values
- ✅ **Time-based IDs** - unique, sortable identifiers with timestamps
- ✅ **Random sampling** - generate random numbers for algorithms
- ✅ **Game development** - random values for game mechanics
- ✅ **Simulations** - random data for statistical simulations

### When Not to Use

- ❌ **Cryptographic keys** - use dedicated crypto libraries (e.g., `crypto/ecdsa`, `crypto/rsa`)
- ❌ **Password hashing** - use `bcrypt`, `argon2`, or similar
- ❌ **Cross-platform UUIDs** - `UUID()` relies on `/dev/urandom` (Unix-only)
- ❌ **Deterministic randomness** - when you need reproducible sequences with custom seeds
- ❌ **High-performance hot paths** - consider inlining `math/rand` calls directly

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package in your Go code:

```go
import "github.com/polarixa/replify/pkg/randn"
```

**Requirements:**

- Go 1.13 or higher
- Unix-based system for UUID functions (Linux, macOS, BSD)

## Usage

### Quick Start

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/randn"
)

func main() {
    // Generate a UUID
    uuid, err := randn.UUID()
    if err != nil {
        panic(err)
    }
    fmt.Println("UUID:", uuid)
    // Output: UUID: a1b2c3d4-e5f6-7890-abcd-ef1234567890

    // Generate a random alphanumeric ID
    id := randn.RandID(16)
    fmt.Println("Random ID:", id)
    // Output: Random ID: aB3dE5fG7hI9jK1m

    // Generate a random integer in range
    num := randn.RandIntr(1, 100)
    fmt.Println("Random number 1-100:", num)
    // Output: Random number 1-100: 42
}
```

## Examples

### 1. UUID Generation

```go
// Standard UUID (with dashes)
uuid, err := randn.UUID()
if err != nil {
    log.Fatalf("Failed to generate UUID: %v", err)
}
fmt.Println(uuid)
// Output: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

// UUID without dashes
uuidNoDash, err := randn.UUIDSep("")
if err != nil {
    log.Fatal(err)
}
fmt.Println(uuidNoDash)
// Output: "a1b2c3d4e5f67890abcdef1234567890"

// UUID with custom delimiter
uuidCustom, err := randn.UUIDSep(":")
fmt.Println(uuidCustom)
// Output: "a1b2c3d4:e5f6:7890:abcd:ef1234567890"

// UUID without error handling (returns empty string on error)
uuid = randn.RandUUID()
if uuid == "" {
    fmt.Println("Failed to generate UUID")
}
```

### 2. Random ID Generation

```go
// Alphanumeric ID (length 16)
id := randn.RandID(16)
fmt.Println("Short ID:", id)
// Output: "aB3dE5fG7hI9jK1m"

// Longer ID for tokens
token := randn.RandID(32)
fmt.Println("Token:", token)
// Output: "aB3dE5fG7hI9jK1mNoP2qR4sT6uV8wX0"

// Cryptographically secure ID
secureID := randn.CryptoID()
fmt.Println("Secure ID:", secureID)
// Output: "a1b2c3d4e5f67890a1b2c3d4e5f67890" (32 hex chars)

// Time-based ID (timestamp + random)
timeID := randn.TimeID()
fmt.Println("Time ID:", timeID)
// Output: "1704067200123456789987654321"
```

### 3. Random Integers

```go
// Random integer (full int range)
anyInt := randn.RandInt()
fmt.Println("Random int:", anyInt)
// Output: Random int: 5577006791947779410

// Random integer in range [1, 100] (inclusive)
dice100 := randn.RandIntr(1, 100)
fmt.Println("1-100:", dice100)
// Output: 1-100: 42

// Simulate dice roll (1-6)
diceRoll := randn.RandIntr(1, 6)
fmt.Println("Dice roll:", diceRoll)
// Output: Dice roll: 4

// Random year (2020-2030)
year := randn.RandIntr(2020, 2030)
fmt.Println("Random year:", year)
// Output: Random year: 2025
```

### 4. Random Floats

```go
// Float64 in [0.0, 1.0)
probability := randn.RandFt64()
fmt.Printf("Probability: %.4f\n", probability)
// Output: Probability: 0.6046

// Float64 in custom range [10.0, 50.0)
temperature := randn.RandFt64r(10.0, 50.0)
fmt.Printf("Temperature: %.2f°C\n", temperature)
// Output: Temperature: 32.45°C

// Float32 in [0.0, 1.0)
smallProb := randn.RandFt32()
fmt.Printf("Small probability: %.4f\n", smallProb)
// Output: Small probability: 0.9451

// Float32 in custom range [-1.0, 1.0)
offset := randn.RandFt32r(-1.0, 1.0)
fmt.Printf("Offset: %.4f\n", offset)
// Output: Offset: -0.3214
```

### 5. Random Bytes

```go
// Generate 16 random bytes
bytes := randn.RandByte(16)
fmt.Printf("Random bytes: %x\n", bytes)
// Output: Random bytes: a1b2c3d4e5f6789012345678abcdef01

// Generate salt for hashing (32 bytes)
salt := randn.RandByte(32)
fmt.Printf("Salt length: %d\n", len(salt))
// Output: Salt length: 32

// Generate IV for encryption (16 bytes)
iv := randn.RandByte(16)
// Use in your encryption scheme
```

### 6. Practical Use Cases

#### Database Record IDs

```go
type User struct {
    ID        string
    Name      string
    CreatedAt time.Time
}

func NewUser(name string) (*User, error) {
    uuid, err := randn.UUID()
    if err != nil {
        return nil, fmt.Errorf("failed to generate user ID: %w", err)
    }

    return &User{
        ID:        uuid,
        Name:      name,
        CreatedAt: time.Now(),
    }, nil
}
```

#### Session Token Generation

```go
type Session struct {
    Token     string
    UserID    string
    ExpiresAt time.Time
}

func CreateSession(userID string) *Session {
    return &Session{
        Token:     randn.CryptoID(), // Cryptographically secure
        UserID:    userID,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
}
```

#### API Key Generation

```go
func GenerateAPIKey() string {
    // Format: prefix + secure random ID
    prefix := "sk"
    randomPart := randn.CryptoID()
    return fmt.Sprintf("%s_%s", prefix, randomPart)
}

// Usage
apiKey := GenerateAPIKey()
fmt.Println(apiKey)
// Output: sk_a1b2c3d4e5f67890a1b2c3d4e5f67890
```

#### Invitation Code Generation

```go
func GenerateInviteCode() string {
    // 8-character alphanumeric code
    return randn.RandID(8)
}

// Usage
code := GenerateInviteCode()
fmt.Println("Invite code:", code)
// Output: Invite code: aB3dE5fG
```

#### Random Test Data

```go
type TestUser struct {
    ID    string
    Age   int
    Score float64
}

func GenerateTestUser() TestUser {
    return TestUser{
        ID:    randn.RandID(12),
        Age:   randn.RandIntr(18, 80),
        Score: randn.RandFt64r(0.0, 100.0),
    }
}
```

#### Time-Ordered IDs

```go
// Generate sortable IDs based on timestamp
func GenerateOrderedID() string {
    return randn.TimeID()
}

// IDs will be naturally sorted by creation time
ids := []string{
    GenerateOrderedID(),
    GenerateOrderedID(),
    GenerateOrderedID(),
}
// ids are naturally in chronological order
```

#### Random Sampling

```go
// Select random items from a list
func RandomSample(items []string, count int) []string {
    if count > len(items) {
        count = len(items)
    }

    selected := make([]string, 0, count)
    used := make(map[int]bool)

    for len(selected) < count {
        idx := randn.RandIntr(0, len(items)-1)
        if !used[idx] {
            selected = append(selected, items[idx])
            used[idx] = true
        }
    }

    return selected
}
```

#### Simulation Data

```go
// Generate random sensor readings
type SensorReading struct {
    ID          string
    Timestamp   time.Time
    Temperature float64
    Humidity    float64
}

func SimulateSensorReading() SensorReading {
    return SensorReading{
        ID:          randn.RandID(8),
        Timestamp:   time.Now(),
        Temperature: randn.RandFt64r(15.0, 35.0),
        Humidity:    randn.RandFt64r(30.0, 90.0),
    }
}
```

## API Reference

### UUID Functions

| Function                                    | Description                          | Returns              | Error Handling                              |
| ------------------------------------------- | ------------------------------------ | -------------------- | ------------------------------------------- |
| `UUID() (string, error)`                    | Generate standard UUID with dashes   | UUID string          | Returns error if `/dev/urandom` unavailable |
| `UUIDSep(delimiter string) (string, error)` | Generate UUID with custom delimiter  | UUID string          | Returns error if `/dev/urandom` unavailable |
| `RandUUID() string`                         | Generate UUID without error handling | UUID string or empty | Returns `""` on error                       |

**UUID Format:** `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` (32 hex digits with dashes)

**Note:** UUID functions require Unix-based systems with `/dev/urandom` (Linux, macOS, BSD).

---

### ID Generation Functions

| Function                    | Description                     | Parameters           | Returns                           |
| --------------------------- | ------------------------------- | -------------------- | --------------------------------- |
| `RandID(length int) string` | Alphanumeric ID                 | `length` - ID length | Random string (A-Z, a-z, 0-9)     |
| `CryptoID() string`         | Cryptographically secure hex ID | None                 | 32-character hex string           |
| `TimeID() string`           | Timestamp-based ID              | None                 | Nanosecond timestamp + random int |

**RandID Characters:** `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`

**CryptoID:** Uses `crypto/rand` for security-critical applications. Suitable for tokens, API keys, secrets.

**TimeID:** Sortable by creation time. Format: `{nanoseconds}{randomint}`

---

### Random Integer Functions

| Function                     | Description             | Parameters               | Returns    | Range          |
| ---------------------------- | ----------------------- | ------------------------ | ---------- | -------------- |
| `RandInt() int`              | Random integer          | None                     | Random int | Full int range |
| `RandIntr(min, max int) int` | Random integer in range | `min`, `max` (inclusive) | Random int | [min, max]     |

**RandIntr Behavior:**

- Both bounds are **inclusive**: `RandIntr(1, 10)` can return 1 or 10
- If `min >= max`, returns `min`
- Automatically reseeds on each call for better randomness

---

### Random Float Functions

| Function                                | Description      | Parameters     | Returns        | Range        |
| --------------------------------------- | ---------------- | -------------- | -------------- | ------------ |
| `RandFt64() float64`                    | Random float64   | None           | Random float64 | [0.0, 1.0)   |
| `RandFt64r(start, end float64) float64` | Float64 in range | `start`, `end` | Random float64 | [start, end) |
| `RandFt32() float32`                    | Random float32   | None           | Random float32 | [0.0, 1.0)   |
| `RandFt32r(start, end float32) float32` | Float32 in range | `start`, `end` | Random float32 | [start, end) |

**Note:** Upper bound is **exclusive** for float functions.

---

### Random Byte Functions

| Function                     | Description           | Parameters                | Returns    |
| ---------------------------- | --------------------- | ------------------------- | ---------- |
| `RandByte(count int) []byte` | Generate random bytes | `count` - number of bytes | Byte slice |

**Use Cases:**

- Encryption IVs (Initialization Vectors)
- Salts for password hashing
- Random binary data
- Nonces for cryptographic operations

## Best Practices & Notes

### ⚠️ Common Pitfalls

1. **UUID Platform Dependency**

   ```go
   // ❌ Won't work on Windows
   uuid, err := randn.UUID()
   // Requires Unix-based system (/dev/urandom)

   // ✅ Always check for errors
   uuid, err := randn.UUID()
   if err != nil {
       // Fallback to another method or handle error
       log.Printf("UUID generation failed: %v", err)
   }
   ```

2. **RandIntr Range Confusion**

   ```go
   // ✅ Correct: both bounds inclusive
   dice := randn.RandIntr(1, 6)  // Can return 1, 2, 3, 4, 5, or 6

   // ❌ Common mistake: forgetting upper bound is inclusive
   index := randn.RandIntr(0, len(slice))  // Wrong! Can exceed slice bounds

   // ✅ Correct: adjust upper bound
   index := randn.RandIntr(0, len(slice)-1)
   ```

3. **Float Range Boundaries**

   ```go
   // Note: upper bound is exclusive for floats
   val := randn.RandFt64r(0.0, 1.0)  // Returns [0.0, 1.0)
   // Can be 0.0, but will never be exactly 1.0
   ```

4. **Security vs Performance**

   ```go
   // ❌ Don't use RandID for security-critical tasks
   apiKey := randn.RandID(32)  // Not cryptographically secure

   // ✅ Use CryptoID for security
   apiKey := randn.CryptoID()  // Cryptographically secure
   ```

### 💡 Recommendations

✅ **Use appropriate functions for the task**

```go
// For database IDs (readability + uniqueness)
userID := randn.UUID()

// For human-readable codes (invites, vouchers)
code := randn.RandID(8)

// For security tokens (sessions, API keys)
token := randn.CryptoID()

// For sortable IDs (logs, events)
logID := randn.TimeID()
```

✅ **Error handling for UUIDs**

```go
// Always handle UUID errors in production
uuid, err := randn.UUID()
if err != nil {
    // Log error and use fallback
    log.Printf("UUID generation failed: %v", err)
    uuid = randn.TimeID() // Fallback
}
```

✅ **Validate ranges before calling RandIntr**

```go
func RandomIndex(sliceLen int) int {
    if sliceLen <= 0 {
        return 0
    }
    return randn.RandIntr(0, sliceLen-1)
}
```

✅ **Use appropriate ID lengths**

```go
// Short codes (human-readable): 6-8 characters
inviteCode := randn.RandID(8)

// Standard IDs (good uniqueness): 16 characters
recordID := randn.RandID(16)

// Long tokens (high uniqueness): 32+ characters
sessionToken := randn.RandID(32)
```

✅ **Prefix IDs for identification**

```go
func GenerateUserID() string {
    return "usr_" + randn.RandID(16)
}

func GenerateOrderID() string {
    return "ord_" + randn.RandID(16)
}
```

### 🔒 Security Considerations

**When to use CryptoID:**

- ✅ API keys
- ✅ Session tokens
- ✅ OAuth secrets
- ✅ Encryption keys
- ✅ Password reset tokens
- ✅ CSRF tokens

**When RandID is sufficient:**

- ✅ Non-sensitive unique IDs
- ✅ Test data
- ✅ Temporary identifiers
- ✅ Display-only codes
- ✅ Gaming/simulation

**Security notes:**

```go
// ❌ Don't use for cryptographic keys
key := randn.RandByte(32)  // Uses math/rand, not crypto/rand

// ✅ Use crypto/rand for sensitive data
import "crypto/rand"
key := make([]byte, 32)
rand.Read(key)
```

### ⚡ Performance Tips

**Fast operations (math/rand):**

- `RandInt()`, `RandIntr()` - Very fast
- `RandFt64()`, `RandFt32()` - Very fast
- `RandID()` - Fast for short lengths
- `RandByte()` - Fast

**Moderate operations:**

- `UUID()`, `RandUUID()` - File I/O overhead
- `UUIDSep()` - File I/O overhead
- `TimeID()` - System call overhead

**Slower operations (crypto/rand):**

- `CryptoID()` - Cryptographically secure, slower

**Optimization strategies:**

```go
// ✅ Generate IDs in batch for better performance
func GenerateBatchIDs(count int) []string {
    ids := make([]string, count)
    for i := 0; i < count; i++ {
        ids[i] = randn.RandID(16)
    }
    return ids
}

// ✅ Reuse byte slices
func ReuseByteSlice() {
    buffer := make([]byte, 16)
    for i := 0; i < 1000; i++ {
        buffer = randn.RandByte(16)
        // Use buffer
    }
}
```

### 🐛 Debugging Tips

**Check UUID generation:**

```go
uuid, err := randn.UUID()
if err != nil {
    fmt.Printf("UUID Error: %v\n", err)
    fmt.Println("Check if /dev/urandom is available")
}
fmt.Printf("UUID: %s (length: %d)\n", uuid, len(uuid))
```

**Verify randomness:**

```go
// Generate multiple values to check distribution
for i := 0; i < 10; i++ {
    fmt.Println(randn.RandIntr(1, 10))
}
```

**Test edge cases:**

```go
// Same min and max
fmt.Println(randn.RandIntr(5, 5))  // Always returns 5

// Invalid range
fmt.Println(randn.RandIntr(10, 5)) // Returns 10 (min)
```

### 🔧 Thread Safety

The package uses a package-level random generator that is **NOT** thread-safe for the custom functions (`RandIntr`).

```go
// ⚠️ Not thread-safe without synchronization
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        randn.RandIntr(1, 100) // Potential race condition
    }()
}

// ✅ Thread-safe with mutex
var mu sync.Mutex
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        mu.Lock()
        val := randn.RandIntr(1, 100)
        mu.Unlock()
        // Use val
    }()
}
```

**Note:** Standard `math/rand` functions (`RandInt`, `RandFt64`, etc.) use the global generator which has internal locking in Go 1.6+.

### 📝 Testing

Example test cases:

```go
func TestUUIDFormat(t *testing.T) {
    uuid, err := randn.UUID()
    if err != nil {
        t.Skip("Skipping UUID test (requires /dev/urandom)")
    }

    // Check format (8-4-4-4-12)
    if len(uuid) != 36 {
        t.Errorf("UUID length = %d, want 36", len(uuid))
    }

    // Check dashes at correct positions
    if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
        t.Errorf("UUID format incorrect: %s", uuid)
    }
}

func TestRandIntrRange(t *testing.T) {
    min, max := 1, 10
    for i := 0; i < 1000; i++ {
        val := randn.RandIntr(min, max)
        if val < min || val > max {
            t.Errorf("RandIntr(%d, %d) = %d, out of range", min, max, val)
        }
    }
}

func TestRandIDLength(t *testing.T) {
    lengths := []int{8, 16, 32, 64}
    for _, length := range lengths {
        id := randn.RandID(length)
        if len(id) != length {
            t.Errorf("RandID(%d) length = %d, want %d", length, len(id), length)
        }
    }
}
```

## Limitations

- **Platform-specific:** UUID functions require Unix-based systems (`/dev/urandom`)
- **Not cryptographically secure:** Most functions use `math/rand` (except `CryptoID`)
- **Thread safety:** `RandIntr` is not thread-safe without external synchronization
- **No UUID version support:** Generated UUIDs don't follow UUID v4 spec exactly
- **No custom seeding:** Cannot set custom seed for reproducibility

## Migration from Standard Library

```go
// Before (math/rand)
import "math/rand"
rand.Seed(time.Now().UnixNano())
val := rand.Intn(100)

// After (randn)
import "github.com/polarixa/replify/pkg/randn"
val := randn.RandIntr(0, 99)
```

## Contributing

Contributions are welcome! Please see the main [replify repository](https://github.com/polarixa/replify) for contribution guidelines.

## License

This library is part of the [replify](https://github.com/polarixa/replify) project.

## Related

Part of the **replify** ecosystem:

- [replify](https://github.com/polarixa/replify) - API response wrapping library
- [conv](https://github.com/polarixa/replify/pkg/conv) - Type conversion utilities
- [hashy](https://github.com/polarixa/replify/pkg/hashy) - Deterministic hashing
- [match](https://github.com/polarixa/replify/pkg/match) - Wildcard pattern matching
- [coll](https://github.com/polarixa/replify/pkg/coll) - Collection utilities
- [strutil](https://github.com/polarixa/replify/pkg/strutil) - String utilities
