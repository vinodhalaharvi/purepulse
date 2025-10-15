package query

import (
	"fmt"
	"strings"
)

// ============================================================================
// SELECT MONOID - Projection
// ============================================================================

// SelectMonoid represents SELECT clause as a monoid
type SelectMonoid struct {
	columns []string
}

// Empty returns identity (SELECT *)
func (SelectMonoid) Empty() SelectMonoid {
	return SelectMonoid{columns: []string{}}
}

// Combine merges two SELECT monoids (union of fields)
func (s SelectMonoid) Combine(other SelectMonoid) SelectMonoid {
	return SelectMonoid{
		columns: append(append([]string{}, s.columns...), other.columns...),
	}
}

// Select creates a new SelectMonoid
func Select(columns ...string) SelectMonoid {
	return SelectMonoid{columns: columns}
}

// Build generates SQL
func (s SelectMonoid) Build() string {
	if len(s.columns) == 0 {
		return "*"
	}
	return strings.Join(s.columns, ", ")
}

// ============================================================================
// WHERE MONOID - Filter
// ============================================================================

// Build generates SQL
func (w WhereMonoid) Build() (string, Params) {
	if len(w.conditions) == 0 {
		return "", Params{}
	}

	sqls := make([]string, len(w.conditions))
	var allParams Params

	for i, cond := range w.conditions {
		sqls[i] = cond.SQL
		allParams = append(allParams, cond.Params...)
	}

	return strings.Join(sqls, " AND "), allParams
}

// IsEmpty checks if monoid is identity
func (w WhereMonoid) IsEmpty() bool {
	return len(w.conditions) == 0
}

// ============================================================================
// JOIN MONOID - Product
// ============================================================================

// JoinMonoid represents JOIN clauses as a monoid
type JoinMonoid struct {
	joins []JoinSpec
}

// JoinSpec represents a single join
type JoinSpec struct {
	Type      string // "INNER", "LEFT", "RIGHT"
	Table     string
	Condition string
	Params    Params
}

// Empty returns identity (no joins)
func (JoinMonoid) Empty() JoinMonoid {
	return JoinMonoid{joins: []JoinSpec{}}
}

// Combine merges two JOIN monoids (compose joins)
func (j JoinMonoid) Combine(other JoinMonoid) JoinMonoid {
	return JoinMonoid{
		joins: append(append([]JoinSpec{}, j.joins...), other.joins...),
	}
}

// InnerJoin creates an INNER JOIN monoid
func InnerJoin(table, condition string) JoinMonoid {
	return JoinMonoid{
		joins: []JoinSpec{{
			Type:      "INNER",
			Table:     table,
			Condition: condition,
		}},
	}
}

// LeftJoin creates a LEFT JOIN monoid
func LeftJoin(table, condition string) JoinMonoid {
	return JoinMonoid{
		joins: []JoinSpec{{
			Type:      "LEFT",
			Table:     table,
			Condition: condition,
		}},
	}
}

// RightJoin creates a RIGHT JOIN monoid
func RightJoin(table, condition string) JoinMonoid {
	return JoinMonoid{
		joins: []JoinSpec{{
			Type:      "RIGHT",
			Table:     table,
			Condition: condition,
		}},
	}
}

// Then chains joins (alias for Combine)
func (j JoinMonoid) Then(other JoinMonoid) JoinMonoid {
	return j.Combine(other)
}

// Build generates SQL
// Build generates SQL and collects all params
func (j JoinMonoid) Build() (string, Params) {
	if len(j.joins) == 0 {
		return "", Params{}
	}

	parts := make([]string, len(j.joins))
	var allParams Params

	for i, join := range j.joins {
		parts[i] = fmt.Sprintf("%s JOIN %s ON %s", join.Type, join.Table, join.Condition)
		allParams = allParams.Append(join.Params)
	}

	return strings.Join(parts, " "), allParams
}

// ============================================================================
// GROUP BY MONOID - Partition
// ============================================================================

// GroupByMonoid represents GROUP BY clause as a monoid
type GroupByMonoid struct {
	columns []string
}

// Empty returns identity (no grouping)
func (GroupByMonoid) Empty() GroupByMonoid {
	return GroupByMonoid{columns: []string{}}
}

// Combine merges two GROUP BY monoids (hierarchical grouping)
func (g GroupByMonoid) Combine(other GroupByMonoid) GroupByMonoid {
	return GroupByMonoid{
		columns: append(append([]string{}, g.columns...), other.columns...),
	}
}

// GroupBy creates a GROUP BY monoid
func GroupBy(columns ...string) GroupByMonoid {
	return GroupByMonoid{columns: columns}
}

// ThenBy adds additional grouping (alias for Combine)
func (g GroupByMonoid) ThenBy(columns ...string) GroupByMonoid {
	return g.Combine(GroupBy(columns...))
}

// Build generates SQL
func (g GroupByMonoid) Build() string {
	if len(g.columns) == 0 {
		return ""
	}
	return strings.Join(g.columns, ", ")
}

// ============================================================================
// AGGREGATION MONOID - Aggregate Functions
// ============================================================================

// AggregationMonoid represents aggregation functions
type AggregationMonoid struct {
	aggregates []Aggregate
}

// Aggregate represents a single aggregate function
type Aggregate struct {
	Function string // "SUM", "COUNT", "AVG", "MIN", "MAX"
	Column   string
	Alias    string
}

// Empty returns identity (no aggregations)
func (AggregationMonoid) Empty() AggregationMonoid {
	return AggregationMonoid{aggregates: []Aggregate{}}
}

// Combine merges two aggregation monoids
func (a AggregationMonoid) Combine(other AggregationMonoid) AggregationMonoid {
	return AggregationMonoid{
		aggregates: append(append([]Aggregate{}, a.aggregates...), other.aggregates...),
	}
}

// Sum creates a SUM aggregation
func Sum(column string) AggregationMonoid {
	return AggregationMonoid{
		aggregates: []Aggregate{{
			Function: "SUM",
			Column:   column,
		}},
	}
}

// Count creates a COUNT aggregation
func Count(column string) AggregationMonoid {
	return AggregationMonoid{
		aggregates: []Aggregate{{
			Function: "COUNT",
			Column:   column,
		}},
	}
}

// Avg creates an AVG aggregation
func Avg(column string) AggregationMonoid {
	return AggregationMonoid{
		aggregates: []Aggregate{{
			Function: "AVG",
			Column:   column,
		}},
	}
}

// Min creates a MIN aggregation
func Min(column string) AggregationMonoid {
	return AggregationMonoid{
		aggregates: []Aggregate{{
			Function: "MIN",
			Column:   column,
		}},
	}
}

// Max creates a MAX aggregation
func Max(column string) AggregationMonoid {
	return AggregationMonoid{
		aggregates: []Aggregate{{
			Function: "MAX",
			Column:   column,
		}},
	}
}

// As sets alias for aggregation
func (a AggregationMonoid) As(alias string) AggregationMonoid {
	if len(a.aggregates) > 0 {
		a.aggregates[len(a.aggregates)-1].Alias = alias
	}
	return a
}

// Build generates SQL
func (a AggregationMonoid) Build() []string {
	if len(a.aggregates) == 0 {
		return []string{}
	}

	result := make([]string, len(a.aggregates))
	for i, agg := range a.aggregates {
		sql := fmt.Sprintf("%s(%s)", agg.Function, agg.Column)
		if agg.Alias != "" {
			sql += " AS " + agg.Alias
		}
		result[i] = sql
	}

	return result
}

// ============================================================================
// ORDER BY MONOID - Sort
// ============================================================================

// OrderByMonoid represents ORDER BY clause as a monoid
type OrderByMonoid struct {
	orders []string
}

// Empty returns identity (no ordering)
func (OrderByMonoid) Empty() OrderByMonoid {
	return OrderByMonoid{orders: []string{}}
}

// Combine merges two ORDER BY monoids (ThenBy)
func (o OrderByMonoid) Combine(other OrderByMonoid) OrderByMonoid {
	return OrderByMonoid{
		orders: append(append([]string{}, o.orders...), other.orders...),
	}
}

// OrderBy creates an ORDER BY monoid
func OrderBy(columns ...string) OrderByMonoid {
	return OrderByMonoid{orders: columns}
}

// Asc adds ASC ordering
func Asc(column string) OrderByMonoid {
	return OrderByMonoid{orders: []string{fmt.Sprintf("%s ASC", column)}}
}

// Desc adds DESC ordering
func Desc(column string) OrderByMonoid {
	return OrderByMonoid{orders: []string{fmt.Sprintf("%s DESC", column)}}
}

// ThenBy adds additional ordering (alias for Combine)
func (o OrderByMonoid) ThenBy(columns ...string) OrderByMonoid {
	return o.Combine(OrderBy(columns...))
}

// Build generates SQL
func (o OrderByMonoid) Build() string {
	if len(o.orders) == 0 {
		return ""
	}
	return strings.Join(o.orders, ", ")
}

// ============================================================================
// LIMIT MONOID - Bound
// ============================================================================

// LimitMonoid represents LIMIT/OFFSET as a monoid
type LimitMonoid struct {
	limit  *int
	offset *int
}

// Empty returns identity (no limit)
func (LimitMonoid) Empty() LimitMonoid {
	return LimitMonoid{}
}

// Combine merges two LIMIT monoids (takes minimum)
func (l LimitMonoid) Combine(other LimitMonoid) LimitMonoid {
	result := LimitMonoid{}

	// Take minimum limit
	if l.limit != nil && other.limit != nil {
		min := *l.limit
		if *other.limit < min {
			min = *other.limit
		}
		result.limit = &min
	} else if l.limit != nil {
		result.limit = l.limit
	} else if other.limit != nil {
		result.limit = other.limit
	}

	// Take maximum offset
	if l.offset != nil && other.offset != nil {
		max := *l.offset
		if *other.offset > max {
			max = *other.offset
		}
		result.offset = &max
	} else if l.offset != nil {
		result.offset = l.offset
	} else if other.offset != nil {
		result.offset = other.offset
	}

	return result
}

// Limit creates a LIMIT monoid
func Limit(n int) LimitMonoid {
	return LimitMonoid{limit: &n}
}

// Offset creates an OFFSET monoid
func Offset(n int) LimitMonoid {
	return LimitMonoid{offset: &n}
}

// Paginate creates LIMIT + OFFSET
func Paginate(page, pageSize int) LimitMonoid {
	offset := (page - 1) * pageSize
	return LimitMonoid{
		limit:  &pageSize,
		offset: &offset,
	}
}

// Build generates SQL
func (l LimitMonoid) Build() string {
	parts := []string{}

	if l.limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *l.limit))
	}

	if l.offset != nil {
		parts = append(parts, fmt.Sprintf("OFFSET %d", *l.offset))
	}

	return strings.Join(parts, " ")
}

// ============================================================================
// WHERE MONOID - Filter (FIXED)
// ============================================================================

// WhereMonoid represents WHERE clause as a monoid
type WhereMonoid struct {
	conditions []Condition
}

// Condition represents a single WHERE condition
type Condition struct {
	SQL    string
	Params Params
}

// Empty returns identity (no filter)
func (WhereMonoid) Empty() WhereMonoid {
	return WhereMonoid{conditions: []Condition{}}
}

// Combine merges two WHERE monoids (AND composition)
func (w WhereMonoid) Combine(other WhereMonoid) WhereMonoid {
	return WhereMonoid{
		conditions: append(append([]Condition{}, w.conditions...), other.conditions...),
	}
}

// Where creates a WHERE monoid with "column = value" format
func Where[V any](column string, value V) WhereMonoid {
	return WhereMonoid{
		conditions: []Condition{{
			SQL:    fmt.Sprintf("%s = ?", column),
			Params: Params{P(value)},
		}},
	}
}

// WhereRaw creates a WHERE monoid with custom SQL (use when SQL has operator)
func WhereRaw[V any](sql string, value V) WhereMonoid {
	return WhereMonoid{
		conditions: []Condition{{
			SQL:    sql,
			Params: Params{P(value)},
		}},
	}
}

// WhereNoParam creates a WHERE monoid with no parameters (for IS NULL, etc)
func WhereNoParam(sql string) WhereMonoid {
	return WhereMonoid{
		conditions: []Condition{{
			SQL:    sql,
			Params: Params{},
		}},
	}
}

// WhereIn creates WHERE IN monoid
func WhereIn[V any](column string, values []V) WhereMonoid {
	params := make(Params, len(values))
	placeholders := make([]string, len(values))

	for i, v := range values {
		params[i] = P(v)
		placeholders[i] = "?"
	}

	return WhereMonoid{
		conditions: []Condition{{
			SQL:    fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", ")),
			Params: params,
		}},
	}
}

// WhereBetween creates BETWEEN monoid
func WhereBetween[V any](column string, start, end V) WhereMonoid {
	return WhereMonoid{
		conditions: []Condition{{
			SQL:    fmt.Sprintf("%s BETWEEN ? AND ?", column),
			Params: Params{P(start), P(end)},
		}},
	}
}

// WhereGT creates > condition
func WhereGT[V any](column string, value V) WhereMonoid {
	return WhereMonoid{
		conditions: []Condition{{
			SQL:    fmt.Sprintf("%s > ?", column),
			Params: Params{P(value)},
		}},
	}
}

// WhereGTE creates >= condition
func WhereGTE[V any](column string, value V) WhereMonoid {
	return WhereMonoid{
		conditions: []Condition{{
			SQL:    fmt.Sprintf("%s >= ?", column),
			Params: Params{P(value)},
		}},
	}
}

// WhereLT creates < condition
func WhereLT[V any](column string, value V) WhereMonoid {
	return WhereMonoid{
		conditions: []Condition{{
			SQL:    fmt.Sprintf("%s < ?", column),
			Params: Params{P(value)},
		}},
	}
}

// WhereLTE creates <= condition
func WhereLTE[V any](column string, value V) WhereMonoid {
	return WhereMonoid{
		conditions: []Condition{{
			SQL:    fmt.Sprintf("%s <= ?", column),
			Params: Params{P(value)},
		}},
	}
}

// WhereIsNull creates IS NULL condition (no parameters!)
func WhereIsNull(column string) WhereMonoid {
	return WhereNoParam(fmt.Sprintf("%s IS NULL", column))
}

// WhereIsNotNull creates IS NOT NULL condition (no parameters!)
func WhereIsNotNull(column string) WhereMonoid {
	return WhereNoParam(fmt.Sprintf("%s IS NOT NULL", column))
}

// WhereLike creates LIKE condition
func WhereLike(column string, pattern string) WhereMonoid {
	return WhereMonoid{
		conditions: []Condition{{
			SQL:    fmt.Sprintf("%s LIKE ?", column),
			Params: Params{P(pattern)},
		}},
	}
}

// ============================================================================
// SUBQUERY MONOID - Nested Query Composition
// ============================================================================

// SubQuery wraps a Query to use as a subquery
type SubQuery[T any] struct {
	query Query[T]
	alias string
}

// AsSubQuery converts a query into a subquery with an alias
func AsSubQuery[T any](q Query[T], alias string) SubQuery[T] {
	return SubQuery[T]{
		query: q,
		alias: alias,
	}
}

// Build generates the subquery SQL
func (sq SubQuery[T]) Build() (string, Params) {
	sql, params := sq.query.Build()

	// Convert []interface{} back to Params
	paramsList := make(Params, len(params))
	for i, p := range params {
		paramsList[i] = Param{Value: p}
	}

	return fmt.Sprintf("(%s) AS %s", sql, sq.alias), paramsList
}

// ============================================================================
// ENHANCED JOIN MONOID - Support SubQuery
// ============================================================================

// InnerJoinSubQuery joins with a subquery
func InnerJoinSubQuery[T any](subquery SubQuery[T], condition string) JoinMonoid {
	sql, params := subquery.Build()
	return JoinMonoid{
		joins: []JoinSpec{{
			Type:      "INNER",
			Table:     sql,
			Condition: condition,
			Params:    params,
		}},
	}
}

// LeftJoinSubQuery joins with a subquery
func LeftJoinSubQuery[T any](subquery SubQuery[T], condition string) JoinMonoid {
	sql, params := subquery.Build()
	return JoinMonoid{
		joins: []JoinSpec{{
			Type:      "LEFT",
			Table:     sql,
			Condition: condition,
			Params:    params,
		}},
	}
}

// RightJoinSubQuery joins with a subquery
func RightJoinSubQuery[T any](subquery SubQuery[T], condition string) JoinMonoid {
	sql, params := subquery.Build()
	return JoinMonoid{
		joins: []JoinSpec{{
			Type:      "RIGHT",
			Table:     sql,
			Condition: condition,
			Params:    params,
		}},
	}
}
