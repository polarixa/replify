# coll

**coll** is a powerful, type-safe Go library providing functional programming utilities for working with collections (slices, maps, sets, and stacks). Built with Go generics, it offers a rich set of operations for transforming, filtering, and manipulating data structures in a clean, expressive way.

## Overview

The `coll` package provides a comprehensive suite of collection utilities inspired by functional programming patterns. It eliminates boilerplate code for common operations like filtering, mapping, reducing, and grouping, while maintaining type safety through Go's generics system.

**Key Features:**

- 🎯 **Type-safe generics** - works with any type using Go 1.18+ generics
- 🔄 **Functional operations** - map, filter, reduce, and more
- 📦 **Multiple data structures** - slices, maps, sets, stacks, and hashmaps
- ⚡ **Zero dependencies** - built on Go standard library
- 🧩 **Composable** - chain operations for complex transformations
- 🎨 **Immutable operations** - most functions return new collections
- 🔍 **Rich API** - 40+ utility functions for collection manipulation

**Problem Solved:** Writing repetitive loops for common collection operations clutters code and increases error risk. `coll` provides tested, optimized functions that make collection manipulation readable and maintainable.

## Use Cases

### When to Use

- ✅ **Data transformation** - convert between types, reshape structures
- ✅ **Filtering and searching** - find, filter, or check elements
- ✅ **Aggregation** - sum, count, group, or reduce data
- ✅ **Set operations** - unique values, intersections, differences
- ✅ **Functional pipelines** - chain operations for clean data processing
- ✅ **Collection utilities** - chunk, flatten, partition, shuffle
- ✅ **Stack/queue operations** - LIFO data structure management
- ✅ **HashMap operations** - type-safe key-value storage

### When Not to Use

- ❌ **Simple iterations** - use native `for` loops for basic iteration
- ❌ **Performance-critical hot paths** - hand-optimized loops may be faster
- ❌ **When mutability is required** - most functions return new collections
- ❌ **Large datasets** - consider streaming or database operations
- ❌ **Complex queries** - use SQL/NoSQL databases for advanced querying

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package in your Go code:

```go
import "github.com/polarixa/replify/pkg/coll"
```

**Requirements:** Go 1.18 or higher (for generics support)

## Usage

### Basic Operations

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/coll"
)

func main() {
    // Filter even numbers
    numbers := []int{1, 2, 3, 4, 5, 6}
    evens := coll.Filter(numbers, func(n int) bool {
        return n%2 == 0
    })
    fmt.Println(evens) // [2, 4, 6]

    // Map: square each number
    squared := coll.Map(numbers, func(n int) int {
        return n * n
    })
    fmt.Println(squared) // [1, 4, 9, 16, 25, 36]

    // Reduce: sum all numbers
    sum := coll.Reduce(numbers, func(acc, n int) int {
        return acc + n
    }, 0)
    fmt.Println(sum) // 21

    // Check if contains
    hasThree := coll.Contains(numbers, 3)
    fmt.Println(hasThree) // true
}
```

## Examples

### 1. Filtering and Transformation

```go
type User struct {
    Name  string
    Age   int
    Active bool
}

users := []User{
    {Name: "Alice", Age: 30, Active: true},
    {Name: "Bob", Age: 25, Active: false},
    {Name: "Charlie", Age: 35, Active: true},
}

// Filter active users
activeUsers := coll.Filter(users, func(u User) bool {
    return u.Active
})

// Extract names
names := coll.Map(activeUsers, func(u User) string {
    return u.Name
})
fmt.Println(names) // ["Alice", "Charlie"]

// Get ages of active users
ages := coll.Map(activeUsers, func(u User) int {
    return u.Age
})
```

### 2. Reducing and Aggregating

```go
// Sum of numbers
numbers := []int{1, 2, 3, 4, 5}
sum := coll.Reduce(numbers, func(acc, n int) int {
    return acc + n
}, 0)
fmt.Println(sum) // 15

// Calculate total with Sum (using transformer)
type Product struct {
    Name  string
    Price float64
}

products := []Product{
    {Name: "Book", Price: 9.99},
    {Name: "Pen", Price: 1.50},
    {Name: "Notebook", Price: 5.99},
}

totalPrice := coll.Sum(products, func(p Product) float64 {
    return p.Price
})
fmt.Println(totalPrice) // 17.48

// Concatenate strings
words := []string{"Hello", "World", "from", "Go"}
sentence := coll.Reduce(words, func(acc, word string) string {
    return acc + " " + word
}, "")
fmt.Println(sentence) // " Hello World from Go"
```

### 3. Finding and Searching

```go
numbers := []int{10, 20, 30, 40, 50}

// Find first element matching condition
first, found := coll.Find(numbers, func(n int) bool {
    return n > 25
})
fmt.Println(first, found) // 30, true

// Find index
index := coll.IndexOf(numbers, 30)
fmt.Println(index) // 2

// Find last index
lastIndex := coll.LastIndexOf([]int{1, 2, 3, 2, 1}, 2)
fmt.Println(lastIndex) // 3

// Check if any element matches
hasLarge := coll.AnyMatch(numbers, func(n int) bool {
    return n > 40
})
fmt.Println(hasLarge) // true

// Check if all elements match
allPositive := coll.AllMatch(numbers, func(n int) bool {
    return n > 0
})
fmt.Println(allPositive) // true
```

### 4. Set Operations

```go
// Remove duplicates
numbers := []int{1, 2, 2, 3, 3, 3, 4, 5, 5}
unique := coll.Unique(numbers)
fmt.Println(unique) // [1, 2, 3, 4, 5]

// Intersection
set1 := []int{1, 2, 3, 4}
set2 := []int{3, 4, 5, 6}
intersection := coll.Intersection(set1, set2)
fmt.Println(intersection) // [3, 4]

// Union
union := coll.Union(set1, set2)
fmt.Println(union) // [1, 2, 3, 4, 5, 6]

// Difference
diff := coll.Difference(set1, set2)
fmt.Println(diff) // [1, 2]
```

### 5. Chunking and Partitioning

```go
// Chunk into smaller slices
numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
chunks := coll.Chunk(numbers, 3)
fmt.Println(chunks) // [[1, 2, 3], [4, 5, 6], [7, 8, 9]]

// Partition by condition
evensAndOdds := coll.Partition(numbers, func(n int) bool {
    return n%2 == 0
})
fmt.Println(evensAndOdds) // [[2, 4, 6, 8], [1, 3, 5, 7, 9]]

// Split at index
left, right := coll.Split(numbers, 5)
fmt.Println(left, right) // [1, 2, 3, 4, 5], [6, 7, 8, 9]
```

### 6. Grouping and Flattening

```go
// Group by key
type Person struct {
    Name string
    Age  int
}

people := []Person{
    {Name: "Alice", Age: 30},
    {Name: "Bob", Age: 25},
    {Name: "Charlie", Age: 30},
}

byAge := coll.GroupBy(people, func(p Person) int {
    return p.Age
})
// Result: map[25:[{Bob 25}] 30:[{Alice 30} {Charlie 30}]]

// Flatten nested slices
nested := [][]int{{1, 2}, {3, 4}, {5}}
flat := coll.Flatten[int](convertToInterfaceSlice(nested))
fmt.Println(flat) // [1, 2, 3, 4, 5]
```

### 7. Sorting and Shuffling

```go
// Sort with custom comparer
numbers := []int{5, 2, 8, 1, 9}
sorted := coll.Sort(numbers, func(a, b int) bool {
    return a < b // ascending
})
fmt.Println(sorted) // [1, 2, 5, 8, 9]

// Sort descending
descending := coll.Sort(numbers, func(a, b int) bool {
    return a > b
})
fmt.Println(descending) // [9, 8, 5, 2, 1]

// Shuffle randomly
shuffled := coll.Shuffle(numbers)
fmt.Println(shuffled) // Random order

// Reverse
reversed := coll.Reverse(numbers)
fmt.Println(reversed) // [9, 1, 8, 2, 5]
```

### 8. Stack Operations

```go
// Create a stack
stack := coll.NewStack[string]()

// Push elements
stack.Push("first")
stack.Push("second")
stack.Push("third")

// Peek at top
top := stack.Peek()
fmt.Println(top) // "third"

// Pop elements
popped := stack.Pop()
fmt.Println(popped) // "third"

// Check if empty
isEmpty := stack.IsEmpty()
fmt.Println(isEmpty) // false

// Get size
size := stack.Size()
fmt.Println(size) // 2

// Clear all
stack.Clear()
```

### 9. HashMap Operations

```go
// Create a HashMap
hashMap := coll.NewHashMap[string, int]()

// Set values
hashMap.Set("apple", 5)
hashMap.Set("banana", 3)
hashMap.Set("orange", 7)

// Get value
count, exists := hashMap.Get("apple")
fmt.Println(count, exists) // 5, true

// Check if key exists
hasKey := hashMap.ContainsKey("banana")
fmt.Println(hasKey) // true

// Remove key
hashMap.Remove("orange")

// Get size
size := hashMap.Size()
fmt.Println(size) // 2

// Get all keys
keys := hashMap.Keys()
fmt.Println(keys) // ["apple", "banana"]

// Clear all
hashMap.Clear()
```

### 10. HashSet Operations

```go
// Create a HashSet
set := coll.NewHashSet[string]()

// Add elements
set.Add("apple")
set.Add("banana")
set.Add("apple") // Duplicate, won't be added

// Check if contains
hasApple := set.Contains("apple")
fmt.Println(hasApple) // true

// Remove element
set.Remove("banana")

// Get size
size := set.Size()
fmt.Println(size) // 1

// Get all values
values := set.Values()
fmt.Println(values) // ["apple"]

// Clear all
set.Clear()
```

### 11. Map Operations

```go
// Check if map contains key
ages := map[string]int{"Alice": 30, "Bob": 25}
hasAlice := coll.ContainsKeyComp(ages, "Alice")
fmt.Println(hasAlice) // true

// Get map keys
keys := coll.Keys(ages)
fmt.Println(keys) // ["Alice", "Bob"]

// Get map values
values := coll.Values(ages)
fmt.Println(values) // [30, 25]

// Merge maps
map1 := map[string]int{"a": 1, "b": 2}
map2 := map[string]int{"b": 3, "c": 4}
merged := coll.MergeMaps(map1, map2)
fmt.Println(merged) // map[a:1 b:3 c:4] (map2 overwrites map1)
```

### 12. Slice Utilities

```go
// Take first N elements
numbers := []int{1, 2, 3, 4, 5}
firstThree := coll.Take(numbers, 3)
fmt.Println(firstThree) // [1, 2, 3]

// Skip first N elements
remaining := coll.Skip(numbers, 2)
fmt.Println(remaining) // [3, 4, 5]

// Slice range
middle := coll.SliceRange(numbers, 1, 4)
fmt.Println(middle) // [2, 3, 4]

// Append element
appended := coll.Push(numbers, 6)
fmt.Println(appended) // [1, 2, 3, 4, 5, 6]

// Remove last element
popped := coll.Pop(numbers)
fmt.Println(popped) // [1, 2, 3, 4]

// Check equality
slice1 := []int{1, 2, 3}
slice2 := []int{1, 2, 3}
areEqual := coll.Equal(slice1, slice2)
fmt.Println(areEqual) // true
```

## API Reference

### Slice Operations

#### Transformation

- `Map[T, U](slice []T, f func(T) U) []U` - Transform each element
- `FlatMap[T, U](slice []T, mapper func(T) []U) []U` - Map and flatten
- `MapWithIndex[T, U](slice []T, mapper func(T, int) U) []U` - Map with index

#### Filtering

- `Filter[T](slice []T, condition func(T) bool) []T` - Keep matching elements
- `Reject[T](slice []T, condition func(T) bool) []T` - Remove matching elements
- `Compact[T](slice []T) []T` - Remove zero values

#### Searching

- `Contains[T](slice []T, item T) bool` - Check if element exists
- `Find[T](slice []T, predicate func(T) bool) (T, bool)` - Find first match
- `IndexOf[T](slice []T, item T) int` - Get index of element
- `LastIndexOf[T](slice []T, item T) int` - Get last index

#### Aggregation

- `Reduce[T, U](slice []T, accumulator func(U, T) U, initial U) U` - Reduce to single value
- `Sum[T](slice []T, transformer func(T) float64) float64` - Sum transformed values

#### Predicates

- `AllMatch[T](slice []T, predicate func(T) bool) bool` - Check if all match
- `AnyMatch[T](slice []T, predicate func(T) bool) bool` - Check if any match
- `NoneMatch[T](slice []T, predicate func(T) bool) bool` - Check if none match

#### Set Operations

- `Unique[T](slice []T) []T` - Remove duplicates
- `Intersection[T](a, b []T) []T` - Common elements
- `Union[T](a, b []T) []T` - All unique elements from both
- `Difference[T](a, b []T) []T` - Elements in a but not in b

#### Partitioning

- `Chunk[T](slice []T, size int) [][]T` - Split into chunks
- `Partition[T](slice []T, predicate func(T) bool) ([]T, []T)` - Split by condition
- `GroupBy[T, K](slice []T, keyFunc func(T) K) map[K][]T` - Group by key
- `Split[T](slice []T, index int) ([]T, []T)` - Split at index

#### Ordering

- `Sort[T](slice []T, comparer func(T, T) bool) []T` - Custom sort
- `Reverse[T](slice []T) []T` - Reverse order
- `Shuffle[T](slice []T) []T` - Random order

#### Utilities

- `Take[T](slice []T, n int) []T` - First N elements
- `Skip[T](slice []T, n int) []T` - Skip first N elements
- `Push[T](slice []T, element T) []T` - Append element
- `Pop[T](slice []T) []T` - Remove last element
- `Flatten[T](slice []any) []T` - Flatten nested slices
- `Equal[T](a, b []T) bool` - Check equality

### Map Operations

- `Map[T, U](slice []T, f func(T) U) []U` - Transform elements
- `ContainsKeyComp[K, V](m map[K]V, key K) bool` - Check if key exists
- `Keys[K, V](m map[K]V) []K` - Get all keys
- `Values[K, V](m map[K]V) []V` - Get all values
- `MergeMaps[K, V](maps ...map[K]V) map[K]V` - Merge multiple maps

### Stack[T comparable]

- `NewStack[T]() *Stack[T]` - Create new stack
- `Push(element T)` - Add to top
- `Pop() T` - Remove and return top
- `Peek() T` - View top without removing
- `IsEmpty() bool` - Check if empty
- `Size() int` - Get number of elements
- `Clear()` - Remove all elements

### HashMap[K, V comparable]

- `NewHashMap[K, V]() *HashMap[K, V]` - Create new hashmap
- `Set(key K, value V)` - Set key-value pair
- `Get(key K) (V, bool)` - Get value by key
- `ContainsKey(key K) bool` - Check if key exists
- `Remove(key K)` - Remove key-value pair
- `Keys() []K` - Get all keys
- `Size() int` - Get number of pairs
- `Clear()` - Remove all pairs

### HashSet[T comparable]

- `NewHashSet[T]() *HashSet[T]` - Create new set
- `Add(item T)` - Add element
- `Contains(item T) bool` - Check if element exists
- `Remove(item T)` - Remove element
- `Values() []T` - Get all values
- `Size() int` - Get number of elements
- `Clear()` - Remove all elements

## Best Practices & Notes

### ⚠️ Common Pitfalls

1. **Mutability**: Most functions return new collections, not modifying originals

   ```go
   // ❌ Wrong: expecting mutation
   numbers := []int{1, 2, 3}
   coll.Filter(numbers, func(n int) bool { return n > 1 })
   // numbers is still [1, 2, 3]

   // ✅ Correct: capture return value
   filtered := coll.Filter(numbers, func(n int) bool { return n > 1 })
   ```

2. **Empty Slices**: Some functions may panic on empty slices

   ```go
   // ❌ Panics
   empty := []int{}
   coll.Pop(empty) // Runtime panic

   // ✅ Check length first
   if len(slice) > 0 {
       popped := coll.Pop(slice)
   }
   ```

3. **Performance**: Creating new slices has memory overhead

   ```go
   // ❌ Inefficient for large datasets
   for i := 0; i < 1000000; i++ {
       result = coll.Push(result, i) // Reallocates each time
   }

   // ✅ Pre-allocate when possible
   result := make([]int, 0, 1000000)
   for i := 0; i < 1000000; i++ {
       result = append(result, i)
   }
   ```

4. **Type Inference**: Sometimes you need explicit type parameters

   ```go
   // ❌ May fail to infer
   result := coll.Map(data, transform)

   // ✅ Explicit types
   result := coll.Map[Input, Output](data, transform)
   ```

### 💡 Recommendations

✅ **Chain operations** for readable pipelines

```go
result := coll.Filter(data, isValid)
result = coll.Map(result, transform)
result = coll.Sort(result, compare)
```

✅ **Use method chaining alternative**

```go
// Consider creating a fluent API wrapper for your use case
type Pipeline[T any] struct {
    data []T
}

func (p Pipeline[T]) Filter(f func(T) bool) Pipeline[T] {
    return Pipeline[T]{coll.Filter(p.data, f)}
}
```

✅ **Leverage type safety** - let compiler catch errors

```go
// Compile-time safety
users := []User{}
names := coll.Map(users, func(u User) string { return u.Name })
// names is []string, enforced by compiler
```

✅ **Document complex transformations** with comments

```go
// Transform users to active user emails
emails := coll.Map(
    coll.Filter(users, func(u User) bool { return u.Active }),
    func(u User) string { return u.Email },
)
```

✅ **Consider performance** for large datasets

- Use native loops for simple iterations
- Profile before optimizing
- Consider streaming for huge datasets

✅ **Use appropriate data structures**

- `Stack` for LIFO operations
- `HashSet` for uniqueness checks
- `HashMap` for key-value lookups
- Slices for ordered data

### 🔒 Thread Safety

**Not thread-safe by default.** All data structures and operations assume single-threaded access.

```go
// ❌ Unsafe concurrent access
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        stack.Push(i) // Race condition!
    }()
}

// ✅ Use mutex for concurrent access
var mu sync.Mutex
stack := coll.NewStack[int]()

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(val int) {
        defer wg.Done()
        mu.Lock()
        stack.Push(val)
        mu.Unlock()
    }(i)
}
```

### ⚡ Performance Tips

**Benchmark before optimizing:**

```go
func BenchmarkCollMap(b *testing.B) {
    data := generateLargeSlice()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        coll.Map(data, transform)
    }
}
```

**Performance characteristics:**

- `Map`, `Filter`: O(n)
- `Contains`, `IndexOf`: O(n)
- `Unique`: O(n) with map
- `Sort`: O(n log n)
- `Intersection`, `Union`: O(n + m)
- `Flatten`: O(n × depth)

**Optimization strategies:**

1. Pre-allocate slices when size is known
2. Use `break` in `for` loops for early exit
3. Avoid nested `Map`/`Filter` - combine logic
4. Use `HashMap` for frequent lookups
5. Consider native loops for hot paths

### 🐛 Debugging Tips

**Print intermediate results:**

```go
filtered := coll.Filter(data, condition)
fmt.Printf("After filter: %v\n", filtered)

mapped := coll.Map(filtered, transform)
fmt.Printf("After map: %v\n", mapped)
```

**Use meaningful function names:**

```go
// ❌ Unclear
coll.Filter(users, func(u User) bool { return u.A > 18 && u.S == "active" })

// ✅ Clear
isEligible := func(u User) bool {
    return u.Age > 18 && u.Status == "active"
}
coll.Filter(users, isEligible)
```

**Test edge cases:**

- Empty slices
- Single element
- All elements match/don't match
- Duplicates
- Nested structures

### 📝 Testing

Example test cases:

```go
func TestFilter(t *testing.T) {
    tests := []struct {
        name   string
        input  []int
        pred   func(int) bool
        want   []int
    }{
        {"empty", []int{}, func(n int) bool { return n > 0 }, []int{}},
        {"all match", []int{1, 2, 3}, func(n int) bool { return n > 0 }, []int{1, 2, 3}},
        {"none match", []int{1, 2, 3}, func(n int) bool { return n > 10 }, []int{}},
        {"some match", []int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 }, []int{2, 4}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := coll.Filter(tt.input, tt.pred)
            if !coll.Equal(got, tt.want) {
                t.Errorf("Filter() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Limitations

- **No lazy evaluation** - all operations are eager (process entire collection)
- **Memory overhead** - most operations create new collections
- **No parallel processing** - all operations are sequential
- **Limited to comparable types** - some structures require comparable types
- **No query optimization** - operations are applied in order given
- **Stack size limits** - recursive operations may hit stack limits on huge nested data

## Performance Considerations

**When to use native loops:**

- Simple iteration with no transformation
- Performance-critical sections (after profiling)
- When you need early termination with complex logic
- Mutation of existing slices

**When to use coll:**

- Complex transformations with multiple steps
- Readability is more important than micro-optimization
- Working with functional patterns
- Rapid prototyping
- Most business logic (not hot paths)

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
- Other sivaosorg utilities

---

**Note:** The search results shown above may be incomplete due to GitHub's result limits. For a complete view of all functions, please visit the [coll package on GitHub](https://github.com/polarixa/replify/tree/master/pkg/coll).
