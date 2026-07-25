package coll

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/polarixa/replify/pkg/encoding"
)

// Contains checks if a specified item is present within a given slice.
//
// This function iterates over a slice of any type that supports comparison
// and checks if the specified `item` exists within it. It returns `true`
// if the `item` is found and `false` otherwise.
//
// The function is generic, so it can be used with any comparable type,
// including strings, integers, floats, or custom types that implement the
// comparable interface.
//
// Parameters:
//   - `array`: The slice of elements to search through. This slice can
//     contain any type `T` that supports comparison (e.g., int, string).
//   - `item`: The item to search for within `array`. It should be of the
//     same type `T` as the elements in `array`.
//
// Returns:
//   - `true` if `item` is found within `array`, `false` otherwise.
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	isPresent := Contains(numbers, 3) // isPresent will be true as 3 is in the slice
//
//	names := []string{"Alice", "Bob", "Charlie"}
//	isPresent := Contains(names, "Eve") // isPresent will be false as "Eve" is not in the slice
func Contains[T comparable](array []T, item T) bool {
	if len(array) == 0 {
		return false
	}
	for _, v := range array {
		if v == item {
			return true
		}
	}
	return false
}

// Filter returns a new slice containing only the elements from the input slice
// that satisfy a specified condition.
//
// This function iterates over each element in the input slice `list` and applies
// the provided `condition` function to it. If the `condition` function returns `true`
// for an element, that element is added to the `filtered` slice. At the end, the
// function returns the `filtered` slice containing only the elements that met the
// condition.
//
// The function is generic, allowing it to work with slices of any type `T` and
// any condition function that takes a `T` and returns a boolean.
//
// Parameters:
//   - `list`: The slice of elements to filter. It can contain elements of any type `T`.
//   - `condition`: A function that defines the filtering criteria. It takes an element
//     of type `T` as input and returns `true` if the element should be included in
//     the result, or `false` if it should be excluded.
//
// Returns:
//   - A new slice of type `[]T` containing only the elements from `list` for which
//     the `condition` function returned `true`.
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	oddNumbers := Filter(numbers, func(n int) bool { return n%2 != 0 })
//	// oddNumbers will be []int{1, 3, 5} as only the odd numbers satisfy the condition
//
//	words := []string{"apple", "banana", "cherry"}
//	longWords := Filter(words, func(word string) bool { return len(word) > 5 })
//	// longWords will be []string{"banana", "cherry"} as they are longer than 5 characters
func Filter[T any](list []T, condition func(T) bool) []T {
	filtered := make([]T, 0)
	for _, item := range list {
		if condition(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// Concat returns a new slice that is the result of concatenating multiple input slices
// into a single slice.
//
// This function takes a variable number of slices as input and combines them into
// one contiguous slice. It first calculates the total length needed for the resulting
// slice, then copies each input slice into the appropriate position within the
// resulting slice.
//
// The function is generic, allowing it to concatenate slices of any type `T`.
//
// Parameters:
//   - `slices`: A variadic parameter representing the slices to concatenate. Each slice
//     can contain elements of any type `T`, and they will be concatenated in the order
//     they are provided.
//
// Returns:
//   - A new slice of type `[]T` containing all elements from each input slice in sequence.
//
// Example:
//
//	// Concatenating integer slices
//	a := []int{1, 2}
//	b := []int{3, 4}
//	c := []int{5, 6}
//	combined := Concat(a, b, c)
//	// combined will be []int{1, 2, 3, 4, 5, 6}
//
//	// Concatenating string slices
//	words1 := []string{"hello", "world"}
//	words2 := []string{"go", "lang"}
//	concatenatedWords := Concat(words1, words2)
//	// concatenatedWords will be []string{"hello", "world", "go", "lang"}
func Concat[T any](slices ...[]T) []T {
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}
	result := make([]T, totalLen)
	i := 0
	for _, s := range slices {
		copy(result[i:], s)
		i += len(s)
	}
	return result
}

// Sum calculates the sum of elements in a slice after transforming each element to a float64.
//
// This function iterates over each element in the input slice `slice`, applies a transformation
// function `transformer` to convert the element to a float64, and adds the result to a running
// total. The final sum is returned as a float64.
//
// The function is generic, allowing it to operate on slices of any type `T`. The `transformer`
// function is used to convert each element to a float64, enabling flexible summation of
// different types (e.g., integers, custom types with numeric properties).
//
// Parameters:
//   - `slice`: The slice of elements to sum. It can contain elements of any type `T`.
//   - `transformer`: A function that takes an element of type `T` and returns a float64
//     representation, which will be used in the summation.
//
// Returns:
//   - A float64 representing the sum of the transformed elements.
//
// Example:
//
//	// Summing integer slice values
//	numbers := []int{1, 2, 3, 4}
//	total := Sum(numbers, func(n int) float64 { return float64(n) })
//	// total will be 10.0 as each integer is converted to float64 and summed
//
//	// Summing custom struct values
//	type Product struct {
//	    Price float64
//	}
//	products := []Product{{Price: 9.99}, {Price: 19.99}, {Price: 5.0}}
//	totalPrice := Sum(products, func(p Product) float64 { return p.Price })
//	// totalPrice will be 34.98 as the prices are summed
func Sum[T any](slice []T, transformer func(T) float64) float64 {
	sum := 0.0
	for _, item := range slice {
		sum += transformer(item)
	}
	return sum
}

// Equal checks if two slices are equal in both length and elements.
//
// This function compares two slices `a` and `b` of any comparable type `T`. It first
// checks if the lengths of the two slices are the same. If they are not, it returns `false`.
// If the lengths match, it then iterates through each element in `a` and `b` to check
// if corresponding elements are equal. If all elements are equal, the function returns `true`;
// otherwise, it returns `false`.
//
// The function is generic and can be used with slices of any comparable type, such as
// integers, strings, or other types that support equality comparison.
//
// Parameters:
//   - `a`: The first slice to compare. It should contain elements of a comparable type `T`.
//   - `b`: The second slice to compare. It should also contain elements of type `T`.
//
// Returns:
//   - `true` if both slices have the same length and identical elements at each position;
//     `false` otherwise.
//
// Example:
//
//	// Comparing integer slices
//	a := []int{1, 2, 3}
//	b := []int{1, 2, 3}
//	isEqual := Equal(a, b)
//	// isEqual will be true as both slices contain the same elements in the same order
//
//	c := []int{1, 2, 4}
//	isEqual = Equal(a, c)
//	// isEqual will be false as the elements differ
//
//	// Comparing string slices
//	names1 := []string{"Alice", "Bob"}
//	names2 := []string{"Alice", "Bob"}
//	isEqual = Equal(names1, names2)
//	// isEqual will be true since the slices have identical elements
func Equal[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Reduce applies an accumulator function over a slice, producing a single accumulated result.
//
// This function iterates over each element in the input slice `slice`, applying the
// `accumulator` function to combine each element with an accumulated result. It starts
// with an initial value `initialValue` and successively updates the result by applying
// `accumulator` to each element in `slice`. The final accumulated result is returned
// once all elements have been processed.
//
// The function is generic, allowing it to operate on slices of any type `T` and produce
// an output of any type `U`. This enables flexible aggregation operations such as
// summing, counting, or accumulating data into more complex structures.
//
// Parameters:
//   - `slice`: The slice of elements to reduce. It can contain elements of any type `T`.
//   - `accumulator`: A function that takes the current accumulated result of type `U`
//     and an element of type `T`, then returns the updated accumulated result of type `U`.
//   - `initialValue`: The initial value for the accumulator, of type `U`. This is the
//     starting point for the reduction process.
//
// Returns:
//   - The final accumulated result of type `U` after applying `accumulator` to each element
//     in `slice`.
//
// Example:
//
//	// Summing integer values in a slice
//	numbers := []int{1, 2, 3, 4}
//	sum := Reduce(numbers, func(acc, n int) int { return acc + n }, 0)
//	// sum will be 10 as each integer is added to the accumulated result
//
//	// Concatenating strings in a slice
//	words := []string{"go", "is", "fun"}
//	sentence := Reduce(words, func(acc, word string) string { return acc + " " + word }, "")
//	// sentence will be " go is fun" (note leading space due to initial value being "")
//
//	// Using a custom struct and custom accumulator
//	type Product struct {
//	    Name  string
//	    Price float64
//	}
//	products := []Product{{Name: "apple", Price: 0.99}, {Name: "banana", Price: 1.29}}
//	totalPrice := Reduce(products, func(total float64, p Product) float64 { return total + p.Price }, 0.0)
//	// totalPrice will be 2.28 as each product's price is added to the accumulated total
func Reduce[T any, U any](slice []T, accumulator func(U, T) U, initialValue U) U {
	result := initialValue
	for _, item := range slice {
		result = accumulator(result, item)
	}
	return result
}

// IndexOf searches for a specific element in a slice and returns its index if found.
//
// This function iterates over each element in the input slice `slice` to find the first
// occurrence of the specified `item`. If `item` is found, the function returns the index
// of `item` within `slice`. If `item` is not present in the slice, it returns -1.
//
// The function is generic, allowing it to operate on slices of any comparable type `T`
// (e.g., int, string, or other types that support equality comparison).
//
// Parameters:
//   - `slice`: The slice in which to search for `item`. It can contain elements of any
//     comparable type `T`.
//   - `item`: The item to search for within `slice`. It should be of the same type `T`
//     as the elements in `slice`.
//
// Returns:
//   - The zero-based index of `item` in `slice` if it exists; otherwise, -1.
//
// Example:
//
//	// Searching for an integer in a slice
//	numbers := []int{1, 2, 3, 4}
//	index := IndexOf(numbers, 3)
//	// index will be 2, as 3 is located at index 2 in the slice
//
//	// Searching for a string in a slice
//	words := []string{"apple", "banana", "cherry"}
//	index = IndexOf(words, "banana")
//	// index will be 1, as "banana" is at index 1 in the slice
//
//	// Item not found in the slice
//	index = IndexOf(words, "date")
//	// index will be -1, as "date" is not in the slice
func IndexOf[T comparable](slice []T, item T) int {
	for i, value := range slice {
		if value == item {
			return i
		}
	}
	return -1
}

// Unique returns a new slice containing only the unique elements from the input slice,
// preserving their original order.
//
// This function iterates over each element in the input slice `slice` and uses a map
// `uniqueMap` to track elements that have already been encountered. If an element has
// not been seen before, it is added to both the `uniqueValues` result slice and the
// map. This ensures that only the first occurrence of each unique element is kept in
// the final slice, while duplicates are ignored.
//
// The function is generic, allowing it to operate on slices of any comparable type `T`.
// The elements must be of a comparable type to allow them to be used as keys in the map.
//
// Parameters:
//   - `slice`: The input slice from which unique elements are extracted. It can contain
//     elements of any comparable type `T`.
//
// Returns:
//   - A new slice of type `[]T` containing only the unique elements from `slice` in the
//     order of their first appearance.
//
// Example:
//
//	// Extracting unique integers from a slice
//	numbers := []int{1, 2, 2, 3, 4, 4, 5}
//	uniqueNumbers := Unique(numbers)
//	// uniqueNumbers will be []int{1, 2, 3, 4, 5}
//
//	// Extracting unique strings from a slice
//	words := []string{"apple", "banana", "apple", "cherry"}
//	uniqueWords := Unique(words)
//	// uniqueWords will be []string{"apple", "banana", "cherry"}
//
//	// An empty slice will return an empty result
//	empty := []int{}
//	uniqueEmpty := Unique(empty)
//	// uniqueEmpty will be []int{}
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{}, len(slice))
	uniqueValues := make([]T, 0, len(slice))
	for _, value := range slice {
		if _, found := seen[value]; !found {
			uniqueValues = append(uniqueValues, value)
			seen[value] = struct{}{}
		}
	}
	return uniqueValues
}

// Flatten takes a slice of potentially nested elements and returns a new slice
// containing all elements of type `T` in a flat structure.
//
// This function recursively processes each element in the input slice `s`, checking if
// it is a nested slice (`[]interface{}`). If a nested slice is found, `Flatten` is called
// recursively to flatten it and append its elements to the `result` slice. If an element
// is of type `T`, it is directly appended to `result`. Elements that are neither `[]interface{}`
// nor of type `T` are ignored.
//
// The function is generic, allowing it to work with any element type `T`, which must be
// specified when calling the function. This makes `Flatten` useful for flattening slices
// with nested structures while filtering only the elements of a specified type.
//
// Parameters:
//   - `s`: A slice of `interface{}`, which can contain nested slices (`[]interface{}`) or
//     elements of any type. Nested slices may contain more nested slices at arbitrary depths.
//
// Returns:
//   - A new slice of type `[]T` containing all elements of type `T` from `s`, flattened
//     into a single level.
//
// Example:
//
//	// Flattening a nested slice of integers
//	nestedInts := []interface{}{1, []interface{}{2, 3}, []interface{}{[]interface{}{4, 5}}}
//	flatInts := Flatten[int](nestedInts)
//	// flatInts will be []int{1, 2, 3, 4, 5}
//
//	// Flattening a nested slice with mixed types, extracting only strings
//	mixedNested := []interface{}{"apple", []interface{}{"banana", 1, []interface{}{"cherry"}}}
//	flatStrings := Flatten[string](mixedNested)
//	// flatStrings will be []string{"apple", "banana", "cherry"}
//
//	// Flattening an empty slice
//	empty := []interface{}{}
//	flatEmpty := Flatten[int](empty)
//	// flatEmpty will be []int{}
func Flatten[T any](s []any) []T {
	result := make([]T, 0)
	for _, v := range s {
		switch val := v.(type) {
		case []any:
			result = append(result, Flatten[T](val)...)
		default:
			if typed, ok := val.(T); ok {
				result = append(result, typed)
			}
		}
	}
	return result
}

// FlattenDeep takes a nested structure of arbitrary depth and returns a flat slice
// containing all elements in a single level.
//
// This function recursively processes each element in `arr`. If an element is itself a
// slice (`[]interface{}`), `FlattenDeep` calls itself to flatten that nested slice and
// appends its elements to the `result` slice. If the element is not a slice, it is directly
// added to `result`. The function allows flattening of complex nested structures while
// maintaining all elements in a single-level output.
//
// This function operates with values of type `interface{}`, making it flexible enough
// to handle mixed types in the input. It returns a slice of `interface{}`, which may
// contain elements of varying types from the original nested structure.
//
// Parameters:
//   - `arr`: The input slice, which can contain nested slices of arbitrary depth and elements
//     of any type.
//
// Returns:
//   - A slice of `[]interface{}` containing all elements from `arr` flattened into a single level.
//
// Example:
//
//	// Flattening a nested structure of mixed values
//	nested := []interface{}{1, []interface{}{2, 3, []interface{}{4, []interface{}{5}}}}
//	flat := FlattenDeep(nested)
//	// flat will be []interface{}{1, 2, 3, 4, 5}
//
//	// Flattening a deeply nested structure with varied types
//	mixedNested := []interface{}{"apple", []interface{}{"banana", 1, []interface{}{"cherry"}}}
//	flatMixed := FlattenDeep(mixedNested)
//	// flatMixed will be []interface{}{"apple", "banana", 1, "cherry"}
//
//	// Flattening a non-nested input returns the input as-is
//	nonNested := 5
//	flatNonNested := FlattenDeep(nonNested)
//	// flatNonNested will be []interface{}{5}
func FlattenDeep(arr any) []any {
	result := make([]any, 0)
	switch v := arr.(type) {
	case []any:
		for _, val := range v {
			result = append(result, FlattenDeep(val)...)
		}
	case any:
		result = append(result, v)
	}
	return result
}

// GroupBy groups elements of a slice into a map based on a specified key.
//
// This function iterates over each element in the input slice `slice`, applies the
// provided `getKey` function to extract a key for each element, and groups elements
// that share the same key into a slice. The function then returns a map where each
// key maps to a slice of elements that correspond to that key.
//
// The function is generic, allowing it to work with slices of any type `T` and to
// generate keys of any comparable type `K`. This makes `GroupBy` useful for organizing
// data based on shared attributes, such as grouping items by category or organizing
// records by a specific field.
//
// Parameters:
//   - `slice`: The input slice containing elements to be grouped. It can be of any type `T`.
//   - `getKey`: A function that takes an element of type `T` and returns a key of type `K`,
//     which is used to group the element in the resulting map.
//
// Returns:
//   - A map of type `map[K][]T`, where each key is associated with a slice of elements
//     that share that key.
//
// Example:
//
//	// Grouping integers by even and odd
//	numbers := []int{1, 2, 3, 4, 5}
//	grouped := GroupBy(numbers, func(n int) string {
//		if n%2 == 0 {
//			return "even"
//		}
//		return "odd"
//	})
//	// grouped will be map[string][]int{"even": {2, 4}, "odd": {1, 3, 5}}
//
//	// Grouping people by age
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//	people := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}, {Name: "Charlie", Age: 30}}
//	groupedByAge := GroupBy(people, func(p Person) int { return p.Age })
//	// groupedByAge will be map[int][]Person{30: {{Name: "Alice", Age: 30}, {Name: "Charlie", Age: 30}}, 25: {{Name: "Bob", Age: 25}}}
//
//	// Grouping strings by their length
//	words := []string{"apple", "pear", "banana", "peach"}
//	groupedByLength := GroupBy(words, func(word string) int { return len(word) })
//	// groupedByLength will be map[int][]string{5: {"apple", "peach"}, 4: {"pear"}, 6: {"banana"}}
func GroupBy[T any, K comparable](slice []T, getKey func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range slice {
		key := getKey(item)
		result[key] = append(result[key], item)
	}
	return result
}

// Join concatenates the string representation of each element in a slice into a single
// string, with a specified separator between each element.
//
// This function iterates over each element in the input slice `slice`, converts each
// element to a string using `fmt.Sprintf` with the `%v` format, and appends it to the
// `result` string. A separator string `separator` is inserted between elements in the
// final concatenated result. If the slice has only one element, no separator is added.
// The function is generic and can work with slices containing elements of any type `T`.
//
// Parameters:
//   - `slice`: The input slice containing elements to be joined. It can contain elements
//     of any type `T`.
//   - `separator`: A string that will be inserted between each element in the final result.
//
// Returns:
//   - A single string that is the result of concatenating all elements in `slice` with the
//     specified `separator` in between.
//
// Example:
//
//	// Joining integers with a comma separator
//	numbers := []int{1, 2, 3}
//	joinedNumbers := Join(numbers, ", ")
//	// joinedNumbers will be "1, 2, 3"
//
//	// Joining strings with a space separator
//	words := []string{"Go", "is", "awesome"}
//	joinedWords := Join(words, " ")
//	// joinedWords will be "Go is awesome"
//
//	// Joining an empty slice returns an empty string
//	emptySlice := []int{}
//	joinedEmpty := Join(emptySlice, ",")
//	// joinedEmpty will be ""
func Join[T any](slice []T, separator string) string {
	var b strings.Builder
	for i, item := range slice {
		if i > 0 {
			b.WriteString(separator)
		}
		fmt.Fprintf(&b, "%v", encoding.JSON(item))
	}
	return b.String()
}

// ReverseN reverses the order of elements in the input slice and returns a new slice
// containing the elements in reverse order.
//
// This function creates a new slice `reversed` with the same length as the input slice `slice`.
// It then uses a two-pointer approach to swap the elements of `slice` from both ends toward the center,
// effectively reversing the slice. The result is a new slice with elements in the opposite order.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slice`: The input slice whose elements are to be reversed. It can contain elements of any type `T`.
//
// Returns:
//   - A new slice of type `[]T` containing the elements of `slice` in reverse order.
//
// Example:
//
//	// Reversing a slice of integers
//	numbers := []int{1, 2, 3, 4}
//	reversedNumbers := ReverseN(numbers)
//	// reversedNumbers will be []int{4, 3, 2, 1}
//
//	// Reversing a slice of strings
//	words := []string{"apple", "banana", "cherry"}
//	reversedWords := ReverseN(words)
//	// reversedWords will be []string{"cherry", "banana", "apple"}
//
//	// Reversing an empty slice returns an empty slice
//	empty := []int{}
//	reversedEmpty := ReverseN(empty)
//	// reversedEmpty will be []int{}
func ReverseN[T any](slice []T) []T {
	reversed := make([]T, len(slice))
	for i, j := 0, len(slice)-1; i <= j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = slice[j], slice[i]
	}
	return reversed
}

// FindIndex searches for the first occurrence of a target element in a slice
// and returns its index. If the element is not found, it returns -1.
//
// This function iterates over each element in the input slice `slice` and compares
// each element to the specified `target`. When the first occurrence of `target` is found,
// the function returns the index of that element. If the element is not found, the function
// returns -1 to indicate that the target is not present in the slice.
//
// The function is generic, allowing it to work with slices of any comparable type `T`,
// such as integers, strings, or other types that support equality comparison.
//
// Parameters:
//   - `slice`: The input slice in which to search for the target element. It can contain
//     elements of any comparable type `T`.
//   - `target`: The element to search for within the slice. It should be of the same type `T`
//     as the elements in `slice`.
//
// Returns:
//   - The zero-based index of the first occurrence of `target` in the slice if it exists;
//     otherwise, -1 if the target is not found.
//
// Example:
//
//	// Searching for an integer in a slice
//	numbers := []int{1, 2, 3, 4}
//	index := FindIndex(numbers, 3)
//	// index will be 2, as 3 is located at index 2 in the slice
//
//	// Searching for a string in a slice
//	words := []string{"apple", "banana", "cherry"}
//	index = FindIndex(words, "banana")
//	// index will be 1, as "banana" is at index 1 in the slice
//
//	// Item not found in the slice
//	index = FindIndex(words, "date")
//	// index will be -1, as "date" is not in the slice
func FindIndex[T comparable](slice []T, target T) int {
	for i, item := range slice {
		if item == target {
			return i
		}
	}
	return -1
}

// Chunk splits a slice into smaller slices (chunks) of the specified size.
//
// This function takes an input slice `slice` and a `chunkSize` and splits the input slice into
// smaller slices, each containing up to `chunkSize` elements. The function returns a slice of slices
// containing the chunked elements. If the `chunkSize` is greater than the length of the input slice,
// the entire slice will be returned as a single chunk. If the `chunkSize` is less than or equal to 0,
// the function returns `nil`.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slice`: The input slice to be split into chunks. It can contain elements of any type `T`.
//   - `chunkSize`: The size of each chunk. If this value is less than or equal to 0, the function returns `nil`.
//
// Returns:
//   - A slice of slices (`[][]T`), where each inner slice contains up to `chunkSize` elements from
//     the original slice. If the slice cannot be split into even chunks, the last chunk may contain
//     fewer elements than `chunkSize`.
//
// Example:
//
//	// Chunking a slice of integers into chunks of size 2
//	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
//	chunks := Chunk(numbers, 2)
//	// chunks will be [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9}}
//
//	// Chunking a slice of strings into chunks of size 3
//	words := []string{"apple", "banana", "cherry", "date", "elderberry", "fig"}
//	chunksWords := Chunk(words, 3)
//	// chunksWords will be [][]string{{"apple", "banana", "cherry"}, {"date", "elderberry", "fig"}}
//
//	// Chunking an empty slice returns an empty slice of slices
//	empty := []int{}
//	chunksEmpty := Chunk(empty, 3)
//	// chunksEmpty will be [][]int{}
//
//	// If chunkSize is 0 or negative, return nil
//	chunksInvalid := Chunk(numbers, -1)
//	// chunksInvalid will be nil
func Chunk[T any](slice []T, chunkSize int) [][]T {
	if chunkSize <= 0 {
		return nil
	}
	var chunks [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// Shuffle randomly shuffles the elements of a slice and returns a new slice with the shuffled elements.
//
// This function takes an input slice `slice` and shuffles its elements using a random permutation.
// It creates a new slice `shuffledSlice` and populates it by selecting elements from the original slice
// according to a random permutation of indices. The resulting `shuffledSlice` contains the same elements
// as the input slice, but in a random order. The function uses a seeded random generator to ensure different
// results each time it is called.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slice`: The input slice to be shuffled. It can contain elements of any type `T`.
//
// Returns:
//   - A new slice of type `[]T`, containing the shuffled elements of the input slice.
//
// Example:
//
//	// Shuffling a slice of integers
//	numbers := []int{1, 2, 3, 4, 5}
//	shuffledNumbers := Shuffle(numbers)
//	// shuffledNumbers will be a random permutation of [1, 2, 3, 4, 5]
//
//	// Shuffling a slice of strings
//	words := []string{"apple", "banana", "cherry"}
//	shuffledWords := Shuffle(words)
//	// shuffledWords will be a random permutation of ["apple", "banana", "cherry"]
//
//	// Shuffling an empty slice returns an empty slice
//	empty := []int{}
//	shuffledEmpty := Shuffle(empty)
//	// shuffledEmpty will be []int{}
func Shuffle[T any](slice []T) []T {
	shuffledSlice := make([]T, len(slice))
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	perm := r.Perm(len(slice))
	for i, randIndex := range perm {
		shuffledSlice[i] = slice[randIndex]
	}
	return shuffledSlice
}

// Cartesian computes the Cartesian product of multiple slices and returns the result as a slice of slices.
//
// This function takes multiple slices of type `[]T` and computes their Cartesian product. The Cartesian
// product of two or more sets (or slices in this case) is the set of all possible combinations where each
// combination consists of one element from each slice. The function recursively computes the product of the
// slices, starting from the second slice and combining it with each element of the first slice. The result is
// a slice of slices, where each inner slice is a combination of elements from the input slices.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slices`: A variadic parameter that represents multiple slices to compute the Cartesian product of.
//     Each slice can contain elements of any type `T`.
//
// Returns:
//   - A slice of slices (`[][]T`), where each inner slice represents a unique combination of elements
//     from the input slices.
//
// Example:
//
//	// Cartesian product of two slices of integers
//	slice1 := []int{1, 2}
//	slice2 := []int{3, 4}
//	product := Cartesian(slice1, slice2)
//	// product will be [][]int{{1, 3}, {1, 4}, {2, 3}, {2, 4}}
//
//	// Cartesian product of three slices of strings
//	slice3 := []string{"a", "b"}
//	slice4 := []string{"x", "y"}
//	slice5 := []string{"1", "2"}
//	productStrings := Cartesian(slice3, slice4, slice5)
//	// productStrings will be [][]string{
//	//   {"a", "x", "1"}, {"a", "x", "2"},
//	//   {"a", "y", "1"}, {"a", "y", "2"},
//	//   {"b", "x", "1"}, {"b", "x", "2"},
//	//   {"b", "y", "1"}, {"b", "y", "2"}
//	// }
//
//	// Cartesian product of an empty slice returns an empty slice
//	empty := []int{}
//	productEmpty := Cartesian(empty)
//	// productEmpty will be [][]int{{}}
func Cartesian[T any](slices ...[]T) [][]T {
	n := len(slices)
	if n == 0 {
		return [][]T{{}}
	}
	if n == 1 {
		product := make([][]T, len(slices[0]))
		for i, item := range slices[0] {
			product[i] = []T{item}
		}
		return product
	}
	tailProduct := Cartesian(slices[1:]...)
	product := make([][]T, 0, len(slices[0])*len(tailProduct))
	for _, head := range slices[0] {
		for _, tail := range tailProduct {
			product = append(product, append([]T{head}, tail...))
		}
	}
	return product
}

// Sort sorts a slice based on a custom comparison function and returns a new sorted slice.
//
// This function takes an input slice `slice` and a comparison function `comparer` that defines the
// sorting order. The comparison function takes two elements of type `T` and returns `true` if the
// first element should come before the second one (i.e., if it should be sorted earlier). The function
// creates a new slice `sortedSlice` by copying the elements of the original slice, then sorts it in place
// using the provided `comparer`. The resulting `sortedSlice` will be a new slice containing the elements
// from the original slice, arranged in the order specified by the comparison function.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slice`: The input slice to be sorted. It can contain elements of any type `T`.
//   - `comparer`: A comparison function that takes two elements of type `T` and returns a boolean value.
//     It determines the order of the elements: it should return `true` if the first element should come
//     before the second element in the sorted order.
//
// Returns:
//   - A new slice of type `[]T`, containing the elements from the original slice sorted according to
//     the provided comparison function.
//
// Example:
//
//	// Sorting a slice of integers in ascending order
//	numbers := []int{5, 3, 8, 1, 2}
//	sortedNumbers := Sort(numbers, func(a, b int) bool {
//		return a < b
//	})
//	// sortedNumbers will be []int{1, 2, 3, 5, 8}
//
//	// Sorting a slice of strings in descending order
//	words := []string{"apple", "banana", "cherry"}
//	sortedWords := Sort(words, func(a, b string) bool {
//		return a > b
//	})
//	// sortedWords will be []string{"cherry", "banana", "apple"}
//
//	// Sorting an empty slice returns an empty slice
//	empty := []int{}
//	sortedEmpty := Sort(empty, func(a, b int) bool {
//		return a < b
//	})
//	// sortedEmpty will be []int{}
func Sort[T any](slice []T, comparer func(T, T) bool) []T {
	sortedSlice := make([]T, len(slice))
	copy(sortedSlice, slice)
	sort.Slice(sortedSlice, func(i, j int) bool {
		return comparer(sortedSlice[i], sortedSlice[j])
	})
	return sortedSlice
}

// AllMatch checks if all elements in a slice satisfy a given condition and returns a boolean result.
//
// This function takes an input slice `slice` and a predicate function `predicate`. It iterates over
// each element in the slice, applying the predicate function to each one. If any element does not
// satisfy the predicate (i.e., the predicate returns `false`), the function immediately returns `false`.
// If all elements satisfy the predicate, the function returns `true`.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slice`: The input slice whose elements will be checked. It can contain elements of any type `T`.
//   - `predicate`: A function that takes an element of type `T` and returns a boolean. This function
//     represents the condition that each element must satisfy. If the predicate returns `true` for an
//     element, the element meets the condition.
//
// Returns:
//   - `true` if all elements in the slice satisfy the predicate; `false` if any element does not.
//
// Example:
//
//	// Checking if all integers in a slice are positive
//	numbers := []int{2, 4, 6, 8}
//	allPositive := AllMatch(numbers, func(n int) bool {
//		return n > 0
//	})
//	// allPositive will be true
//
//	// Checking if all strings in a slice have a length greater than 3
//	words := []string{"apple", "banana", "pear"}
//	allLong := AllMatch(words, func(s string) bool {
//		return len(s) > 3
//	})
//	// allLong will be true
//
//	// If the slice is empty, returns true since no elements violate the predicate
//	empty := []int{}
//	allMatchEmpty := AllMatch(empty, func(n int) bool {
//		return n > 0
//	})
//	// allMatchEmpty will be true
func AllMatch[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// AnyMatch checks if any element in a slice satisfies a given condition and returns a boolean result.
//
// This function takes an input slice `slice` and a predicate function `predicate`. It iterates over
// each element in the slice, applying the predicate function to each one. If the predicate returns
// `true` for any element, the function immediately returns `true`. If no elements satisfy the predicate,
// the function returns `false`.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slice`: The input slice whose elements will be checked. It can contain elements of any type `T`.
//   - `predicate`: A function that takes an element of type `T` and returns a boolean. This function
//     represents the condition that an element must satisfy. If the predicate returns `true` for an
//     element, the element meets the condition.
//
// Returns:
//   - `true` if at least one element in the slice satisfies the predicate; `false` if no elements do.
//
// Example:
//
//	// Checking if any integers in a slice are even
//	numbers := []int{1, 3, 5, 6}
//	anyEven := AnyMatch(numbers, func(n int) bool {
//		return n%2 == 0
//	})
//	// anyEven will be true because 6 is even
//
//	// Checking if any strings in a slice contain the letter "a"
//	words := []string{"apple", "banana", "cherry"}
//	containsA := AnyMatch(words, func(s string) bool {
//		return strings.Contains(s, "a")
//	})
//	// containsA will be true because "apple" and "banana" contain "a"
//
//	// Checking an empty slice returns false since no elements satisfy the predicate
//	empty := []int{}
//	anyMatchEmpty := AnyMatch(empty, func(n int) bool {
//		return n > 0
//	})
//	// anyMatchEmpty will be false
func AnyMatch[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if predicate(item) {
			return true
		}
	}
	return false
}

// Push appends an element to the end of a slice and returns the resulting slice.
//
// This function takes an input slice `slice` and an element `element`, and appends
// the element to the end of the slice using the built-in `append` function. The
// function returns a new slice containing the original elements followed by the new
// element. This function is useful for adding elements dynamically to a slice.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slice`: The input slice to which the element will be appended. It can contain elements of any type `T`.
//   - `element`: The element to be appended to the end of the slice. It is of type `T`.
//
// Returns:
//   - A new slice of type `[]T`, containing the original elements in `slice` with `element` appended at the end.
//
// Example:
//
//	// Appending an integer to a slice of integers
//	numbers := []int{1, 2, 3}
//	updatedNumbers := Push(numbers, 4)
//	// updatedNumbers will be []int{1, 2, 3, 4}
//
//	// Appending a string to a slice of strings
//	words := []string{"apple", "banana"}
//	updatedWords := Push(words, "cherry")
//	// updatedWords will be []string{"apple", "banana", "cherry"}
//
//	// Appending to an empty slice
//	var emptySlice []int
//	newSlice := Push(emptySlice, 1)
//	// newSlice will be []int{1}
func Push[T any](slice []T, element T) []T {
	return append(slice, element)
}

// Pop removes the last element from a slice and returns the resulting slice.
//
// This function takes an input slice `slice` and removes its last element by creating
// a new slice that excludes the last element. The function uses slicing to return a
// portion of the original slice that ends before the last element. If the input slice
// is empty, calling this function will result in a runtime panic.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slice`: The input slice from which the last element will be removed. It can contain elements of any type `T`.
//
// Returns:
//   - A new slice of type `[]T`, containing all elements from the original slice except the last one.
//
// Example:
//
//	// Removing the last element from a slice of integers
//	numbers := []int{1, 2, 3, 4}
//	updatedNumbers := Pop(numbers)
//	// updatedNumbers will be []int{1, 2, 3}
//
//	// Removing the last element from a slice of strings
//	words := []string{"apple", "banana", "cherry"}
//	updatedWords := Pop(words)
//	// updatedWords will be []string{"apple", "banana"}
//
//	// Attempting to pop from an empty slice will cause a runtime panic
//	var emptySlice []int
//	// updatedEmpty := Pop(emptySlice) // This will panic
func Pop[T any](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}
	return slice[:len(slice)-1]
}

// Unshift inserts an element at the beginning of a slice and returns the resulting slice.
//
// This function takes an input slice `slice` and an element `element`, then creates
// a new slice by appending the `element` at the start, followed by the elements of
// the original slice. The function uses the built-in `append` function to combine a
// new slice containing just the `element` with the original slice, effectively adding
// the element to the beginning.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slice`: The input slice to which the element will be prepended. It can contain elements of any type `T`.
//   - `element`: The element to be inserted at the beginning of the slice. It is of type `T`.
//
// Returns:
//   - A new slice of type `[]T`, containing `element` followed by all the elements of the original `slice`.
//
// Example:
//
//	// Adding an integer to the beginning of a slice of integers
//	numbers := []int{2, 3, 4}
//	updatedNumbers := Unshift(numbers, 1)
//	// updatedNumbers will be []int{1, 2, 3, 4}
//
//	// Adding a string to the beginning of a slice of strings
//	words := []string{"banana", "cherry"}
//	updatedWords := Unshift(words, "apple")
//	// updatedWords will be []string{"apple", "banana", "cherry"}
//
//	// Adding to an empty slice
//	var emptySlice []int
//	newSlice := Unshift(emptySlice, 1)
//	// newSlice will be []int{1}
func Unshift[T any](slice []T, element T) []T {
	return append([]T{element}, slice...)
}

// Shift removes the first element from a slice and returns the resulting slice.
//
// This function takes an input slice `slice` and removes its first element by creating
// a new slice that starts from the second element of the original slice. The function
// achieves this using slicing, effectively returning a view of the original slice that
// excludes the first element. If the input slice is empty, calling this function will
// result in a runtime panic due to out-of-bounds access.
//
// The function is generic, allowing it to work with slices of any type `T`.
//
// Parameters:
//   - `slice`: The input slice from which the first element will be removed. It can contain elements of any type `T`.
//
// Returns:
//   - A new slice of type `[]T`, containing all elements from the original slice except the first one.
//
// Example:
//
//	// Removing the first element from a slice of integers
//	numbers := []int{1, 2, 3, 4}
//	updatedNumbers := Shift(numbers)
//	// updatedNumbers will be []int{2, 3, 4}
//
//	// Removing the first element from a slice of strings
//	words := []string{"apple", "banana", "cherry"}
//	updatedWords := Shift(words)
//	// updatedWords will be []string{"banana", "cherry"}
func Shift[T any](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}
	return slice[1:]
}

// AppendIf appends an element to a slice if it is not already present.
//
// This function takes an input slice `slice` and an element `element`. It first checks
// if the element is already in the slice by calling the helper function `ContainsN`.
// If the element is not found in the slice, the function appends it to the end of the slice.
// If the element is already present, the original slice is returned unchanged.
//
// The function is generic and requires that the type `T` be `comparable`, allowing the
// function to use the `==` operator in `ContainsN` to check for equality.
//
// Parameters:
//   - `slice`: The input slice to which the element might be appended. It can contain elements of any comparable type `T`.
//   - `element`: The element to be appended if it is not already in `slice`. It is of type `T`.
//
// Returns:
//   - A new slice of type `[]T` containing the original elements and, if missing, the appended `element`.
//
// Example:
//
//	// Adding a missing integer to a slice
//	numbers := []int{1, 2, 3}
//	updatedNumbers := AppendIf(numbers, 4)
//	// updatedNumbers will be []int{1, 2, 3, 4}
//
//	// Trying to add an existing integer to a slice
//	updatedNumbers = AppendIf(numbers, 3)
//	// updatedNumbers will be []int{1, 2, 3} (unchanged)
//
//	// Adding a missing string to a slice
//	words := []string{"apple", "banana"}
//	updatedWords := AppendIf(words, "cherry")
//	// updatedWords will be []string{"apple", "banana", "cherry"}
func AppendIf[T comparable](slice []T, element T) []T {
	if !Contains(slice, element) {
		return append(slice, element)
	}
	return slice
}

// Intersect returns a new slice containing elements that are present in both input slices.
//
// This function takes two input slices, `slice1` and `slice2`, and identifies elements
// that are present in both slices. It uses a map to track the elements of `slice1`,
// then iterates over `slice2` to find common elements. If an element from `slice2` is
// found in the map (indicating it exists in `slice1`), it is added to the result slice.
//
// The function is generic, allowing it to work with slices of any `comparable` type `T`.
//
// Parameters:
//   - `slice1`: The first input slice containing elements of any comparable type `T`.
//   - `slice2`: The second input slice containing elements of any comparable type `T`.
//
// Returns:
//   - A new slice of type `[]T` that contains the elements found in both `slice1` and `slice2`.
//     Each element in the result slice will appear only once, even if it is duplicated in the input slices.
//
// Example:
//
//	// Finding common integers between two slices
//	numbers1 := []int{1, 2, 3, 4}
//	numbers2 := []int{3, 4, 5, 6}
//	commonNumbers := Intersect(numbers1, numbers2)
//	// commonNumbers will be []int{3, 4}
//
//	// Finding common strings between two slices
//	words1 := []string{"apple", "banana", "cherry"}
//	words2 := []string{"banana", "cherry", "date"}
//	commonWords := Intersect(words1, words2)
//	// commonWords will be []string{"banana", "cherry"}
//
//	// Intersecting with an empty slice results in an empty slice
//	empty := []int{}
//	intersectEmpty := Intersect(numbers1, empty)
//	// intersectEmpty will be []int{}
func Intersect[T comparable](slice1, slice2 []T) []T {
	set := make(map[T]struct{}, len(slice1))
	for _, item := range slice1 {
		set[item] = struct{}{}
	}
	seen := make(map[T]struct{})
	result := make([]T, 0)
	for _, item := range slice2 {
		if _, ok := set[item]; ok {
			if _, dup := seen[item]; !dup {
				result = append(result, item)
				seen[item] = struct{}{}
			}
		}
	}
	return result
}

// Difference returns a new slice containing elements that are unique to each of the two input slices.
//
// This function takes two input slices, `slice1` and `slice2`, and identifies elements
// that are present in either slice but not both. It creates a map to track the elements
// of `slice1`, then checks for unique elements in `slice2` by confirming that they are
// not present in `slice1`. Finally, it appends any unique elements from `slice1` to ensure
// that the result includes all elements unique to either slice.
//
// The function is generic, allowing it to work with slices of any `comparable` type `T`.
//
// Parameters:
//   - `slice1`: The first input slice containing elements of any comparable type `T`.
//   - `slice2`: The second input slice containing elements of any comparable type `T`.
//
// Returns:
//   - A new slice of type `[]T` that contains elements unique to either `slice1` or `slice2`.
//     If an element appears in both slices, it will not appear in the result.
//
// Example:
//
//	// Finding unique integers between two slices
//	numbers1 := []int{1, 2, 3, 4}
//	numbers2 := []int{3, 4, 5, 6}
//	uniqueNumbers := Difference(numbers1, numbers2)
//	// uniqueNumbers will be []int{1, 2, 5, 6}
//
//	// Finding unique strings between two slices
//	words1 := []string{"apple", "banana", "cherry"}
//	words2 := []string{"banana", "date"}
//	uniqueWords := Difference(words1, words2)
//	// uniqueWords will be []string{"apple", "cherry", "date"}
//
//	// Difference with an empty slice results in the original slice
//	empty := []int{}
//	uniqueFromEmpty := Difference(numbers1, empty)
//	// uniqueFromEmpty will be []int{1, 2, 3, 4}
func Difference[T comparable](slice1, slice2 []T) []T {
	set1 := make(map[T]struct{}, len(slice1))
	for _, item := range slice1 {
		set1[item] = struct{}{}
	}
	set2 := make(map[T]struct{}, len(slice2))
	for _, item := range slice2 {
		set2[item] = struct{}{}
	}
	result := make([]T, 0)
	for _, item := range slice2 {
		if _, ok := set1[item]; !ok {
			result = append(result, item)
		}
	}
	for _, item := range slice1 {
		if _, ok := set2[item]; !ok {
			result = append(result, item)
		}
	}
	return result
}

// IndexExists checks whether a given index is valid for the provided slice.
//
// This function takes a slice `slice` of any type `T` and an integer `index`, and returns a boolean indicating
// whether the specified index is within the valid range for the slice. A valid index is one that is greater than
// or equal to 0 and less than the length of the slice.
//
// Parameters:
//   - `slice`: The input slice of any type `T` that is being checked.
//   - `index`: The index to check for validity in the slice.
//
// Returns:
//   - `true` if the index is within the bounds of the slice (i.e., 0 <= index < len(slice)).
//   - `false` otherwise, such as when the index is negative or greater than or equal to the length of the slice.
//
// Example:
//
//	// Checking if an index exists in a slice
//	numbers := []int{1, 2, 3, 4}
//	exists := IndexExists(numbers, 2)
//	// exists will be true, as index 2 is valid for the slice
//
//	// Checking an invalid index
//	exists = IndexExists(numbers, 5)
//	// exists will be false, as index 5 is out of bounds for the slice
func IndexExists[T any](slice []T, index int) bool {
	return index >= 0 && index < len(slice)
}
