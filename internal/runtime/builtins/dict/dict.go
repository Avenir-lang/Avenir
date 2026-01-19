package dict

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerLength()
	registerKeys()
	registerValues()
	registerHas()
	registerGet()
	registerSet()
	registerRemove()
}

func registerLength() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.DictLength,
			Name:       "length",
			Arity:      1,
			ParamNames: []string{"self"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeDict,
			MethodName:   "length",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			dictVal, err := requireDict(args, "dict.length")
			if err != nil {
				return value.Value{}, err
			}
			return value.Int(int64(len(dictVal.Dict))), nil
		},
	})
}

func registerKeys() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.DictKeys,
			Name:       "keys",
			Arity:      1,
			ParamNames: []string{"self"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeString}}},
			ReceiverType: builtins.TypeDict,
			MethodName:   "keys",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			dictVal, err := requireDict(args, "dict.keys")
			if err != nil {
				return value.Value{}, err
			}
			keys := make([]value.Value, 0, len(dictVal.Dict))
			for k := range dictVal.Dict {
				keys = append(keys, value.Str(k))
			}
			return value.List(keys), nil
		},
	})
}

func registerValues() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.DictValues,
			Name:       "values",
			Arity:      1,
			ParamNames: []string{"self"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeDict,
			MethodName:   "values",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			dictVal, err := requireDict(args, "dict.values")
			if err != nil {
				return value.Value{}, err
			}
			values := make([]value.Value, 0, len(dictVal.Dict))
			for _, v := range dictVal.Dict {
				values = append(values, v)
			}
			return value.List(values), nil
		},
	})
}

func registerHas() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.DictHas,
			Name:       "has",
			Arity:      2,
			ParamNames: []string{"self", "key"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBool},
			ReceiverType: builtins.TypeDict,
			MethodName:   "has",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			dictVal, err := requireDict(args, "dict.has")
			if err != nil {
				return value.Value{}, err
			}
			keyVal := args[1].(value.Value)
			if keyVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("dict.has expects key as string")
			}
			_, ok := dictVal.Dict[keyVal.Str]
			return value.Bool(ok), nil
		},
	})
}

func registerGet() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.DictGet,
			Name:       "get",
			Arity:      2,
			ParamNames: []string{"self", "key"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny}, // checker specializes to T?
			ReceiverType: builtins.TypeDict,
			MethodName:   "get",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			dictVal, err := requireDict(args, "dict.get")
			if err != nil {
				return value.Value{}, err
			}
			keyVal := args[1].(value.Value)
			if keyVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("dict.get expects key as string")
			}
			val, ok := dictVal.Dict[keyVal.Str]
			if !ok {
				return value.None(), nil
			}
			return value.Some(val), nil
		},
	})
}

func registerSet() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.DictSet,
			Name:       "set",
			Arity:      3,
			ParamNames: []string{"self", "key", "value"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeDict,
			MethodName:   "set",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			dictVal, err := requireDict(args, "dict.set")
			if err != nil {
				return value.Value{}, err
			}
			if dictVal.Dict == nil {
				return value.Value{}, fmt.Errorf("dict.set called on nil dict")
			}
			keyVal := args[1].(value.Value)
			if keyVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("dict.set expects key as string")
			}
			dictVal.Dict[keyVal.Str] = args[2].(value.Value)
			return value.Value{}, nil
		},
	})
}

func registerRemove() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.DictRemove,
			Name:       "remove",
			Arity:      2,
			ParamNames: []string{"self", "key"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBool},
			ReceiverType: builtins.TypeDict,
			MethodName:   "remove",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			dictVal, err := requireDict(args, "dict.remove")
			if err != nil {
				return value.Value{}, err
			}
			keyVal := args[1].(value.Value)
			if keyVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("dict.remove expects key as string")
			}
			if dictVal.Dict == nil {
				return value.Bool(false), nil
			}
			_, ok := dictVal.Dict[keyVal.Str]
			delete(dictVal.Dict, keyVal.Str)
			return value.Bool(ok), nil
		},
	})
}

func requireDict(args []interface{}, name string) (value.Value, error) {
	if len(args) < 1 {
		return value.Value{}, fmt.Errorf("%s expects receiver", name)
	}
	receiver := args[0].(value.Value)
	if receiver.Kind != value.KindDict {
		return value.Value{}, fmt.Errorf("%s called on non-dict type %v", name, receiver.Kind)
	}
	return receiver, nil
}
