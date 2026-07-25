package coll_test

import (
	"sort"
	"testing"

	"github.com/polarixa/replify/pkg/coll"
)

// ─── HashMap ────────────────────────────────────────────────────────────────

// TestHashMap_KeySet_AllKeysReturned verifies the previously broken KeySet
// method now returns every key, not just the last one written (i was never
// incremented).
func TestHashMap_KeySet_AllKeysReturned(t *testing.T) {
	tests := []struct {
		name string
		puts map[string]int
	}{
		{"empty", map[string]int{}},
		{"single", map[string]int{"a": 1}},
		{"multiple", map[string]int{"a": 1, "b": 2, "c": 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := coll.NewHashMap[string, int]()
			for k, v := range tc.puts {
				m.Put(k, v)
			}
			keys := m.KeySet()
			if len(keys) != len(tc.puts) {
				t.Fatalf("KeySet() len = %d; want %d", len(keys), len(tc.puts))
			}
			got := make(map[string]bool, len(keys))
			for _, k := range keys {
				got[k] = true
			}
			for k := range tc.puts {
				if !got[k] {
					t.Errorf("KeySet() missing key %q", k)
				}
			}
		})
	}
}

// TestHashMap_BasicOps exercises Put/Get/Remove/ContainsKey/Size/IsEmpty/Clear.
func TestHashMap_BasicOps(t *testing.T) {
	m := coll.NewHashMap[string, int]()
	if !m.IsEmpty() {
		t.Fatal("new map should be empty")
	}
	m.Put("x", 10)
	m.Put("y", 20)
	if m.Size() != 2 {
		t.Fatalf("Size() = %d; want 2", m.Size())
	}
	if v := m.Get("x"); v != 10 {
		t.Errorf("Get(x) = %d; want 10", v)
	}
	if !m.ContainsKey("y") {
		t.Error("ContainsKey(y) should be true")
	}
	m.Remove("x")
	if m.ContainsKey("x") {
		t.Error("ContainsKey(x) should be false after Remove")
	}
	m.Clear()
	if !m.IsEmpty() {
		t.Error("IsEmpty() should be true after Clear")
	}
}

// ─── HashSet ─────────────────────────────────────────────────────────────────

func TestHashSet_BasicOps(t *testing.T) {
	s := coll.NewHashSet(1, 2, 3)
	if s.Size() != 3 {
		t.Fatalf("Size() = %d; want 3", s.Size())
	}
	s.Add(3) // duplicate — size must not change
	if s.Size() != 3 {
		t.Errorf("Add(duplicate) Size() = %d; want 3", s.Size())
	}
	if !s.Contains(2) {
		t.Error("Contains(2) should be true")
	}
	s.Remove(2)
	if s.Contains(2) {
		t.Error("Contains(2) should be false after Remove")
	}
	s.Clear()
	if !s.IsEmpty() {
		t.Error("IsEmpty() should be true after Clear")
	}
}

func TestHashSet_SetAlgebra(t *testing.T) {
	a := coll.NewHashSet(1, 2, 3)
	b := coll.NewHashSet(2, 3, 4)

	inter := a.Intersection(b)
	if inter.Size() != 2 || !inter.Contains(2) || !inter.Contains(3) {
		t.Errorf("Intersection = %v; want {2,3}", inter.Slice())
	}

	union := a.Union(b)
	if union.Size() != 4 {
		t.Errorf("Union size = %d; want 4", union.Size())
	}

	diff := a.Difference(b)
	if diff.Size() != 1 || !diff.Contains(1) {
		t.Errorf("Difference = %v; want {1}", diff.Slice())
	}
}

func TestHashSet_String_NotEmpty(t *testing.T) {
	s := coll.NewHashSet("a")
	str := s.String()
	if str == "" {
		t.Error("String() should not be empty for non-empty set")
	}
}

// ─── Stack ───────────────────────────────────────────────────────────────────

func TestStack_PushPopPeek(t *testing.T) {
	st := coll.NewStack[int]()
	if !st.IsEmpty() {
		t.Fatal("new stack should be empty")
	}
	st.Push(1)
	st.Push(2)
	st.Push(3)
	if st.Size() != 3 {
		t.Fatalf("Size() = %d; want 3", st.Size())
	}
	if top := st.Peek(); top != 3 {
		t.Errorf("Peek() = %d; want 3", top)
	}
	if popped := st.Pop(); popped != 3 {
		t.Errorf("Pop() = %d; want 3", popped)
	}
	if st.Size() != 2 {
		t.Errorf("Size() after Pop = %d; want 2", st.Size())
	}
	st.Clear()
	if !st.IsEmpty() {
		t.Error("IsEmpty() should be true after Clear")
	}
	// Pop/Peek from empty stack should return zero value, not panic
	if v := st.Pop(); v != 0 {
		t.Errorf("Pop() on empty = %d; want 0", v)
	}
	if v := st.Peek(); v != 0 {
		t.Errorf("Peek() on empty = %d; want 0", v)
	}
}

// ─── Slice utilities ─────────────────────────────────────────────────────────

func TestDifference_SymmetricDifference(t *testing.T) {
	tests := []struct {
		name    string
		s1      []int
		s2      []int
		wantIn  []int // all of these must be present
		wantNot []int // none of these should be present
	}{
		{
			name:    "disjoint",
			s1:      []int{1, 2},
			s2:      []int{3, 4},
			wantIn:  []int{1, 2, 3, 4},
			wantNot: nil,
		},
		{
			name:    "partial overlap",
			s1:      []int{1, 2, 3, 4},
			s2:      []int{3, 4, 5, 6},
			wantIn:  []int{1, 2, 5, 6},
			wantNot: []int{3, 4},
		},
		{
			name:    "identical slices",
			s1:      []int{1, 2, 3},
			s2:      []int{1, 2, 3},
			wantIn:  nil,
			wantNot: []int{1, 2, 3},
		},
		{
			name:    "empty s2",
			s1:      []int{1, 2, 3},
			s2:      []int{},
			wantIn:  []int{1, 2, 3},
			wantNot: nil,
		},
		{
			name:    "empty s1",
			s1:      []int{},
			s2:      []int{1, 2, 3},
			wantIn:  []int{1, 2, 3},
			wantNot: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := coll.Difference(tc.s1, tc.s2)
			gotSet := make(map[int]bool, len(got))
			for _, v := range got {
				gotSet[v] = true
			}
			for _, v := range tc.wantIn {
				if !gotSet[v] {
					t.Errorf("Difference() missing %d; got %v", v, got)
				}
			}
			for _, v := range tc.wantNot {
				if gotSet[v] {
					t.Errorf("Difference() should not contain %d; got %v", v, got)
				}
			}
		})
	}
}

func TestIntersect_NoDuplicates(t *testing.T) {
	tests := []struct {
		name    string
		s1      []int
		s2      []int
		wantLen int
		wantIn  []int
	}{
		{"basic", []int{1, 2, 3, 4}, []int{3, 4, 5, 6}, 2, []int{3, 4}},
		{"duplicates in s2", []int{1, 2, 3}, []int{2, 2, 3, 3}, 2, []int{2, 3}},
		{"no overlap", []int{1, 2}, []int{3, 4}, 0, nil},
		{"empty s1", []int{}, []int{1, 2}, 0, nil},
		{"empty s2", []int{1, 2}, []int{}, 0, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := coll.Intersect(tc.s1, tc.s2)
			if len(got) != tc.wantLen {
				t.Errorf("Intersect() len = %d; want %d (got %v)", len(got), tc.wantLen, got)
			}
			gotSet := make(map[int]bool, len(got))
			for _, v := range got {
				gotSet[v] = true
			}
			for _, v := range tc.wantIn {
				if !gotSet[v] {
					t.Errorf("Intersect() missing %d; got %v", v, got)
				}
			}
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"no duplicates", []int{1, 2, 3}, []int{1, 2, 3}},
		{"all duplicates", []int{1, 1, 1}, []int{1}},
		{"mixed", []int{1, 2, 2, 3, 1}, []int{1, 2, 3}},
		{"empty", []int{}, []int{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := coll.Unique(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("Unique() len = %d; want %d (got %v)", len(got), len(tc.want), got)
			}
			// Order is preserved (first occurrence).
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("Unique()[%d] = %d; want %d", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestShuffle_ContainsSameElements(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	got := coll.Shuffle(input)
	if len(got) != len(input) {
		t.Fatalf("Shuffle() len = %d; want %d", len(got), len(input))
	}
	sortedInput := make([]int, len(input))
	copy(sortedInput, input)
	sort.Ints(sortedInput)
	sortedGot := make([]int, len(got))
	copy(sortedGot, got)
	sort.Ints(sortedGot)
	for i := range sortedInput {
		if sortedInput[i] != sortedGot[i] {
			t.Errorf("Shuffle() changed element at sorted[%d]: got %d, want %d", i, sortedGot[i], sortedInput[i])
		}
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		separator string
		want      string
	}{
		{"empty", []int{}, ",", ""},
		{"single", []int{1}, ",", "1"},
		{"multiple", []int{1, 2, 3}, ", ", "1, 2, 3"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := coll.Join(tc.input, tc.separator)
			if got != tc.want {
				t.Errorf("Join() = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name  string
		input []any
		want  []int
	}{
		{"empty", []any{}, []int{}},
		{"flat", []any{1, 2, 3}, []int{1, 2, 3}},
		{"nested", []any{1, []any{2, 3}, []any{[]any{4, 5}}}, []int{1, 2, 3, 4, 5}},
		{"mixed types", []any{1, "skip", []any{2}}, []int{1, 2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := coll.Flatten[int](tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("Flatten() len = %d; want %d (got %v)", len(got), len(tc.want), got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("Flatten()[%d] = %d; want %d", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// ─── Map utilities ───────────────────────────────────────────────────────────

func TestJoinKeySep(t *testing.T) {
	tests := []struct {
		name      string
		m         map[string]int
		separator string
		wantKeys  []string // all keys must appear in the joined result
	}{
		{"empty", map[string]int{}, ",", nil},
		{"single", map[string]int{"a": 1}, ",", []string{"a"}},
		{"multiple", map[string]int{"a": 1, "b": 2}, ", ", []string{"a", "b"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := coll.JoinKeySep(tc.m, tc.separator)
			for _, k := range tc.wantKeys {
				found := false
				// Simple substring check is enough — order is non-deterministic.
				for i := 0; i+len(k) <= len(got); i++ {
					if got[i:i+len(k)] == k {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("JoinKeySep() = %q; expected to contain key %q", got, k)
				}
			}
		})
	}
}

func TestMergeComp(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 3, "c": 4}
	got := coll.MergeComp(m1, m2)
	if got["a"] != 1 || got["b"] != 3 || got["c"] != 4 {
		t.Errorf("MergeComp() = %v; want {a:1 b:3 c:4}", got)
	}
}

func TestPickComp(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := coll.PickComp(m, "a", "c")
	if len(got) != 2 || got["a"] != 1 || got["c"] != 3 {
		t.Errorf("PickComp() = %v; want {a:1 c:3}", got)
	}
}

func TestOmitComp(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := coll.OmitComp(m, "b")
	if len(got) != 2 || got["a"] != 1 || got["c"] != 3 {
		t.Errorf("OmitComp() = %v; want {a:1 c:3}", got)
	}
}

func TestInvertComp(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	got := coll.InvertComp(m)
	if got[1] != "a" || got[2] != "b" {
		t.Errorf("InvertComp() = %v; want {1:a 2:b}", got)
	}
}

func TestGetOrDefault(t *testing.T) {
	m := map[string]int{"a": 1}
	if v := coll.GetOrDefault(m, "a", 99); v != 1 {
		t.Errorf("GetOrDefault(existing) = %d; want 1", v)
	}
	if v := coll.GetOrDefault(m, "z", 99); v != 99 {
		t.Errorf("GetOrDefault(missing) = %d; want 99", v)
	}
}

func TestDeepMerge(t *testing.T) {
	dst := map[string]any{
		"fruit": map[string]any{"apple": 5, "banana": 10},
		"count": 1,
	}
	src := map[string]any{
		"fruit": map[string]any{"banana": 7, "orange": 2},
		"count": 2,
	}
	got := coll.DeepMerge(dst, src)
	fruit, ok := got["fruit"].(map[string]any)
	if !ok {
		t.Fatal("DeepMerge() fruit should be map[string]any")
	}
	if fruit["apple"] != 5 {
		t.Errorf("apple = %v; want 5", fruit["apple"])
	}
	if fruit["banana"] != 7 {
		t.Errorf("banana = %v; want 7", fruit["banana"])
	}
	if fruit["orange"] != 2 {
		t.Errorf("orange = %v; want 2", fruit["orange"])
	}
	if got["count"] != 2 {
		t.Errorf("count = %v; want 2", got["count"])
	}
}

func TestFlattenMap(t *testing.T) {
	nested := map[string]any{
		"user": map[string]any{
			"name": "Alice",
			"addr": map[string]any{"city": "Wonderland"},
		},
		"age": 30,
	}
	got := coll.FlattenMap(nested, "")
	if got["user.name"] != "Alice" {
		t.Errorf("FlattenMap() user.name = %v; want Alice", got["user.name"])
	}
	if got["user.addr.city"] != "Wonderland" {
		t.Errorf("FlattenMap() user.addr.city = %v; want Wonderland", got["user.addr.city"])
	}
	if got["age"] != 30 {
		t.Errorf("FlattenMap() age = %v; want 30", got["age"])
	}
}

func TestUnflattenMap(t *testing.T) {
	flat := map[string]any{
		"user.name":      "Alice",
		"user.addr.city": "Wonderland",
		"age":            30,
	}
	got := coll.UnflattenMap(flat, "")
	user, ok := got["user"].(map[string]any)
	if !ok {
		t.Fatal("UnflattenMap() user should be map[string]any")
	}
	if user["name"] != "Alice" {
		t.Errorf("user.name = %v; want Alice", user["name"])
	}
	addr, ok := user["addr"].(map[string]any)
	if !ok {
		t.Fatal("UnflattenMap() user.addr should be map[string]any")
	}
	if addr["city"] != "Wonderland" {
		t.Errorf("user.addr.city = %v; want Wonderland", addr["city"])
	}
	if got["age"] != 30 {
		t.Errorf("age = %v; want 30", got["age"])
	}
}

// ─── HashMap nil-receiver safety ────────────────────────────────────────────

func TestHashMap_NilReceiver(t *testing.T) {
	var m *coll.HashMap[string, int]

	t.Run("Get", func(t *testing.T) {
		if v := m.Get("key"); v != 0 {
			t.Errorf("Get() on nil = %d; want 0", v)
		}
	})
	t.Run("Size", func(t *testing.T) {
		if v := m.Size(); v != 0 {
			t.Errorf("Size() on nil = %d; want 0", v)
		}
	})
	t.Run("IsEmpty", func(t *testing.T) {
		if !m.IsEmpty() {
			t.Error("IsEmpty() on nil = false; want true")
		}
	})
	t.Run("ContainsKey", func(t *testing.T) {
		if m.ContainsKey("key") {
			t.Error("ContainsKey() on nil = true; want false")
		}
	})
	t.Run("KeySet", func(t *testing.T) {
		if ks := m.KeySet(); ks != nil {
			t.Errorf("KeySet() on nil = %v; want nil", ks)
		}
	})
}
