# fj

**fj** (_Fast JSON_) is a Go package that provides a fast and simple way to retrieve, query, and transform values from a JSON document without unmarshalling the entire structure into Go types.

## Overview

The `fj` package uses a dot-notation path syntax that supports wildcards, array indexing, conditional queries, multi-selectors, pipe operators, and a rich set of built-in transformers. Custom transformers can be registered at runtime without modifying the core library.

**Key Features:**

- ⚡ **Zero-allocation hot paths** — `Get` and `GetBytes` minimize heap pressure
- 🔍 **Rich path syntax** — wildcards, array indexing, queries, multi-selectors, literals
- 🔧 **Built-in transformers** — 28+ transformers including `@pretty`, `@minify`, `@reverse`, `@flatten`, `@join`, `@snakeCase`, `@camelCase`, and more
- 🧩 **Extensible** — register custom transformers via `AddTransformer`
- 🎨 **JSON color output** — 27 named color styles for terminal display
- 🔒 **Concurrency-safe** — all public API functions are safe for concurrent goroutine use
- 📁 **File and stream support** — parse JSON from files and `io.Reader` sources

## Use Cases

### When to Use

- ✅ Extracting specific fields from large JSON documents without full unmarshalling
- ✅ Building lightweight JSON pipelines with composable transformers
- ✅ Quickly querying deeply nested or dynamic JSON structures
- ✅ Colorizing JSON output for terminal applications
- ✅ Validating JSON at the edge before heavy processing

### When Not to Use

- ❌ When you need full struct binding (use `encoding/json` or `json-iterator`)
- ❌ When you need to _write_ or _modify_ JSON (use `encoding/json`)
- ❌ When JSON schema validation is required (use a dedicated schema library)

## Getting Started

### Requirements

Go version **1.19** or higher.

### Installation

```bash
go get github.com/polarixa/replify
```

Import the sub-package in your Go code:

```go
import "github.com/polarixa/replify/pkg/fj"
```

## Usage

### Basic Field Access

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/fj"
)

func main() {
    json := `{
        "user": {
            "name": "Alice",
            "age": 30,
            "active": true,
            "roles": ["Admin", "Editor"]
        }
    }`

    fmt.Println(fj.Get(json, "user.name").String())    // Alice
    fmt.Println(fj.Get(json, "user.age").Int64())      // 30
    fmt.Println(fj.Get(json, "user.active").Bool())    // true
    fmt.Println(fj.Get(json, "user.roles.0").String()) // Admin
    fmt.Println(fj.Get(json, "user.roles.#").Int64())  // 2 (array length)
}
```

### Array Iteration

```go
json := `{"roles":[{"name":"Admin"},{"name":"Editor"}]}`

// Collect all names
ctx := fj.Get(json, "roles.#.name")
fmt.Println(ctx.String()) // ["Admin","Editor"]

// Iterate with Foreach
fj.Get(json, "roles").Foreach(func(_, val fj.Context) bool {
    fmt.Println(val.Get("name").String())
    return true
})
```

### Queries

```go
json := `{"items":[
    {"name":"apple","price":1.2},
    {"name":"banana","price":0.8},
    {"name":"cherry","price":3.5}
]}`

// First item with price > 1.0
fmt.Println(fj.Get(json, `items.#(price>1.0).name`).String()) // apple

// All items with price > 1.0
fmt.Println(fj.Get(json, `items.#(price>1.0)#.name`).String()) // ["apple","cherry"]
```

### Multi-Selectors

```go
json := `{"id":1,"name":"Alice","email":"alice@example.com","age":30}`

// Build a new object with selected fields
fmt.Println(fj.Get(json, `{id,name}`).String())
// {"id":1,"name":"Alice"}

// Build a new array with selected fields
fmt.Println(fj.Get(json, `[id,name]`).String())
// [1,"Alice"]
```

### Transformers

```go
json := `{"message":"  Hello, World!  "}`

// Trim whitespace
fmt.Println(fj.Get(json, "message.@trim"))

// Convert to uppercase
fmt.Println(fj.Get(json, "message.@uppercase"))

// Pretty-print JSON
data := `{"b":2,"a":1}`
fmt.Println(fj.Get(data, "@pretty"))

// Pretty-print with sorted keys
fmt.Println(fj.Get(data, `@pretty:{"sort_keys":true}`))

// Minify
fmt.Println(fj.Get(data, "@minify"))
```

### Working with Bytes

```go
data := []byte(`{"status":"ok","code":200}`)
ctx := fj.GetBytes(data, "code")
fmt.Println(ctx.Int64()) // 200
```

### Reading from a File

```go
content, err := fj.ParseJSONFile("/path/to/data.json")
if err != nil {
    log.Fatal(err)
}
ctx := fj.Get(content, "user.name")
fmt.Println(ctx.String())
```

### Validation

```go
fmt.Println(fj.IsValidJSON(`{"key":"value"}`)) // true
fmt.Println(fj.IsValidJSON(`{bad json}`))       // false
```

### Chaining

```go
json := `{"a":{"b":{"c":"deep"}}}`
ctx := fj.Get(json, "a").Get("b").Get("c")
fmt.Println(ctx.String()) // deep
```

### Working with Bytes (no-allocation)

When JSON is already in a `[]byte` slice, use `GetBytes` to avoid converting to string. To further avoid converting `ctx.Raw()` back to `[]byte`, use `ctx.Index()` as a zero-allocation sub-slice:

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/fj"
)

var data = []byte(`{"user":{"id":"12345","roles":[{"roleId":"1","roleName":"Admin"},{"roleId":"2","roleName":"Editor"}]}}`)

func main() {
    ctx := fj.GetBytes(data, "user.roles.#.roleName")

    // Zero-allocation sub-slice of the original []byte
    var raw []byte
    if ctx.Index() > 0 {
        raw = data[ctx.Index() : ctx.Index()+len(ctx.Raw())]
    } else {
        raw = []byte(ctx.Raw())
    }
    fmt.Println(string(raw)) // ["Admin","Editor"]
}
```

### Existence

Use `Exists()` to distinguish a missing path from an explicit JSON `null`:

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/fj"
)

var data = []byte(`{"user":{"id":"12345","name":{"firstName":"John","lastName":"Doe"}}}`)

func main() {
    ctx := fj.GetBytes(data, "user.name.firstName")
    if ctx.Exists() {
        fmt.Println(ctx.String()) // John
    } else {
        fmt.Println("not found")
    }
}
```

### Loop

`Foreach` iterates over array elements or object key-value pairs. Return `false` from the callback to stop early:

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/fj"
)

var data = []byte(`{"user":{"roles":[{"roleId":"1","roleName":"Admin"},{"roleId":"2","roleName":"Editor"}]}}`)

func main() {
    fj.GetBytes(data, "user.roles").Foreach(func(_, value fj.Context) bool {
        fmt.Println(value.Get("roleName").String())
        return true
    })
    // Admin
    // Editor
}
```

### Unmarshal

Use `Value()` to obtain a native Go type via a type assertion:

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/fj"
)

var data = []byte(`{"user":{"id":12345,"name":{"firstName":"John","lastName":"Doe"}}}`)

func main() {
    // As map[string]interface{}
    name, ok := fj.GetBytes(data, "user.name").Value().(map[string]interface{})
    if ok {
        fmt.Println(name) // map[firstName:John lastName:Doe]
    }

    // As float64
    id, ok := fj.GetBytes(data, "user.id").Value().(float64)
    if ok {
        fmt.Println(id) // 12345
    }

    // Direct typed accessors are often more ergonomic
    fmt.Println(fj.GetBytes(data, "user.id").Int64()) // 12345
}
```

### Parse & Get

`ParseBytes` parses JSON once; subsequent `Get` calls navigate inside the parsed result without re-scanning the entire document:

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/fj"
)

var data = []byte(`{"user":{"id":12345,"name":{"firstName":"John","lastName":"Doe"}}}`)

func main() {
    // All three are equivalent
    fmt.Println(fj.ParseBytes(data).Get("user").Get("id").Int64()) // 12345
    fmt.Println(fj.GetBytes(data, "user.id").Int64())              // 12345
    fmt.Println(fj.GetBytes(data, "user").Get("id").Int64())       // 12345
}
```

### JSON Lines (extended)

JSON Lines support uses the `..` prefix to treat a multi-line document as an array. Requires [JSON Lines](https://jsonlines.org/) format (one JSON object per line).

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/fj"
)

var lines = []byte(`
    {"roleId":"1","roleName":"Admin","permissions":[{"permissionId":"101","permissionName":"View Reports"},{"permissionId":"102","permissionName":"Manage Users"}]}
    {"roleId":"2","roleName":"Editor","permissions":[{"permissionId":"201","permissionName":"Edit Content"},{"permissionId":"202","permissionName":"View Analytics"}]}
`)

func main() {
    // Count lines
    fmt.Println(fj.ParseBytes(lines).Get("..#").String()) // 2

    // Get a specific line by index
    fmt.Println(fj.GetBytes(lines, "..0").Get("roleName").String()) // Admin

    // Pluck a field from every line
    fmt.Println(fj.GetBytes(lines, "..#.roleName").String()) // ["Admin","Editor"]

    // Nested query across all lines
    fmt.Println(fj.GetBytes(lines, `..#.permissions.#(permissionId=="101").permissionName`).String())
    // ["View Reports"]
}
```

## Path Syntax Reference

A `fj` path is a sequence of elements separated by `.`. In addition to `.`, several characters carry special meaning: `|`, `#`, `@`, `\`, `*`, `!`, and `?`.

| Syntax                              | Description                         | Example                               |
| ----------------------------------- | ----------------------------------- | ------------------------------------- |
| `field`                             | Object field access                 | `user.name`                           |
| `field.N`                           | Array index access (0-based)        | `roles.0`                             |
| `field.#`                           | Array length                        | `roles.#` → `3`                       |
| `field.#.key`                       | Pluck field from all elements       | `roles.#.name` → `["Admin","Editor"]` |
| `field.*`                           | Wildcard — matches any characters   | `us*.name`                            |
| `field.?`                           | Wildcard — matches exactly one char | `us?.name`                            |
| `field\.key`                        | Escape special character `.`        | `a\.b`                                |
| `field.#(k==v)`                     | Query: first match                  | `#(active==true).name`                |
| `field.#(k==v)#`                    | Query: all matches                  | `#(price>1)#.name`                    |
| `k%"pat"`                           | Query: like (wildcard match)        | `#(name%"A*")`                        |
| `k!%"pat"`                          | Query: not like                     | `#(name!%"A*")`                       |
| `~true` / `~false` / `~null` / `~*` | Tilde — boolean coercion in queries | `#(active==~true)`                    |
| `a.b` / `a\|b`                      | Dot and pipe separators             | `user.name` or `user\|name`           |
| `{f1,f2}`                           | Multi-selector → new object         | `{id,name}`                           |
| `[f1,f2]`                           | Multi-selector → new array          | `[id,name]`                           |
| `!"value"`                          | JSON literal in multi-selector      | `{"active":!true}`                    |
| `..field`                           | JSON Lines — query each line        | `..user.name`                         |

### Object & Array Examples

```
# Basic field access
id                              → "http://subs/base-sample-schema.json"
properties.alias.description    → "An unique identifier in a submission."
properties.alias.minLength      → 1
required                        → ["alias","taxonId","releaseDate"]
required.0                      → "alias"
required.1                      → "taxonId"
oneOf.0.required.1              → "team"

# Wildcards (* matches any sequence, ? matches one character)
anim*ls.1.name                  → "Barky"
*nimals.1.name                  → "Barky"

# Escape special characters
properties.alias\.description   → "An unique identifier in a submission."

# Array length and pluck
animals.#                       → 3
animals.#.name                  → ["Meowsy","Barky","Purrpaws"]
```

### Query Examples

Queries support `==`, `!=`, `<`, `<=`, `>`, `>=`, `%` (like), and `!%` (not like):

```
stock.#(price_2002==56.27).symbol                → "MMM"
stock.#(company=="Amazon.com").symbol             → "AMZN"
stock.#(initial_price>=10)#.symbol               → ["MMM","AMZN","CPB","DIS","DOW","XOM","F","GPS","GIS"]
stock.#(company%"D*")#.symbol                    → ["DIS","DOW"]
stock.#(company!%"D*")#.symbol                   → ["MMM","AMZN","CPB","XOM","F","GPS","GIS"]
required.#(%"*as*")#                             → ["alias","releaseDate"]
required.#(%"*as*")                              → "alias"
animals.#(foods.likes.#(%"*a*"))#.name           → ["Meowsy","Barky"]
```

### Tilde Operator

The `~` operator evaluates a value as a boolean before comparison. The most recent value that does not exist is treated as `false`:

```
~true    → interprets truthy values as true
~false   → interprets falsy and undefined values as true
~null    → interprets null and undefined values as true
~*       → interprets any defined value as true
```

```
bank.#(isActive==~true)#.name   → ["Davis Wade","Oneill Everett"]
bank.#(isActive==~false)#.name  → ["Stark Jenkins","Odonnell Rollins","Rachelle Chang","Dalton Waters"]
bank.#(eyeColor==~null)#.name   → ["Dalton Waters"]
bank.#(company==~*)#.name       → ["Stark Jenkins","Odonnell Rollins","Rachelle Chang","Davis Wade","Oneill Everett","Dalton Waters"]
```

### Dot & Pipe

The `.` is the default separator; `|` can also be used. They behave identically except after `#` inside array/query contexts:

```
bank.0.balance                         → "$1,404.23"
bank|0.balance                         → "$1,404.23"
bank.0|balance                         → "$1,404.23"
bank.#                                 → 6
bank.#(gender=="female")#|#            → 2
bank.#(gender=="female")#.name         → ["Rachelle Chang","Davis Wade"]
bank.#(gender=="female")#|name         → not-present
bank.#(gender=="female")#|0            → first female object
```

### Multi-Selectors

Comma-separated selectors inside `{...}` create a new object; inside `[...]` create a new array:

```
{version,author,type,"top":stock.#(price_2007>=10)#.symbol}
→ {"version":"1.0.0","author":"subs","type":"object","top":["MMM","AMZN","CPB","DIS","DOW","XOM","GPS","GIS"]}
```

### Literals

JSON literals are prefixed with `!` and are useful when building new objects with [Multi-Selectors](#multi-selectors):

```
{version,author,"marked":!true,"scope":!"static"}
→ {"version":"1.0.0","author":"subs","marked":true,"scope":"static"}
```

## Built-in Transformers

Transformers are applied with the `@` prefix inside a path expression and receive the current JSON value as input. An optional argument is passed after a `:` separator.

```
path.@transformerName
path.@transformerName:argument
path.@transformerName:{"key":"value"}
```

### Core transformers

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

### String transformers

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

### Object transformers

| Transformer | Description                                                                                                                                                     |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `@project`  | Pick and/or rename fields from an object. Arg: `{"pick":["f1","f2"],"rename":{"f1":"newName"}}`. Omit `pick` to keep all fields; omit `rename` for no renaming. |
| `@default`  | Inject fallback values for fields that are absent or `null`. Arg: `{"field":"defaultValue",...}`. Existing non-null fields are never overwritten.               |

### Array transformers

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

### Value normalization transformers

| Transformer | Description                                                                                                                                  |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `@coerce`   | Convert a scalar to a target type. Arg: `{"to":"string"}`, `{"to":"number"}`, or `{"to":"bool"}`. Objects and arrays are returned unchanged. |

### Transformer Examples

```go
package main

import (
	"fmt"
	"github.com/polarixa/replify/pkg/fj"
)

func main() {
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
	fmt.Println(fj.Get(json, "@pretty").String())      // indented JSON
	fmt.Println(fj.Get(json, "@minify").String())      // compact JSON
	fmt.Println(fj.Get(json, "user.@keys").String())   // ["name","role","age","city"]
	fmt.Println(fj.Get(json, "user.@values").String()) // ["Alice",null,30,"NY"]
	fmt.Println(fj.Get(json, "user.@valid").String())  // "true"

	// ── String ───────────────────────────────────────────────────────────────────
	fmt.Println(fj.Get(json, "user.name.@uppercase").String())                                // "ALICE"
	fmt.Println(fj.Get(json, "user.name.@reverse").String())                                  // "ecilA"
	fmt.Println(fj.Get(json, "user.name.@snakecase").String())                                // "alice"
	fmt.Println(fj.Get(json, "user.city.@padLeft:{\"padding\":\"0\",\"length\":6}").String()) // "000 NY"

	// ── Object ───────────────────────────────────────────────────────────────────

	// Project: keep only name and age, rename age → years
	fmt.Println(fj.Get(json, `user.@project:{"pick":["name","age"],"rename":{"age":"years"}}`).Raw())
	// → {"name":"Alice","years":30}

	// Default: fill in missing / null fields
	fmt.Println(fj.Get(json, `user.@default:{"role":"viewer","active":true}`).Raw())
	// → {"name":"Alice","role":"viewer","age":30,"city":"NY","active":true}

	// ── Array ────────────────────────────────────────────────────────────────────

	// Filter: keep only active users
	fmt.Println(fj.Get(json, `users.@filter:{"key":"active","value":true}`).Raw())
	// → [{"name":"Alice","active":true,...},{"name":"Carol","active":true,...}]

	// Pluck: extract the city from every user's address
	fmt.Println(fj.Get(json, `users.@pluck:addr.city`).Raw())
	// → ["NY","LA","NY"]

	// Aggregation helpers
	fmt.Println(fj.Get(json, "scores.@first").Raw()) // 95
	fmt.Println(fj.Get(json, "scores.@last").Raw())  // 78
	fmt.Println(fj.Get(json, "scores.@count").Raw()) // 4
	fmt.Println(fj.Get(json, "scores.@sum").Raw())   // 352
	fmt.Println(fj.Get(json, "scores.@min").Raw())   // 78
	fmt.Println(fj.Get(json, "scores.@max").Raw())   // 95

	// ── Coerce ───────────────────────────────────────────────────────────────────
	fmt.Println(fj.Get(`42`, `@coerce:{"to":"string"}`).Raw())   // "42"
	fmt.Println(fj.Get(`"99"`, `@coerce:{"to":"number"}`).Raw()) // 99
	fmt.Println(fj.Get(`1`, `@coerce:{"to":"bool"}`).Raw())      // true
}
```

### Composing transformers

Transformers can be chained using the `|` pipe operator or dot notation:

```go
	// First filter the array, then count the remaining elements
	fmt.Println(fj.Get(json, `users.@filter:{"key":"active","value":true}|@count`).Raw())
	// → 2

	// Pluck names, then reverse the resulting array
	fmt.Println(fj.Get(json, `users.@pluck:name|@reverse`).Raw())
	// → ["Carol","Bob","Alice"]
```

### Complex real-world examples

The following scenarios demonstrate how to combine multiple transformers into a single expression to process realistic JSON payloads.

---

**Example 1 — E-commerce product catalog: filter, aggregate, and shape**

```go
package main

import (
	"fmt"
	"github.com/polarixa/replify/pkg/fj"
)

func main() {
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
	fmt.Println(fj.Get(catalog, `products.@filter:{"key":"category","value":"electronics"}|@filter:{"key":"stock","op":"gt","value":0}|@pluck:name`).Raw())
	// → ["Laptop Pro","USB-C Hub","Webcam HD"]

	// Count of in-stock products
	fmt.Println(fj.Get(catalog, `products.@filter:{"key":"stock","op":"gt","value":0}|@count`).Raw())
	// → 4

	// Price range of in-stock products
	fmt.Println(fj.Get(catalog, `products.@filter:{"key":"stock","op":"gt","value":0}|@pluck:price|@min`).Raw())
	// → 49.99
	fmt.Println(fj.Get(catalog, `products.@filter:{"key":"stock","op":"gt","value":0}|@pluck:price|@max`).Raw())
	// → 1299.99

	// Project the first in-stock product as a display card (pick and rename fields)
	first := fj.Get(catalog, `products.@filter:{"key":"stock","op":"gt","value":0}|@first`).Raw()
	fmt.Println(fj.Get(first, `@project:{"pick":["name","price"],"rename":{"name":"title","price":"cost"}}`).Raw())
	// → {"title":"Laptop Pro","cost":1299.99}
}
```

---

**Example 2 — API response normalization: fill defaults then project and rename**

```go
	// Raw user record from an external API with null / absent fields
	rawUser := `{"id":"u1","name":"Alice","role":null,"verified":null}`

	// One-shot normalization: fill nulls → keep only safe fields → rename id for the frontend
	fmt.Println(fj.Get(rawUser, `@default:{"role":"viewer","verified":false}|@project:{"pick":["id","name","role","verified"],"rename":{"id":"userId"}}`).Raw())
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
	fmt.Println(fj.Get(logs, `@filter:{"key":"level","value":"error"}|@count`).Raw())
	// → 2

	// All error messages
	fmt.Println(fj.Get(logs, `@filter:{"key":"level","value":"error"}|@pluck:msg`).Raw())
	// → ["Connection refused","Timeout exceeded"]

	// Most recent error entry (last in the filtered array)
	fmt.Println(fj.Get(logs, `@filter:{"key":"level","value":"error"}|@last`).Raw())
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
	fmt.Println(fj.Get(teamData, `teams.@filter:{"key":"active","value":true}|@pluck:monthly_revenue|@flatten|@sum`).Raw())
	// → 78000   (Alpha: 33000 + Gamma: 45000)
```

---

**Example 5 — URL-slug generation from a display name**

```go
	// Multi-word title with duplicate internal spaces → URL-safe kebab-case slug
	fmt.Println(fj.Get(`"My   Blog Post Title"`, `@trim|@lowercase|@kebabcase`).Raw())
	// → "my-blog-post-title"

	// Author name to lowercase slug
	fmt.Println(fj.Get(`"John Doe"`, `@lowercase|@replace:{"target":" ","replacement":"-"}`).Raw())
	// → "john-doe"
```

---

**Example 6 — Config merging and introspection**

```go
	// Merge two partial config objects; later values overwrite earlier ones for duplicate keys
	overrides := `[{"host":"localhost","port":5432},{"port":5433,"ssl":true}]`

	merged := fj.Get(overrides, `@join`).Raw()
	// → {"host":"localhost","port":5433,"ssl":true}
	fmt.Println(merged)

	// Inspect which keys are present after the merge
	fmt.Println(fj.Get(merged, `@keys`).Raw())
	// → ["host","port","ssl"]

	// Count the merged keys
	fmt.Println(fj.Get(merged, `@count`).Raw())
	// → 3

	// Project only the connection-relevant subset and rename for the driver
	fmt.Println(fj.Get(merged, `@project:{"pick":["host","port"],"rename":{"port":"dbPort"}}`).Raw())
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
	fmt.Println(grouped)

	// Players with a score of 85 or above
	fmt.Println(fj.Get(grouped, `@filter:{"key":"score","op":"gte","value":85}|@pluck:player`).Raw())
	// → ["Alice","Carol","Dave"]

	// Top player's full record
	fmt.Println(fj.Get(grouped, `@filter:{"key":"score","op":"gte","value":95}|@first`).Raw())
	// → {"player":"Alice","score":98}
```

---

Transformers can be disabled globally with `fj.DisableTransformers = true`.

## Custom Transformers

Register a custom transformer with `AddTransformer`. Registrations are thread-safe and take effect immediately for all subsequent calls.

```go
package main

import (
    "fmt"
    "strings"
    "github.com/polarixa/replify/pkg/fj"
)

func init() {
	// @shout appends "!!!" to the string value
	fj.AddTransformer("shout", fj.TransformerFunc(func(json, arg string) string {
		return strings.Trim(json, `"`) + "!!!"
	}))

	// @repeat repeats the value N times (arg is the count as a string)
	fj.AddTransformer("repeat", fj.TransformerFunc(func(json, arg string) string {
		n := fj.Parse(arg).Int64()
		if n <= 0 {
			return json
		}
		v := strings.Trim(json, `"`)
		return strings.Repeat(v, int(n))
	}))
}

func main() {
    json := `{"greeting":"hello"}`

    fmt.Println(fj.Get(json, "greeting.@shout"))
    // hello!!!

    fmt.Println(fj.Get(json, "greeting.@repeat:3"))
    // hellohellohello
}
```

## JSON Color Styles

Use `Context.StringColored()` for default ANSI coloring or `Context.WithStringColored(style)` for a custom style.

```go
ctx := fj.Get(json, "user")
fmt.Println(ctx.StringColored())                         // default style
fmt.Println(ctx.WithStringColored(fj.NeonStyle))         // neon style
fmt.Println(ctx.WithStringColored(fj.DarkStyle))         // dark style
```

Available named style variables:

| Variable            | Description                                             |
| ------------------- | ------------------------------------------------------- |
| `DarkStyle`         | Dark tones — navy blue, dark green, amber, dark magenta |
| `NeonStyle`         | Vibrant neon — bright cyan, lime, yellow                |
| `PastelStyle`       | Soft pastel tones                                       |
| `HighContrastStyle` | High-contrast for accessibility                         |
| `VintageStyle`      | Muted vintage palette                                   |
| `CyberpunkStyle`    | Futuristic cyberpunk neons                              |
| `OceanStyle`        | Cool ocean blues and cyans                              |
| `FieryStyle`        | Warm reds and oranges                                   |
| `GalaxyStyle`       | Deep-space purples and blues                            |
| `SunsetStyle`       | Warm oranges and pinks                                  |
| `JungleStyle`       | Lush jungle greens                                      |
| `MonochromeStyle`   | Grayscale only                                          |
| `ForestStyle`       | Earthy forest greens and browns                         |
| `IceStyle`          | Cold icy blues and whites                               |
| `RetroStyle`        | Retro terminal amber/green                              |
| `AutumnStyle`       | Browns, oranges, and reds                               |
| `GothicStyle`       | Dark purples and blacks                                 |
| `VaporWaveStyle`    | Aesthetic vaporwave pinks and purples                   |
| `VampireStyle`      | Deep blood reds and blacks                              |
| `CarnivalStyle`     | Bright carnival multicolor                              |
| `SteampunkStyle`    | Brass and copper tones                                  |
| `WoodlandStyle`     | Natural woodland tans and greens                        |
| `CandyStyle`        | Bright candy pastels                                    |
| `TwilightStyle`     | Dusk purples and navies                                 |
| `EarthStyle`        | Warm earth tones                                        |
| `ElectricStyle`     | Electric blues and greens                               |
| `WitchingHourStyle` | Dark witching-hour palette                              |
| `MidnightStyle`     | Deep midnight navy and silver                           |

## API Reference

### Top-level Functions

| Function           | Signature                                                       | Description                                                 |
| ------------------ | --------------------------------------------------------------- | ----------------------------------------------------------- |
| `Get`              | `Get(json, path string) Context`                                | Search JSON for a dot-notation path; return first match.    |
| `GetBytes`         | `GetBytes(json []byte, path string) Context`                    | Same as `Get` but accepts a byte slice.                     |
| `Parse`            | `Parse(json string) Context`                                    | Parse a JSON string into a `Context` without path querying. |
| `ParseBytes`       | `ParseBytes(json []byte) Context`                               | Same as `Parse` but accepts a byte slice.                   |
| `ParseReader`      | `ParseReader(in io.Reader) (string, error)`                     | Read all data from an `io.Reader` and return as a string.   |
| `ParseJSONFile`    | `ParseJSONFile(filepath string) (string, error)`                | Read a JSON file and return its contents as a string.       |
| `IsValidJSON`      | `IsValidJSON(json string) bool`                                 | Report whether a string is valid JSON.                      |
| `IsValidJSONBytes` | `IsValidJSONBytes(json []byte) bool`                            | Report whether a byte slice is valid JSON.                  |
| `AddTransformer`   | `AddTransformer(name string, fn func(json, arg string) string)` | Register a named transformer.                               |

### Search Engine Functions

| Function             | Signature                                                       | Description                                                                                                             |
| -------------------- | --------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `Search`             | `Search(json, keyword string) []Context`                        | Full-tree scan — return all scalar leaves whose string value contains `keyword`.                                        |
| `SearchMatch`        | `SearchMatch(json, pattern string) []Context`                   | Full-tree wildcard scan — return all scalar leaves whose string value matches `pattern` (`*`, `?`). Uses `match.Match`. |
| `SearchByKey`        | `SearchByKey(json string, keys ...string) []Context`            | Return all values stored under the given key name(s) at any depth (exact names).                                        |
| `SearchByKeyPattern` | `SearchByKeyPattern(json, keyPattern string) []Context`         | Return all values stored under object keys that match the wildcard `keyPattern`. Uses `match.Match`.                    |
| `Contains`           | `Contains(json, path, target string) bool`                      | Report whether the value at `path` contains the substring `target`.                                                     |
| `ContainsMatch`      | `ContainsMatch(json, path, pattern string) bool`                | Report whether the value at `path` matches the wildcard `pattern`. Uses `match.Match`.                                  |
| `FindPath`           | `FindPath(json, value string) string`                           | Return the first dot-notation path at which a scalar equals `value`.                                                    |
| `FindPaths`          | `FindPaths(json, value string) []string`                        | Return all dot-notation paths at which a scalar equals `value`.                                                         |
| `FindPathMatch`      | `FindPathMatch(json, valuePattern string) string`               | Return the first dot-notation path at which a scalar matches the wildcard `valuePattern`.                               |
| `FindPathsMatch`     | `FindPathsMatch(json, valuePattern string) []string`            | Return all paths at which a scalar matches the wildcard `valuePattern`.                                                 |
| `Count`              | `Count(json, path string) int`                                  | Count elements at `path` (array length or 1 for scalars; 0 when missing).                                               |
| `Sum`                | `Sum(json, path string) float64`                                | Sum of all numeric values at `path`.                                                                                    |
| `Min`                | `Min(json, path string) (float64, bool)`                        | Minimum numeric value at `path`.                                                                                        |
| `Max`                | `Max(json, path string) (float64, bool)`                        | Maximum numeric value at `path`.                                                                                        |
| `Avg`                | `Avg(json, path string) (float64, bool)`                        | Arithmetic mean of numeric values at `path`.                                                                            |
| `CollectFloat64`     | `CollectFloat64(json, path string) []float64`                   | Collect numeric values at `path` using `conv.Float64` (handles string-encoded numbers).                                 |
| `Filter`             | `Filter(json, path string, fn func(Context) bool) []Context`    | Keep only elements at `path` for which `fn` returns true.                                                               |
| `First`              | `First(json, path string, fn func(Context) bool) Context`       | First element at `path` for which `fn` returns true.                                                                    |
| `Distinct`           | `Distinct(json, path string) []Context`                         | Unique values at `path` (first-occurrence order).                                                                       |
| `Pluck`              | `Pluck(json, path string, fields ...string) []Context`          | Extract named fields from each object in the array at `path`.                                                           |
| `GroupBy`            | `GroupBy(json, path, keyField string) map[string][]Context`     | Group array elements by the string value of `keyField`, using `conv.String` for key normalization.                      |
| `SortBy`             | `SortBy(json, path, keyField string, ascending bool) []Context` | Sort array elements by a sub-field using `conv`-powered numeric and string comparison.                                  |
| `CoerceTo`           | `CoerceTo(ctx Context, into any) error`                         | Coerce a Context's value into any Go typed variable via `conv.Infer`.                                                   |

### `Context` Methods

| Method              | Signature                                         | Description                                                                 |
| ------------------- | ------------------------------------------------- | --------------------------------------------------------------------------- |
| `Kind`              | `Kind() Type`                                     | Return the JSON type (`Null`, `False`, `Number`, `String`, `True`, `JSON`). |
| `Raw`               | `Raw() string`                                    | Return the raw unprocessed JSON fragment.                                   |
| `Number`            | `Number() float64`                                | Return the numeric value (for `Number` type).                               |
| `Index`             | `Index() int`                                     | Return the byte offset of this value in the original JSON.                  |
| `Indexes`           | `Indexes() []int`                                 | Return positions of all `#`-matched elements.                               |
| `String`            | `String() string`                                 | Return a string representation of the value.                                |
| `StringColored`     | `StringColored() string`                          | Return the string with default ANSI color styling.                          |
| `WithStringColored` | `WithStringColored(style *unify4g.Style) string`  | Return the string with a custom ANSI color style.                           |
| `Bool`              | `Bool() bool`                                     | Return the boolean value.                                                   |
| `Int64`             | `Int64() int64`                                   | Return the value as `int64`.                                                |
| `Uint64`            | `Uint64() uint64`                                 | Return the value as `uint64`.                                               |
| `Float64`           | `Float64() float64`                               | Return the value as `float64`.                                              |
| `Float32`           | `Float32() float32`                               | Return the value as `float32`.                                              |
| `Time`              | `Time() time.Time`                                | Parse the value as `time.Time` using RFC 3339.                              |
| `WithTime`          | `WithTime(layout string) time.Time`               | Parse the value as `time.Time` with a custom layout.                        |
| `Array`             | `Array() []Context`                               | Return all array elements as a `[]Context`.                                 |
| `IsObject`          | `IsObject() bool`                                 | Report whether the value is a JSON object.                                  |
| `IsArray`           | `IsArray() bool`                                  | Report whether the value is a JSON array.                                   |
| `IsBool`            | `IsBool() bool`                                   | Report whether the value is a JSON boolean.                                 |
| `Exists`            | `Exists() bool`                                   | Report whether the path was found in the JSON.                              |
| `Value`             | `Value() interface{}`                             | Return the value as a native Go type (`map`, `[]interface{}`, etc.).        |
| `Map`               | `Map() map[string]Context`                        | Return the value as a `map[string]Context` (for JSON objects).              |
| `Foreach`           | `Foreach(iterator func(key, value Context) bool)` | Iterate over array elements or object key-value pairs.                      |
| `Get`               | `Get(path string) Context`                        | Query a sub-path (enables chaining).                                        |
| `GetMulti`          | `GetMulti(paths ...string) []Context`             | Query multiple paths simultaneously.                                        |
| `Path`              | `Path(json string) string`                        | Return the dot-notation path that produced this context.                    |
| `Paths`             | `Paths(json string) []string`                     | Return paths for each element in an array result.                           |
| `Less`              | `Less(token Context, caseSensitive bool) bool`    | Report whether this value is less than `token`.                             |

### JSON Type Constants

| Constant | Description          |
| -------- | -------------------- |
| `Null`   | JSON `null`          |
| `False`  | JSON `false`         |
| `Number` | JSON number          |
| `String` | JSON string          |
| `True`   | JSON `true`          |
| `JSON`   | JSON object or array |

## Examples

### 1. Nested Object Access

```go
json := `{"store":{"book":{"title":"Go Programming","price":29.99}}}`
title := fj.Get(json, "store.book.title").String()
price := fj.Get(json, "store.book.price").Float64()
fmt.Printf("%s costs $%.2f\n", title, price) // Go Programming costs $29.99
```

### 2. Array Queries

```go
json := `{"users":[
    {"name":"Alice","role":"admin","active":true},
    {"name":"Bob","role":"user","active":false},
    {"name":"Carol","role":"admin","active":true}
]}`

// All active admin names
admins := fj.Get(json, `users.#(role=="admin")#.name`)
fmt.Println(admins.String()) // ["Alice","Carol"]
```

### 3. Transformer Pipeline

```go
json := `{"tags":["Go","json","fast"]}`
result := fj.Get(json, "tags.@reverse")
fmt.Println(result.String()) // ["fast","json","Go"]

// Chain transformers using pipe
reversed := fj.Get(json, "tags.@reverse|0")
fmt.Println(reversed.String()) // "fast"
```

### 4. Multi-Selector with Literals

```go
json := `{"version":"1.0","author":"Alice","items":[1,2,3]}`
result := fj.Get(json, `{version,author,"count":items.#,"published":!true}`)
fmt.Println(result.String())
// {"version":"1.0","author":"Alice","count":3,"published":true}
```

### 5. Struct Unmarshalling via Value()

```go
json := `{"name":"Alice","scores":[95,87,91]}`
m := fj.Parse(json).Map()
name := m["name"].String()
scores := m["scores"].Array()
fmt.Println(name)                           // Alice
fmt.Println(len(scores), scores[0].Int64()) // 3  95
```

### 6. Search Engine

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/fj"
)

func main() {
    json := `{
        "store": {
            "books": [
                {"id":1,"title":"The Go Programming Language","author":"Donovan","price":34.99,"genre":"tech"},
                {"id":2,"title":"Clean Code","author":"Martin","price":29.99,"genre":"tech"},
                {"id":3,"title":"Harry Potter","author":"Rowling","price":14.99,"genre":"fiction"},
                {"id":4,"title":"Dune","author":"Herbert","price":12.99,"genre":"fiction"}
            ],
            "owner": "Alice"
        },
        "ratings": [5, 3, 4, 5, 1, 4],
        "tags": ["go","json","fast","go","json"]
    }`

    // --- Full-tree substring search ---
    matches := fj.Search(json, "tech")
    fmt.Println(len(matches)) // 2

    // --- Full-tree wildcard search (match.Match) ---
    wildMatches := fj.SearchMatch(json, "D*")
    fmt.Println(len(wildMatches)) // 2: "Donovan", "Dune"

    // --- Search by exact key name ---
    authors := fj.SearchByKey(json, "author")
    for _, a := range authors {
        fmt.Println(a.String()) // Donovan, Martin, Rowling, Herbert
    }

    // --- Search by key wildcard pattern (match.Match) ---
    keyMatches := fj.SearchByKeyPattern(json, "auth*")
    fmt.Println(len(keyMatches)) // 4 (author fields)

    // --- Substring contains ---
    fmt.Println(fj.Contains(json, "store.owner", "Ali")) // true

    // --- Wildcard contains (match.Match) ---
    fmt.Println(fj.ContainsMatch(json, "store.owner", "Al*")) // true

    // --- Path discovery (exact) ---
    fmt.Println(fj.FindPath(json, "Rowling")) // store.books.2.author

    // --- Path discovery (wildcard, match.Match) ---
    fmt.Println(fj.FindPathMatch(json, "Row*")) // store.books.2.author

    // --- All paths matching a wildcard ---
    paths := fj.FindPathsMatch(json, "D*")
    fmt.Println(paths) // [store.books.0.author store.books.3.title]

    // --- Aggregate functions ---
    fmt.Println(fj.Sum(json, "ratings"))       // 22
    fmt.Println(fj.Count(json, "store.books")) // 4
    v, _ := fj.Min(json, "ratings")
    fmt.Println(v) // 1
    v, _ = fj.Max(json, "ratings")
    fmt.Println(v) // 5
    avg, _ := fj.Avg(json, "ratings")
    fmt.Printf("%.4f\n", avg) // 3.6667

    // --- Collect numbers with conv.Float64 (handles string-encoded numbers too) ---
    nums := fj.CollectFloat64(json, "ratings")
    fmt.Println(nums) // [5 3 4 5 1 4]

    // --- Filter ---
    fiction := fj.Filter(json, "store.books", func(ctx fj.Context) bool {
        return ctx.Get("genre").String() == "fiction"
    })
    fmt.Println(len(fiction)) // 2

    // --- First ---
    cheap := fj.First(json, "store.books", func(ctx fj.Context) bool {
        return ctx.Get("price").Float64() < 20
    })
    fmt.Println(cheap.Get("title").String()) // Harry Potter

    // --- Distinct ---
    unique := fj.Distinct(json, "tags")
    fmt.Println(len(unique)) // 3: go, json, fast

    // --- Pluck ---
    projected := fj.Pluck(json, "store.books", "id", "title")
    for _, p := range projected {
        fmt.Println(p.String())
    }

    // --- GroupBy (uses conv.String for key normalization) ---
    groups := fj.GroupBy(json, "store.books", "genre")
    fmt.Println(len(groups["tech"]))    // 2
    fmt.Println(len(groups["fiction"])) // 2

    // --- SortBy (uses conv.Float64/conv.String for comparison) ---
    byPrice := fj.SortBy(json, "store.books", "price", true)
    fmt.Println(byPrice[0].Get("title").String()) // Dune (cheapest)

    // --- CoerceTo (uses conv.Infer) ---
    ctx := fj.Get(json, "store.books.0.price")
    var price float64
    if err := fj.CoerceTo(ctx, &price); err == nil {
        fmt.Printf("%.2f\n", price) // 34.99
    }
}
```

## Best Practices

### ✅ Do's

1. **Check `Exists()` before using a value** to distinguish a missing path from a JSON `null`:

   ```go
   if ctx := fj.Get(json, "user.email"); ctx.Exists() {
       sendEmail(ctx.String())
   }
   ```

2. **Use `GetBytes` when JSON is already `[]byte`** to avoid an extra allocation:

   ```go
   ctx := fj.GetBytes(rawBytes, "user.id")
   ```

3. **Register custom transformers once at startup** (e.g., in `init()`) to avoid race conditions:

   ```go
   func init() {
       fj.AddTransformer("myTransform", myFn)
   }
   ```

4. **Use `Foreach` instead of `Array()`** when you only need to process elements one at a time:

   ```go
   fj.Get(json, "items").Foreach(func(_, val fj.Context) bool {
       process(val)
       return true
   })
   ```

5. **Validate untrusted input** with `IsValidJSON` before heavy processing:

   ```go
   if !fj.IsValidJSON(input) {
       return errors.New("invalid JSON")
   }
   ```

### ❌ Don'ts

1. **Don't assume a zero-value `Context` means JSON `null`** — it may mean the path was not found:

   ```go
   // ❌ Bad
   val := fj.Get(json, "missing.field").String() // returns "" even if not null

   // ✅ Good
   if ctx := fj.Get(json, "missing.field"); ctx.Exists() {
       val := ctx.String()
   }
   ```

2. **Don't mutate the `[]byte` passed to `GetBytes`** while a query is in progress.

3. **Don't use `@replace` when you intend to replace all occurrences** — use `@replaceAll` instead:

   ```go
   // @replace replaces only the FIRST occurrence
   fj.Get(`"foo foo foo"`, `@replace:{"target":"foo","replacement":"bar"}`)
   // → bar foo foo

   // @replaceAll replaces ALL occurrences
   fj.Get(`"foo foo foo"`, `@replaceAll:{"target":"foo","replacement":"bar"}`)
   // → bar bar bar
   ```

4. **Don't call `Map()` on a non-object** without first checking `IsObject()`.

5. **Don't call `Array()` on a non-array** without first checking `IsArray()`.

## Thread Safety

All public API functions (`Get`, `GetBytes`, `Parse*`, `IsValid*`, `AddTransformer`) are safe for concurrent use by multiple goroutines. The transformer registry uses a `sync.RWMutex` for safe concurrent reads and writes.

## Contributing

To contribute to this project, follow these steps:

1. Clone the repository:

   ```bash
   git clone --depth 1 https://github.com/polarixa/replify.git
   ```

2. Navigate to the project directory:

   ```bash
   cd replify
   ```

3. Prepare the project environment:

   ```bash
   go mod tidy
   ```

4. Run the tests:

   ```bash
   go test ./pkg/fj/...
   ```

5. Submit a pull request.
