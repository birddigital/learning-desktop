module github.com/birddigital/learning-desktop

go 1.25.6

require (
	github.com/birddigital/htmx-r v0.1.0
	github.com/google/uuid v1.6.0
)

// Use local htmx-r during development
replace github.com/birddigital/htmx-r => ../htmx-r
