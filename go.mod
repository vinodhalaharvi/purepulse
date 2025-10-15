module github.com/vinodhalaharvi/purepulse

go 1.25.0

// Use local purekernels for development
replace github.com/vinodhalaharvi/purekernels => ../purekernels

require (
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/vinodhalaharvi/purekernels v0.0.0-00010101000000-000000000000
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20251009144603-d2f985daa21b // indirect
)
