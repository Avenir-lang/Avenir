package meta

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/types"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TypeOf,
			Name:       "typeOf",
			Arity:      1,
			ParamNames: []string{"value"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("typeOf expects 1 argument, got %d", len(args))
			}
			arg := args[0].(value.Value)
			typ, err := typeFromValue(arg, env)
			if err != nil {
				return value.Value{}, err
			}
			return value.Str(typ.String()), nil
		},
	})
}

func typeFromValue(val value.Value, env builtins.Env) (types.Type, error) {
	switch val.Kind {
	case value.KindInt:
		return types.Int, nil
	case value.KindFloat:
		return types.Float, nil
	case value.KindString:
		return types.String, nil
	case value.KindBool:
		return types.Bool, nil
	case value.KindBytes:
		return types.Bytes, nil
	case value.KindError:
		return types.ErrorType, nil
	case value.KindOptional:
		if val.Optional == nil || !val.Optional.IsSome {
			return &types.Optional{Inner: types.Any}, nil
		}
		inner, err := typeFromValue(val.Optional.Value, env)
		if err != nil {
			return nil, err
		}
		if types.IsInvalid(inner) {
			inner = types.Any
		}
		return &types.Optional{Inner: inner}, nil
	case value.KindList:
		elemTypes := make([]types.Type, 0)
		for _, el := range val.List {
			et, err := typeFromValue(el, env)
			if err != nil {
				return nil, err
			}
			if types.IsInvalid(et) {
				continue
			}
			already := false
			for _, existing := range elemTypes {
				if types.Equal(existing, et) {
					already = true
					break
				}
			}
			if !already {
				elemTypes = append(elemTypes, et)
			}
		}
		if len(elemTypes) == 0 {
			elemTypes = []types.Type{types.Any}
		}
		return &types.List{ElementTypes: elemTypes}, nil
	case value.KindStruct:
		if val.Struct == nil {
			return nil, fmt.Errorf("typeOf: nil struct value")
		}
		if env == nil {
			return nil, fmt.Errorf("typeOf: runtime env is nil")
		}
		if name, ok := env.StructTypeName(val.Struct.TypeIndex); ok {
			return &types.Struct{Name: name}, nil
		}
		return nil, fmt.Errorf("typeOf: unknown struct type index %d", val.Struct.TypeIndex)
	case value.KindDict:
		valueTypes := make([]types.Type, 0)
		for _, v := range val.Dict {
			vt, err := typeFromValue(v, env)
			if err != nil {
				return nil, err
			}
			if types.IsInvalid(vt) {
				continue
			}
			already := false
			for _, existing := range valueTypes {
				if types.Equal(existing, vt) {
					already = true
					break
				}
			}
			if !already {
				valueTypes = append(valueTypes, vt)
			}
		}
		var valueType types.Type
		switch len(valueTypes) {
		case 0:
			valueType = types.Any
		case 1:
			valueType = valueTypes[0]
		default:
			valueType = &types.Union{Variants: valueTypes}
		}
		return &types.Dict{ValueType: valueType}, nil
	case value.KindClosure:
		return &types.Func{ParamTypes: []types.Type{}, Result: types.Any}, nil
	case value.KindInvalid:
		return nil, fmt.Errorf("typeOf: invalid value")
	default:
		return nil, fmt.Errorf("typeOf: unsupported value kind %v", val.Kind)
	}
}
