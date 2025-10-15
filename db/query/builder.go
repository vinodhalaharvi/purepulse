package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// ============================================================================
// QUERY - COMPOSITION OF ALL MONOIDS
// ============================================================================

// Query composes all SQL monoids
type Query[T any] struct {
	table      string
	selectM    SelectMonoid
	whereM     WhereMonoid
	joinM      JoinMonoid
	groupByM   GroupByMonoid
	havingM    WhereMonoid // HAVING is also a filter monoid
	aggregateM AggregationMonoid
	orderByM   OrderByMonoid
	limitM     LimitMonoid
}

// Table creates a new query (identity for all monoids)
func Table[T any](name string) Query[T] {
	return Query[T]{
		table:      name,
		selectM:    SelectMonoid{}.Empty(),
		whereM:     WhereMonoid{}.Empty(),
		joinM:      JoinMonoid{}.Empty(),
		groupByM:   GroupByMonoid{}.Empty(),
		havingM:    WhereMonoid{}.Empty(),
		aggregateM: AggregationMonoid{}.Empty(),
		orderByM:   OrderByMonoid{}.Empty(),
		limitM:     LimitMonoid{}.Empty(),
	}
}

// ============================================================================
// MONOID COMBINATORS
// ============================================================================

// Project adds SELECT monoid
func (q Query[T]) Project(s SelectMonoid) Query[T] {
	q.selectM = q.selectM.Combine(s)
	return q
}

// Filter adds WHERE monoid
func (q Query[T]) Filter(w WhereMonoid) Query[T] {
	q.whereM = q.whereM.Combine(w)
	return q
}

// Join adds JOIN monoid
func (q Query[T]) ApplyJoins(j JoinMonoid) Query[T] {
	q.joinM = q.joinM.Combine(j)
	return q
}

// Group adds GROUP BY monoid
func (q Query[T]) Group(g GroupByMonoid) Query[T] {
	q.groupByM = q.groupByM.Combine(g)
	return q
}

// Aggregate adds aggregation monoid
func (q Query[T]) Aggregate(a AggregationMonoid) Query[T] {
	q.aggregateM = q.aggregateM.Combine(a)
	return q
}

// Sort adds ORDER BY monoid
func (q Query[T]) Sort(o OrderByMonoid) Query[T] {
	q.orderByM = q.orderByM.Combine(o)
	return q
}

// Bound adds LIMIT monoid
func (q Query[T]) Bound(l LimitMonoid) Query[T] {
	q.limitM = q.limitM.Combine(l)
	return q
}

// Having adds HAVING monoid
func (q Query[T]) Having(w WhereMonoid) Query[T] {
	q.havingM = q.havingM.Combine(w)
	return q
}

// ============================================================================
// SQL GENERATION
// ============================================================================

// Build generates final SQL
func (q Query[T]) Build() (string, []interface{}) {
	var parts []string
	var allParams Params

	// SELECT clause
	selectSQL := q.selectM.Build()

	// Add aggregations to SELECT if present
	if aggregates := q.aggregateM.Build(); len(aggregates) > 0 {
		if selectSQL == "*" {
			selectSQL = strings.Join(aggregates, ", ")
		} else {
			selectSQL = selectSQL + ", " + strings.Join(aggregates, ", ")
		}
	}

	parts = append(parts, fmt.Sprintf("SELECT %s", selectSQL))

	// FROM clause
	parts = append(parts, fmt.Sprintf("FROM %s", q.table))

	// JOIN clause
	if joinSQL := q.joinM.Build(); joinSQL != "" {
		parts = append(parts, joinSQL)
	}

	// WHERE clause
	if !q.whereM.IsEmpty() {
		whereSQL, whereParams := q.whereM.Build()
		parts = append(parts, fmt.Sprintf("WHERE %s", whereSQL))
		allParams = append(allParams, whereParams...)
	}

	// GROUP BY clause
	if groupSQL := q.groupByM.Build(); groupSQL != "" {
		parts = append(parts, fmt.Sprintf("GROUP BY %s", groupSQL))
	}

	// HAVING clause
	if !q.havingM.IsEmpty() {
		havingSQL, havingParams := q.havingM.Build()
		parts = append(parts, fmt.Sprintf("HAVING %s", havingSQL))
		allParams = append(allParams, havingParams...)
	}

	// ORDER BY clause
	if orderSQL := q.orderByM.Build(); orderSQL != "" {
		parts = append(parts, fmt.Sprintf("ORDER BY %s", orderSQL))
	}

	// LIMIT/OFFSET clause
	if limitSQL := q.limitM.Build(); limitSQL != "" {
		parts = append(parts, limitSQL)
	}

	return strings.Join(parts, " "), allParams.Values()
}

// ============================================================================
// EXECUTION
// ============================================================================

// Executor interface
type Executor interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// All executes and returns all rows
func (q Query[T]) All(ctx context.Context, db Executor, scanner func(*sql.Rows) (T, error)) ([]T, error) {
	query, params := q.Build()
	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, rows.Err()
}

// First executes and returns first row
func (q Query[T]) First(ctx context.Context, db Executor, scanner func(*sql.Row) (T, error)) (T, error) {
	query, params := q.Build()
	row := db.QueryRowContext(ctx, query, params...)
	return scanner(row)
}

// Count executes COUNT(*)
func (q Query[T]) Count(ctx context.Context, db Executor) (int, error) {
	countQuery := q
	countQuery.selectM = Select("COUNT(*)")
	countQuery.aggregateM = AggregationMonoid{}.Empty()

	query, params := countQuery.Build()
	row := db.QueryRowContext(ctx, query, params...)

	var count int
	err := row.Scan(&count)
	return count, err
}

// Exists checks if any row matches
func (q Query[T]) Exists(ctx context.Context, db Executor) (bool, error) {
	count, err := q.Bound(Limit(1)).Count(ctx, db)
	return count > 0, err
}

type Param struct {
	Value interface{} // Unavoidable for SQL driver, but contained
}

// P creates a type-safe parameter (short alias)
func P[T any](value T) Param {
	return Param{Value: value}
}

// Params is a typed list of parameters
type Params []Param

// Values extracts raw values for sql.DB (unavoidable interface conversion)
func (p Params) Values() []interface{} {
	vals := make([]interface{}, len(p))
	for i, param := range p {
		vals[i] = param.Value
	}
	return vals
}
