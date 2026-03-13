package sql

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerAsyncPgConnect()
	registerAsyncPgClose()
	registerAsyncPgQuery()
	registerAsyncPgExec()
	registerAsyncPgBegin()
	registerAsyncPgCommit()
	registerAsyncPgRollback()
	registerAsyncSqliteConnect()
	registerAsyncSqliteClose()
	registerAsyncSqliteQuery()
	registerAsyncSqliteExec()
	registerAsyncSqliteBegin()
	registerAsyncSqliteCommit()
	registerAsyncSqliteRollback()
}

func requireHandle(v value.Value) ([]byte, error) {
	if v.Kind == value.KindBytes {
		return v.Bytes, nil
	}
	return nil, fmt.Errorf("expected handle (bytes), got %v", v.Kind)
}

func extractParams(v value.Value) ([]interface{}, error) {
	if v.Kind != value.KindList {
		return nil, fmt.Errorf("expected list for params, got %v", v.Kind)
	}
	params := make([]interface{}, len(v.List))
	for i, item := range v.List {
		switch item.Kind {
		case value.KindInt:
			params[i] = item.Int
		case value.KindFloat:
			params[i] = item.Float
		case value.KindString:
			params[i] = item.Str
		case value.KindBool:
			params[i] = item.Bool
		case value.KindBytes:
			params[i] = item.Bytes
		case value.KindOptional:
			if item.Optional == nil || !item.Optional.IsSome {
				params[i] = nil
			} else {
				inner := item.Optional.Value
				switch inner.Kind {
				case value.KindInt:
					params[i] = inner.Int
				case value.KindFloat:
					params[i] = inner.Float
				case value.KindString:
					params[i] = inner.Str
				case value.KindBool:
					params[i] = inner.Bool
				default:
					params[i] = nil
				}
			}
		default:
			params[i] = nil
		}
	}
	return params, nil
}

func sqlResultToValue(result *builtins.SQLResultData) value.Value {
	entries := make(map[string]value.Value)

	if result.Columns != nil {
		cols := make([]value.Value, len(result.Columns))
		for i, c := range result.Columns {
			cols[i] = value.Str(c)
		}
		entries["columns"] = value.List(cols)
	} else {
		entries["columns"] = value.List(nil)
	}

	if result.Rows != nil {
		rows := make([]value.Value, len(result.Rows))
		for i, row := range result.Rows {
			rowDict := make(map[string]value.Value, len(row))
			for k, v := range row {
				rowDict[k] = goToValue(v)
			}
			rows[i] = value.Dict(rowDict)
		}
		entries["rows"] = value.List(rows)
	} else {
		entries["rows"] = value.List(nil)
	}

	entries["rowsAffected"] = value.Int(result.RowsAffected)
	entries["lastInsertId"] = value.Int(result.LastInsertID)

	return value.Dict(entries)
}

func goToValue(v interface{}) value.Value {
	if v == nil {
		return value.None()
	}
	switch val := v.(type) {
	case int64:
		return value.Int(val)
	case float64:
		return value.Float(val)
	case bool:
		return value.Bool(val)
	case string:
		return value.Str(val)
	case []byte:
		return value.Bytes(val)
	default:
		return value.Str(fmt.Sprintf("%v", val))
	}
}

func registerAsyncPgConnect() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLPgConnect,
			Name:       "__builtin_async_sql_pg_connect",
			Arity:      5,
			ParamNames: []string{"host", "port", "user", "password", "database"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 5 {
				return nil, fmt.Errorf("__builtin_async_sql_pg_connect expects 5 arguments, got %d", len(args))
			}
			if env == nil || env.SQL() == nil {
				return nil, fmt.Errorf("runtime env sql is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			userVal := args[2].(value.Value)
			passVal := args[3].(value.Value)
			dbVal := args[4].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindString ||
				userVal.Kind != value.KindString || passVal.Kind != value.KindString ||
				dbVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_sql_pg_connect expects all string arguments")
			}
			host, port, user, pass, database := hostVal.Str, portVal.Str, userVal.Str, passVal.Str, dbVal.Str
			sqlSvc := env.SQL()
			return builtins.RunAsync(func() (interface{}, error) {
				handle, err := sqlSvc.PgConnect(host, port, user, pass, database)
				if err != nil {
					return nil, err
				}
				return value.Bytes(append([]byte(nil), handle...)), nil
			}), nil
		},
	})
}

func registerAsyncPgClose() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLPgClose,
			Name:       "__builtin_async_sql_pg_close",
			Arity:      1,
			ParamNames: []string{"handle"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_sql_pg_close expects 1 argument, got %d", len(args))
			}
			if env == nil || env.SQL() == nil {
				return nil, fmt.Errorf("runtime env sql is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			sqlSvc := env.SQL()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := sqlSvc.PgClose(handle); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncPgQuery() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLPgQuery,
			Name:       "__builtin_async_sql_pg_query",
			Arity:      3,
			ParamNames: []string{"handle", "query", "params"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("__builtin_async_sql_pg_query expects 3 arguments, got %d", len(args))
			}
			if env == nil || env.SQL() == nil {
				return nil, fmt.Errorf("runtime env sql is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			queryVal := args[1].(value.Value)
			if queryVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_sql_pg_query expects query as string")
			}
			paramsVal := args[2].(value.Value)
			params, err := extractParams(paramsVal)
			if err != nil {
				return nil, err
			}
			query := queryVal.Str
			sqlSvc := env.SQL()
			return builtins.RunAsync(func() (interface{}, error) {
				result, err := sqlSvc.PgQuery(handle, query, params)
				if err != nil {
					return nil, err
				}
				return sqlResultToValue(result), nil
			}), nil
		},
	})
}

func registerAsyncPgExec() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLPgExec,
			Name:       "__builtin_async_sql_pg_exec",
			Arity:      3,
			ParamNames: []string{"handle", "query", "params"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("__builtin_async_sql_pg_exec expects 3 arguments, got %d", len(args))
			}
			if env == nil || env.SQL() == nil {
				return nil, fmt.Errorf("runtime env sql is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			queryVal := args[1].(value.Value)
			if queryVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_sql_pg_exec expects query as string")
			}
			paramsVal := args[2].(value.Value)
			params, err := extractParams(paramsVal)
			if err != nil {
				return nil, err
			}
			query := queryVal.Str
			sqlSvc := env.SQL()
			return builtins.RunAsync(func() (interface{}, error) {
				result, err := sqlSvc.PgExec(handle, query, params)
				if err != nil {
					return nil, err
				}
				return sqlResultToValue(result), nil
			}), nil
		},
	})
}

func registerAsyncPgBegin() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLPgBegin,
			Name:       "__builtin_async_sql_pg_begin",
			Arity:      1,
			ParamNames: []string{"handle"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_sql_pg_begin expects 1 argument, got %d", len(args))
			}
			if env == nil || env.SQL() == nil {
				return nil, fmt.Errorf("runtime env sql is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			sqlSvc := env.SQL()
			return builtins.RunAsync(func() (interface{}, error) {
				txHandle, err := sqlSvc.PgBegin(handle)
				if err != nil {
					return nil, err
				}
				return value.Bytes(append([]byte(nil), txHandle...)), nil
			}), nil
		},
	})
}

func registerAsyncPgCommit() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLPgCommit,
			Name:       "__builtin_async_sql_pg_commit",
			Arity:      1,
			ParamNames: []string{"txHandle"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_sql_pg_commit expects 1 argument, got %d", len(args))
			}
			if env == nil || env.SQL() == nil {
				return nil, fmt.Errorf("runtime env sql is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			sqlSvc := env.SQL()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := sqlSvc.PgCommit(handle); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncPgRollback() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLPgRollback,
			Name:       "__builtin_async_sql_pg_rollback",
			Arity:      1,
			ParamNames: []string{"txHandle"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_sql_pg_rollback expects 1 argument, got %d", len(args))
			}
			if env == nil || env.SQL() == nil {
				return nil, fmt.Errorf("runtime env sql is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			sqlSvc := env.SQL()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := sqlSvc.PgRollback(handle); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}
