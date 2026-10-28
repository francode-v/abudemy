// Package store is the repository layer. It owns every SQL query and is the
// only place that talks to PostgreSQL. Business rules live in the service
// package, which builds on top of this package.
package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the subset of pgx shared by the pool and a transaction.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Store provides the data access methods for every aggregate.
type Store struct {
	q    DB
	pool *pgxpool.Pool
}

// New returns a Store backed by the given pool.
func New(pool *pgxpool.Pool) *Store { return &Store{q: pool, pool: pool} }

// InTx runs fn inside a transaction. If the Store is already transactional it
// simply invokes fn with the same Store, so nested calls are safe.
func (s *Store) InTx(ctx context.Context, fn func(tx *Store) error) error {
	if _, ok := s.q.(pgx.Tx); ok {
		return fn(s)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(&Store{q: tx, pool: s.pool}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// ErrNotFound is returned when a query expected a row but found none.
var ErrNotFound = errors.New("resource not found")

// ErrConflict is returned when a unique or check constraint is violated.
var ErrConflict = errors.New("resource conflict")

// ErrForeignKey is returned when a referenced row does not exist.
var ErrForeignKey = errors.New("referenced resource does not exist")

// TranslateError maps low-level driver errors onto the repository error model.
func TranslateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%w: %s", ErrConflict, pgErr.ConstraintName)
		case "23503": // foreign_key_violation
			return fmt.Errorf("%w: %s", ErrForeignKey, pgErr.ConstraintName)
		case "23514": // check_violation
			return fmt.Errorf("%w: %s", ErrConflict, pgErr.ConstraintName)
		}
	}
	return err
}

// scanRow is a helper that translates pgx.ErrNoRows into ErrNotFound.
func scanRow(row pgx.Row, dest ...any) error {
	if err := row.Scan(dest...); err != nil {
		return TranslateError(err)
	}
	return nil
}

// qb is a tiny parameterized query builder. It never interpolates values into
// SQL; only placeholder numbers are generated.
type qb struct {
	conds []string
	args  []any
}

// add appends a condition. Use %s for each placeholder, e.g.
// b.add("comuna = %s", comuna).
func (b *qb) add(expr string, vals ...any) {
	ph := make([]any, len(vals))
	for i, v := range vals {
		b.args = append(b.args, v)
		ph[i] = "$" + strconv.Itoa(len(b.args))
	}
	b.conds = append(b.conds, fmt.Sprintf(expr, ph...))
}

// where renders the WHERE clause (without the keyword) or an empty string.
func (b *qb) where() string {
	if len(b.conds) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(b.conds, " AND ")
}

// limitOffset appends LIMIT/OFFSET placeholders and returns the clause.
func (b *qb) limitOffset(limit, offset int) string {
	b.args = append(b.args, limit)
	lim := "$" + strconv.Itoa(len(b.args))
	b.args = append(b.args, offset)
	off := "$" + strconv.Itoa(len(b.args))
	return " LIMIT " + lim + " OFFSET " + off
}
