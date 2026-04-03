package moltark

import (
	"testing"

	"go.starlark.net/starlark"
)

func TestStarlarkStringListRejectsDicts(t *testing.T) {
	dict := starlark.NewDict(1)
	if err := dict.SetKey(starlark.String("name"), starlark.String("demo")); err != nil {
		t.Fatalf("set dict key: %v", err)
	}

	_, err := starlarkStringList(dict, "label")
	if err == nil {
		t.Fatal("expected dict input to fail")
	}
}

func TestStarlarkStringListAcceptsTuples(t *testing.T) {
	values, err := starlarkStringList(starlark.Tuple{starlark.String("one"), starlark.String("two")}, "label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 2 || values[0] != "one" || values[1] != "two" {
		t.Fatalf("unexpected values: %#v", values)
	}
}
