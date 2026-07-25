# ref

**ref** is a Go utility library providing elegant and type-safe pointer operations. It simplifies working with pointer types in structs, API requests, optional fields, and configuration management through a clean, functional API.

## Overview

The `ref` package solves common pointer-related challenges in Go by providing:

- **Pointer Creation**: Create pointers to literals inline without temporary variables
- **Safe Dereferencing**: Safely access pointer values with zero-value or custom defaults
- **Nil Checking**: Elegant nil validation and conditional logic
- **Functional Operations**: Map, filter, and transform pointer values functionally
- **Validation**: Built-in support for pointer value validation
- **Fallback Chains**: Coalesce multiple pointers with fallback logic

**Problem Solved:** Go's pointer syntax can be verbose, especially when working with optional struct fields, API requests with `omitempty`, or configuration with defaults. This package provides idiomatic utilities that make pointer operations clean, safe, and expressive.

## Use Cases

### When to Use

- ✅ **API requests/responses** - optional fields with `omitempty`
- ✅ **Database models** - nullable fields that can be NULL
- ✅ **Configuration management** - optional settings with defaults
- ✅ **Partial updates** - PATCH requests with selective field updates
- ✅ **Optional struct fields** - clean initialization without temp variables
- ✅ **Validation logic** - required field checking
- ✅ **Fallback chains** - multiple sources with precedence (user > env > default)
- ✅ **Functional transformations** - map/filter operations on optional values

### When Not to Use

- ❌ **Simple non-optional values** - use regular values instead
- ❌ **Always-present data** - pointers add unnecessary complexity
- ❌ **Performance-critical hot paths** - pointer operations have overhead
- ❌ **When nil has no semantic meaning** - pointers should represent optionality

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package in your Go code:

```go
import "github.com/polarixa/replify/pkg/ref"
```

**Requirements:** Go 1.18 or higher (for generics support)

## Usage

### Quick Start

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/ref"
)

func main() {
    // Create pointers inline
    name := ref.Ptr("John Doe")
    age := ref.Ptr(30)

    // Safe dereferencing
    fmt.Println(ref.Deref(name))     // "John Doe"
    fmt.Println(ref.Deref((*string)(nil))) // "" (zero value)

    // With custom defaults
    fmt.Println(ref.DerefOr((*int)(nil), 100)) // 100

    // Nil checking
    if ref.IsNotNil(name) {
        fmt.Println("Name is provided:", *name)
    }

    // Transform pointer values
    upper := ref.Map(name, strings.ToUpper)
    fmt.Println(*upper) // "JOHN DOE"
}
```

## Examples

### 1. API Requests with Optional Fields

```go
type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty"`
    Email *string `json:"email,omitempty"`
    Age   *int    `json:"age,omitempty"`
}

// ❌ Without ref: verbose temporary variables
func createRequestOld() UpdateUserRequest {
    name := "John Doe"
    email := "john@example.com"
    return UpdateUserRequest{
        Name:  &name,
        Email: &email,
    }
}

// ✅ With ref: clean and concise
func createRequest() UpdateUserRequest {
    return UpdateUserRequest{
        Name:  ref.Ptr("John Doe"),
        Email: ref.Ptr("john@example.com"),
        Age:   ref.Ptr(30),
    }
}
```

### 2. Configuration with Defaults

```go
type ServerConfig struct {
    Host     *string
    Port     *int
    Timeout  *time.Duration
    MaxConns *int
}

func NewServer(config ServerConfig) *Server {
    return &Server{
        Host:     ref.DerefOr(config.Host, "0.0.0.0"),
        Port:     ref.DerefOr(config.Port, 8080),
        Timeout:  ref.DerefOr(config.Timeout, 30*time.Second),
        MaxConns: ref.DerefOr(config.MaxConns, 100),
    }
}

// Usage
server := NewServer(ServerConfig{
    Host: ref.Ptr("localhost"),
    Port: ref.Ptr(3000),
    // Timeout and MaxConns use defaults
})
```

### 3. Safe Dereferencing

```go
type User struct {
    ID    int
    Name  string
    Email *string // Optional
    Phone *string // Optional
}

func DisplayUser(user User) {
    fmt.Printf("User: %s\n", user.Name)

    // Safe dereferencing with zero values
    email := ref.Deref(user.Email) // "" if nil
    phone := ref.Deref(user.Phone) // "" if nil

    if email != "" {
        fmt.Printf("Email: %s\n", email)
    }
    if phone != "" {
        fmt.Printf("Phone: %s\n", phone)
    }
}
```

### 4. Validation

```go
type CreateUserRequest struct {
    Name     *string
    Email    *string
    Password *string
}

func ValidateCreateUser(req CreateUserRequest) error {
    // Check all required fields are present
    if !ref.All(req.Name, req.Email, req.Password) {
        return errors.New("all fields are required")
    }

    // Safe to dereference after validation
    name := ref.Must(req.Name)
    email := ref.Must(req.Email)
    password := ref.Must(req.Password)

    // Additional validation...
    if len(name) < 2 {
        return errors.New("name too short")
    }

    return nil
}
```

### 5. Partial Updates

```go
type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty"`
    Email *string `json:"email,omitempty"`
    Age   *int    `json:"age,omitempty"`
}

func UpdateUser(userID int, req UpdateUserRequest) error {
    // Only update provided fields
    updates := make(map[string]interface{})

    if ref.IsNotNil(req.Name) {
        updates["name"] = *req.Name
    }
    if ref.IsNotNil(req.Email) {
        updates["email"] = *req.Email
    }
    if ref.IsNotNil(req.Age) {
        updates["age"] = *req.Age
    }

    if len(updates) == 0 {
        return errors.New("no fields to update")
    }

    return db.Model(&User{}).Where("id = ?", userID).Updates(updates).Error
}
```

### 6. Functional Transformations

```go
// Map: Transform pointer values
name := ref.Ptr("john doe")
upperName := ref.Map(name, strings.ToUpper)
fmt.Println(*upperName) // "JOHN DOE"

// Filter: Conditional filtering
age := ref.Ptr(15)
adultAge := ref.Filter(age, func(a int) bool {
    return a >= 18
})
fmt.Println(adultAge) // nil (15 < 18)

// MapOr: Transform with default
var nilName *string
formatted := ref.MapOr(nilName,
    func(n string) string {
        return strings.ToUpper(n)
    },
    "ANONYMOUS",
)
fmt.Println(formatted) // "ANONYMOUS"
```

### 7. Fallback Chains

```go
type Config struct {
    Timeout *time.Duration
}

userConfig := Config{} // nil
envConfig := Config{Timeout: ref.Ptr(10 * time.Second)}
defaultConfig := Config{Timeout: ref.Ptr(30 * time.Second)}

// Use first non-nil value
timeout := ref.Coalesce(
    userConfig.Timeout,
    envConfig.Timeout,
    defaultConfig.Timeout,
) // Returns 10 * time.Second
```

### 8. Conditional Processing

```go
type User struct {
    Email *string
    Phone *string
}

user := User{
    Email: ref.Ptr("user@example.com"),
}

// Execute callback only if non-nil
ref.If(user.Email, func(email string) {
    sendEmail(email, "Welcome!")
})

ref.If(user.Phone, func(phone string) {
    sendSMS(phone, "Welcome!") // Not executed (Phone is nil)
})

// If-else branching
ref.IfElse(user.Phone,
    func(phone string) {
        fmt.Println("Phone:", phone)
    },
    func() {
        fmt.Println("No phone provided")
    },
)
```

### 9. Lazy Evaluation

```go
// OrElseGet: Compute fallback only when needed
func GetUserName(userID int) *string {
    // Try cache first (fast)
    cached := getFromCache(userID)

    // Only query database if cache miss (expensive)
    return ref.OrElseGet(cached, func() *string {
        return queryDatabase(userID)
    })
}

// Generate UUID only if needed
sessionID := ref.OrElseGet(existingSession, func() *string {
    return ref.Ptr(uuid.New().String())
})
```

### 10. Comparison

```go
oldConfig := Config{
    Timeout: ref.Ptr(30 * time.Second),
    MaxRetries: ref.Ptr(3),
}

newConfig := Config{
    Timeout: ref.Ptr(30 * time.Second),
    MaxRetries: ref.Ptr(5),
}

// Compare pointer values
if ref.Equal(oldConfig.Timeout, newConfig.Timeout) {
    fmt.Println("Timeout unchanged")
}

if !ref.Equal(oldConfig.MaxRetries, newConfig.MaxRetries) {
    fmt.Println("MaxRetries changed")
}
```

### 11. Copy and Clone

```go
originalConfig := Config{
    Host: ref.Ptr("localhost"),
    Port: ref.Ptr(8080),
}

// Create independent copy
configCopy := Config{
    Host: ref.Copy(originalConfig.Host),
    Port: ref.Copy(originalConfig.Port),
}

// Modifying copy doesn't affect original
*configCopy.Host = "example.com"
fmt.Println(*originalConfig.Host) // Still "localhost"
```

### 12. Validation with Custom Logic

```go
func isValidEmail(email string) error {
    if !strings.Contains(email, "@") {
        return errors.New("invalid email format")
    }
    return nil
}

email := ref.Ptr("user@example.com")
validEmail := ref.Validate(email, isValidEmail)

if ref.IsNotNil(validEmail) {
    sendWelcomeEmail(*validEmail)
}

badEmail := ref.Ptr("invalid-email")
invalidEmail := ref.Validate(badEmail, isValidEmail)
// invalidEmail is nil
```

### 13. Omitting Zero Values

```go
type APIRequest struct {
    Name  *string `json:"name,omitempty"`
    Age   *int    `json:"age,omitempty"`
    Score *int    `json:"score,omitempty"`
}

name := "John"
age := 0      // Zero value
score := 100

req := APIRequest{
    Name:  ref.ToPtr(name),   // &"John"
    Age:   ref.ToPtr(age),    // nil (zero value omitted)
    Score: ref.ToPtr(score),  // &100
}

// JSON: {"name":"John","score":100}
// Age is omitted because it's nil
```

### 14. FlatMap for Nested Optionals

```go
type Address struct {
    City    string
    ZipCode *string
}

type User struct {
    Name    string
    Address *Address
}

user := User{
    Name:    "John",
    Address: ref.Ptr(Address{
        City: "NYC",
        ZipCode: ref.Ptr("10001"),
    }),
}

// Safe nested access
zipCode := ref.FlatMap(user.Address, func(addr Address) *string {
    return addr.ZipCode
})

if ref.IsNotNil(zipCode) {
    fmt.Println("Zip:", *zipCode)
}
```

## API Reference

### Pointer Creation

| Function                      | Signature                              | Description             |
| ----------------------------- | -------------------------------------- | ----------------------- |
| `Ptr[T any](v T) *T`          | Creates pointer to value               | Inline pointer creation |
| `ToPtr[T comparable](v T) *T` | Creates pointer only if not zero value | Omit zero values        |
| `Copy[T any](ptr *T) *T`      | Creates independent copy               | Deep copy pointer value |

**Examples:**

```go
name := ref.Ptr("John")          // &"John"
age := ref.ToPtr(0)              // nil (zero value)
clone := ref.Copy(name)          // Independent copy
```

---

### Safe Dereferencing

| Function                               | Signature                            | Description                    |
| -------------------------------------- | ------------------------------------ | ------------------------------ |
| `Deref[T any](ptr *T) T`               | Returns value or zero value          | Safe deref with zero default   |
| `DerefOr[T any](ptr *T, def T) T`      | Returns value or custom default      | Safe deref with custom default |
| `Must[T any](ptr *T) T`                | Returns value or panics              | For validated non-nil pointers |
| `MustPtr[T any](ptr *T, msg string) T` | Returns value or panics with message | With custom panic message      |

**Examples:**

```go
value := ref.Deref(ptr)          // Zero value if nil
value := ref.DerefOr(ptr, 100)   // 100 if nil
value := ref.Must(ptr)           // Panics if nil
value := ref.MustPtr(ptr, "required") // Panics with message
```

---

### Nil Checking

| Function                            | Signature                    | Description              |
| ----------------------------------- | ---------------------------- | ------------------------ |
| `IsNil[T any](ptr *T) bool`         | Checks if pointer is nil     | Explicit nil check       |
| `IsNotNil[T any](ptr *T) bool`      | Checks if pointer is not nil | Positive condition check |
| `Equal[T comparable](a, b *T) bool` | Compares pointer values      | Deep equality check      |

**Examples:**

```go
if ref.IsNil(ptr) { }
if ref.IsNotNil(ptr) { }
if ref.Equal(ptr1, ptr2) { }
```

---

### Functional Operations

| Function                                          | Signature                 | Description                 |
| ------------------------------------------------- | ------------------------- | --------------------------- |
| `Map[T, R](ptr *T, fn func(T) R) *R`              | Transform pointer value   | Apply function if non-nil   |
| `FlatMap[T, R](ptr *T, fn func(T) *R) *R`         | Chain optional operations | For nested optionals        |
| `Filter[T](ptr *T, pred func(T) bool) *T`         | Filter by predicate       | Keep if condition true      |
| `MapOr[T, R](ptr *T, fn func(T) R, def R) R`      | Map with default          | Transform or return default |
| `Validate[T](ptr *T, validator func(T) error) *T` | Validate pointer value    | Filter by validation        |

**Examples:**

```go
upper := ref.Map(name, strings.ToUpper)
adult := ref.Filter(age, func(a int) bool { return a >= 18 })
result := ref.MapOr(name, transform, "default")
```

---

### Fallback and Coalescing

| Function                                | Signature             | Description                     |
| --------------------------------------- | --------------------- | ------------------------------- |
| `CoalescePtr[T](ptrs ...*T) *T`         | First non-nil pointer | Returns pointer                 |
| `Coalesce[T](ptrs ...*T) T`             | First non-nil value   | Returns dereferenced value      |
| `OrElse[T](ptr, alt *T) *T`             | Fallback pointer      | Returns ptr or alternative      |
| `OrElseGet[T](ptr *T, fn func() *T) *T` | Lazy fallback         | Compute fallback only if needed |

**Examples:**

```go
ptr := ref.CoalescePtr(user, env, default)
val := ref.Coalesce(user, env, default)
ptr := ref.OrElse(primary, fallback)
ptr := ref.OrElseGet(cached, fetchFromDB)
```

---

### Conditional Execution

| Function                                                | Signature          | Description               |
| ------------------------------------------------------- | ------------------ | ------------------------- |
| `If[T](ptr *T, fn func(T)) bool`                        | Execute if non-nil | Side effect on value      |
| `IfElse[T](ptr *T, onPresent func(T), onAbsent func())` | If-else branching  | Execute based on presence |

**Examples:**

```go
executed := ref.If(email, sendEmail)
ref.IfElse(phone, sendSMS, logMissing)
```

---

### Bulk Operations

| Function                  | Signature         | Description              |
| ------------------------- | ----------------- | ------------------------ |
| `All[T](ptrs ...*T) bool` | Check all non-nil | Validate required fields |
| `Any[T](ptrs ...*T) bool` | Check any non-nil | At least one present     |

**Examples:**

```go
if ref.All(name, email, password) { /* all required */ }
if ref.Any(email, phone, address) { /* at least one */ }
```

## Best Practices & Notes

### ⚠️ Common Pitfalls

1. **Overusing Must**

   ```go
   // ❌ Bad: May panic in production
   name := ref.Must(req.Name)

   // ✅ Good: Validate first or use safe alternatives
   if ref.IsNil(req.Name) {
       return errors.New("name required")
   }
   name := *req.Name

   // ✅ Better: Use with defaults
   name := ref.DerefOr(req.Name, "Anonymous")
   ```

2. **Not Checking Validation Results**

   ```go
   // ❌ Bad: Assuming validation succeeded
   validEmail := ref.Validate(email, isValidEmail)
   sendEmail(*validEmail) // May panic if nil

   // ✅ Good: Check result
   if validEmail := ref.Validate(email, isValidEmail); ref.IsNotNil(validEmail) {
       sendEmail(*validEmail)
   }
   ```

3. **Unnecessary Pointer Usage**

   ```go
   // ❌ Bad: Pointer for always-present field
   type User struct {
       ID   *int    // Always present, shouldn't be pointer
       Name *string // Always present, shouldn't be pointer
   }

   // ✅ Good: Only optional fields are pointers
   type User struct {
       ID    int
       Name  string
       Email *string // Optional
       Phone *string // Optional
   }
   ```

4. **Modifying Shared Pointers**

   ```go
   // ❌ Dangerous: Modifying shared pointer
   ptr1 := ref.Ptr("original")
   ptr2 := ptr1
   *ptr2 = "modified" // Affects ptr1 too!

   // ✅ Safe: Use Copy for independent values
   ptr1 := ref.Ptr("original")
   ptr2 := ref.Copy(ptr1)
   *ptr2 = "modified" // ptr1 unchanged
   ```

### 💡 Recommendations

✅ **Use for optional fields**

```go
type Config struct {
    Host     string  // Required
    Port     int     // Required
    Timeout  *int    // Optional
    MaxConns *int    // Optional
}
```

✅ **Use ToPtr for omitting zero values**

```go
req := APIRequest{
    Name:  ref.ToPtr(name),  // Omit if ""
    Age:   ref.ToPtr(age),   // Omit if 0
    Score: ref.ToPtr(score), // Omit if 0
}
```

✅ **Validate before Must**

```go
if !ref.All(req.Name, req.Email) {
    return errors.New("missing required fields")
}
// Safe to use Must now
name := ref.Must(req.Name)
email := ref.Must(req.Email)
```

✅ **Use functional operations for transformations**

```go
// Clean and expressive
normalized := ref.Map(email, normalizeEmail)
adult := ref.Filter(age, isAdult)
```

✅ **Leverage OrElseGet for expensive operations**

```go
// Only query DB if cache miss
data := ref.OrElseGet(cached, queryDatabase)
```

### 🔒 Thread Safety

All functions are **thread-safe** as they don't maintain state. However, be careful with concurrent pointer modifications:

```go
// ⚠️ Not safe: Concurrent modification
ptr := ref.Ptr("value")
go func() { *ptr = "modified1" }()
go func() { *ptr = "modified2" }() // Race condition!

// ✅ Safe: Use synchronization
var mu sync.Mutex
ptr := ref.Ptr("value")
go func() {
    mu.Lock()
    *ptr = "modified1"
    mu.Unlock()
}()
```

### ⚡ Performance Considerations

**Pointer overhead:**

- Pointers add memory indirection (slower access)
- Heap allocations (GC pressure)
- Use only when optionality is needed

**Optimization tips:**

```go
// ❌ Inefficient: Unnecessary pointers
type Point struct {
    X *float64
    Y *float64
}

// ✅ Efficient: Direct values
type Point struct {
    X float64
    Y float64
}

// ✅ Use pointers only for optional fields
type User struct {
    ID    int      // Always present
    Name  string   // Always present
    Bio   *string  // Optional
}
```

### 🐛 Debugging Tips

**Check for nil before dereferencing:**

```go
if ref.IsNotNil(ptr) {
    value := *ptr
} else {
    log.Println("Pointer is nil")
}
```

**Use Must in tests, avoid in production:**

```go
// ✅ OK in tests
func TestUser(t *testing.T) {
    user := createTestUser()
    name := ref.Must(user.Name)
    assert.Equal(t, "Test", name)
}

// ❌ Avoid in production
func HandleRequest(req Request) {
    name := ref.Must(req.Name) // May panic!
}
```

### 📝 Testing

Example tests:

```go
func TestPtr(t *testing.T) {
    value := 42
    ptr := ref.Ptr(value)

    assert.NotNil(t, ptr)
    assert.Equal(t, 42, *ptr)
}

func TestDerefOr(t *testing.T) {
    var nilPtr *int
    result := ref.DerefOr(nilPtr, 100)
    assert.Equal(t, 100, result)

    ptr := ref.Ptr(42)
    result = ref.DerefOr(ptr, 100)
    assert.Equal(t, 42, result)
}
```

## Limitations

- **Not for concurrent modification** - requires external synchronization
- **Pointer overhead** - adds memory indirection and GC pressure
- **Can hide nil pointer bugs** - be explicit about optionality
- **Zero value vs nil distinction** - understand the difference
- **Not suitable for high-performance code** - direct access is faster

## When to Use vs. Explicit Nil Checks

**Use `ref` when:**

- Working with optional struct fields
- Building APIs with `omitempty`
- Configuration with defaults
- Functional transformations

**Use explicit nil checks when:**

- Simple one-off nil checks
- Performance-critical paths
- When clarity is more important than brevity

## Contributing

Contributions are welcome! Please see the main [replify repository](https://github.com/polarixa/replify) for contribution guidelines.

## License

This library is part of the [replify](https://github.com/polarixa/replify) project.

## Related

Part of the **replify** ecosystem:

- [replify](https://github.com/polarixa/replify) - API response wrapping library
- [conv](https://github.com/polarixa/replify/pkg/conv) - Type conversion utilities
- [coll](https://github.com/polarixa/replify/pkg/coll) - Type-safe collection utilities
- [common](https://github.com/polarixa/replify/pkg/common) - Reflection-based utilities
- [hashy](https://github.com/polarixa/replify/pkg/hashy) - Deterministic hashing
- [match](https://github.com/polarixa/replify/pkg/match) - Wildcard pattern matching
- [strutil](https://github.com/polarixa/replify/pkg/strutil) - String utilities
- [randn](https://github.com/polarixa/replify/pkg/randn) - Random data generation
- [encoding](https://github.com/polarixa/replify/pkg/encoding) - JSON encoding utilities
