package runtime

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"

	"avenir/internal/runtime/builtins"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

type sqlService struct {
	nextID uint64
	mu     sync.Mutex
	conns  map[uint64]*sql.DB
	txs    map[uint64]*sql.Tx
}

func newSQLService() *sqlService {
	return &sqlService{
		conns: make(map[uint64]*sql.DB),
		txs:   make(map[uint64]*sql.Tx),
	}
}

func (s *sqlService) nextHandle() uint64 {
	return atomic.AddUint64(&s.nextID, 1)
}

func (s *sqlService) PgConnect(host, port, user, password, database string) ([]byte, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql: connection error: %w", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("sql: connection error: %w", err)
	}
	id := s.nextHandle()
	s.mu.Lock()
	s.conns[id] = db
	s.mu.Unlock()
	return encodeHandle(id), nil
}

func (s *sqlService) PgClose(handle []byte) error {
	id, err := decodeHandle(handle)
	if err != nil {
		return err
	}
	s.mu.Lock()
	db, ok := s.conns[id]
	if ok {
		delete(s.conns, id)
	}
	s.mu.Unlock()
	if !ok || db == nil {
		return fmt.Errorf("sql: invalid connection handle")
	}
	return db.Close()
}

func (s *sqlService) getConn(handle []byte) (*sql.DB, error) {
	id, err := decodeHandle(handle)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	db := s.conns[id]
	s.mu.Unlock()
	if db == nil {
		return nil, fmt.Errorf("sql: invalid connection handle")
	}
	return db, nil
}

type queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (s *sqlService) getQueryable(handle []byte) (queryable, error) {
	id, err := decodeHandle(handle)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	if tx := s.txs[id]; tx != nil {
		s.mu.Unlock()
		return tx, nil
	}
	db := s.conns[id]
	s.mu.Unlock()
	if db == nil {
		return nil, fmt.Errorf("sql: invalid connection or transaction handle")
	}
	return db, nil
}

func (s *sqlService) PgQuery(handle []byte, query string, params []interface{}) (*builtins.SQLResultData, error) {
	q, err := s.getQueryable(handle)
	if err != nil {
		return nil, err
	}
	rows, err := q.QueryContext(context.Background(), query, params...)
	if err != nil {
		return nil, fmt.Errorf("sql: query error: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (s *sqlService) PgExec(handle []byte, query string, params []interface{}) (*builtins.SQLResultData, error) {
	q, err := s.getQueryable(handle)
	if err != nil {
		return nil, err
	}
	result, err := q.ExecContext(context.Background(), query, params...)
	if err != nil {
		return nil, fmt.Errorf("sql: query error: %w", err)
	}
	affected, _ := result.RowsAffected()
	lastID, _ := result.LastInsertId()
	return &builtins.SQLResultData{
		Columns:      nil,
		Rows:         nil,
		RowsAffected: affected,
		LastInsertID: lastID,
	}, nil
}

func (s *sqlService) PgBegin(handle []byte) ([]byte, error) {
	db, err := s.getConn(handle)
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("sql: transaction error: %w", err)
	}
	id := s.nextHandle()
	s.mu.Lock()
	s.txs[id] = tx
	s.mu.Unlock()
	return encodeHandle(id), nil
}

func (s *sqlService) getTx(txHandle []byte) (*sql.Tx, uint64, error) {
	id, err := decodeHandle(txHandle)
	if err != nil {
		return nil, 0, err
	}
	s.mu.Lock()
	tx := s.txs[id]
	s.mu.Unlock()
	if tx == nil {
		return nil, 0, fmt.Errorf("sql: invalid transaction handle")
	}
	return tx, id, nil
}

func (s *sqlService) PgCommit(txHandle []byte) error {
	tx, id, err := s.getTx(txHandle)
	if err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.txs, id)
	s.mu.Unlock()
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sql: transaction error: %w", err)
	}
	return nil
}

func (s *sqlService) PgRollback(txHandle []byte) error {
	tx, id, err := s.getTx(txHandle)
	if err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.txs, id)
	s.mu.Unlock()
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("sql: transaction error: %w", err)
	}
	return nil
}

func (s *sqlService) SqliteConnect(path string) ([]byte, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sql: connection error: %w", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("sql: connection error: %w", err)
	}
	id := s.nextHandle()
	s.mu.Lock()
	s.conns[id] = db
	s.mu.Unlock()
	return encodeHandle(id), nil
}

func (s *sqlService) SqliteClose(handle []byte) error {
	return s.PgClose(handle)
}

func (s *sqlService) SqliteQuery(handle []byte, query string, params []interface{}) (*builtins.SQLResultData, error) {
	return s.PgQuery(handle, query, params)
}

func (s *sqlService) SqliteExec(handle []byte, query string, params []interface{}) (*builtins.SQLResultData, error) {
	return s.PgExec(handle, query, params)
}

func (s *sqlService) SqliteBegin(handle []byte) ([]byte, error) {
	return s.PgBegin(handle)
}

func (s *sqlService) SqliteCommit(txHandle []byte) error {
	return s.PgCommit(txHandle)
}

func (s *sqlService) SqliteRollback(txHandle []byte) error {
	return s.PgRollback(txHandle)
}

func scanRows(rows *sql.Rows) (*builtins.SQLResultData, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("sql: query error: %w", err)
	}
	var resultRows []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("sql: query error: %w", err)
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			row[col] = normalizeValue(values[i])
		}
		resultRows = append(resultRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sql: query error: %w", err)
	}
	return &builtins.SQLResultData{
		Columns: cols,
		Rows:    resultRows,
	}, nil
}

func normalizeValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return val
	case bool:
		return val
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
