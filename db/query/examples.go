package query

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/vinodhalaharvi/purepulse/pkg/types"
)

// ============================================================================
// EXAMPLE 1: Combine SELECT Monoids
// ============================================================================

func Example1() {
	users := Table[User]("users")

	// Create independent SELECT monoids
	basicFields := Select("id", "name")
	contactFields := Select("email", "phone")
	addressFields := Select("street", "city", "country")

	// Monoid combine = union of fields
	allFields := basicFields.
		Combine(contactFields).
		Combine(addressFields)

	// Apply to query
	query := users.Project(allFields)

	PrintQuery(query)

	// SQL: SELECT id, name, email, phone, street, city, country FROM users
}

// PrintQuery prints the SQL and parameters for debugging
func PrintQuery[T any](q Query[T]) {
	sql, params := q.Build()

	fmt.Println("SQL:")
	fmt.Println(sql)
	fmt.Println()

	if len(params) > 0 {
		fmt.Println("Parameters:")
		for i, param := range params {
			fmt.Printf("  $%d: %v (%T)\n", i+1, param, param)
		}
	} else {
		fmt.Println("Parameters: none")
	}
	fmt.Println()
}

// ============================================================================
// EXAMPLE 2: Compose JOIN Monoids
// ============================================================================

func Example2() {
	users := Table[User]("users")

	// Independent join specifications
	userOrders := InnerJoin("orders", "users.id = orders.user_id")
	orderProducts := InnerJoin("products", "orders.product_id = products.id")

	// Compose joins (associative!)
	allJoins := userOrders.Then(orderProducts)

	query := users.ApplyJoins(allJoins)

	PrintQuery(query)

	// SQL: SELECT * FROM users
	//      INNER JOIN orders ON users.id = orders.user_id
	//      INNER JOIN products ON orders.product_id = products.id
}

// ============================================================================
// EXAMPLE 3: Combine WHERE Monoids
// ============================================================================

func Example2Better() {
	users := Table[User]("users")

	// Create filter monoids - very readable!
	filters := Where("active", true). // active = true
						Combine(WhereGTE("age", 18)).       // AND age >= 18
						Combine(WhereIsNotNull("email")).   // AND email IS NOT NULL
						Combine(WhereLike("name", "John%")) // AND name LIKE 'John%'

	query := users.Filter(filters)

	PrintQuery(query)

	// SQL:
	// SELECT * FROM users
	// WHERE active = ? AND age >= ? AND email IS NOT NULL AND name LIKE ?
	//
	// Parameters:
	//   $1: true (bool)
	//   $2: 18 (int)
	//   $3: John% (string)
}

func Example3() {
	users := Table[User]("users")

	// Independent filter monoids
	isActive := Where("active", true)
	isAdult := Where("age >=", 18)
	hasEmail := WhereIsNotNull("email")

	// Combine filters (monoid!)
	allFilters := isActive.
		Combine(isAdult).
		Combine(hasEmail)

	query := users.Filter(allFilters)

	PrintQuery(query)

	// SQL: SELECT * FROM users
	//      WHERE active = ? AND age >= ? AND email IS NOT NULL
}

// ============================================================================
// EXAMPLE 4: GROUP BY + AGGREGATION Monoids
// ============================================================================

func Example4() {
	orders := Table[Order]("orders")

	// Aggregation monoids
	sumRevenue := Sum("total").As("revenue")
	countOrders := Count("*").As("order_count")
	avgOrderValue := Avg("total").As("avg_value")

	// Combine aggregations
	aggregations := sumRevenue.
		Combine(countOrders).
		Combine(avgOrderValue)

	query := orders.
		Group(GroupBy("user_id")).
		Aggregate(aggregations)

	PrintQuery(query)

	// SQL: SELECT user_id, SUM(total) AS revenue, COUNT(*) AS order_count, AVG(total) AS avg_value
	//      FROM orders
	//      GROUP BY user_id
}

// ============================================================================
// EXAMPLE 5: Full Composition - All Monoids Together
// ============================================================================

func Example5() {
	orders := Table[Order]("orders")

	// Define all monoids independently
	selectFields := Select("user_id")
	joinUsers := LeftJoin("users", "orders.user_id = users.id")
	filterActive := Where("orders.status", "active")
	grouping := GroupBy("user_id")
	aggregates := Sum("total").As("revenue").
		Combine(Count("*").As("order_count"))
	havingFilter := Where("SUM(total) >", 1000)
	sorting := Desc("revenue")
	pagination := Paginate(1, 20)

	// Compose everything (order doesn't matter - monoids!)
	query := orders.
		Project(selectFields).
		ApplyJoins(joinUsers).
		Filter(filterActive).
		Group(grouping).
		Aggregate(aggregates).
		Having(havingFilter).
		Sort(sorting).
		Bound(pagination)

	PrintQuery[Order](query)

	// SQL: SELECT user_id, SUM(total) AS revenue, COUNT(*) AS order_count
	//      FROM orders
	//      LEFT JOIN users ON orders.user_id = users.id
	//      WHERE orders.status = ?
	//      GROUP BY user_id
	//      HAVING SUM(total) > ?
	//      ORDER BY revenue DESC
	//      LIMIT 20 OFFSET 0
}

// ============================================================================
// EXAMPLE 6: Reusable Query Fragments as Monoids
// ============================================================================

// Define reusable monoid fragments
var (
	ActiveFilter     = Where("active", true)
	NotDeletedFilter = Where[string]("deleted_at IS NULL", "")
	RecentSort       = Desc("created_at")
)

func Example6() {
	users := Table[User]("users")

	// Compose reusable fragments
	query := users.
		Filter(ActiveFilter.Combine(NotDeletedFilter)).
		Sort(RecentSort).
		Bound(Limit(10))

	fmt.Println(query)

	// SQL: SELECT * FROM users
	//      WHERE active = ? AND deleted_at IS NULL
	//      ORDER BY created_at DESC
	//      LIMIT 10
}

// ============================================================================
// USER MODEL
// ============================================================================

// User represents a user in the system
type User struct {
	ID          types.UserID `db:"id" json:"id"`
	DisplayName string       `db:"display_name" json:"display_name"`
	Email       string       `db:"email" json:"email"`
	Timezone    string       `db:"timezone" json:"timezone"`

	// Platform identities
	SlackID       sql.NullString `db:"slack_id" json:"slack_id,omitempty"`
	GitHubLogin   sql.NullString `db:"github_login" json:"github_login,omitempty"`
	JiraAccountID sql.NullString `db:"jira_account_id" json:"jira_account_id,omitempty"`
	ZoomEmail     sql.NullString `db:"zoom_email" json:"zoom_email,omitempty"`

	// Metadata
	Active       bool         `db:"active" json:"active"`
	LastActivity sql.NullTime `db:"last_activity" json:"last_activity,omitempty"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt    sql.NullTime `db:"deleted_at" json:"deleted_at,omitempty"`
}

// ScanUser scans a sql.Row into User
func ScanUser(row *sql.Row) (User, error) {
	var u User
	err := row.Scan(
		&u.ID,
		&u.DisplayName,
		&u.Email,
		&u.Timezone,
		&u.SlackID,
		&u.GitHubLogin,
		&u.JiraAccountID,
		&u.ZoomEmail,
		&u.Active,
		&u.LastActivity,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	return u, err
}

// ScanUsers scans sql.Rows into []User
func ScanUsers(rows *sql.Rows) (User, error) {
	var u User
	err := rows.Scan(
		&u.ID,
		&u.DisplayName,
		&u.Email,
		&u.Timezone,
		&u.SlackID,
		&u.GitHubLogin,
		&u.JiraAccountID,
		&u.ZoomEmail,
		&u.Active,
		&u.LastActivity,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	return u, err
}

// Order represents an example order (not in schema, just for query examples)
type Order struct {
	ID        string       `db:"id" json:"id"`
	UserID    string       `db:"user_id" json:"user_id"`
	ProductID string       `db:"product_id" json:"product_id"`
	Total     float64      `db:"total" json:"total"`
	Status    string       `db:"status" json:"status"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"deleted_at,omitempty"`
}

// ScanOrder scans a sql.Row into Order
func ScanOrder(row *sql.Row) (Order, error) {
	var o Order
	err := row.Scan(
		&o.ID,
		&o.UserID,
		&o.ProductID,
		&o.Total,
		&o.Status,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.DeletedAt,
	)
	return o, err
}

// ScanOrders scans sql.Rows into Order
func ScanOrders(rows *sql.Rows) (Order, error) {
	var o Order
	err := rows.Scan(
		&o.ID,
		&o.UserID,
		&o.ProductID,
		&o.Total,
		&o.Status,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.DeletedAt,
	)
	return o, err
}

func Example6SubQueryComposition() {
	// Build the subquery as a monoid!
	userStatsSubQuery := AsSubQuery(
		Table[Order]("orders").
			Project(Select(
				"user_id",
				"COUNT(*) as total_orders",
			)).
			Filter(Where("status", "completed")).
			Group(GroupBy("user_id")).
			Aggregate(Sum("total").As("lifetime_value")).
			Having(WhereRaw("SUM(total) >", 5000)),
		"user_stats", // alias
	)

	// Another subquery for top products
	topProductsSubQuery := AsSubQuery(
		Table[Order]("order_items").
			Project(Select("product_id")).
			Group(GroupBy("product_id")).
			Aggregate(Sum("quantity").As("total_qty")).
			Having(WhereRaw("SUM(quantity) >", 100)),
		"top_products",
	)

	// Main query using subqueries - ALL MONOIDS!
	query := Table[Order]("orders").
		Project(Select(
			"orders.user_id",
			"users.name AS user_name",
			"users.email",
			"user_stats.lifetime_value",
			"user_stats.total_orders",
		)).
		ApplyJoins(
			LeftJoin("users", "orders.user_id = users.id").
				Then(InnerJoinSubQuery(userStatsSubQuery, "orders.user_id = user_stats.user_id")).
				Then(InnerJoinSubQuery(topProductsSubQuery, "orders.product_id = top_products.product_id")),
		).
		Filter(
			Where("orders.status", "active").
				Combine(WhereGTE("user_stats.lifetime_value", 10000)),
		).
		Sort(Desc("user_stats.lifetime_value")).
		Bound(Limit(20))

	PrintQuery[Order](query)

	/* Expected Output:
	SQL:
	SELECT orders.user_id, users.name AS user_name, users.email, user_stats.lifetime_value, user_stats.total_orders
	FROM orders
	LEFT JOIN users ON orders.user_id = users.id
	INNER JOIN (SELECT user_id, COUNT(*) as total_orders, SUM(total) AS lifetime_value FROM orders WHERE status = ? GROUP BY user_id HAVING SUM(total) > ?) AS user_stats ON orders.user_id = user_stats.user_id
	INNER JOIN (SELECT product_id, SUM(quantity) AS total_qty FROM order_items GROUP BY product_id HAVING SUM(quantity) > ?) AS top_products ON orders.product_id = top_products.product_id
	WHERE orders.status = ? AND user_stats.lifetime_value >= ?
	ORDER BY user_stats.lifetime_value DESC
	LIMIT 20

	Parameters:
	  $1: completed (string)        -- from subquery 1 WHERE
	  $2: 5000 (int)                -- from subquery 1 HAVING
	  $3: 100 (int)                 -- from subquery 2 HAVING
	  $4: active (string)           -- from main query WHERE
	  $5: 10000 (int)               -- from main query WHERE
	*/
}
