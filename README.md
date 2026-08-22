# replify

**replify** is a Go library designed to simplify and standardize API response wrapping for RESTful services. It leverages the Decorator Pattern to dynamically add error handling, metadata, pagination, and other response features in a clean and human-readable format.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-GPL-green)](LICENSE)

## Overview

Building RESTful APIs often requires repetitive boilerplate code for standardizing responses. **replify** eliminates this by providing a fluent, chainable API that ensures consistent response formats across all your endpoints.

### What Problems Does It Solve?

- ❌ **Inconsistent response formats** across different endpoints
- ❌ **Repetitive error handling** boilerplate in every handler
- ❌ **Manual metadata management** (request IDs, timestamps, versions)
- ❌ **Complex pagination logic** scattered throughout the codebase
- ❌ **Debugging difficulties** in production vs development environments

### The Solution

✅ **Standardized response structure** - One format for all endpoints  
✅ **Fluent API** - Chainable methods for building responses  
✅ **Built-in pagination** - Complete pagination support out of the box  
✅ **Metadata management** - Request IDs, timestamps, API versions, locales  
✅ **Conditional debugging** - Development-only debug information  
✅ **Error handling** - Stack traces, error wrapping, contextual messages  
✅ **Type safety** - Full type safety with Go generics  
✅ **Zero dependencies** - Only uses Go standard library

## Features

### Core Capabilities

- 🎯 **Standardized JSON Format** - Consistent structure across all API responses
- 🔗 **Fluent Builder Pattern** - Chain methods to construct complex responses
- 📄 **Pagination Support** - Built-in page, per_page, total_items, total_pages, is_last
- 🔍 **Request Tracing** - Track requests with unique IDs across microservices
- 🌍 **Internationalization** - Locale support for multi-language APIs
- 🐛 **Debug Mode** - Conditional debugging information for development
- ⚡ **Error Handling** - Rich error information with stack traces
- 📊 **Metadata** - API version, custom fields, timestamps
- ✅ **Status Helpers** - IsSuccess(), IsClientError(), IsServerError()
- 🔄 **JSON Parsing** - Parse JSON strings back to wrapper objects

## Requirements

- Go version 1.23 or higher

## Installation

### Install Package

> Latest version

```bash
go get github.com/polarixa/replify@latest
```

> Specific version

```bash
go get github.com/polarixa/replify@v0.1.0
```

### Import in Code

```go
import "github.com/polarixa/replify"
```

With [Go's module support](https://go.dev/wiki/Modules#how-to-use-modules), `go [build|run|test]` automatically fetches the necessary dependencies when you add the import.

## Quick Start

### Native `net/http` Handler

Use `Write` to send a Replify response directly to an `http.ResponseWriter`.
`Write` dispatches automatically: if a filepath is configured it calls `WriteFile`,
if the data is a `[]byte` it calls `WriteBinary`, and otherwise it calls `WriteJSON`.
No external helper is needed — the correct Content-Type and status code are set for you.

```go
package main

import (
    "log"
    "net/http"

    "github.com/polarixa/replify"
)

func GetUser(response http.ResponseWriter, request *http.Request) {
	user := map[string]string{"id": "123", "name": "John Doe"}

	w := replify.New().
		OK().
		WithBody(user).
		WithMessage("User retrieved successfully").
		Write(response)

	w.Slogging()

	if w.IsErrorPresent() {
		w.S().Errorf("Error writing response: %v", w.Error())
	}
}

func main() {
    http.HandleFunc("/user", GetUser)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Write Responses

Replify provides a generic `Write` method that dispatches to the right writer based on wrapper state, plus dedicated writers for JSON, files, and binary data. All writers are framework-independent and accept the standard `http.ResponseWriter`, so they work naturally with `net/http`, Gin, Echo, Fiber, and any other framework whose response writer implements that interface.

> **Zero dependencies** — the implementation uses only Go's standard library (`net/http`, `mime`, `path/filepath`, `os`, `io`).

### JSON

```go
replify.New().
    WithStatusCode(http.StatusOK).
    WithBody(user).
    WithMessage("User retrieved successfully").
    Write(w)
```

### File

```go
replify.New().
    File("/tmp/report.pdf").
    Write(w)
```

### File Attachment

Serves the file with a `Content-Disposition: attachment` header so the browser offers a download dialog:

```go
replify.New().
    FileAttachment("/tmp/report.pdf", "report.pdf").
    Write(w)
```

### Binary

```go
replify.New().
    Binary(data).
    Write(w)
```

### Binary with Filename

The filename drives MIME-type detection and the `Content-Disposition` header:

```go
replify.New().
    Binary(data).
    WithFilename("report.pdf").
    Write(w)
```

### Custom Status Code

```go
replify.New().
    WithStatusCode(http.StatusCreated).
    Binary(data).
    WithFilename("result.json").
    Write(w)
```

### Gin

Replify does **not** import Gin. It only depends on the standard `http.ResponseWriter` interface. Because Gin's `c.Writer` implements `http.ResponseWriter`, you can pass it directly:

```go
func Download(c *gin.Context) {
    replify.New().
        FileAttachment("/tmp/report.pdf", "report.pdf").
        Write(c.Writer)
}
```

### Dispatch Rules

`Write` inspects wrapper state and calls the appropriate terminal writer:

| Condition          | Writer called |
| ------------------ | ------------- |
| `filepath` is set  | `WriteFile`   |
| `data` is `[]byte` | `WriteBinary` |
| otherwise          | `WriteJSON`   |

No response-type field is introduced. The decision is made purely from the existing wrapper state.

### Safe Filename Handling

Filenames used in `Content-Disposition` headers are validated and encoded by the standard `mime` package (RFC 5987). Filenames containing CR, LF, or null bytes are rejected with an error to prevent HTTP response-header injection.

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify"
)

func main() {
    // Create a simple success response
    response := replify.New().
        WithHeader(replify.OK).
        WithMessage("User retrieved successfully").
        WithBody(map[string]string{
            "id":   "123",
            "name": "John Doe",
        })

    fmt.Println(response.JSONPretty())
}
```

**Output:**

```json
{
  "data": {
    "id": "123",
    "name": "John Doe"
  },
  "header": {
    "code": 200,
    "text": "OK"
  },
  "message": "User retrieved successfully",
  "meta": {
    "api_version": "v0.0.1",
    "locale": "en_US",
    "request_id": "d7e5ce24b796da94770911db36565bf9",
    "requested_time": "2026-01-29T10:07:05.751501+07:00"
  },
  "status_code": 200,
  "total": 0
}
```

## Standard Response Format

The library produces responses in this standardized format:

```json
{
  "status_code": 200,
  "message": "Resource retrieved successfully",
  "path": "/api/v1/users",
  "data": [
    // abstract data (can be array or object)
    {
      "id": "user_01J6G7W9K2M4X7V5P8B3Q2Z1NS",
      "username": "jdoe_dev",
      "email": "j.doe@example.com",
      "role": "administrator",
      "status": "active",
      "created_at": "2025-01-15T08:30:00Z",
      "last_login": "2026-02-26T14:15:22Z"
    },
    {
      "id": "user_01J6G7W9K2M4X7V5P8B3Q2Z1NT",
      "username": "s_smith",
      "email": "sarah.smith@example.com",
      "role": "editor",
      "status": "active",
      "created_at": "2025-02-01T10:15:00Z",
      "last_login": "2026-02-25T09:45:10Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 2,
    "total_items": 120,
    "total_pages": 60,
    "is_last": false
  },
  "cursor": {
    "next": "eyJsYXN0X2lkIjoiMTAwMSJ9",
    "previous": "eyJsYabc...",
    "has_next": true,
    "has_previous": false,
    "limit": 20
  },
  "meta": {
    "request_id": "req_80eafc6a1655ec5a06595d155f1e6951",
    "api_version": "v1.0.4",
    "locale": "en_US",
    "requested_time": "2026-02-26T17:30:28.983Z",
    "custom_fields": {
      // custom fields
      "trace_id": "80eafc6a1655ec5a06595d155f1e6951",
      "origin_region": "us-east-1"
    }
  },
  "debug": {
    // custom fields
    "trace_session_id": "4919e84fc26881e9fe790f5d07465db4",
    "execution_time_ms": 42
  }
}
```

### Field Descriptions

| Field                    | Type          | Description                                             |
| ------------------------ | ------------- | ------------------------------------------------------- |
| `data`                   | `interface{}` | The primary data payload of the response                |
| `status_code`            | `int`         | HTTP status code for the response                       |
| `message`                | `string`      | Human-readable message providing context                |
| `total`                  | `int`         | Total number of items (used in non-paginated responses) |
| `path`                   | `string`      | Request path for which the response is generated        |
| `meta`                   | `object`      | Metadata about the API response                         |
| `meta.request_id`        | `string`      | Unique identifier for the request, useful for debugging |
| `meta.api_version`       | `string`      | API version used for the request                        |
| `meta.locale`            | `string`      | Locale used for the request (e.g., "en_US")             |
| `meta.requested_time`    | `string`      | Timestamp when the request was made (ISO 8601)          |
| `meta.custom_fields`     | `object`      | Additional custom metadata fields                       |
| `pagination`             | `object`      | Pagination details, if applicable                       |
| `pagination.page`        | `int`         | Current page number                                     |
| `pagination.per_page`    | `int`         | Number of items per page                                |
| `pagination.total_items` | `int`         | Total number of items available                         |
| `pagination.total_pages` | `int`         | Total number of pages                                   |
| `pagination.is_last`     | `bool`        | Indicates whether this is the last page                 |
| `debug`                  | `object`      | Debugging information (useful for development)          |

## Usage

### 1. Creating Basic Responses

#### Success Response

```go
response := replify.New().
    WithStatusCode(200).
    WithMessage("Operation successful").
    WithBody(data)
```

#### Error Response

```go
response := replify.New().
    WithStatusCode(400).
    WithError("Invalid input: email is required").
    WithMessage("Validation failed")
```

#### Response with Metadata

```go
response := replify.New().
    WithStatusCode(200).
    WithBody(users).
    WithRequestID("req-123-456").
    WithApiVersion("v1.0.0").
    WithLocale("en_US").
    WithPath("/api/v1/users")
```

### 2. Pagination

#### Creating Pagination

```go
pagination := replify.Pages().
    WithPage(1).
    WithPerPage(20).
    WithTotalItems(150).
    WithTotalPages(8).
    WithIsLast(false)

response := replify.New().
    WithStatusCode(200).
    WithBody(users).
    WithPagination(pagination).
    WithTotal(20)
```

### 3. Debugging Information

```go
response := replify.New().
    WithStatusCode(500).
    WithError("Database connection failed").
    WithDebuggingKV("query", "SELECT * FROM users").
    WithDebuggingKV("error_code", "CONN_TIMEOUT").
    WithDebuggingKV("retry_count", 3)
```

### 4. Complete Example

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify"
    "github.com/polarixa/replify/pkg/randn"
)

func main() {
    // Create pagination
    p := replify.Pages().
        WithIsLast(true).
        WithPage(1000).
        WithTotalItems(120).
        WithTotalPages(34).
        WithPerPage(2)

    // Create response
    w := replify.New().
        WithStatusCode(200).
        WithTotal(1).
        WithMessagef("How are you? %v", "I'm good").
        WithDebuggingKV("refer", 1234).
        WithDebuggingKVf("___abc", "trace sessions_id: %v", randn.CryptoID()).
        WithBody("response body here").
        WithPath("/api/v1/users").
        WithCustomFieldKVf("fields", "userID: %v", 103).
        WithPagination(p)

    if !w.Available() {
        return
    }

    // Access response properties
    fmt.Println(w.JSON())
    fmt.Println(w.StatusCode())
    fmt.Println(w.StatusText())
    fmt.Println(w.Message())
    fmt.Println(w.Body())
    fmt.Println(w.IsSuccess())
    fmt.Println(w.Respond())

    // Check metadata
    fmt.Println(w.Meta().IsCustomPresent())
    fmt.Println(w.Meta().IsApiVersionPresent())
    fmt.Println(w.Meta().IsRequestIDPresent())
    fmt.Println(w.Meta().IsRequestedTimePresent())
}
```

### 5. Parsing JSON to Response

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/polarixa/replify"
)

func main() {
    jsonStr := `{
        "data": "response body here",
        "debug": {
          "___abc": "trace sessions_id: 4919e84fc26881e9fe790f5d07465db4",
          "refer": 1234
        },
        "message": "How do you do? I'm good",
        "meta": {
          "api_version": "v0.0.1",
          "custom_fields": {
            "fields": "userID: 103"
          },
          "locale": "en_US",
          "request_id": "80eafc6a1655ec5a06595d155f1e6951",
          "requested_time": "2024-12-14T20:24:23.983839+07:00"
        },
        "pagination": {
          "is_last": true,
          "page": 1000,
          "per_page": 2,
          "total_items": 120,
          "total_pages": 34
        },
        "path": "/api/v1/users",
        "status_code": 200,
        "total": 1
    }`

    t := time.Now()
    w, err := replify.UnwrapJSON(jsonStr)
    diff := time.Since(t)

    if err != nil {
        log.Fatalf("Error parsing JSON: %v", err)
    }

    fmt.Printf("Exe time: %+v\n", diff.String())
    fmt.Printf("%+v\n", w.OnDebugging("___abc"))
    fmt.Printf("%+v\n", w.JSONPretty())
}
```

## Practical Examples

### Example 1: RESTful CRUD API

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/polarixa/replify"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// GET /users/:id
func GetUser(w http.ResponseWriter, r *http.Request) {
    id := getIDFromPath(r)
    user, err := findUserByID(id)

    var response *replify.R
    if err != nil {
        response = replify.New().
            WithStatusCode(404).
            WithError(err.Error()).
            WithMessage("User not found").
            WithRequestID(r.Header.Get("X-Request-ID"))
    } else {
        response = replify.New().
            WithStatusCode(200).
            WithBody(user).
            WithMessage("User retrieved successfully").
            WithRequestID(r.Header.Get("X-Request-ID"))
    }

    respondJSON(w, response)
}

// POST /users
func CreateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        response := replify.New().
            WithStatusCode(400).
            WithError(err.Error()).
            WithMessage("Invalid request body")
        respondJSON(w, response)
        return
    }

    if err := validateUser(user); err != nil {
        response := replify.New().
            WithStatusCode(422).
            WithError(err.Error()).
            WithMessage("Validation failed")
        respondJSON(w, response)
        return
    }

    createdUser, err := createUser(user)
    if err != nil {
        response := replify.New().
            WithStatusCode(500).
            WithErrorAck(err).
            WithMessage("Failed to create user")
        respondJSON(w, response)
        return
    }

    response := replify.New().
        WithStatusCode(201).
        WithBody(createdUser).
        WithMessage("User created successfully")
    respondJSON(w, response)
}

func respondJSON(w http.ResponseWriter, response *replify.R) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(response.StatusCode())
    w.Write([]byte(response.JSON()))
}
```

### Example 2: Paginated List API

```go
func ListUsers(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    page := getQueryInt(r, "page", 1)
    perPage := getQueryInt(r, "per_page", 10)
    search := r.URL.Query().Get("search")

    // Fetch users with pagination
    users, total, err := db.FindUsers(search, page, perPage)
    if err != nil {
        response := replify.New().
            WithStatusCode(500).
            WithErrorAck(err).
            WithMessage("Failed to fetch users").
            WithDebuggingKV("search", search).
            WithDebuggingKV("page", page)
        respondJSON(w, response)
        return
    }

    // Calculate pagination metadata
    totalPages := (total + perPage - 1) / perPage
    isLast := page >= totalPages

    pagination := replify.Pages().
        WithPage(page).
        WithPerPage(perPage).
        WithTotalItems(total).
        WithTotalPages(totalPages).
        WithIsLast(isLast)

    response := replify.New().
        WithStatusCode(200).
        WithBody(users).
        WithPagination(pagination).
        WithTotal(len(users)).
        WithMessage("Users retrieved successfully").
        WithPath(r.URL.Path).
        WithRequestID(r.Header.Get("X-Request-ID"))

    respondJSON(w, response)
}
```

### Example 3: Error Handling with Stack Traces

```go
func ProcessOrder(w http.ResponseWriter, r *http.Request) {
    order, err := processOrderLogic(r)

    response := replify.New()

    if err != nil {
        response.
            WithStatusCode(500).
            WithErrorAck(err).
            WithMessage("Order processing failed")

        // Add debug info in development
        if os.Getenv("ENV") == "development" {
            response.
                WithDebuggingKV("timestamp", time.Now()).
                WithDebuggingKV("stack_trace", err.Error()).
                WithDebuggingKV("order_data", order)
        }
    } else {
        response.
            WithStatusCode(200).
            WithBody(order).
            WithMessage("Order processed successfully")
    }

    respondJSON(w, response)
}
```

## API Reference

### Wrapper Type (R)

```go
type R struct {
    *wrapper
}
```

The `R` type is a high-level abstraction providing a simplified interface for handling API responses.

### Core Functions

| Function                                       | Description                     |
| ---------------------------------------------- | ------------------------------- |
| `New() *wrapper`                               | Creates a new response wrapper  |
| `Pages() *pagination`                          | Creates a new pagination object |
| `UnwrapJSON(jsonStr string) (*wrapper, error)` | Parses JSON string to wrapper   |

### Configuration Methods

#### Response Configuration

| Method                                            | Description                        |
| ------------------------------------------------- | ---------------------------------- |
| `WithStatusCode(code int)`                        | Sets HTTP status code              |
| `WithBody(v interface{})`                         | Sets response body/data            |
| `WithMessage(message string)`                     | Sets response message              |
| `WithMessagef(format string, args...)`            | Sets formatted message             |
| `WithError(message string)`                       | Sets error message                 |
| `WithErrorf(format string, args...)`              | Sets formatted error               |
| `WithErrorAck(err error)`                         | Sets error with stack trace        |
| `AppendError(err error, message string)`          | Wraps error with context           |
| `AppendErrorf(err error, format string, args...)` | Wraps error with formatted context |
| `WithPath(v string)`                              | Sets request path                  |
| `WithPathf(v string, args...)`                    | Sets formatted request path        |
| `WithTotal(total int)`                            | Sets total items count             |

#### Metadata Methods

| Method                                             | Description                 |
| -------------------------------------------------- | --------------------------- |
| `WithRequestID(v string)`                          | Sets request ID             |
| `WithRequestIDf(format string, args...)`           | Sets formatted request ID   |
| `WithApiVersion(v string)`                         | Sets API version            |
| `WithApiVersionf(format string, args...)`          | Sets formatted API version  |
| `WithLocale(v string)`                             | Sets locale (e.g., "en_US") |
| `WithRequestedTime(v time.Time)`                   | Sets request timestamp      |
| `WithCustomFieldKV(key string, value interface{})` | Adds custom metadata field  |
| `WithCustomFieldKVf(key, format string, args...)`  | Adds formatted custom field |
| `WithCustomFields(values map[string]interface{})`  | Sets multiple custom fields |
| `WithMeta(v *meta)`                                | Sets entire metadata object |
| `WithHeader(v *header)`                            | Sets the header             |

#### Pagination Methods

| Method                          | Description                  |
| ------------------------------- | ---------------------------- |
| `WithPagination(v *pagination)` | Sets pagination object       |
| `WithPage(v int)`               | Sets current page number     |
| `WithPerPage(v int)`            | Sets items per page          |
| `WithTotalItems(v int)`         | Sets total items count       |
| `WithTotalPages(v int)`         | Sets total pages count       |
| `WithIsLast(v bool)`            | Sets if current page is last |

#### Debugging Methods

| Method                                           | Description                 |
| ------------------------------------------------ | --------------------------- |
| `WithDebugging(v map[string]interface{})`        | Sets debug information map  |
| `WithDebuggingKV(key string, value interface{})` | Adds single debug key-value |
| `WithDebuggingKVf(key, format string, args...)`  | Adds formatted debug value  |

### Query Methods

| Method                    | Returns                  | Description                   |
| ------------------------- | ------------------------ | ----------------------------- |
| `Available()`             | `bool`                   | Checks if wrapper is non-nil  |
| `StatusCode()`            | `int`                    | Gets HTTP status code         |
| `StatusText()`            | `string`                 | Gets status text (e.g., "OK") |
| `Body()`                  | `interface{}`            | Gets response body            |
| `Message()`               | `string`                 | Gets response message         |
| `Error()`                 | `string`                 | Gets error message            |
| `Cause()`                 | `error`                  | Gets underlying error cause   |
| `Total()`                 | `int`                    | Gets total items              |
| `Meta()`                  | `*meta`                  | Gets metadata object          |
| `Header()`                | `*header`                | Gets header object            |
| `Pagination()`            | `*pagination`            | Gets pagination object        |
| `Debugging()`             | `map[string]interface{}` | Gets debug information        |
| `OnDebugging(key string)` | `interface{}`            | Gets specific debug value     |

### Conditional Check Methods

| Method                              | Returns | Description                                 |
| ----------------------------------- | ------- | ------------------------------------------- |
| `IsSuccess()`                       | `bool`  | Checks if status is 2xx                     |
| `IsClientError()`                   | `bool`  | Checks if status is 4xx                     |
| `IsServerError()`                   | `bool`  | Checks if status is 5xx                     |
| `IsRedirection()`                   | `bool`  | Checks if status is 3xx                     |
| `IsError()`                         | `bool`  | Checks if error exists or status is 4xx/5xx |
| `IsErrorPresent()`                  | `bool`  | Checks if error field exists                |
| `IsBodyPresent()`                   | `bool`  | Checks if body exists                       |
| `IsPagingPresent()`                 | `bool`  | Checks if pagination exists                 |
| `IsMetaPresent()`                   | `bool`  | Checks if metadata exists                   |
| `IsHeaderPresent()`                 | `bool`  | Checks if header exists                     |
| `IsDebuggingPresent()`              | `bool`  | Checks if debug info exists                 |
| `IsDebuggingKeyPresent(key string)` | `bool`  | Checks if specific debug key exists         |
| `IsLastPage()`                      | `bool`  | Checks if current page is last              |
| `IsStatusCodePresent()`             | `bool`  | Checks if valid status code exists          |
| `IsTotalPresent()`                  | `bool`  | Checks if total count exists                |

### Serialization Methods

| Method         | Returns                  | Description                 |
| -------------- | ------------------------ | --------------------------- |
| `JSON()`       | `string`                 | Returns compact JSON string |
| `JSONPretty()` | `string`                 | Returns pretty-printed JSON |
| `Respond()`    | `map[string]interface{}` | Returns map representation  |
| `Reply()`      | `R`                      | Returns R wrapper           |

## HTTP Status Codes Reference

### Common API Scenarios

| **Scenario**                      | **HTTP Status Codes**                              | **Example**                             |
| --------------------------------- | -------------------------------------------------- | --------------------------------------- |
| **Successful Resource Retrieval** | 200 OK, 304 Not Modified                           | `GET /users/123` - Returns user data    |
| **Resource Creation**             | 201 Created                                        | `POST /users` - Creates a new user      |
| **Asynchronous Processing**       | 202 Accepted                                       | `POST /large-file` - File upload starts |
| **Validation Errors**             | 400 Bad Request                                    | `POST /users` - Missing required field  |
| **Authentication Issues**         | 401 Unauthorized, 403 Forbidden                    | Invalid credentials or permissions      |
| **Rate Limiting**                 | 429 Too Many Requests                              | Exceeded API request limits             |
| **Missing Resource**              | 404 Not Found                                      | `GET /users/999` - User not found       |
| **Server Failures**               | 500 Internal Server Error, 503 Service Unavailable | Database failure or maintenance         |
| **Version Conflicts**             | 409 Conflict                                       | Outdated version causing conflict       |

### Detailed Status Codes

#### Success (2xx)

| Code | Status          | Use Case                           |
| ---- | --------------- | ---------------------------------- |
| 200  | OK              | Successful GET, PUT, PATCH         |
| 201  | Created         | Successful POST (resource created) |
| 202  | Accepted        | Async processing started           |
| 204  | No Content      | Successful DELETE                  |
| 206  | Partial Content | Video streaming, range requests    |

#### Redirection (3xx)

| Code | Status             | Use Case                              |
| ---- | ------------------ | ------------------------------------- |
| 301  | Moved Permanently  | Resource permanently moved            |
| 302  | Found              | Temporary redirect                    |
| 304  | Not Modified       | Cached content still valid            |
| 307  | Temporary Redirect | POST redirect maintaining method      |
| 308  | Permanent Redirect | Permanent redirect maintaining method |

#### Client Errors (4xx)

| Code | Status                 | Use Case                       |
| ---- | ---------------------- | ------------------------------ |
| 400  | Bad Request            | Invalid request format/data    |
| 401  | Unauthorized           | Missing/invalid authentication |
| 403  | Forbidden              | Insufficient permissions       |
| 404  | Not Found              | Resource doesn't exist         |
| 409  | Conflict               | Resource conflict (duplicate)  |
| 413  | Payload Too Large      | Request body too large         |
| 415  | Unsupported Media Type | Invalid content type           |
| 422  | Unprocessable Entity   | Validation errors              |
| 429  | Too Many Requests      | Rate limiting                  |

#### Server Errors (5xx)

| Code | Status                | Use Case                 |
| ---- | --------------------- | ------------------------ |
| 500  | Internal Server Error | Unexpected server error  |
| 501  | Not Implemented       | Feature not implemented  |
| 502  | Bad Gateway           | Upstream service error   |
| 503  | Service Unavailable   | Service down/maintenance |
| 504  | Gateway Timeout       | Upstream timeout         |

## Best Practices

### ✅ Do's

1. **Always set status codes**

   ```go
   response := replify.New().
       WithStatusCode(200).
       WithBody(data)
   ```

2. **Use request IDs for tracing**

   ```go
   response := replify.New().
       WithRequestID(r.Header.Get("X-Request-ID")).
       WithBody(data)
   ```

3. **Include API version**

   ```go
   response := replify.New().
       WithApiVersion("v1.0.0").
       WithBody(data)
   ```

4. **Use WithErrorAck for stack traces**

   ```go
   response := replify.New().
       WithStatusCode(500).
       WithErrorAck(err)
   ```

5. **Check response status before processing**

   ```go
   if response.IsSuccess() {
       processData(response.Body())
   }
   ```

6. **Use pagination for list endpoints**
   ```go
   pagination := replify.Pages().
       WithPage(page).
       WithPerPage(perPage).
       WithTotalItems(total)
   ```

### ❌ Don'ts

1. **Don't forget to set status codes**

   ```go
   // ❌ Bad
   response := replify.New().WithBody(data)

   // ✅ Good
   response := replify.New().WithStatusCode(200).WithBody(data)
   ```

2. **Don't expose sensitive debug info in production**

   ```go
   // ❌ Bad
   response := replify.New().
       WithDebuggingKV("database_password", dbPass)

   // ✅ Good
   if os.Getenv("ENV") == "development" {
       response.WithDebuggingKV("query", sqlQuery)
   }
   ```

3. **Don't use generic error messages**

   ```go
   // ❌ Bad
   WithError("Error occurred")

   // ✅ Good
   WithError("Failed to create user: email already exists")
   ```

4. **Don't ignore error checking**

   ```go
   // ❌ Bad
   wrapper, _ := replify.UnwrapJSON(jsonStr)

   // ✅ Good
   wrapper, err := replify.UnwrapJSON(jsonStr)
   if err != nil {
       log.Printf("Failed to parse JSON: %v", err)
   }
   ```

## Use Cases

### ✅ When to Use

- **RESTful API Development** - Standardizing API responses
- **Microservices** - Consistent responses across services
- **API Versioning** - Including version metadata
- **Error Standardization** - Consistent error formats
- **Pagination** - APIs returning paginated results
- **Multi-tenant APIs** - Including tenant/locale information
- **Request Tracing** - Tracking requests across services
- **Development Debugging** - Conditional debug information

### ❌ When Not to Use

- **GraphQL APIs** - GraphQL has its own response format
- **gRPC Services** - Protocol Buffers define the structure
- **WebSocket APIs** - Real-time bidirectional communication
- **Simple CLIs** - Overkill for command-line tools
- **Internal Services** - Where custom formats are required
- **High-Performance** - Direct JSON encoding may be faster

## fj Usage Guide

`fj` (_Fast JSON_) is the JSON path-extraction engine embedded in **replify**. It lets you read, query, and transform values from a JSON document **without unmarshalling the entire structure** into Go types. It lives in `pkg/fj` and is exposed through the `wrapper` type in `parser.go`.

### Purpose in the replify Architecture

When a `wrapper` carries a JSON body, `fj` powers every field-level query on that body. Instead of decoding the whole payload into a `map[string]any` or a concrete struct, `fj` walks the raw string just far enough to locate the requested path. This keeps allocations low and throughput high on hot request paths.

```
HTTP Request → wrapper.WithBody(data) → wrapper.QueryJSONBody("user.name")
                                                 ↓
                                    fj.Get(jsonString, "user.name")
                                                 ↓
                                          fj.Context  ← single value, no full decode
```

### When to Use fj Instead of encoding/json

| Scenario                                                | Recommended approach               |
| ------------------------------------------------------- | ---------------------------------- |
| Extract one or a few fields from a large response body  | `fj` / `QueryJSONBody`             |
| Validate that the body is well-formed JSON              | `fj.IsValidJSON` / `ValidJSONBody` |
| Search leaf values or keys across an unknown schema     | `fj.Search` / `SearchJSONBody*`    |
| Apply streaming transforms (pretty-print, minify, etc.) | `fj` transformers                  |
| Bind the full payload into a typed struct               | `encoding/json` or `json-iterator` |
| Write or modify JSON                                    | `encoding/json`                    |
| JSON schema validation                                  | a dedicated schema library         |

### Path Syntax Quick Reference

```
user.name              field access
roles.0                array index
roles.#                array length
roles.#.name           collect field from every element
roles.#(role=="admin") first element where role == "admin"
roles.#(role=="admin")# all elements where role == "admin"
{id,name}              multi-selector → new object
[id,name]              multi-selector → new array
name.@uppercase        built-in transformer
name.@word:upper       transformer with argument
..title                recursive descent (JSON Lines / deep scan)
```

Dots and wildcards in key names can be escaped with a backslash (`\`).

### Core API

#### Direct fj usage

```go
import "github.com/polarixa/replify/pkg/fj"

json := `{
    "user": {"name": "Alice", "age": 30, "active": true},
    "roles": ["admin", "editor"],
    "scores": [95, 87, 92]
}`

// Single path
name := fj.Get(json, "user.name").String()    // "Alice"
age  := fj.Get(json, "user.age").Int64()      // 30
ok   := fj.Get(json, "user.active").Bool()    // true
n    := fj.Get(json, "roles.#").Int()         // 2 (array length)

// Multiple paths in one pass
results := fj.GetMulti(json, "user.name", "user.age", "roles.#")
// results[0].String() == "Alice", results[1].Int64() == 30, results[2].Int() == 2

// Check presence before use
if ctx := fj.Get(json, "user.email"); ctx.Exists() {
    fmt.Println(ctx.String())
}

// Parse a document once, query multiple times (avoids re-parsing)
doc := fj.Parse(json)
fmt.Println(doc.Get("user.name").String())
fmt.Println(doc.Get("roles.0").String())
```

#### Zero-copy byte-slice access

`GetBytes` is preferred when you already hold a `[]byte`. It uses `unsafe` pointer operations internally to avoid an extra string allocation:

```go
rawBytes := []byte(`{"id":42,"status":"active"}`)

id     := fj.GetBytes(rawBytes, "id").Int()        // 42
status := fj.GetBytes(rawBytes, "status").String() // "active"

// Multiple paths from bytes
res := fj.GetBytesMulti(rawBytes, "id", "status")
```

> **Memory note**: `fj.Context.Raw()` returns a substring view of the original string without copying. Do not hold a reference to the `Context` after the source string has been released; the backing memory will be reclaimed.

### Wrapper Integration (parser.go)

The `wrapper` type exposes all `fj` operations without requiring you to import `pkg/fj` directly in most cases:

```go
	response := replify.New().
		WithStatusCode(200).
		WithBody(map[string]any{
			"user": map[string]any{"name": "Alice", "role": "admin"},
			"items": []map[string]any{
				{"id": 1, "price": 9.99},
				{"id": 2, "price": 4.50},
			},
		})

	// Single path query
	name := response.QueryJSONBody("user.name").String() // "Alice"

	// Multiple paths in one call (one JSON serialization)
	fields := response.QueryJSONBodyMulti("user.name", "user.role")
	fmt.Println(fields[0].String(), fields[1].String()) // Alice admin

	// Parse the body once and chain subsequent queries
	ctx := response.JSONBodyParser()
	fmt.Println(ctx.Get("user.name").String())
	fmt.Println(ctx.Get("items.#").Int()) // array length

	// Validate the body
	if !response.ValidJSONBody() {
		log.Println("body is not valid JSON")
		return
	}

	// Aggregate helpers
	total := response.SumJSONBody("items.#.price")  // 14.49
	min, _ := response.MinJSONBody("items.#.price") // 4.50
	max, _ := response.MaxJSONBody("items.#.price") // 9.99
	avg, _ := response.AvgJSONBody("items.#.price") // 7.245

	fmt.Println(name)
	fmt.Println(fields[0].String(), fields[1].String())
	fmt.Println(ctx.Get("user.name").String())
	fmt.Println(ctx.Get("items.#").Int())
	fmt.Println(total)
	fmt.Println(min)
	fmt.Println(max)
	fmt.Println(avg)
```

> **Performance tip**: `QueryJSONBody` serializes the body on every call. For repeated queries on the same body, call `JSONBodyParser()` once and reuse the returned `fj.Context`.

### Context Value Extraction

A `fj.Context` is returned by every query. Always call `.Exists()` before using the value if the path might be absent.

```go
ctx := fj.Get(json, "optional.field")

ctx.Exists()   // false when path is missing
ctx.Kind()     // fj.Null | fj.String | fj.Number | fj.True | fj.False | fj.JSON
ctx.String()   // string representation
ctx.Bool()     // bool
ctx.Int()      // int
ctx.Int64()    // int64
ctx.Float64()  // float64
ctx.Raw()      // raw JSON token (no allocation)
ctx.IsArray()  // true when kind == JSON and raw starts with '['
ctx.IsObject() // true when kind == JSON and raw starts with '{'
ctx.IsError()  // true if parsing produced an error
ctx.Cause()    // error string, or "" if no error

// Iterate array values
ctx.Foreach(func(key, val fj.Context) bool {
    fmt.Println(val.String())
    return true // return false to stop
})
```

### Transformers

Transformers are applied with the `@` prefix inside a path expression and receive the current JSON value as input. An optional argument is passed after a `:` separator.

```
path.@transformerName
path.@transformerName:argument
path.@transformerName:{"key":"value"}
```

#### Core transformers

| Transformer | Alias(es) | Input            | Description                                                                                               |
| ----------- | --------- | ---------------- | --------------------------------------------------------------------------------------------------------- |
| `@pretty`   | —         | any              | Pretty-print (indented) JSON. Accepts optional `{"sort_keys":true,"indent":"\t","prefix":"","width":80}`. |
| `@minify`   | `@ugly`   | any              | Compact single-line JSON (all whitespace removed).                                                        |
| `@valid`    | —         | any              | Returns `"true"` / `"false"` — whether the input is valid JSON.                                           |
| `@this`     | —         | any              | Identity — returns the input unchanged.                                                                   |
| `@reverse`  | —         | array \| object  | Reverses element order (array) or key order (object).                                                     |
| `@flatten`  | —         | array            | Shallow-flatten nested arrays. Pass `{"deep":true}` to recurse.                                           |
| `@join`     | —         | array of objects | Merge an array of objects into one object. Pass `{"preserve":true}` to keep duplicate keys.               |
| `@keys`     | —         | object           | Return a JSON array of the object's keys.                                                                 |
| `@values`   | —         | object           | Return a JSON array of the object's values.                                                               |
| `@group`    | —         | object of arrays | Zip object-of-arrays into an array-of-objects.                                                            |
| `@search`   | —         | any              | `@search:path` — collect all values reachable at `path` anywhere in the tree.                             |
| `@json`     | —         | string           | Parse the string as JSON and return the value.                                                            |
| `@string`   | —         | any              | Encode the value as a JSON string literal.                                                                |

#### String transformers

| Transformer   | Alias(es)              | Description                                                                   |
| ------------- | ---------------------- | ----------------------------------------------------------------------------- |
| `@uppercase`  | `@upper`               | Convert all characters to upper-case.                                         |
| `@lowercase`  | `@lower`               | Convert all characters to lower-case.                                         |
| `@flip`       | —                      | Reverse the characters of the string.                                         |
| `@trim`       | —                      | Strip leading/trailing whitespace.                                            |
| `@snakecase`  | `@snake`, `@snakeCase` | Convert to `snake_case`.                                                      |
| `@camelcase`  | `@camel`, `@camelCase` | Convert to `camelCase`.                                                       |
| `@kebabcase`  | `@kebab`, `@kebabCase` | Convert to `kebab-case`.                                                      |
| `@replace`    | —                      | `@replace:{"target":"old","replacement":"new"}` — replace first occurrence.   |
| `@replaceAll` | —                      | `@replaceAll:{"target":"old","replacement":"new"}` — replace all occurrences. |
| `@hex`        | —                      | Hex-encode the value.                                                         |
| `@bin`        | —                      | Binary-encode the value.                                                      |
| `@insertAt`   | —                      | `@insertAt:{"index":5,"insert":"XYZ"}` — insert a substring at position.      |
| `@wc`         | —                      | Return the word-count of a string as an integer.                              |
| `@padLeft`    | —                      | `@padLeft:{"padding":"*","length":10}` — left-pad to a fixed width.           |
| `@padRight`   | —                      | `@padRight:{"padding":"*","length":10}` — right-pad to a fixed width.         |

#### Object transformers

| Transformer | Description                                                                                                                                                     |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `@project`  | Pick and/or rename fields from an object. Arg: `{"pick":["f1","f2"],"rename":{"f1":"newName"}}`. Omit `pick` to keep all fields; omit `rename` for no renaming. |
| `@default`  | Inject fallback values for fields that are absent or `null`. Arg: `{"field":"defaultValue",...}`. Existing non-null fields are never overwritten.               |

#### Array transformers

| Transformer | Description                                                                                                                                                   |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `@filter`   | Keep only elements matching a condition. Arg: `{"key":"field","op":"eq","value":val}`. Operators: `eq` (default), `ne`, `gt`, `gte`, `lt`, `lte`, `contains`. |
| `@pluck`    | Extract a named field (supports dot-notation paths) from every element. Arg: field path string, e.g. `@pluck:name` or `@pluck:addr.city`.                     |
| `@first`    | Return the first element of the array, or `null` if empty.                                                                                                    |
| `@last`     | Return the last element of the array, or `null` if empty.                                                                                                     |
| `@count`    | Return the number of elements (array) or key-value pairs (object) as an integer. Scalars return `0`.                                                          |
| `@sum`      | Sum all numeric values in the array; non-numeric elements are skipped. Returns `0` for empty arrays.                                                          |
| `@min`      | Return the minimum numeric value in the array. Returns `null` when no numbers are present.                                                                    |
| `@max`      | Return the maximum numeric value in the array. Returns `null` when no numbers are present.                                                                    |

#### Value normalization transformers

| Transformer | Description                                                                                                                                  |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `@coerce`   | Convert a scalar to a target type. Arg: `{"to":"string"}`, `{"to":"number"}`, or `{"to":"bool"}`. Objects and arrays are returned unchanged. |

#### Examples

```go
json := `{
    "user": {"name": "Alice", "role": null, "age": 30, "city": "NY"},
    "scores": [95, 87, 92, 78],
    "users": [
        {"name": "Alice", "active": true,  "addr": {"city": "NY"}},
        {"name": "Bob",   "active": false, "addr": {"city": "LA"}},
        {"name": "Carol", "active": true,  "addr": {"city": "NY"}}
    ]
}`

// ── Core ─────────────────────────────────────────────────────────────────────
fj.Get(json, "@pretty").String()             // indented JSON
fj.Get(json, "@minify").String()             // compact JSON
fj.Get(json, "user.@keys").String()          // ["name","role","age","city"]
fj.Get(json, "user.@values").String()        // ["Alice",null,30,"NY"]
fj.Get(json, "user.@valid").String()         // "true"

// ── String ───────────────────────────────────────────────────────────────────
fj.Get(json, "user.name.@uppercase").String()   // "ALICE"
fj.Get(json, "user.name.@reverse").String()     // "ecilA"
fj.Get(json, "user.name.@snakecase").String()   // "alice"
fj.Get(json, "user.city.@padLeft:{\"padding\":\"0\",\"length\":6}").String() // "000 NY"

// ── Object ───────────────────────────────────────────────────────────────────

// Project: keep only name and age, rename age → years
fj.Get(json, `user.@project:{"pick":["name","age"],"rename":{"age":"years"}}`).Raw()
// → {"name":"Alice","years":30}

// Default: fill in missing / null fields
fj.Get(json, `user.@default:{"role":"viewer","active":true}`).Raw()
// → {"name":"Alice","role":"viewer","age":30,"city":"NY","active":true}

// ── Array ────────────────────────────────────────────────────────────────────

// Filter: keep only active users
fj.Get(json, `users.@filter:{"key":"active","value":true}`).Raw()
// → [{"name":"Alice","active":true,...},{"name":"Carol","active":true,...}]

// Pluck: extract the city from every user's address
fj.Get(json, `users.@pluck:addr.city`).Raw()
// → ["NY","LA","NY"]

// Aggregation helpers
fj.Get(json, "scores.@first").Raw()    // 95
fj.Get(json, "scores.@last").Raw()     // 78
fj.Get(json, "scores.@count").Raw()    // 4
fj.Get(json, "scores.@sum").Raw()      // 352
fj.Get(json, "scores.@min").Raw()      // 78
fj.Get(json, "scores.@max").Raw()      // 95

// ── Coerce ───────────────────────────────────────────────────────────────────
fj.Get(`42`,   `@coerce:{"to":"string"}`).Raw()  // "42"
fj.Get(`"99"`, `@coerce:{"to":"number"}`).Raw()  // 99
fj.Get(`1`,    `@coerce:{"to":"bool"}`).Raw()    // true
```

#### Composing transformers

Transformers can be chained using the `|` pipe operator or dot notation:

```go
// First filter the array, then count the remaining elements
fj.Get(json, `users.@filter:{"key":"active","value":true}|@count`).Raw()
// → 2

// Pluck names, then reverse the resulting array
fj.Get(json, `users.@pluck:name|@reverse`).Raw()
// → ["Carol","Bob","Alice"]
```

#### Complex real-world examples

The following scenarios demonstrate how to combine multiple transformers into a single expression to process realistic JSON payloads.

---

**Example 1 — E-commerce product catalog: filter, aggregate, and shape**

```go
catalog := `{
    "products": [
        {"id":"p1","name":"Laptop Pro",    "category":"electronics","price":1299.99,"stock":5},
        {"id":"p2","name":"USB-C Hub",     "category":"electronics","price":49.99,  "stock":120},
        {"id":"p3","name":"Desk Chair",    "category":"furniture",  "price":349.00, "stock":0},
        {"id":"p4","name":"Standing Desk", "category":"furniture",  "price":699.00, "stock":3},
        {"id":"p5","name":"Webcam HD",     "category":"electronics","price":89.99,  "stock":45}
    ]
}`

// All in-stock electronics names
fj.Get(catalog, `products.@filter:{"key":"category","value":"electronics"}|@filter:{"key":"stock","op":"gt","value":0}|@pluck:name`).Raw()
// → ["Laptop Pro","USB-C Hub","Webcam HD"]

// Count of in-stock products
fj.Get(catalog, `products.@filter:{"key":"stock","op":"gt","value":0}|@count`).Raw()
// → 4

// Price range of in-stock products
fj.Get(catalog, `products.@filter:{"key":"stock","op":"gt","value":0}|@pluck:price|@min`).Raw()
// → 49.99
fj.Get(catalog, `products.@filter:{"key":"stock","op":"gt","value":0}|@pluck:price|@max`).Raw()
// → 1299.99

// Project the first in-stock product as a display card (pick and rename fields)
first := fj.Get(catalog, `products.@filter:{"key":"stock","op":"gt","value":0}|@first`).Raw()
fj.Get(first, `@project:{"pick":["name","price"],"rename":{"name":"title","price":"cost"}}`).Raw()
// → {"title":"Laptop Pro","cost":1299.99}
```

---

**Example 2 — API response normalization: fill defaults then project and rename**

```go
// Raw user record from an external API with null / absent fields
rawUser := `{"id":"u1","name":"Alice","role":null,"verified":null}`

// One-shot normalization: fill nulls → keep only safe fields → rename id for the frontend
fj.Get(rawUser, `@default:{"role":"viewer","verified":false}|@project:{"pick":["id","name","role","verified"],"rename":{"id":"userId"}}`).Raw()
// → {"userId":"u1","name":"Alice","role":"viewer","verified":false}
```

---

**Example 3 — Log processing: filter, count, and retrieve the latest entry**

```go
logs := `[
    {"level":"error","msg":"Connection refused","ts":1700001},
    {"level":"info", "msg":"Server started",    "ts":1700002},
    {"level":"error","msg":"Timeout exceeded",  "ts":1700003},
    {"level":"warn", "msg":"High memory",       "ts":1700004}
]`

// How many errors?
fj.Get(logs, `@filter:{"key":"level","value":"error"}|@count`).Raw()
// → 2

// All error messages
fj.Get(logs, `@filter:{"key":"level","value":"error"}|@pluck:msg`).Raw()
// → ["Connection refused","Timeout exceeded"]

// Most recent error entry (last in the filtered array)
fj.Get(logs, `@filter:{"key":"level","value":"error"}|@last`).Raw()
// → {"level":"error","msg":"Timeout exceeded","ts":1700003}
```

---

**Example 4 — Nested data aggregation: filter → pluck → flatten → sum**

```go
teamData := `{
    "teams": [
        {"name":"Alpha","active":true, "monthly_revenue":[10000,12000,11000]},
        {"name":"Beta", "active":false,"monthly_revenue":[8000,9000,8500]},
        {"name":"Gamma","active":true, "monthly_revenue":[15000,16000,14000]}
    ]
}`

// Total revenue across all active teams, flattening the per-team monthly arrays first
fj.Get(teamData, `teams.@filter:{"key":"active","value":true}|@pluck:monthly_revenue|@flatten|@sum`).Raw()
// → 78000   (Alpha: 33000 + Gamma: 45000)
```

---

**Example 5 — URL-slug generation from a display name**

```go
// Multi-word title with duplicate internal spaces → URL-safe kebab-case slug
fj.Get(`"My   Blog Post Title"`, `@trim|@lowercase|@kebabcase`).Raw()
// → "my-blog-post-title"

// Author name to lowercase slug
fj.Get(`"John Doe"`, `@lowercase|@replace:{"target":" ","replacement":"-"}`).Raw()
// → "john-doe"
```

---

**Example 6 — Config merging and introspection**

```go
// Merge two partial config objects; later values overwrite earlier ones for duplicate keys
overrides := `[{"host":"localhost","port":5432},{"port":5433,"ssl":true}]`

merged := fj.Get(overrides, `@join`).Raw()
// → {"host":"localhost","port":5433,"ssl":true}

// Inspect which keys are present after the merge
fj.Get(merged, `@keys`).Raw()
// → ["host","port","ssl"]

// Count the merged keys
fj.Get(merged, `@count`).Raw()
// → 3

// Project only the connection-relevant subset and rename for the driver
fj.Get(merged, `@project:{"pick":["host","port"],"rename":{"port":"dbPort"}}`).Raw()
// → {"host":"localhost","dbPort":5433}
```

---

**Example 7 — Leaderboard: zip parallel arrays, filter, and pluck**

```go
// Two parallel arrays zipped via @group into an array-of-objects, then filtered and plucked
leaderboard := `{"player":["Alice","Bob","Carol","Dave"],"score":[98,72,85,91]}`

// Zip the parallel arrays into objects
grouped := fj.Get(leaderboard, `@group`).Raw()
// → [{"player":"Alice","score":98},{"player":"Bob","score":72},
//    {"player":"Carol","score":85},{"player":"Dave","score":91}]

// Players with a score of 85 or above
fj.Get(grouped, `@filter:{"key":"score","op":"gte","value":85}|@pluck:player`).Raw()
// → ["Alice","Carol","Dave"]

// Top player's full record
fj.Get(grouped, `@filter:{"key":"score","op":"gte","value":95}|@first`).Raw()
// → {"player":"Alice","score":98}
```

---

#### Registering custom transformers

```go
func init() {
    fj.AddTransformer("redact", fj.TransformerFunc(func(json, arg string) string {
        return `"[REDACTED]"`
    }))
}

// Usage in path
fj.Get(json, "user.password.@redact").String() // "[REDACTED]"
```

Transformers can be disabled globally with `fj.DisableTransformers = true`.

### Search and Scan Helpers

```go
// Full-tree substring search across all leaf values
hits := response.SearchJSONBody("admin")

// Wildcard scan of leaf values
hits = response.SearchJSONBodyMatch("err*")

// Find all values stored under specific key names
emails := response.SearchJSONBodyByKey("email")

// Find all values under keys matching a wildcard
hits = response.SearchJSONBodyByKeyPattern("user*")

// Substring / wildcard check at a specific path
response.JSONBodyContains("user.role", "admin")
response.JSONBodyContainsMatch("user.email", "*@example.com")

// Return the dot-notation path where a value first appears
path := response.FindJSONBodyPath("alice@example.com")

// All paths where value matches a pattern
paths := response.FindJSONBodyPathsMatch("err*")
```

### Data Manipulation Helpers

```go
import "github.com/polarixa/replify/pkg/fj"

// Count elements at a path
n := response.CountJSONBody("items")

// Filter array elements by predicate
active := response.FilterJSONBody("users", func(ctx fj.Context) bool {
    return ctx.Get("active").Bool()
})

// First match
admin := response.FirstJSONBody("users", func(ctx fj.Context) bool {
    return ctx.Get("role").String() == "admin"
})

// Deduplicate (first-occurrence order preserved)
tags := response.DistinctJSONBody("tags")

// Project fields from an array of objects
rows := response.PluckJSONBody("users", "id", "email")

// Group by a key field
byRole := response.GroupByJSONBody("users", "role")

// Sort array by a field (numeric or string comparison)
sorted := response.SortJSONBody("products", "price", true)
```

### Limitations

- **Read-only**: `fj` cannot write or modify JSON. Use `encoding/json` for serialization.
- **No schema validation**: For strict schema enforcement use a dedicated library.
- **No struct binding**: `fj` returns `Context` values, not typed Go structs. Use `encoding/json` when binding is required.
- **`Raw()` lifetime**: The raw string returned by `Context.Raw()` is a zero-copy view into the source JSON string. It must not outlive the original string.
- **`UnsafeBytes`**: The byte slice returned by `fj.UnsafeBytes` shares memory with the source string. Never mutate it, as this violates Go's string immutability guarantees and can cause undefined behavior.
- **Malformed input**: `fj` does not validate JSON before parsing. Pass untrusted input through `fj.IsValidJSON` or `ValidJSONBody()` first.
- **Transformers are global**: `AddTransformer` writes to a package-level registry. Register all transformers during program initialization (e.g., in `init()` functions) before concurrent access begins to avoid data races.

### Best Practices

1. **Check existence before use**

   ```go
   if ctx := response.QueryJSONBody("optional.key"); ctx.Exists() {
       process(ctx.String())
   }
   ```

2. **Parse once, query many times**

   ```go
   doc := response.JSONBodyParser()
   id    := doc.Get("user.id").String()
   email := doc.Get("user.email").String()
   role  := doc.Get("user.role").String()
   ```

3. **Prefer `GetBytes` for byte-slice payloads**

   ```go
   // ✅ avoids string conversion allocation
   ctx := fj.GetBytes(rawBytes, "user.name")

   // ❌ unnecessary allocation
   ctx = fj.Get(string(rawBytes), "user.name")
   ```

4. **Validate untrusted input first**

   ```go
   if !response.ValidJSONBody() {
       return errors.New("invalid JSON body")
   }
   ```

5. **Register custom transformers in `init()`**

   ```go
   func init() {
       fj.AddTransformer("mask", func(json, arg string) string {
           return `"***"`
       })
   }
   ```

6. **Never mutate `UnsafeBytes` output**

   ```go
   b := fj.UnsafeBytes(someString)
   // ✅ read-only access
   _ = b[0]
   // ❌ mutating b corrupts the original string
   ```

## Contributing

To contribute to this project, follow these steps:

1. **Clone the repository**

   ```bash
   git clone --depth 1 https://github.com/polarixa/replify.git
   ```

2. **Navigate to the project directory**

   ```bash
   cd replify
   ```

3. **Prepare the project environment**

   ```bash
   go mod tidy
   ```

4. **Make your changes**
   - Follow Go best practices
   - Add tests for new features
   - Update documentation

5. **Run tests**

   ```bash
   go test ./...
   ```

6. **Submit a pull request**

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Packages

Part of the **replify** ecosystem:

- [replify](https://github.com/polarixa/replify) - API response wrapping library (this package)
- [conv](https://github.com/polarixa/replify/pkg/conv) - Type conversion utilities
- [coll](https://github.com/polarixa/replify/pkg/coll) - Type-safe collection utilities
- [common](https://github.com/polarixa/replify/pkg/common) - Reflection-based utilities
- [encoding](https://github.com/polarixa/replify/pkg/encoding) - JSON encoding utilities
- [hashy](https://github.com/polarixa/replify/pkg/hashy) - Deterministic hashing
- [match](https://github.com/polarixa/replify/pkg/match) - Wildcard pattern matching
- [msort](https://github.com/polarixa/replify/pkg/msort) - Map sorting utilities
- [randn](https://github.com/polarixa/replify/pkg/randn) - Random data generation
- [ref](https://github.com/polarixa/replify/pkg/ref) - Pointer utilities
- [strutil](https://github.com/polarixa/replify/pkg/strutil) - String utilities
- [truncate](https://github.com/polarixa/replify/pkg/truncate) - String truncation utilities

## Support

- **Issues**: [GitHub Issues](https://github.com/polarixa/replify/issues)
- **Discussions**: [GitHub Discussions](https://github.com/polarixa/replify/discussions)

## Acknowledgments

Built with ❤️ for the Go community.
