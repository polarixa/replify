# msort

**msort** (Map Sort) is a type-safe Go library for sorting maps by keys or values. It leverages Go 1.18+ generics to provide efficient, compile-time type-checked map sorting with a clean, functional API.

## Overview

The `msort` package solves the problem of sorting Go maps, which are inherently unordered. It provides:

- **Type-Safe Sorting**: Compile-time type checking with generics
- **Flexible Sorting**: Sort by keys, values, or custom logic
- **Time Support**: Specialized functions for `time.Time` values
- **Stable Sorting**: Preserve order of equal elements
- **Utility Methods**: Extract keys, values, top-N results, or convert back to maps
- **Zero Dependencies**: No external dependencies beyond Go standard library

**Problem Solved:** Go maps are unordered by design. When you need deterministic iteration order (for display, testing, or processing), you must manually extract entries, sort them, and iterate. This package provides a clean, reusable solution with zero boilerplate.

## Use Cases

### When to Use

- ✅ **Displaying sorted data** - show users data in alphabetical or numerical order
- ✅ **Top-N queries** - get top products, highest scores, latest events
- ✅ **Deterministic iteration** - consistent ordering for testing or logging
- ✅ **Leaderboards** - sort scores, rankings, or metrics
- ✅ **Timeline display** - chronological ordering of events
- ✅ **Alphabetical lists** - sort by names, categories, or labels
- ✅ **Value-based ranking** - sort by price, count, rating, etc.
- ✅ **Custom sorting logic** - complex multi-criteria sorting

### When Not to Use

- ❌ **Simple slice sorting** - use `sort.Slice` directly for slices
- ❌ **Large maps (>100k entries)** - consider database sorting instead
- ❌ **Frequent modifications** - maintain a sorted data structure instead
- ❌ **When order doesn't matter** - iterate maps directly
- ❌ **Real-time streaming data** - use priority queues or heaps

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package in your Go code:

```go
import "github.com/polarixa/replify/pkg/msort"
```

**Requirements:** Go 1.18 or higher (for generics support)

## Usage

### Quick Start

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/msort"
)

func main() {
    // Sample data: product ratings
    ratings := map[string]int{
        "Product A": 4,
        "Product B": 5,
        "Product C": 3,
        "Product D": 5,
    }

    // Sort by value (rating) - descending
    sorted := msort.SortValueDesc(ratings)

    // Display sorted results
    for _, item := range sorted {
        fmt.Printf("%s: %d stars\n", item.Key, item.Value)
    }
    // Output:
    // Product B: 5 stars
    // Product D: 5 stars
    // Product A: 4 stars
    // Product C: 3 stars
}
```

## Examples

### 1. Sort by Keys

```go
// Alphabetically sort by keys (ascending)
users := map[string]int{
    "charlie": 30,
    "alice":   25,
    "bob":     35,
}

sorted := msort.SortKey(users)
for _, item := range sorted {
    fmt.Printf("%s: %d years old\n", item.Key, item.Value)
}
// Output:
// alice: 25 years old
// bob: 35 years old
// charlie: 30 years old

// Reverse alphabetical order (descending)
reversed := msort.SortKeyDesc(users)
for _, item := range reversed {
    fmt.Printf("%s: %d years old\n", item.Key, item.Value)
}
// Output:
// charlie: 30 years old
// bob: 35 years old
// alice: 25 years old
```

### 2. Sort by Values

```go
// Sort by values (ascending)
scores := map[string]int{
    "player1": 150,
    "player2": 300,
    "player3": 200,
}

sorted := msort.SortValue(scores)
for _, item := range sorted {
    fmt.Printf("%s: %d points\n", item.Key, item.Value)
}
// Output:
// player1: 150 points
// player3: 200 points
// player2: 300 points

// Sort by values (descending) - leaderboard
leaderboard := msort.SortValueDesc(scores)
for i, item := range leaderboard {
    fmt.Printf("#%d: %s - %d points\n", i+1, item.Key, item.Value)
}
// Output:
// #1: player2 - 300 points
// #2: player3 - 200 points
// #3: player1 - 150 points
```

### 3. Custom Sorting

```go
type Product struct {
    Name  string
    Price float64
    Stock int
}

products := map[string]Product{
    "A": {Name: "Widget A", Price: 19.99, Stock: 10},
    "B": {Name: "Widget B", Price: 29.99, Stock: 5},
    "C": {Name: "Widget C", Price: 19.99, Stock: 20},
}

// Sort by price (ascending), then by stock (descending)
sorted := msort.SortFunc(products, func(x, y msort.item[string, Product]) bool {
    if x.Value.Price != y.Value.Price {
        return x.Value.Price < y.Value.Price
    }
    return x.Value.Stock > y.Value.Stock
})

for _, item := range sorted {
    p := item.Value
    fmt.Printf("%s: $%.2f (%d in stock)\n", p.Name, p.Price, p.Stock)
}
// Output:
// Widget C: $19.99 (20 in stock)
// Widget A: $19.99 (10 in stock)
// Widget B: $29.99 (5 in stock)
```

### 4. Top-N Results

```go
// Get top 3 highest scores
scores := map[string]int{
    "player1": 150,
    "player2": 300,
    "player3": 200,
    "player4": 400,
    "player5": 100,
}

top3 := msort.SortValueDesc(scores).Top(3)
fmt.Println("Top 3 Players:")
for i, item := range top3 {
    fmt.Printf("%d. %s: %d points\n", i+1, item.Key, item.Value)
}
// Output:
// Top 3 Players:
// 1. player4: 400 points
// 2. player2: 300 points
// 3. player3: 200 points
```

### 5. Working with Time Values

```go
events := map[string]time.Time{
    "launch":    time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
    "beta":      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    "feature_x": time.Date(2024, 2, 10, 14, 30, 0, 0, time.UTC),
}

// Sort chronologically (earliest first)
timeline := msort.SortTimeValue(events)
fmt.Println("Project Timeline:")
for _, item := range timeline {
    fmt.Printf("%s: %s\n", item.Key, item.Value.Format("2006-01-02"))
}
// Output:
// Project Timeline:
// beta: 2024-01-01
// feature_x: 2024-02-10
// launch: 2024-03-15

// Most recent first
recent := msort.SortTimeValueDesc(events)
fmt.Println("\nRecent Events:")
for _, item := range recent {
    fmt.Printf("%s: %s\n", item.Key, item.Value.Format("2006-01-02"))
}
// Output:
// Recent Events:
// launch: 2024-03-15
// feature_x: 2024-02-10
// beta: 2024-01-01
```

### 6. Extract Keys or Values Only

```go
prices := map[string]float64{
    "apple":  1.50,
    "banana": 0.75,
    "orange": 1.25,
}

// Get sorted keys only
sortedNames := msort.SortKey(prices).Keys()
fmt.Println("Products:", sortedNames)
// Output: Products: [apple banana orange]

// Get values only (sorted by key)
sortedPrices := msort.SortKey(prices).Values()
fmt.Println("Prices:", sortedPrices)
// Output: Prices: [1.5 0.75 1.25]

// Get values sorted by value
pricesSorted := msort.SortValue(prices).Values()
fmt.Println("Prices (sorted):", pricesSorted)
// Output: Prices (sorted): [0.75 1.25 1.5]
```

### 7. Convert Back to Map

```go
// Sort and filter, then convert back to map
scores := map[string]int{
    "player1": 150,
    "player2": 300,
    "player3": 200,
    "player4": 400,
    "player5": 100,
}

// Get top 3 as a map
top3Map := msort.SortValueDesc(scores).Top(3).ToMap()
fmt.Println(top3Map)
// Output: map[player2:300 player3:200 player4:400]

// Note: map is unordered, so sorting information is lost
for player, score := range top3Map {
    fmt.Printf("%s: %d\n", player, score)
}
```

### 8. Stable Sorting

```go
// When multiple keys have the same value, preserve their relative order
grades := map[string]int{
    "alice":   85,
    "bob":     90,
    "charlie": 85,
    "dave":    90,
    "eve":     85,
}

// Stable sort maintains original iteration order for equal values
sorted := msort.SortValueStable(grades)
fmt.Println("Grades (stable sort):")
for _, item := range sorted {
    fmt.Printf("%s: %d\n", item.Key, item.Value)
}
// Students with same grade maintain relative order
```

### 9. Practical: Leaderboard

```go
func DisplayLeaderboard(scores map[string]int, limit int) {
    sorted := msort.SortValueDesc(scores).Top(limit)

    fmt.Printf("🏆 Top %d Players\n", limit)
    fmt.Println(strings.Repeat("=", 40))

    for rank, item := range sorted {
        medal := ""
        switch rank {
        case 0:
            medal = "🥇"
        case 1:
            medal = "🥈"
        case 2:
            medal = "🥉"
        default:
            medal = fmt.Sprintf("%2d.", rank+1)
        }
        fmt.Printf("%s %-20s %6d pts\n", medal, item.Key, item.Value)
    }
}

// Usage
scores := map[string]int{
    "Alice":   1250,
    "Bob":     980,
    "Charlie": 1500,
    "Dave":    1100,
    "Eve":     890,
}
DisplayLeaderboard(scores, 3)
```

### 10. Practical: Product Catalog

```go
func SortProductsByPrice(products map[string]float64, ascending bool) {
    var sorted msort.items[string, float64]

    if ascending {
        sorted = msort.SortValue(products)
    } else {
        sorted = msort.SortValueDesc(products)
    }

    fmt.Println("Product Catalog:")
    for _, item := range sorted {
        fmt.Printf("%-20s $%.2f\n", item.Key, item.Value)
    }
}

// Usage
catalog := map[string]float64{
    "Laptop":     899.99,
    "Mouse":      19.99,
    "Keyboard":   79.99,
    "Monitor":    249.99,
    "Headphones": 149.99,
}

SortProductsByPrice(catalog, true)  // Cheapest first
SortProductsByPrice(catalog, false) // Most expensive first
```

### 11. Practical: Event Timeline

```go
func DisplayTimeline(events map[string]time.Time) {
    sorted := msort.SortTimeValue(events)

    fmt.Println("📅 Event Timeline")
    fmt.Println(strings.Repeat("-", 50))

    for _, item := range sorted {
        fmt.Printf("%s | %s\n",
            item.Value.Format("2006-01-02 15:04"),
            item.Key)
    }
}

// Usage
events := map[string]time.Time{
    "Project Kickoff": time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
    "Alpha Release":   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
    "Beta Release":    time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
    "Final Launch":    time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
}

DisplayTimeline(events)
```

## API Reference

### Core Sorting Functions

| Function        | Signature                                                     | Description                 |
| --------------- | ------------------------------------------------------------- | --------------------------- |
| `SortKey`       | `[K Ordered, V any](m map[K]V) items[K, V]`                   | Sort by keys (ascending)    |
| `SortKeyDesc`   | `[K Ordered, V any](m map[K]V) items[K, V]`                   | Sort by keys (descending)   |
| `SortValue`     | `[K comparable, V Ordered](m map[K]V) items[K, V]`            | Sort by values (ascending)  |
| `SortValueDesc` | `[K comparable, V Ordered](m map[K]V) items[K, V]`            | Sort by values (descending) |
| `SortFunc`      | `[K comparable, V any](m map[K]V, less LessFunc) items[K, V]` | Custom sorting logic        |

**Examples:**

```go
sorted := msort.SortKey(myMap)
sorted := msort.SortValueDesc(scores)
sorted := msort.SortFunc(data, customCompare)
```

---

### Time Sorting Functions

| Function            | Signature                                        | Description                          |
| ------------------- | ------------------------------------------------ | ------------------------------------ |
| `SortTimeValue`     | `[K comparable](m map[K]time.Time) timeItems[K]` | Sort by time (chronological)         |
| `SortTimeValueDesc` | `[K comparable](m map[K]time.Time) timeItems[K]` | Sort by time (reverse chronological) |

**Examples:**

```go
timeline := msort.SortTimeValue(events)      // Earliest first
recent := msort.SortTimeValueDesc(events)    // Latest first
```

---

### Stable Sorting Functions

| Function          | Signature                                          | Description           |
| ----------------- | -------------------------------------------------- | --------------------- |
| `SortKeyStable`   | `[K Ordered, V any](m map[K]V) items[K, V]`        | Stable sort by keys   |
| `SortValueStable` | `[K comparable, V Ordered](m map[K]V) items[K, V]` | Stable sort by values |

**Use Case:** Preserve relative order of equal elements for deterministic results.

---

### Utility Methods

#### items[K, V] Methods

| Method       | Signature     | Description         |
| ------------ | ------------- | ------------------- |
| `Top(n int)` | `items[K, V]` | Get first n items   |
| `ToMap()`    | `map[K]V`     | Convert back to map |
| `Keys()`     | `[]K`         | Extract keys only   |
| `Values()`   | `[]V`         | Extract values only |

**Examples:**

```go
top5 := sorted.Top(5)
resultMap := sorted.ToMap()
keys := sorted.Keys()
values := sorted.Values()
```

#### timeItems[K] Methods

Same methods as `items[K, V]` but for time-based sorting:

- `Top(n int) timeItems[K]`
- `ToMap() map[K]time.Time`
- `Keys() []K`
- `Values() []time.Time`

---

### Types

#### Ordered Constraint

```go
type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
    ~float32 | ~float64 | ~string
}
```

Types that support comparison operators (`<`, `>`, `<=`, `>=`).

#### LessFunc

```go
type LessFunc[K comparable, V any] func(x, y item[K, V]) bool
```

Comparison function for custom sorting. Returns `true` if `x` should come before `y`.

#### item[K, V]

```go
type item[K comparable, V any] struct {
    Key   K
    Value V
}
```

Represents a key-value pair from the map.

## Best Practices & Notes

### ⚠️ Common Pitfalls

1. **Converting Back to Map Loses Order**

   ```go
   // ❌ Order is lost when converting back
   sorted := msort.SortValue(m)
   resultMap := sorted.ToMap() // Map is unordered!

   // ✅ Iterate over sorted items directly
   for _, item := range sorted {
       fmt.Printf("%s: %v\n", item.Key, item.Value)
   }
   ```

2. **Type Constraints Must Match**

   ```go
   // ❌ Won't compile: struct not Ordered
   type Person struct { Name string }
   m := map[string]Person{}
   msort.SortKey(m) // Error: Person not comparable

   // ✅ Use SortFunc for custom types
   msort.SortFunc(m, func(x, y item[string, Person]) bool {
       return x.Value.Name < y.Value.Name
   })
   ```

3. **Stable vs Unstable Sort**

   ```go
   // With duplicate values, order may vary
   m := map[string]int{"a": 1, "b": 1, "c": 1}

   // Use stable sort for deterministic results
   sorted := msort.SortValueStable(m)
   ```

4. **Top() Doesn't Modify Original**

   ```go
   sorted := msort.SortValue(m)
   top3 := sorted.Top(3)

   fmt.Println(len(sorted)) // Original unchanged
   fmt.Println(len(top3))   // Slice of first 3
   ```

### 💡 Recommendations

✅ **Cache sorting results if reused**

```go
// ❌ Inefficient: sorting on every access
func getTopScores() []item[string, int] {
    return msort.SortValueDesc(scores).Top(10)
}

// ✅ Cache if accessed frequently
var cachedTop10 = msort.SortValueDesc(scores).Top(10)
```

✅ **Use appropriate sort function**

```go
// For simple key/value sorting
sorted := msort.SortKey(m)
sorted := msort.SortValue(m)

// For complex logic
sorted := msort.SortFunc(m, customLogic)

// For time values
sorted := msort.SortTimeValue(events)
```

✅ **Extract only what you need**

```go
// If you only need keys
keys := msort.SortValue(m).Keys()

// If you only need values
values := msort.SortValue(m).Values()

// If you need full items
items := msort.SortValue(m)
```

✅ **Use Top() for pagination**

```go
func GetPage(m map[string]int, page, pageSize int) items[string, int] {
    sorted := msort.SortValueDesc(m)
    start := page * pageSize
    if start >= len(sorted) {
        return nil
    }
    end := start + pageSize
    if end > len(sorted) {
        end = len(sorted)
    }
    return sorted[start:end]
}
```

### 🔒 Thread Safety

Sorting functions are **safe for concurrent reads** but **not for concurrent writes**:

```go
// ✅ Safe: concurrent sorts on different maps
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(m map[string]int) {
        defer wg.Done()
        sorted := msort.SortValue(m)
        // Process sorted...
    }(getCopyOfMap())
}

// ❌ Unsafe: concurrent modification
go func() { m["key"] = 1 }()
go func() { msort.SortValue(m) }() // Race condition!

// ✅ Safe: use mutex
var mu sync.RWMutex
mu.RLock()
sorted := msort.SortValue(m)
mu.RUnlock()
```

### ⚡ Performance Tips

**Time Complexity:** O(n log n) for all sorting operations
**Space Complexity:** O(n) - creates a slice of all entries

**Optimization strategies:**

1. **Avoid repeated sorting**

   ```go
   // ❌ Sorts on every call
   func display() {
       for _, item := range msort.SortValue(m) { }
   }

   // ✅ Sort once
   sorted := msort.SortValue(m)
   func display() {
       for _, item := range sorted { }
   }
   ```

2. **Use Top() for partial results**

   ```go
   // ❌ Sorts everything, takes first 10
   top10 := msort.SortValue(largeMap)[:10]

   // ✅ Same result (both O(n log n))
   top10 := msort.SortValue(largeMap).Top(10)
   ```

3. **Consider alternatives for very large maps**
   ```go
   // For >100k entries, consider:
   // - Database-level sorting (ORDER BY)
   // - Streaming/incremental processing
   // - Heap for top-N queries (O(n log k))
   ```

### 🐛 Debugging Tips

**Print sorted results:**

```go
sorted := msort.SortValue(m)
for i, item := range sorted {
    fmt.Printf("[%d] %v: %v\n", i, item.Key, item.Value)
}
```

**Verify sort order:**

```go
sorted := msort.SortValue(m)
for i := 1; i < len(sorted); i++ {
    if sorted[i-1].Value > sorted[i].Value {
        fmt.Printf("Sort error at index %d\n", i)
    }
}
```

### 📝 Testing

Example tests:

```go
func TestSortValue(t *testing.T) {
    m := map[string]int{
        "c": 3,
        "a": 1,
        "b": 2,
    }

    sorted := msort.SortValue(m)

    expected := []int{1, 2, 3}
    for i, item := range sorted {
        if item.Value != expected[i] {
            t.Errorf("Index %d: got %d, want %d", i, item.Value, expected[i])
        }
    }
}

func TestTop(t *testing.T) {
    m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
    top2 := msort.SortValueDesc(m).Top(2)

    if len(top2) != 2 {
        t.Errorf("Expected 2 items, got %d", len(top2))
    }
    if top2[0].Value != 4 {
        t.Errorf("Expected first value 4, got %d", top2[0].Value)
    }
}
```

## Limitations

- **Not for large datasets** - O(n log n) sorting in memory
- **Map must fit in memory** - all entries loaded into slice
- **No lazy evaluation** - entire map sorted upfront
- **Order lost when converting back** - maps are unordered
- **No partial sorting** - always sorts all entries (use Top() after)
- **Time.Time requires special functions** - not part of Ordered constraint

## Performance Characteristics

| Operation       | Time Complexity | Space Complexity |
| --------------- | --------------- | ---------------- |
| SortKey         | O(n log n)      | O(n)             |
| SortValue       | O(n log n)      | O(n)             |
| SortFunc        | O(n log n)      | O(n)             |
| Top(k)          | O(1)            | O(1) (slice)     |
| ToMap()         | O(n)            | O(n)             |
| Keys()/Values() | O(n)            | O(n)             |

Where n = number of map entries

## When to Use vs. Alternatives

**Use `msort` when:**

- Displaying sorted results to users
- Small to medium maps (<10k entries)
- One-time or infrequent sorting
- Need simple, readable API

**Consider alternatives when:**

- Frequent sorting: Use sorted data structure (B-tree, sorted slice)
- Large datasets: Use database `ORDER BY`
- Top-N only: Use heap (O(n log k) vs O(n log n))
- Real-time: Use priority queue

## Contributing

Contributions are welcome! Please see the main [replify repository](https://github.com/polarixa/replify) for contribution guidelines.

## License

This library is part of the [replify](https://github.com/polarixa/replify) project.

## Related

Part of the **replify** ecosystem:

- [replify](https://github.com/polarixa/replify) - API response wrapping library
- [coll](https://github.com/polarixa/replify/pkg/coll) - Type-safe collection utilities
- [common](https://github.com/polarixa/replify/pkg/common) - Reflection-based utilities
- [conv](https://github.com/polarixa/replify/pkg/conv) - Type conversion utilities
- [ref](https://github.com/polarixa/replify/pkg/ref) - Pointer utilities
- [hashy](https://github.com/polarixa/replify/pkg/hashy) - Deterministic hashing
- [match](https://github.com/polarixa/replify/pkg/match) - Wildcard pattern matching
- [strutil](https://github.com/polarixa/replify/pkg/strutil) - String utilities
- [randn](https://github.com/polarixa/replify/pkg/randn) - Random data generation
- [encoding](https://github.com/polarixa/replify/pkg/encoding) - JSON encoding utilities
