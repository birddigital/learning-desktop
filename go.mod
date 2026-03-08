module github.com/birddigital/learning-desktop

go 1.25.6

require (
	github.com/birddigital/htmx-r v0.1.0
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/jackc/pgx/v5 v5.5.0
	github.com/jackc/pgx/v5/stdlib v1.0.0-rc.1
)

// Use local htmx-r during development
replace github.com/birddigital/htmx-r => ../htmx-r
