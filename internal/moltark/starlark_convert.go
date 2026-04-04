package moltark

import (
	"fmt"

	"go.starlark.net/starlark"
)

func starlarkValueToGo(value starlark.Value) (any, error) {
	switch typed := value.(type) {
	case starlark.NoneType:
		return nil, nil
	case starlark.String:
		return string(typed), nil
	case starlark.Bool:
		return bool(typed), nil
	case starlark.Int:
		if intValue, ok := typed.Int64(); ok {
			return intValue, nil
		}
		return nil, fmt.Errorf("integer %s is out of range", typed.String())
	case starlark.Float:
		return float64(typed), nil
	case *starlark.List:
		values := make([]any, 0, typed.Len())
		iter := typed.Iterate()
		defer iter.Done()
		var item starlark.Value
		for iter.Next(&item) {
			converted, err := starlarkValueToGo(item)
			if err != nil {
				return nil, err
			}
			values = append(values, converted)
		}
		return values, nil
	case starlark.Tuple:
		values := make([]any, 0, len(typed))
		for _, item := range typed {
			converted, err := starlarkValueToGo(item)
			if err != nil {
				return nil, err
			}
			values = append(values, converted)
		}
		return values, nil
	case *starlark.Dict:
		values := make(map[string]any, typed.Len())
		for _, entry := range typed.Items() {
			key, ok := starlark.AsString(entry[0])
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			converted, err := starlarkValueToGo(entry[1])
			if err != nil {
				return nil, err
			}
			values[key] = converted
		}
		return values, nil
	case *factValueRefValue:
		return FactValueRef{
			TargetProjectID: typed.targetProjectID,
			Name:            typed.name,
			Path:            typed.path,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported value type %s", value.Type())
	}
}
