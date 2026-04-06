package module

import (
	"fmt"
	"hash/fnv"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type factValueRefValue struct {
	targetProjectID string
	name            string
	path            string
}

func (r *factValueRefValue) String() string {
	return fmt.Sprintf("fact_value(target=%q, name=%q, path=%q)", r.targetProjectID, r.name, r.path)
}

func (r *factValueRefValue) Type() string {
	return "fact_value_ref"
}

func (r *factValueRefValue) Freeze() {}

func (r *factValueRefValue) Truth() starlark.Bool {
	return starlark.True
}

func (r *factValueRefValue) Hash() (uint32, error) {
	h := fnv.New32a()
	_, _ = h.Write([]byte(r.targetProjectID))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(r.name))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(r.path))
	return h.Sum32(), nil
}

func (r *factValueRefValue) Attr(name string) (starlark.Value, error) {
	switch name {
	case "target":
		return starlark.String(r.targetProjectID), nil
	case "name":
		return starlark.String(r.name), nil
	case "path":
		return starlark.String(r.path), nil
	default:
		return nil, nil
	}
}

func (r *factValueRefValue) AttrNames() []string {
	return []string{"target", "name", "path"}
}

func (r *factValueRefValue) CompareSameType(op syntax.Token, y_ starlark.Value, depth int) (bool, error) {
	y, ok := y_.(*factValueRefValue)
	if !ok {
		return false, fmt.Errorf("comparison against non-fact reference")
	}
	left := r.targetProjectID + "\x00" + r.name + "\x00" + r.path
	right := y.targetProjectID + "\x00" + y.name + "\x00" + y.path
	return starlark.CompareDepth(op, starlark.String(left), starlark.String(right), depth)
}
