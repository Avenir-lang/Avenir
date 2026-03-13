package sql

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func registerAsyncSqliteConnect() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLSqliteConnect,
			Name:       "__builtin_async_sql_sqlite_connect",
			Arity:      1,
			ParamNames: []string{"path"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_sql_sqlite_connect expects 1 argument, got %d", len(args))
			}
			if env == nil || env.SQL() == nil {
				return nil, fmt.Errorf("runtime env sql is nil")
			}
			pathVal := args[0].(value.Value)
			if pathVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_sql_sqlite_connect expects string argument")
			}
			path := pathVal.Str
			sqlSvc := env.SQL()
			return builtins.RunAsync(func() (interface{}, error) {
				handle, err := sqlSvc.SqliteConnect(path)
				if err != nil {
					return nil, err
				}
				return value.Bytes(append([]byte(nil), handle...)), nil
			}), nil
		},
	})
}

func registerAsyncSqliteClose() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLSqliteClose,
			Name:       "__builtin_async_sql_sqlite_close",
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
				return nil, fmt.Errorf("__builtin_async_sql_sqlite_close expects 1 argument, got %d", len(args))
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
				if err := sqlSvc.SqliteClose(handle); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncSqliteQuery() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLSqliteQuery,
			Name:       "__builtin_async_sql_sqlite_query",
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
				return nil, fmt.Errorf("__builtin_async_sql_sqlite_query expects 3 arguments, got %d", len(args))
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
				return nil, fmt.Errorf("__builtin_async_sql_sqlite_query expects query as string")
			}
			paramsVal := args[2].(value.Value)
			params, err := extractParams(paramsVal)
			if err != nil {
				return nil, err
			}
			query := queryVal.Str
			sqlSvc := env.SQL()
			return builtins.RunAsync(func() (interface{}, error) {
				result, err := sqlSvc.SqliteQuery(handle, query, params)
				if err != nil {
					return nil, err
				}
				return sqlResultToValue(result), nil
			}), nil
		},
	})
}

func registerAsyncSqliteExec() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLSqliteExec,
			Name:       "__builtin_async_sql_sqlite_exec",
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
				return nil, fmt.Errorf("__builtin_async_sql_sqlite_exec expects 3 arguments, got %d", len(args))
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
				return nil, fmt.Errorf("__builtin_async_sql_sqlite_exec expects query as string")
			}
			paramsVal := args[2].(value.Value)
			params, err := extractParams(paramsVal)
			if err != nil {
				return nil, err
			}
			query := queryVal.Str
			sqlSvc := env.SQL()
			return builtins.RunAsync(func() (interface{}, error) {
				result, err := sqlSvc.SqliteExec(handle, query, params)
				if err != nil {
					return nil, err
				}
				return sqlResultToValue(result), nil
			}), nil
		},
	})
}

func registerAsyncSqliteBegin() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLSqliteBegin,
			Name:       "__builtin_async_sql_sqlite_begin",
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
				return nil, fmt.Errorf("__builtin_async_sql_sqlite_begin expects 1 argument, got %d", len(args))
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
				txHandle, err := sqlSvc.SqliteBegin(handle)
				if err != nil {
					return nil, err
				}
				return value.Bytes(append([]byte(nil), txHandle...)), nil
			}), nil
		},
	})
}

func registerAsyncSqliteCommit() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLSqliteCommit,
			Name:       "__builtin_async_sql_sqlite_commit",
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
				return nil, fmt.Errorf("__builtin_async_sql_sqlite_commit expects 1 argument, got %d", len(args))
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
				if err := sqlSvc.SqliteCommit(handle); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncSqliteRollback() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSQLSqliteRollback,
			Name:       "__builtin_async_sql_sqlite_rollback",
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
				return nil, fmt.Errorf("__builtin_async_sql_sqlite_rollback expects 1 argument, got %d", len(args))
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
				if err := sqlSvc.SqliteRollback(handle); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}
