package mapbuilder

import "testing"

func TestNewAndFrom(t *testing.T) {
	b := New()
	if b.data == nil {
		t.Fatal("New() should initialize data")
	}

	src := map[string]any{
		"a": 1,
		"b": map[string]any{"c": 2},
	}
	b2 := From(src)
	if b2.data["a"] != 1 {
		t.Fatal("From() should copy values")
	}

	// Check independence
	src["a"] = 99
	src["b"].(map[string]any)["c"] = 99
	if b2.data["a"] == 99 {
		t.Fatal("From() should deep copy, modifying src shouldn't affect builder")
	}
}

func TestAddAndAddAll(t *testing.T) {
	b := New().Add("a", 1).Add("b.c", "literal")
	res, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if res["a"] != 1 || res["b.c"] != "literal" {
		t.Fatal("Add failed")
	}

	b = New().AddAll(map[string]any{"a": 1, "b": 2})
	res, err = b.Build()
	if err != nil || res["a"] != 1 {
		t.Fatal("AddAll failed")
	}

	// Empty key error
	b = New().Add("", 1)
	if b.Cause() == nil {
		t.Fatal("Add empty key should error")
	}

	b = New().AddAll(map[string]any{"a": 1, "": 2})
	if b.Cause() == nil {
		t.Fatal("AddAll empty key should error")
	}
}

func TestAddIfAbsent(t *testing.T) {
	b := New().Add("a", 1).AddIfAbsent("a", 99).AddIfAbsent("b", 2)
	res := b.MustBuild()
	if res["a"] != 1 || res["b"] != 2 {
		t.Fatal("AddIfAbsent failed")
	}

	// Explicit nil counts as present
	b = New().Add("a", nil).AddIfAbsent("a", 99)
	res = b.MustBuild()
	if res["a"] != nil {
		t.Fatal("AddIfAbsent explicit nil failed")
	}
}

func TestPaths(t *testing.T) {
	b := New().
		AddPath("user.profile.name", "Alice").
		AddPath("user.profile.age", 30)
	res, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	profile := res["user"].(map[string]any)["profile"].(map[string]any)
	if profile["name"] != "Alice" || profile["age"] != 30 {
		t.Fatal("AddPath failed")
	}

	// AddPathIfAbsent
	b = New().AddPath("u.n", 1).AddPathIfAbsent("u.n", 2)
	res = b.MustBuild()
	if res["u"].(map[string]any)["n"] != 1 {
		t.Fatal("AddPathIfAbsent failed")
	}

	// Invalid paths
	for _, p := range []string{"", "u.", ".u", "u..n"} {
		b = New().AddPath(p, 1)
		if b.Cause() == nil {
			t.Fatalf("Path %q should be invalid", p)
		}
	}

	// Conflict with scalar
	b = New().Add("u", "scalar").AddPath("u.n", 1)
	if b.Cause() == nil {
		t.Fatal("Conflict should error")
	}
}

func TestAppend(t *testing.T) {
	b := New().
		Append("tags", "go", "backend").
		Append("tags", "opensource").
		AppendPath("user.roles", "admin", "editor")
	res, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}

	tags := res["tags"].([]any)
	if len(tags) != 3 || tags[2] != "opensource" {
		t.Fatal("Append failed")
	}

	roles := res["user"].(map[string]any)["roles"].([]any)
	if len(roles) != 2 || roles[0] != "admin" {
		t.Fatal("AppendPath failed")
	}

	// Zero values is no-op
	b = New().Append("tags")
	res = b.MustBuild()
	if len(res) != 0 {
		t.Fatal("Append zero values should be no-op")
	}

	// Incompatible existing value
	b = New().Add("tags", "scalar").Append("tags", "new")
	if b.Cause() == nil {
		t.Fatal("Append incompatible should error")
	}
}

func TestMerge(t *testing.T) {
	b := New().
		Add("a", 1).
		Add("b", map[string]any{"c": 2}).
		Merge(map[string]any{
			"b": map[string]any{"d": 3},
			"e": 4,
		})
	res, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	bMap := res["b"].(map[string]any)
	if bMap["c"] != 2 || bMap["d"] != 3 {
		t.Fatal("Merge recursive failed")
	}
	if res["e"] != 4 {
		t.Fatal("Merge new key failed")
	}

	// Merge replacement
	b = New().Add("a", 1).Merge(map[string]any{"a": 2})
	res = b.MustBuild()
	if res["a"] != 2 {
		t.Fatal("Merge replace scalar failed")
	}

	// Merge slice replacement
	b = New().Add("s", []any{1}).Merge(map[string]any{"s": []any{2}})
	res = b.MustBuild()
	s := res["s"].([]any)
	if len(s) != 1 || s[0] != 2 {
		t.Fatal("Merge slice replacement failed")
	}

	// Merge independence
	src := map[string]any{"a": map[string]any{"b": 1}}
	b = New().Merge(src)
	res = b.MustBuild()
	res["a"].(map[string]any)["b"] = 99
	if src["a"].(map[string]any)["b"] == 99 {
		t.Fatal("Merge should not mutate source")
	}
}

func TestRemove(t *testing.T) {
	b := New().
		Add("a", 1).
		Add("b", 2).
		Remove("a")
	res := b.MustBuild()
	if res["a"] != nil {
		t.Fatal("Remove failed")
	}

	b = New().
		AddPath("u.p.n", 1).
		AddPath("u.p.a", 2).
		RemovePath("u.p.n")
	res = b.MustBuild()
	p := res["u"].(map[string]any)["p"].(map[string]any)
	if p["n"] != nil || p["a"] != 2 {
		t.Fatal("RemovePath failed")
	}

	// Pruning empty maps
	b = New().
		AddPath("u.p.n", 1).
		RemovePath("u.p.n")
	res = b.MustBuild()
	if _, ok := res["u"]; ok {
		t.Fatal("RemovePath should prune empty maps")
	}
}

func TestBuildAndError(t *testing.T) {
	b := New().Add("", 1).Add("a", 2) // Error on Add
	_, err := b.Build()
	if err == nil {
		t.Fatal("Build should return error")
	}
	if b.Cause() == nil {
		t.Fatal("Cause() should return error")
	}
}

func TestMustBuildPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustBuild should panic on error")
		}
	}()
	b := New().AddPath("u..n", 1)
	b.MustBuild()
}

func TestGetters(t *testing.T) {
	b := New().Add("a", 1).AddPath("u.n", 2)

	v, ok := b.Get("a")
	if !ok || v != 1 {
		t.Fatal("Get failed")
	}

	v, ok = b.GetPath("u.n")
	if !ok || v != 2 {
		t.Fatal("GetPath failed")
	}

	if !b.Has("a") || !b.HasPath("u.n") {
		t.Fatal("Has/HasPath failed")
	}
}
