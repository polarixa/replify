# common

**common** is a Go utility library providing generic collection operations and reflection-based utilities for working with slices, arrays, and maps. It offers functional programming patterns like map, filter, reduce, and more, all without requiring type-specific implementations.

## Overview

The `common` package leverages Go's reflection capabilities to provide type-agnostic utility functions for collections. It solves the problem of repetitive code for common collection operations by offering:

- **Functional Operations**: Transform, filter, reduce, and iterate over collections
- **Collection Utilities**: Find, count, partition, unique, sort, and reverse
- **Set Operations**: Difference, intersection, union-like operations
- **Comparison Utilities**: Deep equality checking for any type
- **Type Checking**: Scalar type validation and empty value detection

**Problem Solved:** Before Go 1.18 generics, developers needed to write type-specific implementations for common collection operations. While generics now exist, this package provides reflection-based alternatives that work with `interface{}`, making it useful for dynamic scenarios where types aren't known at compile time.

## Use Cases

### When to Use

- ✅ **Dynamic data processing** - when types are unknown at compile time
- ✅ **Generic utilities** - building libraries that work with any type
- ✅ **Reflection-based operations** - when working with `interface{}` types
- ✅ **Functional programming patterns** - map/filter/reduce on arbitrary collections
- ✅ **Testing utilities** - comparing complex nested structures
- ✅ **Data transformations** - converting between collection formats
- ✅ **Legacy code** - working with pre-generics codebases

### When Not to Use

- ❌ **Type-safe operations** - use Go 1.18+ generics instead (e.g., `coll` package)
- ❌ **Performance-critical paths** - reflection has overhead
- ❌ **Simple operations** - use standard library or type-specific code
- ❌ **Compile-time safety needed** - reflection bypasses type checking
- ❌ **Large datasets** - consider streaming or database operations

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package in your Go code:

```go
import "github.com/polarixa/replify/pkg/common"
```

**Requirements:** Go 1.13 or higher

## Usage

### Quick Start

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/common"
)

func main() {
    // Filter even numbers
    numbers := []int{1, 2, 3, 4, 5, 6}
    evens := common.Filter(numbers, func(v interface{}) bool {
        return v.(int)%2 == 0
    })
    fmt.Println(evens) // [2 4 6]

    // Transform (map) - square each number
    squared := common.Transform(numbers, func(v interface{}) interface{} {
        return v.(int) * v.(int)
    })
    fmt.Println(squared) // [1 4 9 16 25 36]

    // Find first element > 3
    found := common.Find(numbers, func(v interface{}) bool {
        return v.(int) > 3
    })
    fmt.Println(found) // 4

    // Check if contains
    hasThree := common.Contains(numbers, 3)
    fmt.Println(hasThree) // true
}
```

## Examples

### 1. Functional Operations

#### Transform (Map)

```go
// Square numbers
numbers := []int{1, 2, 3, 4, 5}
squared := common.Transform(numbers, func(v interface{}) interface{} {
    return v.(int) * v.(int)
})
fmt.Println(squared) // [1 4 9 16 25]

// Convert to strings
strings := common.Transform(numbers, func(v interface{}) interface{} {
    return fmt.Sprintf("num-%d", v.(int))
})
fmt.Println(strings) // ["num-1" "num-2" "num-3" "num-4" "num-5"]
```

#### Filter

```go
// Filter even numbers
numbers := []int{1, 2, 3, 4, 5, 6, 7, 8}
evens := common.Filter(numbers, func(v interface{}) bool {
    return v.(int)%2 == 0
})
fmt.Println(evens) // [2 4 6 8]

// Filter strings by length
words := []string{"hi", "hello", "world", "a"}
longWords := common.Filter(words, func(v interface{}) bool {
    return len(v.(string)) > 2
})
fmt.Println(longWords) // ["hello" "world"]
```

#### Reduce

```go
// Sum of numbers
numbers := []int{1, 2, 3, 4, 5}
sum := common.Reduce(numbers, func(acc interface{}, v interface{}) interface{} {
    return acc.(int) + v.(int)
}, 0)
fmt.Println(sum) // 15

// Concatenate strings
words := []string{"Hello", "World", "!"}
sentence := common.Reduce(words, func(acc interface{}, v interface{}) interface{} {
    return acc.(string) + " " + v.(string)
}, "")
fmt.Println(sentence) // " Hello World !"
```

#### ReduceRight

```go
// Reduce from right to left
numbers := []int{1, 2, 3, 4}
result := common.ReduceRight(numbers, func(acc interface{}, v interface{}) interface{} {
    return fmt.Sprintf("%v, %v", acc, v)
}, "start")
fmt.Println(result) // "start, 4, 3, 2, 1"
```

### 2. Collection Queries

#### Find

```go
// Find first even number
numbers := []int{1, 3, 5, 6, 7, 8}
firstEven := common.Find(numbers, func(v interface{}) bool {
    return v.(int)%2 == 0
})
fmt.Println(firstEven) // 6

// Find returns nil if not found
notFound := common.Find([]int{1, 3, 5}, func(v interface{}) bool {
    return v.(int)%2 == 0
})
fmt.Println(notFound) // <nil>
```

#### All

```go
// Check if all are positive
numbers := []int{1, 2, 3, 4, 5}
allPositive := common.All(numbers, func(v interface{}) bool {
    return v.(int) > 0
})
fmt.Println(allPositive) // true

// Check if all are even
allEven := common.All(numbers, func(v interface{}) bool {
    return v.(int)%2 == 0
})
fmt.Println(allEven) // false
```

#### Any

```go
// Check if any are negative
numbers := []int{1, 2, 3, 4, 5}
anyNegative := common.Any(numbers, func(v interface{}) bool {
    return v.(int) < 0
})
fmt.Println(anyNegative) // false

// Check if any are even
anyEven := common.Any(numbers, func(v interface{}) bool {
    return v.(int)%2 == 0
})
fmt.Println(anyEven) // true
```

#### Count

```go
// Count even numbers
numbers := []int{1, 2, 3, 4, 5, 6}
evenCount := common.Count(numbers, func(v interface{}) bool {
    return v.(int)%2 == 0
})
fmt.Println(evenCount) // 3
```

### 3. Collection Manipulation

#### Remove

```go
// Remove even numbers
numbers := []int{1, 2, 3, 4, 5, 6}
odds := common.Remove(numbers, func(v interface{}) bool {
    return v.(int)%2 == 0
})
fmt.Println(odds) // [1 3 5]
```

#### Unique

```go
// Remove duplicates
numbers := []int{1, 2, 2, 3, 4, 4, 5, 5, 5}
unique := common.Unique(numbers)
fmt.Println(unique) // [1 2 3 4 5]
```

#### Reverse

```go
// Reverse in-place
numbers := []int{1, 2, 3, 4, 5}
common.Reverse(numbers)
fmt.Println(numbers) // [5 4 3 2 1]
```

#### Sort

```go
// Custom sort (ascending)
numbers := []int{5, 2, 8, 1, 9}
common.Sort(numbers, func(i, j int) bool {
    return numbers[i] < numbers[j]
})
fmt.Println(numbers) // [1 2 5 8 9]

// Sort descending
common.Sort(numbers, func(i, j int) bool {
    return numbers[i] > numbers[j]
})
fmt.Println(numbers) // [9 8 5 2 1]
```

### 4. Set Operations

#### Contains

```go
numbers := []int{1, 2, 3, 4, 5}
fmt.Println(common.Contains(numbers, 3))  // true
fmt.Println(common.Contains(numbers, 10)) // false
```

#### Difference

```go
// Elements in first collection but not in second
set1 := []int{1, 2, 3, 4, 5}
set2 := []int{3, 4, 5, 6, 7}
diff := common.Difference(set1, set2)
fmt.Println(diff) // [1 2]
```

#### Intersection

```go
// Common elements
set1 := []int{1, 2, 3, 4, 5}
set2 := []int{3, 4, 5, 6, 7}
intersection := common.Intersection(set1, set2)
fmt.Println(intersection) // [3 4 5]
```

### 5. Partitioning and Slicing

#### Partition

```go
// Split into evens and odds
numbers := []int{1, 2, 3, 4, 5, 6}
evens, odds := common.Partition(numbers, func(v interface{}) bool {
    return v.(int)%2 == 0
})
fmt.Println(evens) // [2 4 6]
fmt.Println(odds)  // [1 3 5]
```

#### Slice

```go
// Extract subrange
numbers := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
sub := common.Slice(numbers, 2, 7)
fmt.Println(sub) // [2 3 4 5 6]
```

#### SliceWithIndices

```go
// Extract specific indices
numbers := []int{10, 20, 30, 40, 50}
indices := []int{0, 2, 4}
selected := common.SliceWithIndices(numbers, indices)
fmt.Println(selected) // [10 30 50]
```

### 6. Advanced Operations

#### Zip

```go
// Combine multiple collections
names := []string{"Alice", "Bob", "Charlie"}
ages := []int{30, 25, 35}
zipped := common.Zip(names, ages)
fmt.Println(zipped)
// [[Alice 30] [Bob 25] [Charlie 35]]
```

#### RotateLeft

```go
numbers := []int{1, 2, 3, 4, 5}
rotated := common.RotateLeft(numbers, 2)
fmt.Println(rotated) // [3 4 5 1 2]
```

#### RotateRight

```go
numbers := []int{1, 2, 3, 4, 5}
rotated := common.RotateRight(numbers, 2)
fmt.Println(rotated) // [4 5 1 2 3]
```

#### Iterate

```go
// Custom iteration
numbers := []int{1, 2, 3, 4, 5}
common.Iterate(numbers, func(index int, value interface{}) {
    fmt.Printf("Index: %d, Value: %v\n", index, value)
})
// Output:
// Index: 0, Value: 1
// Index: 1, Value: 2
// Index: 2, Value: 3
// Index: 3, Value: 4
// Index: 4, Value: 5
```

### 7. Comparison and Type Checking

#### DeepEqual

```go
// Compare via JSON serialization
type Person struct {
    Name string
    Age  int
}

p1 := Person{Name: "Alice", Age: 30}
p2 := Person{Name: "Alice", Age: 30}
p3 := Person{Name: "Bob", Age: 25}

fmt.Println(common.DeepEqual(p1, p2)) // true
fmt.Println(common.DeepEqual(p1, p3)) // false
```

#### DeepEqualComp

```go
// Compare comparable types
slice1 := []int{1, 2, 3}
slice2 := []int{1, 2, 3}
slice3 := []int{1, 2, 4}

fmt.Println(common.DeepEqualComp(slice1, slice2)) // true
fmt.Println(common.DeepEqualComp(slice1, slice3)) // false
```

#### IsScalarType

```go
fmt.Println(common.IsScalarType(42))          // true (int)
fmt.Println(common.IsScalarType("hello"))     // true (string)
fmt.Println(common.IsScalarType(3.14))        // true (float64)
fmt.Println(common.IsScalarType(true))        // true (bool)
fmt.Println(common.IsScalarType([]int{1,2}))  // false (slice)
fmt.Println(common.IsScalarType(nil))         // false (nil)
```

#### IsEmptyValue

```go
import "reflect"

// Empty values
fmt.Println(common.IsEmptyValue(reflect.ValueOf("")))        // true
fmt.Println(common.IsEmptyValue(reflect.ValueOf(0)))         // true
fmt.Println(common.IsEmptyValue(reflect.ValueOf(false)))     // true
fmt.Println(common.IsEmptyValue(reflect.ValueOf([]int{})))   // true

// Non-empty values
fmt.Println(common.IsEmptyValue(reflect.ValueOf("hello")))   // false
fmt.Println(common.IsEmptyValue(reflect.ValueOf(42)))        // false
fmt.Println(common.IsEmptyValue(reflect.ValueOf([]int{1})))  // false
```

### 8. Practical Use Cases

#### Data Pipeline

```go
// Multi-step data transformation
type User struct {
    Name   string
    Age    int
    Active bool
}

users := []User{
    {Name: "Alice", Age: 30, Active: true},
    {Name: "Bob", Age: 25, Active: false},
    {Name: "Charlie", Age: 35, Active: true},
}

// Convert to interface{} for common package
usersInterface := make([]interface{}, len(users))
for i, u := range users {
    usersInterface[i] = u
}

// Filter active users
active := common.Filter(usersInterface, func(v interface{}) bool {
    return v.(User).Active
})

// Extract names
names := common.Transform(active, func(v interface{}) interface{} {
    return v.(User).Name
})

fmt.Println(names) // ["Alice" "Charlie"]
```

#### Statistical Operations

```go
numbers := []int{1, 2, 3, 4, 5}

// Convert to interface{}
numbersInterface := make([]interface{}, len(numbers))
for i, n := range numbers {
    numbersInterface[i] = n
}

// Calculate average
sum := common.Reduce(numbersInterface, func(acc interface{}, v interface{}) interface{} {
    return acc.(int) + v.(int)
}, 0).(int)

count := len(numbers)
average := float64(sum) / float64(count)
fmt.Printf("Average: %.2f\n", average) // Average: 3.00
```

## API Reference

### Functional Operations

| Function      | Signature                                               | Description                            |
| ------------- | ------------------------------------------------------- | -------------------------------------- |
| `Transform`   | `(collection any, fn func(any) any) any`                | Apply function to each element         |
| `Filter`      | `(collection any, fn func(any) bool) any`               | Keep elements matching condition       |
| `Reduce`      | `(collection any, fn func(any, any) any, init any) any` | Reduce to single value (left-to-right) |
| `ReduceRight` | `(collection any, fn func(any, any) any, init any) any` | Reduce to single value (right-to-left) |
| `Iterate`     | `(collection any, fn func(int, any))`                   | Iterate with callback                  |

### Collection Queries

| Function   | Signature                                  | Description                  |
| ---------- | ------------------------------------------ | ---------------------------- |
| `Find`     | `(collection any, fn func(any) bool) any`  | Find first matching element  |
| `All`      | `(collection any, fn func(any) bool) bool` | Check if all match condition |
| `Any`      | `(collection any, fn func(any) bool) bool` | Check if any match condition |
| `Count`    | `(collection any, fn func(any) bool) int`  | Count matching elements      |
| `Contains` | `(collection any, element any) bool`       | Check if element exists      |

### Collection Manipulation

| Function  | Signature                                  | Description              |
| --------- | ------------------------------------------ | ------------------------ |
| `Remove`  | `(collection any, fn func(any) bool) any`  | Remove matching elements |
| `Unique`  | `(collection any) any`                     | Remove duplicates        |
| `Reverse` | `(collection any)`                         | Reverse in-place         |
| `Sort`    | `(collection any, fn func(i, j int) bool)` | Custom sort in-place     |

### Set Operations

| Function       | Signature              | Description                  |
| -------------- | ---------------------- | ---------------------------- |
| `Difference`   | `(col1, col2 any) any` | Elements in col1 not in col2 |
| `Intersection` | `(col1, col2 any) any` | Common elements              |

### Slicing and Partitioning

| Function           | Signature                                        | Description         |
| ------------------ | ------------------------------------------------ | ------------------- |
| `Slice`            | `(collection any, start, end int) any`           | Extract subrange    |
| `SliceWithIndices` | `(collection any, indices []int) any`            | Extract by indices  |
| `Partition`        | `(collection any, fn func(any) bool) (any, any)` | Split by condition  |
| `Zip`              | `(collections ...any) []any`                     | Combine collections |
| `RotateLeft`       | `(collection any, positions int) any`            | Rotate left         |
| `RotateRight`      | `(collection any, positions int) any`            | Rotate right        |

### Comparison and Type Checking

| Function        | Signature                     | Description                        |
| --------------- | ----------------------------- | ---------------------------------- |
| `DeepEqual`     | `(a, b any) bool`             | Compare via JSON serialization     |
| `DeepEqualComp` | `[T comparable](a, b T) bool` | Deep equality for comparable types |
| `IsScalarType`  | `(value any) bool`            | Check if primitive type            |
| `IsEmptyValue`  | `(v reflect.Value) bool`      | Check if reflect.Value is empty    |

## Best Practices & Notes

### ⚠️ Common Pitfalls

1. **Type Assertions**

   ```go
   // ❌ Forgetting type assertion
   numbers := []int{1, 2, 3}
   result := common.Transform(numbers, func(v interface{}) interface{} {
       return v * v // Won't compile!
   })

   // ✅ Correct: type assert
   result := common.Transform(numbers, func(v interface{}) interface{} {
       return v.(int) * v.(int)
   })
   ```

2. **Panic on Wrong Type**

   ```go
   // ❌ Dangerous: will panic if type is wrong
   numbers := []interface{}{1, "two", 3}
   squared := common.Transform(numbers, func(v interface{}) interface{} {
       return v.(int) * v.(int) // Panics on "two"
   })

   // ✅ Safe: check type
   squared := common.Transform(numbers, func(v interface{}) interface{} {
       if num, ok := v.(int); ok {
           return num * num
       }
       return v // or handle error
   })
   ```

3. **Modifying Original Collections**

   ```go
   // ⚠️ Reverse modifies in-place
   numbers := []int{1, 2, 3}
   common.Reverse(numbers)
   fmt.Println(numbers) // [3 2 1] - original modified!

   // ⚠️ Sort modifies in-place
   common.Sort(numbers, func(i, j int) bool {
       return numbers[i] < numbers[j]
   })
   ```

4. **Performance Overhead**

   ```go
   // ❌ Slow for performance-critical code
   for i := 0; i < 1000000; i++ {
       result := common.Transform(data, transform)
   }

   // ✅ Use type-specific code or generics
   for i := 0; i < 1000000; i++ {
       for j := range data {
           result[j] = transform(data[j])
       }
   }
   ```

### 💡 Recommendations

✅ **Use for dynamic types**

```go
// When type is unknown at compile time
func ProcessData(data interface{}) {
    filtered := common.Filter(data, condition)
    // ...
}
```

✅ **Consider type-safe alternatives**

```go
// Prefer generics (Go 1.18+) when type is known
import "github.com/polarixa/replify/pkg/coll"

numbers := []int{1, 2, 3, 4, 5}
evens := coll.Filter(numbers, func(n int) bool {
    return n%2 == 0
}) // Type-safe, no assertions needed
```

✅ **Handle type assertion errors**

```go
result := common.Transform(data, func(v interface{}) interface{} {
    if num, ok := v.(int); ok {
        return num * 2
    }
    return nil // or default value
})
```

✅ **Document expected types**

```go
// ProcessNumbers expects a slice or array of integers
func ProcessNumbers(numbers interface{}) interface{} {
    return common.Filter(numbers, func(v interface{}) bool {
        return v.(int) > 0
    })
}
```

✅ **Check collection types before operations**

```go
if reflect.TypeOf(data).Kind() == reflect.Slice {
    result := common.Transform(data, fn)
}
```

### 🔒 Thread Safety

**Not thread-safe by default.** Functions that modify in-place (`Reverse`, `Sort`) should not be called concurrently on the same collection.

```go
// ❌ Unsafe
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        common.Reverse(sharedSlice) // Race condition!
    }()
}

// ✅ Safe: use mutex or create copies
var mu sync.Mutex
mu.Lock()
common.Reverse(sharedSlice)
mu.Unlock()
```

### ⚡ Performance Considerations

**Reflection is slow** compared to type-specific operations:

```go
// Benchmark comparison
// Type-specific:     ~1 ns/op
// Reflection-based:  ~100 ns/op (100x slower)
```

**When to optimize:**

- Hot paths in your application
- Large datasets (>10,000 elements)
- Tight loops with repeated operations

**Optimization strategies:**

1. Use generics instead (Go 1.18+)
2. Write type-specific implementations
3. Profile before optimizing
4. Cache reflection results when possible

### 🐛 Debugging Tips

**Print types:**

```go
fmt.Printf("Type: %T, Value: %v\n", value, value)
```

**Check for nil:**

```go
if value == nil {
    fmt.Println("Value is nil")
}
```

**Inspect reflection values:**

```go
v := reflect.ValueOf(data)
fmt.Printf("Kind: %v, Type: %v, Len: %d\n", v.Kind(), v.Type(), v.Len())
```

### 📝 Testing

Example tests:

```go
func TestTransform(t *testing.T) {
    numbers := []int{1, 2, 3}
    result := common.Transform(numbers, func(v interface{}) interface{} {
        return v.(int) * 2
    })

    expected := []interface{}{2, 4, 6}
    if !common.DeepEqual(result, expected) {
        t.Errorf("Got %v, want %v", result, expected)
    }
}
```

## Limitations

- **Runtime type checks**: No compile-time type safety
- **Performance overhead**: Reflection is slower than direct operations
- **Panic risk**: Type assertions can panic if incorrect
- **No compile-time errors**: Type mistakes discovered at runtime
- **Memory overhead**: Interface{} boxes require heap allocations

## When to Use vs. Generics

**Use `common` when:**

- Type is unknown at compile time
- Working with `interface{}` types
- Need dynamic type handling
- Building flexible libraries

**Use generics when:**

- Type is known at compile time
- Performance matters
- Want type safety
- Using Go 1.18+

## Contributing

Contributions are welcome! Please see the main [replify repository](https://github.com/polarixa/replify) for contribution guidelines.

## License

This library is part of the [replify](https://github.com/polarixa/replify) project.

## Related

Part of the **replify** ecosystem:

- [replify](https://github.com/polarixa/replify) - API response wrapping library
- [conv](https://github.com/polarixa/replify/pkg/conv) - Type conversion utilities
- [coll](https://github.com/polarixa/replify/pkg/coll) - Type-safe collection utilities (generics)
- [hashy](https://github.com/polarixa/replify/pkg/hashy) - Deterministic hashing
- [match](https://github.com/polarixa/replify/pkg/match) - Wildcard pattern matching
- [strutil](https://github.com/polarixa/replify/pkg/strutil) - String utilities
- [randn](https://github.com/polarixa/replify/pkg/randn) - Random data generation
- [encoding](https://github.com/polarixa/replify/pkg/encoding) - JSON encoding utilities
